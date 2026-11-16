package com.playingwithclouds.veil.ui.cams

import com.playingwithclouds.veil.data.CamRepository
import kotlin.math.roundToInt

/** The genders the room listing can narrow to, with their labels; null is all. */
val CAM_GENDERS: List<Pair<String?, String>> = listOf(
    null to "All",
    "female" to "Female",
    "male" to "Male",
    "couple" to "Couples",
    "trans" to "Trans",
)

/** What auto-record can be set to capture: the best rendition or a target height. */
val AUTO_RECORD_QUALITIES = listOf(CamRepository.SOURCE_QUALITY, "1080", "720", "480", "360", "240")

/** Behind the live edge at least this far, the record menu offers to start at the playhead. */
const val RECORD_FROM_HERE_MINIMUM_SECONDS = 10

/** A start in the past is offered when the buffer holds this much less than it, since the buffer trails a little. */
const val RECORD_FROM_PAST_SLACK_SECONDS = 10

/** The fixed starts in the past the record menu offers, in seconds. */
private val RECORD_FROM_PAST_SECONDS = listOf(60, 180)

/** A way to start a recording in the past: its menu label and how many seconds back. */
data class RecordStart(val label: String, val secondsAgo: Int)

/**
 * The starts in the past a room's record menu offers: "From here" at the playhead when it is
 * [lagSeconds] behind the live edge, and one and three minutes back when the room's buffer holds
 * about [bufferedSeconds] of media.
 */
fun recordStarts(bufferedSeconds: Double, lagSeconds: Double): List<RecordStart> {
    val starts = mutableListOf<RecordStart>()
    if (lagSeconds >= RECORD_FROM_HERE_MINIMUM_SECONDS) {
        starts.add(RecordStart("From here", lagSeconds.roundToInt()))
    }
    for (seconds in RECORD_FROM_PAST_SECONDS) {
        if (bufferedSeconds + RECORD_FROM_PAST_SLACK_SECONDS >= seconds) {
            starts.add(RecordStart("From ${seconds / SECONDS_PER_MINUTE} min ago", seconds))
        }
    }
    return starts
}

private const val SECONDS_PER_MINUTE = 60

/** "Source", or "480p" for a stored recording quality. */
fun recordQualityLabel(quality: String): String {
    if (quality == CamRepository.SOURCE_QUALITY) {
        return "Source"
    }
    return "${quality}p"
}

/** The quality to record at while watching at [pinnedHeight]: that height, or the source when the player picks. */
fun recordQualityFor(pinnedHeight: Int?): String {
    if (pinnedHeight == null || pinnedHeight <= 0) {
        return CamRepository.SOURCE_QUALITY
    }
    return pinnedHeight.toString()
}

/** A viewer count as "1.2k" from a thousand on. */
fun formatViewers(viewers: Int): String {
    if (viewers < THOUSAND) {
        return viewers.toString()
    }
    val thousands = viewers / THOUSAND.toDouble()
    if (thousands >= TEN) {
        return "${thousands.roundToInt()}k"
    }
    return "${(thousands * TEN).roundToInt() / TEN.toDouble()}k"
}

private const val THOUSAND = 1000
private const val TEN = 10
