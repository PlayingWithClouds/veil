package com.playingwithclouds.veil.scenes

import android.content.Context
import android.graphics.Bitmap
import android.net.ConnectivityManager
import android.os.PowerManager
import android.util.Log
import coil3.SingletonImageLoader
import coil3.request.ImageRequest
import coil3.request.SuccessResult
import coil3.request.allowHardware
import coil3.toBitmap
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.data.PlayableVideo
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.habits.HabitPreferences
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.withContext

/** Whether the open video is scanned now, or why not. */
enum class OpenScanDecision {
    SCAN,
    NO_MODEL,

    /** Scanned with the installed model already: its segments are there. */
    ALREADY_SCANNED,

    /** Wi-Fi only is on and the phone is on mobile data: the scan reads the stream once more, which costs as much again. */
    WAITING_FOR_WIFI,

    /** Battery saver is on. */
    BATTERY_SAVER,
}

/**
 * When the open video is scanned: with a model, not scanned with it yet, not in battery saver, and
 * on mobile data only when Background → Wi-Fi only is off or the video plays from a local copy
 * (grabbing frames reads the stream about as much as watching it again). No charging needed.
 */
object OpenScanPolicy {

    /** The decision for one video page. */
    fun decide(modelReady: Boolean, alreadyScanned: Boolean, wifiOnly: Boolean, metered: Boolean, localCopy: Boolean, powerSave: Boolean): OpenScanDecision {
        if (!modelReady) {
            return OpenScanDecision.NO_MODEL
        }
        if (alreadyScanned) {
            return OpenScanDecision.ALREADY_SCANNED
        }
        if (powerSave) {
            return OpenScanDecision.BATTERY_SAVER
        }
        if (wifiOnly && metered && !localCopy) {
            return OpenScanDecision.WAITING_FOR_WIFI
        }
        return OpenScanDecision.SCAN
    }

    /** The decision for [videoId] on this phone right now. */
    fun decide(context: Context, videoId: String, localCopy: Boolean): OpenScanDecision {
        val connectivity = context.getSystemService(ConnectivityManager::class.java)
        val power = context.getSystemService(PowerManager::class.java)
        return decide(
            modelReady = SceneClassifier.available(),
            alreadyScanned = videoId in SceneScans.scannedIds(),
            wifiOnly = HabitPreferences.wifiOnly.value,
            metered = connectivity?.isActiveNetworkMetered ?: true,
            localCopy = localCopy,
            powerSave = power?.isPowerSaveMode ?: false,
        )
    }
}

/**
 * How far the open video's scan is and what it found so far; [done] once every frame is read (or
 * when only the rough pass runs). [segments] are the ai segments to show in place of the stored
 * ones (null: keep the stored ones); [broad] are the display-only broad sections ([BroadLabels])
 * that name the stretches the segments leave empty.
 */
data class OpenScanProgress(val progress: Float, val segments: List<TimedSegment>?, val broad: List<TimedSegment>, val done: Boolean)

/** What the rough pass read from the site's preview stills (or the cover) spread over the runtime. */
class RoughReading(val segments: List<TimedSegment>, val broad: List<TimedSegment>)

/**
 * The rough pass's pure parts: which pictures it reads and how the stream scan's results replace
 * its own where the stream is scanned.
 */
object RoughPass {

    /** With fewer stills than this the cover is read too. */
    const val MIN_STILLS = 4

    /**
     * The pictures to read, in the order they're spread over the runtime: the stills, with the
     * cover in the middle when there are fewer than [MIN_STILLS] (a cover usually shows the main act).
     */
    fun pictures(stills: List<String>, cover: String?): List<String> {
        if (stills.size >= MIN_STILLS || cover == null) {
            return stills
        }
        val middle = stills.size / 2
        return stills.subList(0, middle) + cover + stills.subList(middle, stills.size)
    }

    /** The stream's segments, plus the rough ones outside [scannedSpans]; null [rough] adds nothing. */
    fun merge(stream: List<TimedSegment>, rough: List<TimedSegment>?, scannedSpans: List<ClosedFloatingPointRange<Double>>): List<TimedSegment> {
        if (rough == null) {
            return stream
        }
        return SceneSegments.mergeTouching(stream + SceneSegments.outside(rough, scannedSpans))
    }
}

/**
 * Scans the video open on the page. First a rough pass ([rough]) over the site's preview stills,
 * which gives sections within seconds; then the stream beside playback ([ScanPace.BESIDE_PLAYBACK]),
 * from where it plays, whose results replace the rough ones span by span: every batch reports the
 * segments found so far; the stream's (never the rough or broad ones) are saved as the video's ai
 * segments every quarter and at the end, when the video counts as scanned with this model. Leaving
 * the page cancels it; what was saved stays, and the next visit or the background scan starts over.
 */
class OpenVideoScan(context: Context) {

    private val appContext = context.applicationContext
    private val scanner = SceneScanner(appContext)

