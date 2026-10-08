package com.playingwithclouds.veil.util

import java.time.Duration
import java.time.Instant
import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.time.format.DateTimeParseException
import java.util.Locale
import kotlin.math.abs
import kotlin.math.roundToInt

private const val SECONDS_PER_MINUTE = 60
private const val MINUTES_PER_HOUR = 60
private const val HOURS_PER_DAY = 24
private const val HOURS_PER_WEEK = 168
private const val KILOBYTE = 1024.0
private const val MEGABYTE = KILOBYTE * 1024
private const val GIGABYTE = MEGABYTE * 1024

/** Runtime like "37 min" or "1 h 12 min"; null when unknown. */
fun formatDuration(seconds: Int?): String? {
    if (seconds == null || seconds <= 0) {
        return null
    }
    val minutes = maxOf(1, (seconds.toDouble() / SECONDS_PER_MINUTE).roundToInt())
    if (minutes < MINUTES_PER_HOUR) {
        return "$minutes min"
    }
    return "${minutes / MINUTES_PER_HOUR} h ${minutes % MINUTES_PER_HOUR} min"
}

/** Player clock like "4:05" or "1:02:03". */
fun formatClock(totalSeconds: Double): String {
    val wholeSeconds = maxOf(0.0, totalSeconds).toLong()
    val hours = wholeSeconds / 3600
    val minutes = (wholeSeconds % 3600) / 60
    val seconds = wholeSeconds % 60
    if (hours > 0) {
        return String.format(Locale.ROOT, "%d:%02d:%02d", hours, minutes, seconds)
    }
    return String.format(Locale.ROOT, "%d:%02d", minutes, seconds)
}

/** Compact count like "1.2K" for scene counts. */
fun formatCount(count: Int): String {
    if (count >= 1_000_000) {
        return String.format(Locale.ROOT, "%.1fM", count / 1_000_000.0)
    }
    if (count >= 1_000) {
        return String.format(Locale.ROOT, "%.1fK", count / 1_000.0)
    }
    return count.toString()
}

/** "335.3 MB", "1.2 GB", "0 B". */
fun formatBytes(bytes: Double): String {
    if (bytes < 1) {
        return "0 B"
    }
    if (bytes >= GIGABYTE) {
        return String.format(Locale.ROOT, "%.1f GB", bytes / GIGABYTE)
    }
    if (bytes >= MEGABYTE) {
        return String.format(Locale.ROOT, "%.1f MB", bytes / MEGABYTE)
    }
    if (bytes >= KILOBYTE) {
        return String.format(Locale.ROOT, "%.0f KB", bytes / KILOBYTE)
    }
    return "${bytes.roundToInt()} B"
}

/** "2.0 MB/s". */
fun formatRate(bytesPerSecond: Double): String {
    return "${formatBytes(bytesPerSecond)}/s"
}

/** Labels a re-run interval: "Every hour", "Every 6 hours", "Every day", "Every week". */
fun intervalLabel(hours: Int): String {
    if (hours == 1) {
        return "Every hour"
    }
    if (hours == HOURS_PER_DAY) {
        return "Every day"
    }
    if (hours == HOURS_PER_WEEK) {
        return "Every week"
    }
    if (hours % HOURS_PER_WEEK == 0) {
        return "Every ${hours / HOURS_PER_WEEK} weeks"
    }
    if (hours % HOURS_PER_DAY == 0) {
        return "Every ${hours / HOURS_PER_DAY} days"
    }
    return "Every $hours hours"
}

/** Parses an ISO timestamp or plain date; null when it is neither. */
fun parseTimestamp(value: String): Instant? {
    try {
        return Instant.parse(value)
    } catch (error: DateTimeParseException) {
        // Not a UTC timestamp; try the other shapes below.
    }
    try {
        return OffsetDateTime.parse(value).toInstant()
    } catch (error: DateTimeParseException) {
        // Not an offset timestamp either.
    }
    try {
        return LocalDate.parse(value.take(10)).atStartOfDay().toInstant(ZoneOffset.UTC)
    } catch (error: DateTimeParseException) {
        return null
    }
}

/** A time unit with how many seconds it spans, largest first. */
private enum class RelativeUnit(val seconds: Long, val label: String) {
    YEAR(365L * 86400, "year"),
    MONTH(30L * 86400, "month"),
    WEEK(7L * 86400, "week"),
    DAY(86400L, "day"),
    HOUR(3600L, "hour"),
    MINUTE(60L, "minute"),
}

/** "3 hours ago", "in 20 minutes", "just now": the timestamp relative to now, in the largest fitting unit. */
fun formatRelativeTime(timestamp: String, now: Instant = Instant.now()): String {
    val moment = parseTimestamp(timestamp) ?: return timestamp
    val differenceSeconds = Duration.between(now, moment).seconds
    for (unit in RelativeUnit.entries) {
        if (abs(differenceSeconds) < unit.seconds) {
            continue
        }
        val amount = abs((differenceSeconds.toDouble() / unit.seconds).roundToInt())
        var noun = unit.label
        if (amount != 1) {
            noun += "s"
        }
        if (differenceSeconds < 0) {
            return "$amount $noun ago"
        }
        return "in $amount $noun"
    }
    return "just now"
}

/** "3 months ago"-style label for a release date, falling back to the raw value. */
fun formatReleaseDate(value: String?, now: Instant = Instant.now()): String? {
    if (value == null) {
        return null
    }
    return formatRelativeTime(value, now)
}

/** Time left as "3 min left", "1 h 20 min left" or "under a minute left". */
fun formatTimeLeft(seconds: Double): String {
    if (seconds < SECONDS_PER_MINUTE) {
        return "under a minute left"
    }
    val minutes = (seconds / SECONDS_PER_MINUTE).roundToInt()
    if (minutes < MINUTES_PER_HOUR) {
        return "$minutes min left"
    }
    return "${minutes / MINUTES_PER_HOUR} h ${minutes % MINUTES_PER_HOUR} min left"
}

/**
 * A tag name in one consistent casing: lowercase, except short all-caps words that are
 * acronyms ("POV", "MILF"). Sites disagree on casing, so the stored spelling is whatever
 * arrived first.
 */
fun tagLabel(name: String): String {
    return name.split(" ").joinToString(" ") { word ->
        keepOrLowercase(word)
    }
}

/** Keeps short all-caps words (acronyms), lowercases everything else. */
private fun keepOrLowercase(word: String): String {
    val isAcronym = word.length in 2..5 && word == word.uppercase() && word.any { letter -> letter in 'A'..'Z' }
    if (isAcronym) {
        return word
    }
    return word.lowercase()
}

/** Describes when a subscription last ran and runs next: "Ran 5 minutes ago · next in 6 hours". */
fun scheduleLabel(lastRunAt: String?, nextRunAt: String?, enabled: Boolean, now: Instant = Instant.now()): String {
    var lastRun = "Not run yet"
    if (lastRunAt != null) {
        lastRun = "Ran ${formatRelativeTime(lastRunAt, now)}"
    }
    if (!enabled) {
        return "$lastRun · paused"
    }
    if (nextRunAt == null) {
        return lastRun
    }
    return "$lastRun · next ${formatRelativeTime(nextRunAt, now)}"
}

/** "1 video", "2 videos", "1.2K videos": a video count with its noun. */
fun formatVideoCount(count: Int): String {
    if (count == 1) {
        return "1 video"
    }
    return "${formatCount(count)} videos"
}
