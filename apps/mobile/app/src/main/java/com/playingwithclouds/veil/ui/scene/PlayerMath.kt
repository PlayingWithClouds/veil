package com.playingwithclouds.veil.ui.scene

import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.util.formatClock
import java.util.Locale

/** Seconds one double tap on either half of the player jumps. */
const val DOUBLE_TAP_SEEK_SECONDS = 10

/** Playback speeds offered in the player settings. */
val PLAYBACK_SPEEDS = listOf(0.25f, 0.5f, 0.75f, 1f, 1.25f, 1.5f, 2f)

/** Frame rate assumed for stepping when the stream does not report one. */
const val FALLBACK_FRAME_RATE = 30f

private const val MILLISECONDS_PER_SECOND = 1000L

/** Which way a double tap at [x] in a player [width] wide seeks: -1 (back) on the left half, 1 (forward) on the right. */
fun seekDirectionAt(x: Float, width: Float): Int {
    if (x < width / 2f) {
        return -1
    }
    return 1
}

/**
 * The position after jumping [deltaMilliseconds] from [positionMilliseconds], kept inside the
 * video. An unknown duration (zero or negative, like `C.TIME_UNSET`) only bounds the start.
 */
fun clampedSeekTarget(positionMilliseconds: Long, deltaMilliseconds: Long, durationMilliseconds: Long): Long {
    val target = maxOf(0L, positionMilliseconds + deltaMilliseconds)
    if (durationMilliseconds <= 0) {
        return target
    }
    return minOf(target, durationMilliseconds)
}

/** The fraction (0 to 1) of a bar [width] wide at [x]. */
fun fractionAt(x: Float, width: Float): Float {
    if (width <= 0f) {
        return 0f
    }
    return (x / width).coerceIn(0f, 1f)
}

/** How far (0 to 1) [positionMilliseconds] is into the video; 0 while the duration is unknown. */
fun fractionOf(positionMilliseconds: Long, durationMilliseconds: Long): Float {
    if (durationMilliseconds <= 0) {
        return 0f
    }
    return (positionMilliseconds.toFloat() / durationMilliseconds).coerceIn(0f, 1f)
}

/** The position at [fraction] of the video. */
fun positionAt(fraction: Float, durationMilliseconds: Long): Long {
    if (durationMilliseconds <= 0) {
        return 0
    }
    return (fraction.coerceIn(0f, 1f) * durationMilliseconds).toLong()
}

/** A playback position as a clock, like "1:23" or "1:02:03". */
fun formatPlaybackClock(milliseconds: Long): String {
    return formatClock(milliseconds.toDouble() / MILLISECONDS_PER_SECOND)
}

/** Elapsed and total time, like "1:23 / 6:40"; only the elapsed time while the duration is unknown. */
fun formatPlaybackTime(positionMilliseconds: Long, durationMilliseconds: Long): String {
    val elapsed = formatPlaybackClock(positionMilliseconds)
    if (durationMilliseconds <= 0) {
        return elapsed
    }
    return "$elapsed / ${formatPlaybackClock(durationMilliseconds)}"
}

/** A playback speed for the settings menu: "Normal" at 1×, else like "1.25×". */
fun formatSpeed(speed: Float): String {
    if (speed == 1f) {
        return "Normal"
    }
    val number = String.format(Locale.ROOT, "%.2f", speed).trimEnd('0').trimEnd('.')
    return "$number×"
}

/** The running total shown while double taps seek; [id] changes with every tap so its timer restarts. */
data class SeekFeedback(val direction: Int, val seconds: Int, val id: Long)

/** The feedback after another double tap in [direction]: repeated taps the same way add up, a turn starts over. */
fun accumulateSeek(previous: SeekFeedback?, direction: Int): SeekFeedback {
    if (previous == null) {
        return SeekFeedback(direction, DOUBLE_TAP_SEEK_SECONDS, 1)
    }
    if (previous.direction == direction) {
        return previous.copy(seconds = previous.seconds + DOUBLE_TAP_SEEK_SECONDS, id = previous.id + 1)
    }
    return SeekFeedback(direction, DOUBLE_TAP_SEEK_SECONDS, previous.id + 1)
}

/** The feedback label, like "+20 s" or "−10 s". */
fun seekFeedbackLabel(feedback: SeekFeedback): String {
    if (feedback.direction < 0) {
        return "−${feedback.seconds} s"
    }
    return "+${feedback.seconds} s"
}

/** Whether the controls should hide by themselves: shown, playing, and nobody is using them. */
fun shouldAutoHide(visible: Boolean, playing: Boolean, scrubbing: Boolean, menuOpen: Boolean): Boolean {
    return visible && playing && !scrubbing && !menuOpen
}

/**
 * Whether a vertical drag opens the related panel: it starts in the lower half of a player [height]
 * tall and travels at least [thresholdPixels] up ([dragY] is negative upwards).
 */
fun isRelatedPanelSwipe(startY: Float, height: Float, dragY: Float, thresholdPixels: Float): Boolean {
    return startY >= height / 2f && dragY <= -thresholdPixels
}

/** Whether a drag of the related panel by [dragY] (positive downwards) dismisses it. */
fun isPanelDismissSwipe(dragY: Float, thresholdPixels: Float): Boolean {
    return dragY >= thresholdPixels
}

/** One selectable video quality: a track of a group in the player's current tracks. */
data class VideoQuality(val height: Int, val groupIndex: Int, val trackIndex: Int)

