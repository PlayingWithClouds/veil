package com.playingwithclouds.veil.ui.scene

import com.playingwithclouds.veil.data.HeatSpan
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class HeatSessionTest {

    @Test
    fun playingThenPausingMakesOneSpan() {
        val session = HeatSession()
        session.playbackStarted(10.0)
        session.playbackStopped(40.0)
        val report = session.drain(600.0)!!
        assertEquals(listOf(HeatSpan(10.0, 40.0)), report.spans)
        assertEquals(emptyList<Double>(), report.scrubs)
        assertEquals(600.0, report.durationSeconds, 0.0)
    }

    @Test
    fun aJumpWhilePlayingSplitsTheSpanAndCountsAsScrub() {
        val session = HeatSession()
        session.playbackStarted(0.0)
        session.jumped(30.0, 200.0, playing = true, loopStartSeconds = null)
        session.playbackStopped(260.0)
        val report = session.drain(600.0)!!
        assertEquals(listOf(HeatSpan(0.0, 30.0), HeatSpan(200.0, 260.0)), report.spans)
        assertEquals(listOf(200.0), report.scrubs)
    }

    @Test
    fun smallNudgesAndLoopRestartsAreNotScrubs() {
        val session = HeatSession()
        session.playbackStarted(0.0)
        session.jumped(10.0, 12.0, playing = true, loopStartSeconds = null)
        session.jumped(50.0, 20.2, playing = true, loopStartSeconds = 20.0)
        session.playbackStopped(40.0)
        assertEquals(emptyList<Double>(), session.drain(600.0)!!.scrubs)
    }

    @Test
    fun blipsShorterThanASecondAreDropped() {
        val session = HeatSession()
        session.playbackStarted(5.0)
        session.playbackStopped(5.4)
        assertNull(session.drain(600.0))
    }

    @Test
    fun drainingClearsAndNeedsARuntime() {
        val session = HeatSession()
        session.playbackStarted(0.0)
        session.playbackStopped(20.0)
        assertNull(session.drain(0.0))
        assertEquals(1, session.drain(600.0)!!.spans.size)
        assertNull(session.drain(600.0))
    }
}
