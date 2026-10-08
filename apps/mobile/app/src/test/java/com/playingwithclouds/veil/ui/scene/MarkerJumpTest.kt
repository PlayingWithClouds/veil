package com.playingwithclouds.veil.ui.scene

import com.playingwithclouds.veil.data.SceneMarker
import org.junit.Assert.assertEquals
import org.junit.Test

class MarkerJumpTest {

    private fun marker(id: String, seconds: Double, tag: String? = null, label: String? = null): SceneMarker {
        return SceneMarker(id, tag, label, seconds, null, false)
    }

    @Test
    fun chipsFollowPlayOrder() {
        val ordered = markersInPlayOrder(listOf(marker("b", 90.0), marker("a", 10.0), marker("c", 30.0)))
        assertEquals(listOf("a", "c", "b"), ordered.map { marker -> marker.id })
    }

    @Test
    fun activeMarkerIsTheLastOneStarted() {
        val ordered = listOf(marker("a", 10.0), marker("b", 30.0), marker("c", 90.0))
        assertEquals(-1, activeMarkerIndex(ordered, 5.0))
        assertEquals(0, activeMarkerIndex(ordered, 10.0))
        assertEquals(0, activeMarkerIndex(ordered, 28.9))
        assertEquals(1, activeMarkerIndex(ordered, 29.5))
        assertEquals(1, activeMarkerIndex(ordered, 87.0))
        assertEquals(2, activeMarkerIndex(ordered, 500.0))
        assertEquals(-1, activeMarkerIndex(emptyList(), 5.0))
    }

    @Test
    fun chipLabelsHaveTextAndTime() {
        assertEquals("Opening · 1:05", markerChipLabel(marker("a", 65.0, tag = "Opening")))
        assertEquals("Intro - Part 1 · 0:00", markerChipLabel(marker("a", 0.0, tag = "Intro", label = "Part 1")))
        assertEquals("2:00", markerChipLabel(marker("a", 120.0)))
    }

    @Test
    fun tickFractionsSkipMarkersBeyondTheEnd() {
        val markers = listOf(marker("a", 25.0), marker("b", 50.0), marker("c", 500.0))
        assertEquals(listOf(0.25f, 0.5f), markerFractions(markers, 100.0))
        assertEquals(emptyList<Float>(), markerFractions(markers, 0.0))
    }
}
