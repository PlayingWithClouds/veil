package com.playingwithclouds.veil.ui.scene

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class PlayerLoopTest {

    @Test
    fun marksGoAThenBThenClear() {
        val withA = advanceLoop(AbLoop(), 10_000)
        assertEquals(AbLoop(10_000, null), withA)
        assertFalse(withA.isActive)
        val looping = advanceLoop(withA, 25_000)
        assertEquals(AbLoop(10_000, 25_000), looping)
        assertTrue(looping.isActive)
        assertEquals(AbLoop(), advanceLoop(looping, 30_000))
    }

    @Test
    fun aMarkAtOrBeforeAMovesA() {
        assertEquals(AbLoop(5_000, null), advanceLoop(AbLoop(10_000, null), 5_000))
        assertEquals(AbLoop(10_000, null), advanceLoop(AbLoop(10_000, null), 10_000))
    }

    @Test
    fun restartsAtAOnceBIsReached() {
        val loop = AbLoop(10_000, 25_000)
        assertNull(loopRestartTarget(loop, 24_999))
        assertEquals(10_000L, loopRestartTarget(loop, 25_000))
        assertEquals(10_000L, loopRestartTarget(loop, 26_000))
        assertNull(loopRestartTarget(AbLoop(10_000, null), 99_000))
    }

    @Test
    fun labelsFollowTheLoop() {
        assertEquals("Off", loopLabel(AbLoop()))
        assertEquals("A at 0:10", loopLabel(AbLoop(10_000, null)))
        assertEquals("0:10 – 0:25", loopLabel(AbLoop(10_000, 25_000)))
        assertEquals("Set loop start (A)", loopActionLabel(AbLoop()))
        assertEquals("Set loop end (B)", loopActionLabel(AbLoop(10_000, null)))
        assertEquals("Clear loop", loopActionLabel(AbLoop(10_000, 25_000)))
    }

    @Test
    fun frameDurationFollowsTheFrameRate() {
        assertEquals(42L, frameDurationMilliseconds(24f))
        assertEquals(17L, frameDurationMilliseconds(60f))
        assertEquals(33L, frameDurationMilliseconds(0f))
        assertEquals(33L, frameDurationMilliseconds(-1f))
    }

    @Test
    fun verticalVideosAreTallerThanWide() {
        assertTrue(isVerticalVideo(1080, 1920))
        assertFalse(isVerticalVideo(1920, 1080))
        assertFalse(isVerticalVideo(1080, 1080))
        assertFalse(isVerticalVideo(0, 0))
    }
}

class PictureInPictureAspectTest {

    @Test
    fun keepsTheVideosShapeWithinAndroidsLimits() {
        assertEquals(Pair(1920, 1080), pictureInPictureAspect(1920, 1080))
        assertEquals(Pair(1080, 1920), pictureInPictureAspect(1080, 1920))
        assertEquals(Pair(239, 100), pictureInPictureAspect(4000, 1000))
        assertEquals(Pair(100, 239), pictureInPictureAspect(1000, 4000))
        assertEquals(Pair(16, 9), pictureInPictureAspect(0, 0))
    }
}
