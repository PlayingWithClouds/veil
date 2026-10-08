package com.playingwithclouds.veil.ui.scene

import com.playingwithclouds.veil.util.formatClock
import java.util.Locale

/** Seconds one double tap on either half of the player jumps. */
const val DOUBLE_TAP_SEEK_SECONDS = 10

/** Playback speeds offered in the player settings. */
val PLAYBACK_SPEEDS = listOf(0.5f, 0.75f, 1f, 1.25f, 1.5f, 2f)

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