/** The qualities to offer, highest first, one per height. */
fun distinctQualities(qualities: List<VideoQuality>): List<VideoQuality> {
    return qualities
        .filter { quality -> quality.height > 0 }
        .sortedByDescending { quality -> quality.height }
        .distinctBy { quality -> quality.height }
}

/** A quality's menu label, like "1080p". */
fun videoQualityLabel(height: Int): String {
    return "${height}p"
}

/** The A and B points of a loop, in milliseconds; either may still be unset. */
data class AbLoop(val startMilliseconds: Long? = null, val endMilliseconds: Long? = null) {

    /** Both points are set, so playback is looping. */
    val isActive: Boolean
        get() = startMilliseconds != null && endMilliseconds != null
}

/**
 * The loop after the user marks the current position: the first mark sets A, the second sets B (a
 * position at or before A moves A instead), the third clears the loop.
 */
fun advanceLoop(loop: AbLoop, positionMilliseconds: Long): AbLoop {
    val start = loop.startMilliseconds
    if (start == null) {
        return AbLoop(startMilliseconds = positionMilliseconds)
    }
    if (loop.endMilliseconds != null) {
        return AbLoop()
    }
    if (positionMilliseconds <= start) {
        return AbLoop(startMilliseconds = positionMilliseconds)
    }
    return AbLoop(start, positionMilliseconds)
}

/** Where playback jumps back to: A once [positionMilliseconds] has reached B of an active loop, else null. */
fun loopRestartTarget(loop: AbLoop, positionMilliseconds: Long): Long? {
    val start = loop.startMilliseconds
    val end = loop.endMilliseconds
    if (start == null || end == null || positionMilliseconds < end) {
        return null
    }
    return start
}

/** The loop's state for the settings menu, like "Off", "A at 1:23" or "1:23 – 2:05". */
fun loopLabel(loop: AbLoop): String {
    val start = loop.startMilliseconds ?: return "Off"
    val end = loop.endMilliseconds ?: return "A at ${formatPlaybackClock(start)}"
    return "${formatPlaybackClock(start)} – ${formatPlaybackClock(end)}"
}

/** What the loop menu entry does next: set A, set B or clear. */
fun loopActionLabel(loop: AbLoop): String {
    if (loop.startMilliseconds == null) {
        return "Set loop start (A)"
    }
    if (loop.endMilliseconds == null) {
        return "Set loop end (B)"
    }
    return "Clear loop"
}

/** Length of one video frame in milliseconds at [frameRate], falling back to 30 fps when it is unknown. */
fun frameDurationMilliseconds(frameRate: Float): Long {
    var rate = frameRate
    if (rate <= 0f) {
        rate = FALLBACK_FRAME_RATE
    }
    return maxOf(1L, Math.round(MILLISECONDS_PER_SECOND / rate.toDouble()))
}

/** Whether a video of this size is taller than wide, so fullscreen should stay in portrait. */
fun isVerticalVideo(width: Int, height: Int): Boolean {
    return width > 0 && height > width
}

/** The widest or tallest shape Android accepts for a picture-in-picture window. */
private const val MAX_PICTURE_IN_PICTURE_RATIO = 2.39f

/**
 * The window shape (width to height) for a video of this size: its own shape, kept within what
 * Android allows, and 16:9 while the size is unknown.
 */
fun pictureInPictureAspect(width: Int, height: Int): Pair<Int, Int> {
    if (width <= 0 || height <= 0) {
        return Pair(16, 9)
    }
    val ratio = width.toFloat() / height
    if (ratio > MAX_PICTURE_IN_PICTURE_RATIO) {
        return Pair(239, 100)
    }
    if (ratio < 1f / MAX_PICTURE_IN_PICTURE_RATIO) {
        return Pair(100, 239)
    }
    return Pair(width, height)
}

/** The loop's menu entry: what it does next, followed by the marked positions once there are any. */
fun loopMenuLabel(loop: AbLoop): String {
    if (loop.startMilliseconds == null) {
        return loopActionLabel(loop)
    }
    return "${loopActionLabel(loop)} · ${loopLabel(loop)}"
}

/** The markers in playback order, which the chips and [activeMarkerIndex] rely on. */
fun markersInPlayOrder(markers: List<SceneMarker>): List<SceneMarker> {
    return markers.sortedBy { marker -> marker.seconds }
}

/**
 * Index of the marker playback is in: the last one that started at or before [positionSeconds] in
 * [markers] (in play order), or -1 before the first.
 */
fun activeMarkerIndex(markers: List<SceneMarker>, positionSeconds: Double): Int {
    var active = -1
    for ((index, marker) in markers.withIndex()) {
        if (marker.seconds > positionSeconds) {
            break
        }
        active = index
    }
    return active
}

/** A marker chip's text: what the marker says and when, like "Doggy · 12:30"; just the time for a marker without text. */
fun markerChipLabel(marker: SceneMarker): String {
    val time = formatClock(marker.seconds)
    if (marker.title.isBlank()) {
        return time
    }
    return "${marker.title} · $time"
}

/** Where each marker sits on the seek bar, as fractions of a video [durationSeconds] long; markers beyond it are left out. */
fun markerFractions(markers: List<SceneMarker>, durationSeconds: Double): List<Float> {
    if (durationSeconds <= 0) {
        return emptyList()
    }
    return markers
        .filter { marker -> marker.seconds in 0.0..durationSeconds }
        .map { marker -> (marker.seconds / durationSeconds).toFloat() }
}
