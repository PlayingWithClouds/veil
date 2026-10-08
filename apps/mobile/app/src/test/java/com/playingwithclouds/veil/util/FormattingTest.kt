package com.playingwithclouds.veil.util

import java.time.Instant
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Test

class FormattingTest {

    private val now = Instant.parse("2026-10-08T12:00:00Z")

    @Test
    fun durationUnderAnHourIsInMinutes() {
        assertEquals("37 min", formatDuration(37 * 60))
        assertEquals("1 min", formatDuration(20))
    }

    @Test
    fun durationOverAnHourHasHoursAndMinutes() {
        assertEquals("1 h 12 min", formatDuration(72 * 60))
    }

    @Test
    fun unknownDurationIsNull() {
        assertNull(formatDuration(null))
        assertNull(formatDuration(0))
    }

    @Test
    fun clockShowsMinutesAndSeconds() {
        assertEquals("4:05", formatClock(245.0))
        assertEquals("0:00", formatClock(-3.0))
    }

    @Test
    fun clockShowsHoursWhenNeeded() {
        assertEquals("1:02:03", formatClock(3723.0))
    }

    @Test
    fun countsAreCompact() {
        assertEquals("999", formatCount(999))
        assertEquals("1.2K", formatCount(1200))
        assertEquals("3.5M", formatCount(3_500_000))
    }

    @Test
    fun bytesPickTheLargestUnit() {
        assertEquals("0 B", formatBytes(0.0))
        assertEquals("512 B", formatBytes(512.0))
        assertEquals("2 KB", formatBytes(2048.0))
        assertEquals("335.3 MB", formatBytes(335.3 * 1024 * 1024))
        assertEquals("1.2 GB", formatBytes(1.2 * 1024 * 1024 * 1024))
        assertEquals("2.0 MB/s", formatRate(2.0 * 1024 * 1024))
    }

    @Test
    fun timeLeftIsRoundedToMinutes() {
        assertEquals("under a minute left", formatTimeLeft(30.0))
        assertEquals("3 min left", formatTimeLeft(180.0))
        assertEquals("1 h 20 min left", formatTimeLeft(4800.0))
    }

    @Test
    fun intervalsAreLabeledInTheLargestWholeUnit() {
        assertEquals("Every hour", intervalLabel(1))
        assertEquals("Every 6 hours", intervalLabel(6))
        assertEquals("Every day", intervalLabel(24))
        assertEquals("Every 2 days", intervalLabel(48))
        assertEquals("Every week", intervalLabel(168))
        assertEquals("Every 2 weeks", intervalLabel(336))
    }

    @Test
    fun relativeTimeLooksBackAndForward() {
        assertEquals("5 minutes ago", formatRelativeTime("2026-10-08T11:55:00Z", now))
        assertEquals("1 hour ago", formatRelativeTime("2026-10-08T11:00:00Z", now))
        assertEquals("in 6 hours", formatRelativeTime("2026-10-08T18:00:00Z", now))
        assertEquals("2 days ago", formatRelativeTime("2026-10-06T12:00:00Z", now))
        assertEquals("just now", formatRelativeTime("2026-10-08T11:59:30Z", now))
    }

    @Test
    fun relativeTimeAcceptsPlainDatesAndOffsets() {
        assertEquals("1 year ago", formatRelativeTime("2025-10-08", now))
        assertEquals("1 hour ago", formatRelativeTime("2026-10-08T13:00:00+02:00", now))
    }

    @Test
    fun unparseableTimestampsAreShownAsTheyAre() {
        assertEquals("soon", formatRelativeTime("soon", now))
        assertNull(parseTimestamp("soon"))
        assertNotNull(parseTimestamp("2026-10-08T12:00:00Z"))
    }

    @Test
    fun scheduleLabelDescribesLastAndNextRun() {
        val label = scheduleLabel("2026-10-08T11:55:00Z", "2026-10-08T18:00:00Z", enabled = true, now = now)
        assertEquals("Ran 5 minutes ago · next in 6 hours", label)
    }

    @Test
    fun scheduleLabelOfAPausedSubscription() {
        assertEquals("Not run yet · paused", scheduleLabel(null, null, enabled = false, now = now))
    }

    @Test
    fun tagsKeepAcronymsAndLowercaseTheRest() {
        assertEquals("POV blonde", tagLabel("POV Blonde"))
        assertEquals("big tits", tagLabel("Big Tits"))
        assertEquals("MILF", tagLabel("MILF"))
    }
}
