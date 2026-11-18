package com.playingwithclouds.veil.habits

import java.time.DayOfWeek
import java.time.Duration
import java.time.ZonedDateTime
import java.time.temporal.ChronoUnit

/** Minutes per slot of the week the habits are counted in. */
private const val SLOT_MINUTES = 30

private const val MINUTES_PER_DAY = 24 * 60
private const val SLOTS_PER_DAY = MINUTES_PER_DAY / SLOT_MINUTES
private const val SLOTS_PER_WEEK = 7 * SLOTS_PER_DAY

/** An O says far more about the time than opening a video. */
private const val FINISH_WEIGHT = 3.0
private const val OPEN_WEIGHT = 1.0

/** The same time on other weekdays counts this much: habits are partly daily, partly weekly. */
private const val OTHER_DAYS_SHARE = 0.35

/** The slots either side count this much, so 21:50 and 22:10 make one habit. */
private const val NEIGHBOUR_SHARE = 0.5

/** Fewest weighted events a slot needs to be a usual time, so a single late night isn't a habit. */
private const val MIN_SLOT_SCORE = 6.0

/** How far ahead the next usual time is looked for. */
private val LOOKAHEAD: Duration = Duration.ofHours(30)

/** One thing the user did and when: an O ([finish]) or opening a video. */
data class HabitEvent(val at: ZonedDateTime, val finish: Boolean)

/**
 * When the user usually watches, learned from when they opened videos and finished (weighted
 * more), per half hour of the week with some help from the same time on other days. The start of
 * the next usual time is where a run of busy slots begins.
 */
class UsualTimes(events: List<HabitEvent>) {

    private val scores: DoubleArray = smoothed(counted(events))

    /** The start of the next usual time after [from] within the next day or so; null when there's no habit yet. */
    fun nextStart(from: ZonedDateTime): ZonedDateTime? {
        var slotStart = from.truncatedTo(ChronoUnit.HOURS).plusMinutes(((from.minute / SLOT_MINUTES) * SLOT_MINUTES).toLong())
        var previousBusy = isBusy(slotStart.minusMinutes(SLOT_MINUTES.toLong()))
        val until = from.plus(LOOKAHEAD)
        while (slotStart.isBefore(until)) {
            val busy = isBusy(slotStart)
            if (busy && !previousBusy && slotStart.isAfter(from)) {
                return slotStart
            }
            previousBusy = busy
            slotStart = slotStart.plusMinutes(SLOT_MINUTES.toLong())
        }
        return null
    }

    /** Whether the slot starting at [time] is a usual time. */
    private fun isBusy(time: ZonedDateTime): Boolean {
        return scores[slotOf(time)] >= MIN_SLOT_SCORE
    }

    companion object {

        /** The week slot (0 = Monday 00:00) a time falls in. */
        fun slotOf(time: ZonedDateTime): Int {
            val day = time.dayOfWeek.value - DayOfWeek.MONDAY.value
            return day * SLOTS_PER_DAY + (time.hour * 60 + time.minute) / SLOT_MINUTES
        }

        /** Weighted events per week slot. */
        private fun counted(events: List<HabitEvent>): DoubleArray {
            val counts = DoubleArray(SLOTS_PER_WEEK)
            for (event in events) {
                var weight = OPEN_WEIGHT
                if (event.finish) {
                    weight = FINISH_WEIGHT
                }
                counts[slotOf(event.at)] += weight
            }
            return counts
        }

        /** Each slot with a share of its neighbours and of the same time on the other days. */
        private fun smoothed(counts: DoubleArray): DoubleArray {
            val daily = DoubleArray(SLOTS_PER_DAY)
            for (slot in counts.indices) {
                daily[slot % SLOTS_PER_DAY] += counts[slot]
            }
            val out = DoubleArray(SLOTS_PER_WEEK)
            for (slot in counts.indices) {
                val before = counts[(slot - 1 + SLOTS_PER_WEEK) % SLOTS_PER_WEEK]
                val after = counts[(slot + 1) % SLOTS_PER_WEEK]
                val otherDays = daily[slot % SLOTS_PER_DAY] - counts[slot]
                out[slot] = counts[slot] + NEIGHBOUR_SHARE * (before + after) + OTHER_DAYS_SHARE * otherDays
            }
            return out
        }
    }
}