    /**
     * Reads [stills] (paths or URLs; with the cover when there are few) spread evenly over
     * [durationSeconds]: each picture stands for its share of the runtime. Null without a model or
     * when no picture loads.
     */
    suspend fun rough(stills: List<String>, coverUrl: String?, durationSeconds: Double): RoughReading? {
        val manifest = SceneClassifier.manifest() ?: return null
        val started = System.currentTimeMillis()
        val pictures = RoughPass.pictures(stills, coverUrl)
        val bitmaps = coroutineScope { pictures.map { picture -> async { load(picture) } }.awaitAll() }.filterNotNull()
        if (bitmaps.isEmpty() || durationSeconds <= 0) {
            Log.i(SceneScanner.TAG, "rough pass: ${pictures.size} pictures, none loaded")
            return null
        }
        val step = durationSeconds / bitmaps.size
        val timeline = ScanTimeline(List(bitmaps.size) { index -> index * step }, step, 0, SegmentRules.STILLS)
        val read = withContext(ScanPace.BESIDE_PLAYBACK.dispatcher) { readStills(timeline, bitmaps, manifest) }
        if (!read) {
            return null
        }
        val reading = RoughReading(timeline.segments(manifest), timeline.broadSegments(manifest))
        logRough(stills.size, pictures.size > stills.size, timeline, reading, System.currentTimeMillis() - started, manifest)
        return reading
    }

    /**
     * Scans [video] of [videoId] starting at [startSeconds]; [onProgress] gets every batch's result,
     * the stream's in its scanned spans and [rough]'s elsewhere.
     */
    suspend fun run(videoId: String, video: PlayableVideo, startSeconds: Double, rough: RoughReading?, onProgress: (OpenScanProgress) -> Unit) {
        val manifest = SceneClassifier.manifest() ?: return
        var savedQuarter = 0
        val timeline = scanner.scanStream(video, startSeconds, ScanPace.BESIDE_PLAYBACK) { partial ->
            val segments = withContext(Dispatchers.Default) { partial.segments(manifest) }
            val broad = withContext(Dispatchers.Default) { partial.broadSegments(manifest) }
            val spans = partial.scannedSpans()
            onProgress(OpenScanProgress(partial.progress, RoughPass.merge(segments, rough?.segments, spans), RoughPass.merge(broad, rough?.broad, spans), done = false))
            val quarter = (partial.progress * QUARTERS).toInt()
            if (quarter > savedQuarter && !partial.complete) {
                savedQuarter = quarter
                save(videoId, segments)
            }
        } ?: return
        if (!timeline.complete) {
            return
        }
        val segments = timeline.segments(manifest)
        onProgress(OpenScanProgress(1f, segments, timeline.broadSegments(manifest), done = true))
        SceneRepository.replaceSegments(videoId, segments)
        SceneScans.markScanned(videoId)
    }

    /** The stills through the frame graph in batches; false when the model went away. */
    private fun readStills(timeline: ScanTimeline, bitmaps: List<Bitmap>, manifest: SceneModelManifest): Boolean {
        for (batch in bitmaps.indices.chunked(SceneScanner.BATCH_FRAMES)) {
            val outputs = SceneClassifier.frames(batch.map { index -> bitmaps[index] }) ?: return false
            val probabilities = perFrame(outputs.logits, manifest)
            for ((position, index) in batch.withIndex()) {
                timeline.add(index, outputs.embeddings[position], probabilities[position])
            }
        }
        return true
    }

    /** One picture as a software bitmap, small (Coil, its cache first); null when it doesn't load. */
    private suspend fun load(picture: String): Bitmap? {
        val url = ServerSettings.imageUrl(picture) ?: return null
        val request = ImageRequest.Builder(appContext).data(url).size(STILL_SIZE).allowHardware(false).build()
        val result = SingletonImageLoader.get(appContext).execute(request) as? SuccessResult ?: return null
        return result.image.toBitmap()
    }

    /** Logs the rough pass: how many stills, whether the cover was read, how long it took, the strongest labels and what it found. */
    private fun logRough(stills: Int, withCover: Boolean, timeline: ScanTimeline, reading: RoughReading, milliseconds: Long, manifest: SceneModelManifest) {
        val strongest = timeline.strongestOverall(LOGGED_LABELS, manifest).joinToString { (label, probability) -> "$label ${"%.2f".format(probability)}" }
        val found = reading.segments.map { segment -> segment.tag }.distinct().joinToString()
        val broad = reading.broad.map { segment -> segment.tag }.distinct().joinToString()
        Log.i(SceneScanner.TAG, "rough pass: $stills stills, cover $withCover, ${timeline.scannedCount} read in ${milliseconds}ms; top: $strongest; found: $found; broad: $broad")
    }

    /** Saves what's found so far; a failed save waits for the next one. */
    private suspend fun save(videoId: String, segments: List<TimedSegment>) {
        try {
            SceneRepository.replaceSegments(videoId, segments)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // The next quarter or the end saves again.
        }
    }

    private companion object {
        const val QUARTERS = 4

        /** Stills are loaded at about this many pixels on a side; the model reads them smaller still. */
        const val STILL_SIZE = 320

        /** Labels named in the rough pass's log line. */
        const val LOGGED_LABELS = 5
    }
}
