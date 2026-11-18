package com.playingwithclouds.veil.scenes

import android.content.Context
import android.graphics.Bitmap
import android.os.Process
import android.util.Log
import androidx.media3.common.MediaItem
import androidx.media3.common.MimeTypes
import androidx.media3.effect.Presentation
import androidx.media3.exoplayer.SeekParameters
import androidx.media3.inspector.frame.FrameExtractor
import com.google.common.util.concurrent.ListenableFuture
import com.google.common.util.concurrent.MoreExecutors
import com.playingwithclouds.veil.data.PlayableVideo
import com.playingwithclouds.veil.data.SceneRepository
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.NonCancellable
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.delay
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import java.util.concurrent.Executors
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException
import kotlin.math.ceil

/** How a scan ended. */
enum class ScanOutcome {
    /** Segments were stored (possibly none: nothing was recognised). */
    SCANNED,

    /** No source played, or the runtime is unknown; tried again with the next model. */
    UNPLAYABLE,

    /** No model is installed. */
    NO_MODEL,
}

/**
 * How hard a scan works: frames per frame-graph call, a pause after each batch, and the thread the
 * graphs run on. The open video's scan goes gently beside playback; the background one, while
 * charging, goes at full speed.
 */
class ScanPace(val batchFrames: Int, val pauseMilliseconds: Long, val dispatcher: CoroutineDispatcher) {
    companion object {
        /** While charging, nothing else on screen. */
        val BACKGROUND = ScanPace(SceneScanner.BATCH_FRAMES, 0, Dispatchers.Default)

        /** Beside playback: small batches on a background-priority thread, with breaks for the decoder. */
        val BESIDE_PLAYBACK by lazy { ScanPace(FOREGROUND_BATCH_FRAMES, FOREGROUND_PAUSE_MILLISECONDS, lowPriorityDispatcher()) }

        private const val FOREGROUND_BATCH_FRAMES = 4
        private const val FOREGROUND_PAUSE_MILLISECONDS = 150L

        /** One thread at Android's background priority, so the graphs yield to the player and the UI. */
        private fun lowPriorityDispatcher(): CoroutineDispatcher {
            val executor = Executors.newSingleThreadExecutor { runnable ->
                val thread = Thread({
                    Process.setThreadPriority(Process.THREAD_PRIORITY_BACKGROUND)
                    runnable.run()
                }, "scene-scan")
                thread.isDaemon = true
                thread
            }
            return executor.asCoroutineDispatcher()
        }
    }
}

/**
 * Scans one video in small steps: grabs a frame every few seconds across the whole video straight
 * from its stream (Media3's frame extractor seeks to the nearest key frame, scaled down while
 * decoding; no file is downloaded or cut), runs the frame graph in small batches and the temporal
 * graph over windows of the embeddings, turns the per-frame probabilities into labelled segments
 * and replaces the video's ai segments with them.
 */
class SceneScanner(private val context: Context) {

    /** Scans [videoId] from its start; [onProgress] gets the share done, 0..1. */
    suspend fun scan(videoId: String, onProgress: (Float) -> Unit = {}): ScanOutcome {
        val manifest = SceneClassifier.manifest() ?: return ScanOutcome.NO_MODEL
        val video = SceneRepository.playable(videoId, readPages = true) ?: return ScanOutcome.UNPLAYABLE
        val timeline = scanStream(video, 0.0, ScanPace.BACKGROUND) { partial -> onProgress(partial.progress) } ?: return ScanOutcome.UNPLAYABLE
        if (!timeline.complete) {
            return ScanOutcome.NO_MODEL
        }
        SceneRepository.replaceSegments(videoId, timeline.segments(manifest))
        return ScanOutcome.SCANNED
    }

    /**
     * Grabs and reads [video]'s frames starting at [startSeconds] (to the end, then from the start),
     * calling [onBatch] after each batch with what's in so far. Null when the runtime is unknown;
     * a timeline that isn't complete when the model went away midway.
     */
    suspend fun scanStream(video: PlayableVideo, startSeconds: Double, pace: ScanPace, onBatch: suspend (ScanTimeline) -> Unit): ScanTimeline? {
        val manifest = SceneClassifier.manifest() ?: return null
        val duration = video.durationSeconds
        if (duration == null || duration <= 0) {
            return null
        }
        val step = ScanPlan.stepSeconds(duration)
        val times = ScanPlan.times(duration, step)
        var window = 0
        if (manifest.temporal != null) {
            window = manifest.window
        }
        val timeline = ScanTimeline(times, step, window)
        val order = ScanTimeline.order(times.size, ScanTimeline.indexAt(times, startSeconds))
        val extractor = withContext(Dispatchers.Main) { extractorFor(video) }
        try {
            for (batch in order.chunked(pace.batchFrames)) {
                val grabStarted = System.currentTimeMillis()
                val bitmaps = batch.map { index -> grab(extractor, times[index]) }
                val readStarted = System.currentTimeMillis()
                val read = withContext(pace.dispatcher) { readBatch(timeline, batch, bitmaps, manifest) }
                if (!read) {
                    Log.w(TAG, "model went away mid-scan")
                    return timeline
                }
                logBatch(timeline, batch, readStarted - grabStarted, System.currentTimeMillis() - readStarted, manifest)
                onBatch(timeline)
                delay(pace.pauseMilliseconds)
            }
        } finally {
            withContext(Dispatchers.Main + NonCancellable) { extractor.close() }
        }
        return timeline
    }

