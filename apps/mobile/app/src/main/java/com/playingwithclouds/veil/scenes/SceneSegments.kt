package com.playingwithclouds.veil.scenes

import kotlin.math.exp
import kotlin.math.max

/** A run of sampled frames with one label: [start] and [end] are frame indices, both included. */
data class LabelRun(val label: String, val start: Int, val end: Int, val score: Double)

/** A labelled stretch of a video in seconds, its label already a tag name ("Blowjob"). */
data class TimedSegment(val tag: String, val startSeconds: Double, val endSeconds: Double)

/**
 * Per-frame probabilities → labelled time segments, a port of the classifier's `nsr/segments.py`
 * (multi-label heads by hysteresis, exclusive heads by a smoothed argmax), and logits →
 * probabilities as `nsr.model.probabilities` computes them.
 */
object SceneSegments {

    /** Multi-label hysteresis: a label turns on at this probability… */
    const val MULTI_HIGH = 0.6f

    /** …and stays on until it drops below this one. */
    const val MULTI_LOW = 0.4f

    /** Off-runs this many frames long or shorter between two on-runs are closed. */
    const val MAX_GAP_FRAMES = 2

    /** Runs shorter than this many frames are dropped. */
    const val MIN_RUN_FRAMES = 3

    /** Exclusive heads: frames whose best smoothed label is below this get none. */
    const val MIN_CONFIDENCE = 0.4f

    /** Exclusive heads: frames averaged around each frame before the argmax. */
    const val SMOOTH_FRAMES = 5

    /** Each label's on/off runs of a multi-label head; [probabilities] is one array per frame. */
    fun multiLabel(probabilities: List<FloatArray>, names: List<String>): List<LabelRun> {
        val out = mutableListOf<LabelRun>()
        for ((column, name) in names.withIndex()) {
            val values = FloatArray(probabilities.size) { frame -> probabilities[frame][column] }
            val on = closeGaps(hysteresis(values, MULTI_HIGH, MULTI_LOW), MAX_GAP_FRAMES)
            for ((start, end) in runs(on)) {
                if (end - start + 1 >= MIN_RUN_FRAMES) {
                    out.add(LabelRun(name, start, end, mean(values, start, end)))
                }
            }
        }
        return out
    }

    /** One label per frame of an exclusive head, by smoothed argmax, as runs in start order. */
    fun exclusive(probabilities: List<FloatArray>, names: List<String>): List<LabelRun> {
        val best = smoothedArgmax(probabilities, names.size)
        val out = mutableListOf<LabelRun>()
        for ((column, name) in names.withIndex()) {
            val on = BooleanArray(best.size) { frame -> best[frame] == column }
            for ((start, end) in runs(on)) {
                if (end - start + 1 >= MIN_RUN_FRAMES) {
                    val values = FloatArray(probabilities.size) { frame -> probabilities[frame][column] }
                    out.add(LabelRun(name, start, end, mean(values, start, end)))
                }
            }
        }
        return out.sortedBy { run -> run.start }
    }

    /**
     * The tagged segments of a scanned video: each segment head's runs, labels mapped to tag names
     * ([SceneLabels]), frame indices to seconds (frame i was grabbed at [times]`[i]` and stands for
     * the [stepSeconds] after it). Runs of one tag that touch merge, since two labels may share one.
     */
    fun timedSegments(times: List<Double>, stepSeconds: Double, probabilities: Map<String, List<FloatArray>>, manifest: SceneModelManifest): List<TimedSegment> {
        val found = mutableListOf<TimedSegment>()
        for (headName in SceneLabels.segmentHeads) {
            val head = manifest.head(headName) ?: continue
            val frames = probabilities[headName] ?: continue
            found.addAll(headSegments(head, frames, times, stepSeconds))
        }
        return mergeTouching(found)
    }

    /** One head's runs as timed segments, labels that never become segments left out. */
    private fun headSegments(head: ModelHead, frames: List<FloatArray>, times: List<Double>, stepSeconds: Double): List<TimedSegment> {
        val names = head.labels.map { label -> label.name }
        var labelRuns = multiLabel(frames, names)
        if (head.exclusive) {
            labelRuns = exclusive(frames, names)
        }
        return labelRuns.mapNotNull { run ->
            val label = head.labels.first { label -> label.name == run.label }
            val tag = SceneLabels.tagName(head.name, label) ?: return@mapNotNull null
            TimedSegment(tag, times[run.start], times[run.end] + stepSeconds)
        }
    }

    /** Joins the segments of one tag that overlap or touch; the result is in start order. */
    fun mergeTouching(segments: List<TimedSegment>): List<TimedSegment> {
        val out = mutableListOf<TimedSegment>()
        for ((tag, ofTag) in segments.groupBy { segment -> segment.tag }) {
            var current: TimedSegment? = null
            for (segment in ofTag.sortedBy { segment -> segment.startSeconds }) {
                val open = current
                if (open != null && segment.startSeconds <= open.endSeconds) {
                    current = open.copy(endSeconds = max(open.endSeconds, segment.endSeconds))
                    continue
                }
                if (open != null) {
                    out.add(open)
                }
                current = segment
            }
            current?.let(out::add)
        }
        return out.sortedWith(compareBy({ segment -> segment.startSeconds }, { segment -> segment.tag }))
    }

