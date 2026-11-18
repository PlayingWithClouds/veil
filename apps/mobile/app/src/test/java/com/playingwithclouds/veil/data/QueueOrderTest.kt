package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.ui.video.ReorderStep
import com.playingwithclouds.veil.ui.video.reorderStep
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import kotlin.random.Random

class QueueOrderTest {

    private val queue = listOf("a", "b", "c")

    @Test
    fun theQueueGoesOnFromTheVideoPlaying() {
        assertEquals("b", QueueOrder.next(queue, "a", RepeatMode.OFF))
        assertEquals("a", QueueOrder.next(queue, "elsewhere", RepeatMode.OFF))
        assertNull(QueueOrder.next(queue, "c", RepeatMode.OFF))
        assertNull(QueueOrder.next(emptyList(), "x", RepeatMode.ALL))
    }

    @Test
    fun repeatAllStartsOverAndRepeatOnePlaysTheSameAgain() {
        assertEquals("a", QueueOrder.next(queue, "c", RepeatMode.ALL))
        assertEquals("only", QueueOrder.next(listOf("only"), "only", RepeatMode.ALL))
        assertEquals("b", QueueOrder.next(queue, "b", RepeatMode.ONE))
    }

    @Test
    fun theNextButtonMovesOnEvenOnRepeatOne() {
        assertEquals("c", QueueOrder.skip(queue, "b", RepeatMode.ONE))
        assertEquals("a", QueueOrder.skip(queue, "c", RepeatMode.ONE))
        assertNull(QueueOrder.skip(queue, "c", RepeatMode.OFF))
        assertNull(QueueOrder.skip(listOf("only"), "only", RepeatMode.ALL))
    }

    @Test
    fun shufflingKeepsThePlayingVideoFirstAndEveryVideoOnce() {
        val shuffled = QueueOrder.shuffled(listOf("a", "b", "c", "d", "e"), "c", Random(7))
        assertEquals("c", shuffled.first())
        assertEquals(setOf("a", "b", "c", "d", "e"), shuffled.toSet())
        assertEquals(5, shuffled.size)
    }

    @Test
    fun aShuffledOrderFollowsTheQueuesChanges() {
        val shuffled = listOf("c", "a", "b")
        assertEquals(listOf("c", "b", "d"), QueueOrder.reconcile(shuffled, listOf("b", "c", "d")))
        assertEquals(listOf("c", "a", "b"), QueueOrder.playOrder(queue, shuffled))
        assertEquals(queue, QueueOrder.playOrder(queue, null))
    }

    @Test
    fun repeatCyclesOffAllOne() {
        assertEquals(RepeatMode.ALL, RepeatMode.OFF.next())
        assertEquals(RepeatMode.ONE, RepeatMode.ALL.next())
        assertEquals(RepeatMode.OFF, RepeatMode.ONE.next())
    }

    @Test
    fun movingReordersTheQueue() {
        assertEquals(listOf("b", "c", "a"), QueueOrder.moved(queue, 0, 2))
        assertEquals(listOf("c", "a", "b"), QueueOrder.moved(queue, 2, 0))
        assertEquals(queue, QueueOrder.moved(queue, 0, 5))
    }

    @Test
    fun aDraggedRowTakesTheNextSlotPastHalfARow() {
        assertEquals(ReorderStep(1, 0f), reorderStep(1, 0f, 100f, 3))
        assertEquals(ReorderStep(2, -40f), reorderStep(1, 60f, 100f, 3))
        assertEquals(ReorderStep(0, 40f), reorderStep(1, -60f, 100f, 3))
        assertEquals(ReorderStep(3, 80f), reorderStep(3, 80f, 100f, 3))
        assertEquals(ReorderStep(0, -80f), reorderStep(0, -80f, 100f, 3))
    }
}
