package com.playingwithclouds.veil.scenes

/**
 * One video's scan as it fills in, frame by frame and in any order (the open video starts where it
 * plays): each grabbed frame's embedding and image-head probabilities, and the video heads' once
 * the frame's whole temporal window is in. [segments] reads what's there at any point; once
 * [complete], it's what a scan from the start would give, since every window is the same.
 */
class ScanTimeline(val times: List<Double>, val stepSeconds: Double, private val window: Int) {

    private val embeddings = arrayOfNulls<FloatArray>(times.size)
    private val image = arrayOfNulls<Map<String, FloatArray>>(times.size)
    private val video = arrayOfNulls<Map<String, FloatArray>>(times.size)

    /** How many frames are in. */
    var scannedCount = 0
        private set

    /** Whether every frame is in. */
    val complete: Boolean
        get() = scannedCount == times.size

    /** The share of frames in, 0..1. */
    val progress: Float
        get() = scannedCount.toFloat() / maxOf(1, times.size)

    /** Stores one grabbed frame. */
    fun add(index: Int, embedding: FloatArray, imageProbabilities: Map<String, FloatArray>) {
        if (embeddings[index] == null) {
            scannedCount++
        }
        embeddings[index] = embedding
        image[index] = imageProbabilities
    }

    /** The [count] most likely "head.label"s of one frame's picture, strongest first, for the scan log. */
    fun strongestLabels(index: Int, count: Int, manifest: SceneModelManifest): List<Pair<String, Float>> {
        val probabilities = image[index] ?: return emptyList()
        val all = mutableListOf<Pair<String, Float>>()
        for ((headName, values) in probabilities) {
            val head = manifest.head(headName) ?: continue
            for ((labelIndex, value) in values.withIndex()) {
                all.add("$headName.${head.labels.getOrNull(labelIndex)?.name ?: labelIndex}" to value)
            }
        }
        return all.sortedByDescending { (_, value) -> value }.take(count)
    }

    /** Frames whose video heads can be read now: their whole window is in and they weren't read yet. None without a temporal graph. */
    fun pendingWindows(): List<Int> {
        if (window <= 0) {
            return emptyList()
        }
        return times.indices.filter { index -> video[index] == null && windowComplete(index) }
    }

    /**
     * The temporal graph's input for frame [index]: the last [window] embeddings up to it,
     * left-padded with the video's first, as [FrameProbabilities.windows] builds them.
     */
    fun windowOf(index: Int): List<FloatArray> {
        val first = maxOf(0, index - window + 1)
        val recent = (first..index).map { frame -> embeddings[frame] ?: error("frame $frame isn't scanned") }
        return List(window - recent.size) { recent.first() } + recent
    }

    /** Stores the video heads' probabilities of frame [index]. */
    fun setVideo(index: Int, probabilities: Map<String, FloatArray>) {
        video[index] = probabilities
    }

    /**
     * The tagged segments of what's scanned so far: each unbroken stretch of scanned frames read on
     * its own (frames still waiting for their window count by their picture alone), then joined.
     */
    fun segments(manifest: SceneModelManifest): List<TimedSegment> {
        val found = mutableListOf<TimedSegment>()
        for (run in scannedRuns()) {
            val frames = run.map { index -> FrameProbabilities.combine(image[index] ?: emptyMap(), video[index], manifest) }
            found.addAll(SceneSegments.timedSegments(times.subList(run.first, run.last + 1), stepSeconds, byHead(frames), manifest))
        }
        return SceneSegments.mergeTouching(found)
    }

    /** The unbroken stretches of scanned frames, in time order. */
    fun scannedRuns(): List<IntRange> {
        val runs = mutableListOf<IntRange>()
        var start = -1
        for (index in times.indices) {
            val scanned = embeddings[index] != null
            if (scanned && start < 0) {
                start = index
            }
            if (!scanned && start >= 0) {
                runs.add(start until index)
                start = -1
            }
        }
        if (start >= 0) {
            runs.add(start until times.size)
        }
        return runs
    }

    /** Whether every frame of [index]'s window is scanned. */
    private fun windowComplete(index: Int): Boolean {
        val first = maxOf(0, index - window + 1)
        return (first..index).all { frame -> embeddings[frame] != null }
    }

    companion object {

        /** The order frames are grabbed in: from [startIndex] to the end, then from the start up to it. */
        fun order(count: Int, startIndex: Int): List<Int> {
            val start = startIndex.coerceIn(0, maxOf(0, count - 1))
            return (start until count).toList() + (0 until start).toList()
        }

        /** The frame grabbed at or just before [seconds]. */
        fun indexAt(times: List<Double>, seconds: Double): Int {
            val index = times.indexOfLast { time -> time <= seconds }
            return maxOf(0, index)
        }
    }
}
