package com.playingwithclouds.veil.scenes

import android.content.Context
import android.graphics.Bitmap
import androidx.media3.common.MediaItem
import androidx.media3.common.MimeTypes
import androidx.media3.effect.Presentation
import androidx.media3.exoplayer.SeekParameters
import androidx.media3.inspector.frame.FrameExtractor
import com.google.common.util.concurrent.ListenableFuture
import com.google.common.util.concurrent.MoreExecutors
import com.playingwithclouds.veil.data.PlayableVideo
import com.playingwithclouds.veil.data.SceneRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
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
 * Scans one video in small steps: grabs a frame every few seconds across the whole video straight
 * from its stream (Media3's frame extractor seeks to the nearest key frame, scaled down while
 * decoding; no file is downloaded or cut), runs the frame graph in small batches and the temporal
 * graph over windows of the embeddings, turns the per-frame probabilities into labelled segments
 * and replaces the video's ai segments with them.
 */
class SceneScanner(private val context: Context) {

    /** Scans [videoId]; [onProgress] gets the share done, 0..1. */
    suspend fun scan(videoId: String, onProgress: (Float) -> Unit = {}): ScanOutcome {
        val manifest = SceneClassifier.manifest() ?: return ScanOutcome.NO_MODEL
        val video = SceneRepository.playable(videoId, readPages = true) ?: return ScanOutcome.UNPLAYABLE
        val duration = video.durationSeconds
        if (duration == null || duration <= 0) {
            return ScanOutcome.UNPLAYABLE
        }
        val step = ScanPlan.stepSeconds(duration)
        val times = ScanPlan.times(duration, step)
        val probabilities = classify(video, times, manifest, onProgress)
        val segments = SceneSegments.timedSegments(times, step, probabilities, manifest)
        SceneRepository.replaceSegments(videoId, segments)
        return ScanOutcome.SCANNED
    }

    /** Per head, one probability row per grabbed frame. */
    private suspend fun classify(video: PlayableVideo, times: List<Double>, manifest: SceneModelManifest, onProgress: (Float) -> Unit): Map<String, List<FloatArray>> {
        val embeddings = mutableListOf<FloatArray>()
        val imageProbabilities = mutableListOf<Map<String, FloatArray>>()
        val extractor = withContext(Dispatchers.Main) { extractorFor(video) }
        try {
            for (batch in times.chunked(BATCH_FRAMES)) {
                val bitmaps = batch.map { seconds -> grab(extractor, seconds) }
                val outputs = withContext(Dispatchers.Default) { SceneClassifier.frames(bitmaps) } ?: return emptyMap()
                embeddings.addAll(outputs.embeddings)
                imageProbabilities.addAll(perFrame(outputs.logits, manifest))
                onProgress(embeddings.size.toFloat() / times.size)
            }
        } finally {
            withContext(Dispatchers.Main) { extractor.close() }
        }
        val videoProbabilities = withContext(Dispatchers.Default) { temporalProbabilities(embeddings, manifest) }
        return byHead(imageProbabilities.indices.map { frame ->
            FrameProbabilities.combine(imageProbabilities[frame], videoProbabilities?.get(frame), manifest)
        })
    }

    /** The video heads' probabilities per frame, from windows of the embeddings; null without a temporal graph. */
    private fun temporalProbabilities(embeddings: List<FloatArray>, manifest: SceneModelManifest): List<Map<String, FloatArray>>? {
        if (manifest.temporal == null || embeddings.isEmpty()) {
            return null
        }
        val out = mutableListOf<Map<String, FloatArray>>()
        for (windows in FrameProbabilities.windows(embeddings, manifest.window).chunked(BATCH_WINDOWS)) {
            val logits = SceneClassifier.temporal(windows) ?: return null
            out.addAll(perFrame(logits, manifest))
        }
        return out
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
