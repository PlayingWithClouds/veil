package com.playingwithclouds.veil.scenes

import com.playingwithclouds.veil.data.VideoSegment
import kotlin.math.max
import kotlin.math.min

/** A named stretch of the seek bar; an empty [title] is a stretch where nothing was recognised. */
data class Section(val title: String, val startSeconds: Double, val endSeconds: Double) {
    val length: Double get() = endSeconds - startSeconds
}

/**
 * Turns the scene model's segments into chapters the player can show: one named section at a
 * time (at most two labels, "Blowjob · POV"), nothing shorter than [MIN_SECTION_SECONDS], so the
 * seek bar doesn't break into 2-second confetti. The user's and the plugins' segments always win:
 * ai sections overlapping them are left out.
 */
object SceneSections {

    /** Runs of one label this close together count as one. */
    const val MERGE_GAP_SECONDS = 12.0

    /** The shortest section (or unlabelled stretch) kept; shorter ones join a neighbour. */
    const val MIN_SECTION_SECONDS = 20.0

    /** The most labels a section names. */
    private const val MAX_TITLE_LABELS = 2

    /** Separates the labels of one section. */
    private const val TITLE_SEPARATOR = " · "

    /** The origin of the scene model's segments, as the API spells it. */
    const val ORIGIN_AI = "ai"

    /**
     * The chapters to show: the user's and the plugins' segments as they are, then the ai sections
     * around them (cut where a user or plugin stretch runs, left out where one of their markers
     * without an end falls), in start order. Every ai section is titled: where the ai segments
     * recognise nothing, [broad] ([BroadLabels], display only) names the stretch. With
     * [durationSeconds] the sections run from the start to the end.
     */
    fun chapters(segments: List<VideoSegment>, broad: List<TimedSegment> = emptyList(), durationSeconds: Double? = null): List<VideoSegment> {
        val others = segments.filter { segment -> segment.origin != ORIGIN_AI }
        val found = segments.filter { segment -> segment.origin == ORIGIN_AI }.mapNotNull(::timedOf)
        val kept = aroundOthers(sections(found, broad, durationSeconds), others)
        val aiChapters = kept.map { section ->
            VideoSegment("ai-${section.startSeconds}", section.title, section.startSeconds, section.endSeconds, ORIGIN_AI)
        }
        return (others + aiChapters).sortedBy { segment -> segment.seconds }
    }

    /**
     * The video's segments with the open video's scan in place of its stored ai ones: once the
     * scan is [done], or as soon as it found something (until then the stored ones, maybe from an
     * older model, are better than nothing). The user's and the plugins' segments stay.
     */
    fun withScanned(segments: List<VideoSegment>, scanned: List<TimedSegment>?, done: Boolean): List<VideoSegment> {
        if (scanned == null || (scanned.isEmpty() && !done)) {
            return segments
        }
        val others = segments.filter { segment -> segment.origin != ORIGIN_AI }
        val found = scanned.map { segment ->
            VideoSegment("scan-${segment.tag}-${segment.startSeconds}", segment.tag, segment.startSeconds, segment.endSeconds, ORIGIN_AI, segment.tag)
        }
        return others + found
    }

    /** An ai segment as a timed one, named by its tag; null for a marker without an end. */
    private fun timedOf(segment: VideoSegment): TimedSegment? {
        val end = segment.endSeconds ?: return null
        if (end <= segment.seconds) {
            return null
        }
        return TimedSegment(segment.tag ?: segment.title, segment.seconds, end)
    }

    /**
     * [sections] without the ones a user or plugin marker without an end falls in, and cut around
     * the user's and plugins' stretches; pieces left shorter than [MIN_SECTION_SECONDS] go.
     */
    private fun aroundOthers(sections: List<Section>, others: List<VideoSegment>): List<Section> {
        val markers = others.filter { other -> other.endSeconds == null }
        val spans = others.mapNotNull { other -> other.endSeconds?.let { end -> other.seconds..end } }
        val unmarked = sections.filter { section -> markers.none { marker -> marker.seconds >= section.startSeconds && marker.seconds < section.endSeconds } }
        val asSegments = unmarked.map { section -> TimedSegment(section.title, section.startSeconds, section.endSeconds) }
        return SceneSegments.outside(asSegments, spans)
            .map { piece -> Section(piece.tag, piece.startSeconds, piece.endSeconds) }
            .filter { piece -> piece.length >= MIN_SECTION_SECONDS }
    }

    /**
     * The sections, all titled: per stretch the longest one or two labels of [segments] running;
     * where none runs, the [broad] label; equal neighbours joined; short pieces, and stretches
     * nothing names, folded into their longer neighbour. From the first labelled second to the last,
     * or from 0 to [durationSeconds] when it's known.
     */
    fun sections(segments: List<TimedSegment>, broad: List<TimedSegment> = emptyList(), durationSeconds: Double? = null): List<Section> {
        val merged = mergeNear(segments)
        val mergedBroad = mergeNear(broad)
        if (merged.isEmpty() && mergedBroad.isEmpty()) {
            return emptyList()
        }
        val specific = elementaryPieces(merged)
        val pieces = withEnds(specific, mergedBroad, durationSeconds).flatMap { piece -> filled(piece, mergedBroad) }
        return absorbShort(joinEqual(pieces))
    }

