package com.playingwithclouds.veil.scenes

import kotlin.math.max
import kotlin.math.min

/** One stretch the compilation plays: a part of a video whose scenes the user likes, named by [tags]. */
data class CompilationClip(val videoId: String, val startSeconds: Double, val endSeconds: Double, val tags: List<String>)

/** A video to cut a compilation from: its ai segments and runtime (null when unknown). */
data class CompilationSource(val videoId: String, val segments: List<TimedSegment>, val durationSeconds: Double?)

/**
 * Plans the compilation mode: from each video, only the stretches whose scene model tags the user
 * liked, near ones joined so playback doesn't stutter between them, too short ones dropped, long
 * ones cut so one video doesn't fill the compilation. Videos keep their order.
 */
object CompilationPlanner {

    /** Liked stretches this close together play as one. */
    const val JOIN_GAP_SECONDS = 6.0

    /** Shorter stretches aren't worth the jump. */
    const val MIN_CLIP_SECONDS = 8.0

    /** A stretch plays at most this long. */
    const val MAX_CLIP_SECONDS = 150.0

    /** Playback starts this much before a stretch, so the cut doesn't land mid-motion. */
    const val LEAD_IN_SECONDS = 1.0

    /** The clips of every source, in source order. */
    fun plan(sources: List<CompilationSource>, likedTags: Set<String>): List<CompilationClip> {
        return sources.flatMap { source -> clips(source, likedTags) }
    }

    /** One video's clips in time order. Tags compare without case. */
    fun clips(source: CompilationSource, likedTags: Set<String>): List<CompilationClip> {
        val liked = likedTags.map { tag -> tag.lowercase() }.toSet()
        val matching = source.segments.filter { segment -> segment.tag.lowercase() in liked }.sortedBy { segment -> segment.startSeconds }
        val out = mutableListOf<CompilationClip>()
        for (segment in matching) {
            val last = out.lastOrNull()
            if (last != null && segment.startSeconds - last.endSeconds <= JOIN_GAP_SECONDS) {
                out[out.size - 1] = last.copy(endSeconds = max(last.endSeconds, segment.endSeconds), tags = (last.tags + segment.tag).distinct())
                continue
            }
            out.add(CompilationClip(source.videoId, segment.startSeconds, segment.endSeconds, listOf(segment.tag)))
        }
        return out.filter { clip -> clip.endSeconds - clip.startSeconds >= MIN_CLIP_SECONDS }.map { clip -> fitted(clip, source.durationSeconds) }
    }

    /** A clip with its lead-in, cut to [MAX_CLIP_SECONDS] and kept inside the runtime. */
    private fun fitted(clip: CompilationClip, durationSeconds: Double?): CompilationClip {
        val start = max(0.0, clip.startSeconds - LEAD_IN_SECONDS)
        var end = min(clip.endSeconds, start + MAX_CLIP_SECONDS)
        if (durationSeconds != null && durationSeconds > 0) {
            end = min(end, durationSeconds)
        }
        return clip.copy(startSeconds = start, endSeconds = end)
    }
}

/**
 * Picks the moments a card's preview jumps through: 4–6 points across the video, from the ai
 * segments with the tags the user likes most (liked first, then the rest when too few), spread
 * out so the preview shows the video rather than one scene.
 */
object PreviewMoments {

    /** How many moments a preview wants at most. */
    const val MAX_MOMENTS = 6

    /** Fewer moments than this aren't a preview; the stills flipbook or the site's clip is better. */
    const val MIN_MOMENTS = 4

    /** How long each moment plays. */
    const val MOMENT_SECONDS = 2.5

    /** Moments start at least this far apart. */
    private const val MIN_SPACING_SECONDS = 15.0

    /** A segment longer than this offers a second point in its middle. */
    private const val LONG_SEGMENT_SECONDS = 60.0

    /** How far into a segment its moment starts, at most; the first seconds are often a cut. */
    private const val MAX_OFFSET_SECONDS = 5.0

    /** The moments' start seconds in time order; empty when fewer than [MIN_MOMENTS] are found. */
    fun pick(segments: List<TimedSegment>, likedTags: Set<String>): List<Double> {
        val liked = likedTags.map { tag -> tag.lowercase() }.toSet()
        val preferred = segments.filter { segment -> segment.tag.lowercase() in liked }
        val others = segments.filter { segment -> segment.tag.lowercase() !in liked }
        val picked = mutableListOf<Double>()
        addSpread(picked, points(preferred))
        if (picked.size < MIN_MOMENTS) {
            addSpread(picked, points(others))
        }
        if (picked.size < MIN_MOMENTS) {
            return emptyList()
        }
        return picked.sorted()
    }

    /** Candidate points of segments, the longest segments' first. */
    private fun points(segments: List<TimedSegment>): List<Double> {
        return segments.sortedByDescending { segment -> segment.endSeconds - segment.startSeconds }.flatMap { segment ->
            val length = segment.endSeconds - segment.startSeconds
            val first = segment.startSeconds + min(MAX_OFFSET_SECONDS, length / 4)
            if (length > LONG_SEGMENT_SECONDS) {
                listOf(first, segment.startSeconds + length / 2)
            } else {
                listOf(first)
            }
        }
    }

    /** Adds the candidates at least [MIN_SPACING_SECONDS] from every picked point, up to [MAX_MOMENTS]. */
    private fun addSpread(picked: MutableList<Double>, candidates: List<Double>) {
        for (candidate in candidates) {
            if (picked.size >= MAX_MOMENTS) {
                return
            }
            if (picked.all { point -> kotlin.math.abs(point - candidate) >= MIN_SPACING_SECONDS }) {
                picked.add(candidate)
            }
        }
    }
}
