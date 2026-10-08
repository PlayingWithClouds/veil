package com.playingwithclouds.veil.ui.scene

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PlayerMathTest {

    @Test
    fun doubleTapSeeksBackOnTheLeftHalfAndForwardOnTheRight() {
        assertEquals(-1, seekDirectionAt(10f, 400f))
        assertEquals(-1, seekDirectionAt(199f, 400f))
        assertEquals(1, seekDirectionAt(200f, 400f))
        assertEquals(1, seekDirectionAt(390f, 400f))
    }

    @Test
    fun seekTargetStaysInsideTheVideo() {
        assertEquals(15_000L, clampedSeekTarget(5_000, 10_000, 60_000))
        assertEquals(0L, clampedSeekTarget(4_000, -10_000, 60_000))
        assertEquals(60_000L, clampedSeekTarget(55_000, 10_000, 60_000))
    }

    @Test
    fun seekTargetWithUnknownDurationOnlyBoundsTheStart() {
        assertEquals(70_000L, clampedSeekTarget(60_000, 10_000, Long.MIN_VALUE + 1))
        assertEquals(0L, clampedSeekTarget(3_000, -10_000, 0))
    }

    @Test
    fun barFractionIsClampedToTheBar() {
        assertEquals(0.25f, fractionAt(100f, 400f), 0.0001f)
        assertEquals(0f, fractionAt(-20f, 400f), 0.0001f)
        assertEquals(1f, fractionAt(500f, 400f), 0.0001f)
        assertEquals(0f, fractionAt(50f, 0f), 0.0001f)
    }

    @Test
    fun positionAndFractionConvertBothWays() {
        assertEquals(0.5f, fractionOf(30_000, 60_000), 0.0001f)
        assertEquals(0f, fractionOf(30_000, -1), 0.0001f)
        assertEquals(1f, fractionOf(90_000, 60_000), 0.0001f)
        assertEquals(15_000L, positionAt(0.25f, 60_000))
        assertEquals(0L, positionAt(0.25f, 0))
        assertEquals(60_000L, positionAt(1.5f, 60_000))
    }

    @Test
    fun playbackTimeShowsElapsedAndTotal() {
        assertEquals("5:19 / 6:40", formatPlaybackTime(319_400, 400_000))
        assertEquals("1:02:03 / 2:00:00", formatPlaybackTime(3_723_000, 7_200_000))
        assertEquals("0:42", formatPlaybackTime(42_000, -1))
    }

    @Test
    fun speedsAreNamedCompactly() {
        assertEquals("Normal", formatSpeed(1f))
        assertEquals("1.25×", formatSpeed(1.25f))
        assertEquals("0.5×", formatSpeed(0.5f))
        assertEquals("2×", formatSpeed(2f))
    }

    @Test
    fun repeatedDoubleTapsAddUpAndATurnStartsOver() {
        val first = accumulateSeek(null, 1)
        assertEquals(DOUBLE_TAP_SEEK_SECONDS, first.seconds)
        val second = accumulateSeek(first, 1)
        assertEquals(DOUBLE_TAP_SEEK_SECONDS * 2, second.seconds)
        assertTrue(second.id != first.id)
        val turned = accumulateSeek(second, -1)
        assertEquals(-1, turned.direction)
        assertEquals(DOUBLE_TAP_SEEK_SECONDS, turned.seconds)
        assertEquals("−10 s", seekFeedbackLabel(turned))
        assertEquals("+20 s", seekFeedbackLabel(second))
    }

    @Test
    fun controlsHideOnlyWhilePlayingUntouched() {
        assertTrue(shouldAutoHide(visible = true, playing = true, scrubbing = false, menuOpen = false))
        assertFalse(shouldAutoHide(visible = true, playing = false, scrubbing = false, menuOpen = false))
        assertFalse(shouldAutoHide(visible = true, playing = true, scrubbing = true, menuOpen = false))
        assertFalse(shouldAutoHide(visible = true, playing = true, scrubbing = false, menuOpen = true))
        assertFalse(shouldAutoHide(visible = false, playing = true, scrubbing = false, menuOpen = false))
    }

    @Test
    fun relatedPanelOpensOnASwipeUpFromTheLowerHalf() {
        assertTrue(isRelatedPanelSwipe(startY = 300f, height = 400f, dragY = -150f, thresholdPixels = 100f))
        assertFalse(isRelatedPanelSwipe(startY = 100f, height = 400f, dragY = -150f, thresholdPixels = 100f))
        assertFalse(isRelatedPanelSwipe(startY = 300f, height = 400f, dragY = -50f, thresholdPixels = 100f))
        assertFalse(isRelatedPanelSwipe(startY = 300f, height = 400f, dragY = 150f, thresholdPixels = 100f))
    }

    @Test
    fun panelClosesOnALongEnoughDragDown() {
        assertTrue(isPanelDismissSwipe(120f, 100f))
        assertFalse(isPanelDismissSwipe(40f, 100f))
    }

    @Test
    fun qualitiesAreHighestFirstAndOnePerHeight() {
        val qualities = listOf(
            VideoQuality(480, 0, 0),
            VideoQuality(1080, 0, 1),
            VideoQuality(720, 0, 2),
            VideoQuality(1080, 1, 0),
            VideoQuality(0, 0, 3),
        )
        val distinct = distinctQualities(qualities)
        assertEquals(listOf(1080, 720, 480), distinct.map { quality -> quality.height })
        assertEquals(VideoQuality(1080, 0, 1), distinct.first())
        assertEquals("720p", videoQualityLabel(720))
    }
}