    /** Runs one batch through the frame graph, then the temporal graph over the windows that filled; false without a model. */
    private fun readBatch(timeline: ScanTimeline, batch: List<Int>, bitmaps: List<Bitmap>, manifest: SceneModelManifest): Boolean {
        val outputs = SceneClassifier.frames(bitmaps) ?: return false
        val probabilities = perFrame(outputs.logits, manifest)
        for ((position, index) in batch.withIndex()) {
            timeline.add(index, outputs.embeddings[position], probabilities[position])
        }
        for (indices in timeline.pendingWindows().chunked(BATCH_WINDOWS)) {
            val logits = SceneClassifier.temporal(indices.map(timeline::windowOf)) ?: return true
            for ((position, frame) in perFrame(logits, manifest).withIndex()) {
                timeline.setVideo(indices[position], frame)
            }
        }
        return true
    }

    /** Logs a batch: where it was, how long grabbing and reading took, and the strongest labels of its first frame. */
    private fun logBatch(timeline: ScanTimeline, batch: List<Int>, grabMilliseconds: Long, readMilliseconds: Long, manifest: SceneModelManifest) {
        val first = batch.firstOrNull() ?: return
        val strongest = timeline.strongestLabels(first, LOGGED_LABELS, manifest).joinToString { (label, probability) -> "$label ${"%.2f".format(probability)}" }
        Log.i(TAG, "frames ${batch.size} at ${timeline.times[first].toInt()}s: grab ${grabMilliseconds}ms, model ${readMilliseconds}ms, ${(timeline.progress * 100).toInt()}% done; top: $strongest")
    }

    /** Grabs the frame at [seconds] (on the main thread, where the extractor lives). */
    private suspend fun grab(extractor: FrameExtractor, seconds: Double): Bitmap {
        return withContext(Dispatchers.Main) { extractor.getFrame((seconds * MILLISECONDS_PER_SECOND).toLong()).await().bitmap }
    }

    /** A frame extractor over the video's stream that seeks to key frames and scales frames down. */
    private fun extractorFor(video: PlayableVideo): FrameExtractor {
        val item = MediaItem.Builder().setUri(video.stream.url)
        if (video.stream.isHls) {
            item.setMimeType(MimeTypes.APPLICATION_M3U8)
        }
        return FrameExtractor.Builder(context, item.build())
            .setSeekParameters(SeekParameters.CLOSEST_SYNC)
            .setEffects(listOf(Presentation.createForHeight(FRAME_HEIGHT)))
            .build()
    }

    companion object {
        /** Frames per frame-graph call: small, so memory stays low and the phone stays cool. */
        const val BATCH_FRAMES = 8

        /** Windows per temporal-graph call. */
        const val BATCH_WINDOWS = 32

        /** Frames are decoded at this height; the model's input is smaller still. */
        const val FRAME_HEIGHT = 288

        private const val MILLISECONDS_PER_SECOND = 1000.0

        /** Log tag of the scans. */
        const val TAG = "VeilScan"

        /** Labels named per logged batch. */
        private const val LOGGED_LABELS = 5
    }
}

/** Where a video is sampled: cheap enough for a phone, dense enough for the segment rules. */
object ScanPlan {

    /** The most frames one video gets; longer videos are sampled more sparsely. */
    const val MAX_FRAMES = 900

    /** The step between frames: 2 s under 10 minutes, 3 s up to 40, 4 s beyond, more when [MAX_FRAMES] needs it. */
    fun stepSeconds(durationSeconds: Double): Double {
        var step = 4.0
        if (durationSeconds <= 10 * 60) {
            step = 2.0
        } else if (durationSeconds <= 40 * 60) {
            step = 3.0
        }
        return maxOf(step, ceil(durationSeconds / MAX_FRAMES))
    }

    /** The grab times, from the start up to the end. */
    fun times(durationSeconds: Double, stepSeconds: Double): List<Double> {
        val count = maxOf(1, (durationSeconds / stepSeconds).toInt())
        return List(count) { index -> index * stepSeconds }
    }
}

/** Logits per head (one row per item) → probabilities per item. */
internal fun perFrame(logits: Map<String, List<FloatArray>>, manifest: SceneModelManifest): List<Map<String, FloatArray>> {
    val count = logits.values.firstOrNull()?.size ?: 0
    return List(count) { index ->
        SceneSegments.probabilities(logits.mapValues { (_, rows) -> rows[index] }, manifest)
    }
}

/** Per-item maps → one series per head. */
internal fun byHead(frames: List<Map<String, FloatArray>>): Map<String, List<FloatArray>> {
    val names = frames.flatMap { frame -> frame.keys }.toSet()
    return names.associateWith { name -> frames.map { frame -> frame[name] ?: FloatArray(0) } }.filterValues { rows -> rows.all { row -> row.isNotEmpty() } }
}

/** Waits for a future without blocking the thread. */
internal suspend fun <T> ListenableFuture<T>.await(): T {
    return suspendCancellableCoroutine { continuation ->
        addListener({
            try {
                continuation.resume(get())
            } catch (error: java.util.concurrent.ExecutionException) {
                continuation.resumeWithException(error.cause ?: error)
            }
        }, MoreExecutors.directExecutor())
        continuation.invokeOnCancellation { cancel(true) }
    }
}
