package com.playingwithclouds.veil.ui.cams

import org.junit.Assert.assertEquals
import org.junit.Test

class CamRulesTest {

    @Test
    fun recordStartsNeedTheBufferToHoldThem() {
        assertEquals(emptyList<RecordStart>(), recordStarts(bufferedSeconds = 20.0, lagSeconds = 4.0))
        assertEquals(listOf(RecordStart("From 1 min ago", 60)), recordStarts(bufferedSeconds = 55.0, lagSeconds = 4.0))
        val all = recordStarts(bufferedSeconds = 175.0, lagSeconds = 42.4)
        assertEquals(listOf("From here", "From 1 min ago", "From 3 min ago"), all.map { start -> start.label })
        assertEquals(42, all.first().secondsAgo)
    }

    @Test
    fun recordQualityFollowsThePinnedHeight() {
        assertEquals("source", recordQualityFor(null))
        assertEquals("source", recordQualityFor(0))
        assertEquals("720", recordQualityFor(720))
        assertEquals("Source", recordQualityLabel("source"))
        assertEquals("480p", recordQualityLabel("480"))
    }

    @Test
    fun viewerCountsShortenFromAThousand() {
        assertEquals("950", formatViewers(950))
        assertEquals("1.2k", formatViewers(1240))
        assertEquals("15k", formatViewers(15_400))
    }
}
