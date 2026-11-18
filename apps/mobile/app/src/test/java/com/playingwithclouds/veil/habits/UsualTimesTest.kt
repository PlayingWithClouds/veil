package com.playingwithclouds.veil.habits

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import java.time.ZoneId
import java.time.ZonedDateTime

class UsualTimesTest {

    private val zone = ZoneId.of("Europe/Berlin")

    /** A Monday at [hour]:[minute], [weeksAgo] weeks back from the test's week. */
    private fun monday(hour: Int, minute: Int, weeksAgo: Long = 0): ZonedDateTime {
        return ZonedDateTime.of(2026, 11, 16, hour, minute, 0, 0, zone).minusWeeks(weeksAgo)
    }

    @Test
    fun findsTheEveningHabit() {
        val events = (1L..4L).flatMap { week ->
            listOf(HabitEvent(monday(22, 5, week), finish = false), HabitEvent(monday(22, 20, week), finish = true))
        }
        val next = UsualTimes(events).nextStart(monday(12, 0))
        assertEquals(monday(21, 30), next)
    }

    @Test
    fun aSingleNightIsNoHabit() {
        val events = listOf(HabitEvent(monday(3, 0, 1), finish = true))
        assertNull(UsualTimes(events).nextStart(monday(12, 0)))
    }
}
