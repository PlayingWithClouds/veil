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
     * that don't overlap them, each followed by an untitled marker where nothing is recognised for a
     * while, so the chapter label clears. In start order.
     */
    fun chapters(segments: List<VideoSegment>): List<VideoSegment> {
        val others = segments.filter { segment -> segment.origin != ORIGIN_AI }
        val found = segments.filter { segment -> segment.origin == ORIGIN_AI }.mapNotNull(::timedOf)
        val kept = sections(found).filter { section -> others.none { other -> overlaps(section, other) } }
        val aiChapters = kept.map { section ->
            VideoSegment("ai-${section.startSeconds}", section.title, section.startSeconds, section.endSeconds, ORIGIN_AI)
        }.toMutableList()
        val last = kept.lastOrNull()
        if (last != null && last.title.isNotEmpty()) {
            aiChapters.add(VideoSegment("ai-end-${last.endSeconds}", "", last.endSeconds, null, ORIGIN_AI))
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

    /** Whether an ai section overlaps a user or plugin segment (a marker inside it counts). */
    private fun overlaps(section: Section, other: VideoSegment): Boolean {
        val otherEnd = other.endSeconds ?: other.seconds
        return other.seconds < section.endSeconds && otherEnd >= section.startSeconds
    }

    /**
     * The sections from the first labelled second to the last: per stretch the longest one or two
     * labels running, equal neighbours joined, short pieces folded into their longer neighbour.
     * Untitled sections are long gaps.
     */
    fun sections(segments: List<TimedSegment>): List<Section> {
        val merged = mergeNear(segments)
        if (merged.isEmpty()) {
            return emptyList()
        }
        var pieces = elementaryPieces(merged)
        pieces = absorbShort(joinEqual(pieces))
        return pieces
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

    /** Folds the shortest piece under [MIN_SECTION_SECONDS] into its longer neighbour until none is left. */
    private fun absorbShort(start: List<Section>): List<Section> {
        var pieces = start
        while (pieces.size > 1) {
            val shortest = pieces.withIndex().filter { (_, piece) -> piece.length < MIN_SECTION_SECONDS }.minByOrNull { (_, piece) -> piece.length }
            if (shortest == null) {
                break
            }
            pieces = joinEqual(foldInto(pieces, shortest.index))
        }
        return pieces.filter { piece -> piece.title.isNotEmpty() || piece.length >= MIN_SECTION_SECONDS }
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
