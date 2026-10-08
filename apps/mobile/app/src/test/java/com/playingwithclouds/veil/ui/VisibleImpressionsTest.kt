package com.playingwithclouds.veil.ui

import com.playingwithclouds.veil.ui.home.ItemExtent
import com.playingwithclouds.veil.ui.home.VisibleImpressions
import org.junit.Assert.assertEquals
import org.junit.Test

class VisibleImpressionsTest {

    @Test
    fun itemsHalfOnScreenCount() {
        val items = listOf(
            ItemExtent(index = 0, offset = -60, size = 100),
            ItemExtent(index = 1, offset = 40, size = 100),
            ItemExtent(index = 2, offset = 140, size = 100),
        )
        assertEquals(listOf(1), VisibleImpressions.shownIndices(items, viewportStart = 0, viewportEnd = 180))
    }

    @Test
    fun anItemExactlyHalfVisibleCounts() {
        val items = listOf(ItemExtent(index = 3, offset = 150, size = 100))
        assertEquals(listOf(3), VisibleImpressions.shownIndices(items, viewportStart = 0, viewportEnd = 200))
    }

    @Test
    fun emptyAndOffScreenItemsNeverCount() {
        val items = listOf(ItemExtent(0, 10, 0), ItemExtent(1, 500, 100))
        assertEquals(emptyList<Int>(), VisibleImpressions.shownIndices(items, 0, 200))
    }
}
