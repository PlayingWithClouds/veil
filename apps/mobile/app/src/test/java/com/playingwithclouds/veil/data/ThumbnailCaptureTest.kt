package com.playingwithclouds.veil.data

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ThumbnailCaptureTest {

    @Test
    fun needsOneWhenThereIsABestMomentAndNoThumbnail() {
        assertTrue(needsNewThumbnail(bestMomentSeconds = 120.0, thumbnailSeconds = null))
    }

    @Test
    fun needsNoneWithoutABestMoment() {
        assertFalse(needsNewThumbnail(bestMomentSeconds = null, thumbnailSeconds = null))
        assertFalse(needsNewThumbnail(bestMomentSeconds = null, thumbnailSeconds = 50.0))
    }

    @Test
    fun redoesItOnlyWhenTheMomentMovedFar() {
        assertFalse(needsNewThumbnail(bestMomentSeconds = 130.0, thumbnailSeconds = 120.0))
        assertTrue(needsNewThumbnail(bestMomentSeconds = 300.0, thumbnailSeconds = 120.0))
    }
}
