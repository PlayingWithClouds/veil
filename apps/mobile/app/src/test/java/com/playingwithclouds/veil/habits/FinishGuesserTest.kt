package com.playingwithclouds.veil.habits

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class FinishGuesserTest {

    private var clock = 0L
    private val guesser = FinishGuesser { clock }

    /** Plays [id] from [fromSeconds] for [seconds], reporting every 15 s of playing as the watch log does. */
    private fun play(id: String, fromSeconds: Double, seconds: Int) {
        guesser.progress(id, fromSeconds)
        var position = fromSeconds
        repeat(seconds / 15) {
            clock += 15_000
            position += 15
            guesser.progress(id, position)
        }
    }

    @Test
    fun leavingRightAfterALongWatchIsAFinish() {
        play("video", 0.0, 600)
        clock += 30_000
        assertEquals(GuessedFinish("video", 600.0), guesser.appLeft())
    }

    @Test
    fun aShortWatchIsNot() {
        play("video", 0.0, 60)
        assertNull(guesser.appLeft())
    }

    @Test
    fun leavingLongAfterIsNot() {
        play("video", 0.0, 600)
        clock += 20 * 60_000
        assertNull(guesser.appLeft())
    }

    @Test
    fun seekingDoesNotCountAsWatching() {
        guesser.progress("video", 0.0)
        clock += 15_000
        guesser.progress("video", 900.0)
        assertNull(guesser.appLeft())
    }

    @Test
    fun aVideoCountsOnceInAWhile() {
        play("video", 0.0, 600)
        guesser.appLeft()
        play("video", 600.0, 600)
        assertNull(guesser.appLeft())
    }
}
