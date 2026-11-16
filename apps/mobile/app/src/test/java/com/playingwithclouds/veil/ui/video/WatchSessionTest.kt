package com.playingwithclouds.veil.ui.video

import org.junit.Assert.assertEquals
import org.junit.Test

class WatchSessionTest {

    /** Ticks once a second from [fromSeconds] up to [toSeconds], collecting what the session reports. */
    private fun WatchSession.play(fromSeconds: Int, toSeconds: Int): List<WatchEvent> {
        playing(fromSeconds.toDouble())
        return (fromSeconds + 1..toSeconds).flatMap { second -> tick(second.toDouble()) }
    }

    @Test
    fun pausingEndsTheStretchAndLogsTheResumePoint() {
        val session = WatchSession()
        session.play(10, 14)
        assertEquals(listOf(WatchEvent.Span(10.0, 14.0), WatchEvent.Progress(14.0)), session.paused(14.0))
    }

    @Test
    fun progressIsLoggedEveryFifteenSecondsOfPlaying() {
        val session = WatchSession()
        val events = session.play(0, 31)
        assertEquals(listOf(WatchEvent.Progress(15.0), WatchEvent.Progress(30.0)), events)
    }

    @Test
    fun aSeekEndsTheStretchAndDoesNotCountAsPlaying() {
        val session = WatchSession()
        session.play(0, 5)
        assertEquals(listOf(WatchEvent.Span(0.0, 5.0)), session.seeked(5.0, 300.0, isPlaying = true))
        assertEquals(emptyList<WatchEvent>(), session.tick(301.0))
        assertEquals(listOf(WatchEvent.Span(300.0, 301.0), WatchEvent.Progress(301.0)), session.paused(301.0))
    }

    @Test
    fun stretchesShorterThanASecondAreDropped() {
        val session = WatchSession()
        session.playing(40.0)
        assertEquals(listOf(WatchEvent.Progress(40.5)), session.paused(40.5))
    }

    @Test
    fun theEndClosesTheStretchAndFinishes() {
        val session = WatchSession()
        session.play(100, 110)
        assertEquals(listOf(WatchEvent.Span(100.0, 120.0), WatchEvent.Finish(120.0)), session.ended(120.0))
    }

    @Test
    fun closingLogsProgressOnlyWhenSomethingPlayedSinceTheLast() {
        val session = WatchSession()
        session.play(0, 3)
        assertEquals(listOf(WatchEvent.Span(0.0, 3.0), WatchEvent.Progress(3.0)), session.closed())
        assertEquals(emptyList<WatchEvent>(), session.closed())
    }
}