    /**
     * Logits → probabilities per head: softmax for exclusive heads, sigmoid for the others; a
     * conditional head is scaled by its condition label's probability.
     */
    fun probabilities(logits: Map<String, FloatArray>, manifest: SceneModelManifest): Map<String, FloatArray> {
        val out = mutableMapOf<String, FloatArray>()
        for ((name, values) in logits) {
            val head = manifest.head(name) ?: continue
            out[name] = sigmoid(values)
            if (head.exclusive) {
                out[name] = softmax(values)
            }
        }
        for ((name, values) in out.toMap()) {
            out[name] = gated(manifest.head(name)?.condition, values, out, manifest)
        }
        return out
    }

    /** [values] scaled by the condition label's probability; unchanged without a condition. */
    private fun gated(condition: HeadCondition?, values: FloatArray, probabilities: Map<String, FloatArray>, manifest: SceneModelManifest): FloatArray {
        if (condition == null) {
            return values
        }
        val gateHead = manifest.head(condition.head) ?: return values
        val gateValues = probabilities[condition.head] ?: return values
        val index = gateHead.labels.indexOfFirst { label -> label.name == condition.label }
        if (index < 0) {
            return values
        }
        return FloatArray(values.size) { column -> values[column] * gateValues[index] }
    }

    /** Softmax of one row of logits. */
    fun softmax(values: FloatArray): FloatArray {
        val largest = values.maxOrNull() ?: return values
        val exponents = FloatArray(values.size) { index -> exp((values[index] - largest).toDouble()).toFloat() }
        val sum = exponents.sum()
        return FloatArray(values.size) { index -> exponents[index] / sum }
    }

    /** Sigmoid of each logit. */
    fun sigmoid(values: FloatArray): FloatArray {
        return FloatArray(values.size) { index -> (1.0 / (1.0 + exp(-values[index].toDouble()))).toFloat() }
    }

    /** On while above [high], staying on until below [low]. */
    private fun hysteresis(values: FloatArray, high: Float, low: Float): BooleanArray {
        val on = BooleanArray(values.size)
        var state = false
        for ((index, value) in values.withIndex()) {
            state = if (state) value >= low else value >= high
            on[index] = state
        }
        return on
    }

    /** Fills off-gaps of at most [maxGap] frames between two on-runs. */
    private fun closeGaps(on: BooleanArray, maxGap: Int): BooleanArray {
        val out = on.copyOf()
        val found = runs(on)
        for (index in 0 until found.size - 1) {
            val gapStart = found[index].second + 1
            val gapEnd = found[index + 1].first - 1
            if (gapEnd - gapStart + 1 <= maxGap) {
                for (frame in gapStart..gapEnd) {
                    out[frame] = true
                }
            }
        }
        return out
    }

    /** The on-runs as (first, last) frame pairs. */
    private fun runs(on: BooleanArray): List<Pair<Int, Int>> {
        val out = mutableListOf<Pair<Int, Int>>()
        var start = -1
        for ((index, value) in on.withIndex()) {
            if (value && start < 0) {
                start = index
            }
            if (!value && start >= 0) {
                out.add(start to index - 1)
                start = -1
            }
        }
        if (start >= 0) {
            out.add(start to on.size - 1)
        }
        return out
    }

    /** Per frame, the best label of the probabilities averaged over [SMOOTH_FRAMES] around it (edges repeated); -1 below [MIN_CONFIDENCE]. */
    private fun smoothedArgmax(probabilities: List<FloatArray>, labelCount: Int): IntArray {
        val half = SMOOTH_FRAMES / 2
        val last = probabilities.size - 1
        return IntArray(probabilities.size) { frame ->
            val averaged = FloatArray(labelCount)
            for (offset in -half..half) {
                val row = probabilities[(frame + offset).coerceIn(0, last)]
                for (column in 0 until labelCount) {
                    averaged[column] += row[column] / SMOOTH_FRAMES
                }
            }
            bestAbove(averaged, MIN_CONFIDENCE)
        }
    }

    /** Index of the largest value when it reaches [minimum], else -1. */
    private fun bestAbove(values: FloatArray, minimum: Float): Int {
        var best = -1
        for ((index, value) in values.withIndex()) {
            if (best < 0 || value > values[best]) {
                best = index
            }
        }
        if (best < 0 || values[best] < minimum) {
            return -1
        }
        return best
    }

    /** Mean of [values] from [start] to [end], both included. */
    private fun mean(values: FloatArray, start: Int, end: Int): Double {
        var sum = 0.0
        for (index in start..end) {
            sum += values[index]
        }
        return sum / (end - start + 1)
    }
}
