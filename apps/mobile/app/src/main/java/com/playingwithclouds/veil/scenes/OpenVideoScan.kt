package com.playingwithclouds.veil.scenes

import android.content.Context
import android.net.ConnectivityManager
import android.os.PowerManager
import com.playingwithclouds.veil.data.PlayableVideo
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.habits.HabitPreferences
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
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

/** How far the open video's scan is and what it found so far; [done] once every frame is read. */
data class OpenScanProgress(val progress: Float, val segments: List<TimedSegment>, val done: Boolean)

/**
 * Scans the video open on the page beside playback ([ScanPace.BESIDE_PLAYBACK]), from where it
 * plays: every batch reports the segments found so far, which are saved as the video's ai segments
 * every quarter and at the end, when the video counts as scanned with this model. Leaving the page
 * cancels it; what was saved stays, and the next visit or the background scan starts over.
 */
class OpenVideoScan(context: Context) {

    private val scanner = SceneScanner(context.applicationContext)

    /** Scans [video] of [videoId] starting at [startSeconds]; [onProgress] gets every batch's result. */
    suspend fun run(videoId: String, video: PlayableVideo, startSeconds: Double, onProgress: (OpenScanProgress) -> Unit) {
        val manifest = SceneClassifier.manifest() ?: return
        var savedQuarter = 0
        val timeline = scanner.scanStream(video, startSeconds, ScanPace.BESIDE_PLAYBACK) { partial ->
            val segments = withContext(Dispatchers.Default) { partial.segments(manifest) }
            onProgress(OpenScanProgress(partial.progress, segments, done = false))
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
        onProgress(OpenScanProgress(1f, segments, done = true))
        SceneRepository.replaceSegments(videoId, segments)
        SceneScans.markScanned(videoId)
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
    }
}