    /**
     * [pieces] with untitled ones added before the first and after the last, out to 0 and
     * [durationSeconds] when known, else to where [broad] starts and ends.
     */
    private fun withEnds(pieces: List<Section>, broad: List<TimedSegment>, durationSeconds: Double?): List<Section> {
        val all = pieces.map { piece -> piece.startSeconds } + broad.map { segment -> segment.startSeconds }
        val ends = pieces.map { piece -> piece.endSeconds } + broad.map { segment -> segment.endSeconds }
        var start = all.min()
        var end = ends.max()
        if (durationSeconds != null && durationSeconds > 0) {
            start = 0.0
            end = max(end, durationSeconds)
        }
        if (pieces.isEmpty()) {
            return listOf(Section("", start, end))
        }
        val out = mutableListOf<Section>()
        if (start < pieces.first().startSeconds) {
            out.add(Section("", start, pieces.first().startSeconds))
        }
        out.addAll(pieces)
        if (pieces.last().endSeconds < end) {
            out.add(Section("", pieces.last().endSeconds, end))
        }
        return out
    }

    /** A titled piece as it is; an untitled one cut by the [broad] labels running through it (bits none covers stay untitled). */
    private fun filled(piece: Section, broad: List<TimedSegment>): List<Section> {
        if (piece.title.isNotEmpty()) {
            return listOf(piece)
        }
        val inside = broad
            .filter { segment -> segment.startSeconds < piece.endSeconds && segment.endSeconds > piece.startSeconds }
            .map { segment -> segment.copy(startSeconds = max(segment.startSeconds, piece.startSeconds), endSeconds = min(segment.endSeconds, piece.endSeconds)) }
        if (inside.isEmpty()) {
            return listOf(piece)
        }
        val cuts = (listOf(piece.startSeconds, piece.endSeconds) + inside.flatMap { segment -> listOf(segment.startSeconds, segment.endSeconds) }).distinct().sorted()
        return cuts.zipWithNext().map { (start, end) -> Section(broadTitleOf(inside, start, end), start, end) }
    }

    /** The longest broad label running from [start] to [end]; empty when none does. */
    private fun broadTitleOf(broad: List<TimedSegment>, start: Double, end: Double): String {
        val running = broad.filter { segment -> segment.startSeconds < end && segment.endSeconds > start }
        return running.maxByOrNull { segment -> segment.endSeconds - segment.startSeconds }?.tag.orEmpty()
    }

    /** Joins the runs of one tag that are at most [MERGE_GAP_SECONDS] apart. */
    private fun mergeNear(segments: List<TimedSegment>): List<TimedSegment> {
        val bridged = segments.map { segment -> segment.copy(endSeconds = segment.endSeconds + MERGE_GAP_SECONDS) }
        return SceneSegments.mergeTouching(bridged).map { segment -> segment.copy(endSeconds = segment.endSeconds - MERGE_GAP_SECONDS) }
    }

    /** The timeline cut at every start and end, each piece titled by what runs through it. */
    private fun elementaryPieces(segments: List<TimedSegment>): List<Section> {
        val cuts = segments.flatMap { segment -> listOf(segment.startSeconds, segment.endSeconds) }.distinct().sorted()
        val out = mutableListOf<Section>()
        for (index in 0 until cuts.size - 1) {
            val start = cuts[index]
            val end = cuts[index + 1]
            val running = segments.filter { segment -> segment.startSeconds < end && segment.endSeconds > start }
            out.add(Section(titleOf(running), start, end))
        }
        return out
    }

    /** The longest [MAX_TITLE_LABELS] running labels, longest first; empty when none runs. */
    private fun titleOf(running: List<TimedSegment>): String {
        return running
            .sortedWith(compareByDescending<TimedSegment> { segment -> segment.endSeconds - segment.startSeconds }.thenBy { segment -> segment.tag })
            .map { segment -> segment.tag }
            .distinct()
            .take(MAX_TITLE_LABELS)
            .joinToString(TITLE_SEPARATOR)
    }

    /** Joins neighbours with the same title. */
    private fun joinEqual(pieces: List<Section>): List<Section> {
        val out = mutableListOf<Section>()
        for (piece in pieces) {
            val last = out.lastOrNull()
            if (last != null && last.title == piece.title) {
                out[out.size - 1] = last.copy(endSeconds = piece.endSeconds)
                continue
            }
            out.add(piece)
        }
        return out
    }

    /**
     * Folds pieces into their longer neighbour until every piece is titled and at least
     * [MIN_SECTION_SECONDS] long: untitled ones first (whatever their length), then the shortest.
     * Untitled pieces left alone (nothing titled at all) are dropped.
     */
    private fun absorbShort(start: List<Section>): List<Section> {
        var pieces = start
        while (pieces.size > 1) {
            val index = nextToFold(pieces)
            if (index < 0) {
                break
            }
            pieces = joinEqual(foldInto(pieces, index))
        }
        return pieces.filter { piece -> piece.title.isNotEmpty() }
    }

    /** The piece to fold next: the shortest untitled one, else the shortest under [MIN_SECTION_SECONDS]; -1 when none is. */
    private fun nextToFold(pieces: List<Section>): Int {
        val untitled = pieces.withIndex().filter { (_, piece) -> piece.title.isEmpty() }.minByOrNull { (_, piece) -> piece.length }
        if (untitled != null) {
            return untitled.index
        }
        val short = pieces.withIndex().filter { (_, piece) -> piece.length < MIN_SECTION_SECONDS }.minByOrNull { (_, piece) -> piece.length }
        return short?.index ?: -1
    }

    /** [pieces] with the one at [index] taken over by its longer neighbour. */
    private fun foldInto(pieces: List<Section>, index: Int): List<Section> {
        val out = pieces.toMutableList()
        val piece = out.removeAt(index)
        val previous = pieces.getOrNull(index - 1)
        val next = pieces.getOrNull(index + 1)
        if (previous != null && (next == null || previous.length >= next.length)) {
            out[index - 1] = previous.copy(endSeconds = max(previous.endSeconds, piece.endSeconds))
            return out
        }
        if (next != null) {
            out[index] = next.copy(startSeconds = min(next.startSeconds, piece.startSeconds))
        }
        return out
    }
}
