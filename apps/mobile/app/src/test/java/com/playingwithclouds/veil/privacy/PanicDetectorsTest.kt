package com.playingwithclouds.veil.privacy

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PanicDetectorsTest {

    @Test
    fun faceDownFiresOnceAfterTheHoldTime() {
        val detector = FaceDownDetector()
        assertFalse(detector.onSample(0f, 0f, -9.8f, 0))
        assertFalse(detector.onSample(0f, 0f, -9.8f, 300))
        assertTrue(detector.onSample(0f, 0f, -9.8f, 600))
        assertFalse(detector.onSample(0f, 0f, -9.8f, 900))
    }

    @Test
    fun faceDownRearmsAfterTheFlipBack() {
        val detector = FaceDownDetector()
        detector.onSample(0f, 0f, -9.8f, 0)
        assertTrue(detector.onSample(0f, 0f, -9.8f, 600))
        detector.onSample(0f, 0f, 9.8f, 700)
        detector.onSample(0f, 0f, -9.8f, 800)
        assertTrue(detector.onSample(0f, 0f, -9.8f, 1400))
    }

    @Test
    fun aTiltedOrFaceUpPhoneNeverFires() {
        val detector = FaceDownDetector()
        assertFalse(detector.onSample(0f, 0f, 9.8f, 0))
        assertFalse(detector.onSample(0f, 0f, 9.8f, 1000))
        assertFalse(detector.onSample(6f, 0f, -8.9f, 2000))
        assertFalse(detector.onSample(6f, 0f, -8.9f, 3000))
    }

    @Test
    fun movingBeforeTheHoldTimeResetsIt() {
        val detector = FaceDownDetector()
        detector.onSample(0f, 0f, -9.8f, 0)
        detector.onSample(0f, 0f, 2f, 300)
        assertFalse(detector.onSample(0f, 0f, -9.8f, 400))
        assertFalse(detector.onSample(0f, 0f, -9.8f, 800))
    }

    @Test
    fun twoQuickTapsMakeADoubleTap() {
        val detector = DoubleTapDetector()
        assertFalse(detector.onTap(1000))
        assertTrue(detector.onTap(1300))
        assertFalse(detector.onTap(1400))
    }

    @Test
    fun slowTapsDoNot() {
        val detector = DoubleTapDetector()
        detector.onTap(1000)
        assertFalse(detector.onTap(1500))
    }
}
