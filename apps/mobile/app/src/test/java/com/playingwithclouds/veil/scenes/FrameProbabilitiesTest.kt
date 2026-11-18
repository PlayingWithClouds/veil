package com.playingwithclouds.veil.scenes

import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Test

class FrameProbabilitiesTest {

    @Test
    fun windowsArePaddedWithTheFirstFrame() {
        val embeddings = List(3) { frame -> floatArrayOf(frame.toFloat()) }
        val windows = FrameProbabilities.windows(embeddings, 2)
        assertEquals(listOf(0f, 0f), windows[0].map { embedding -> embedding[0] })
        assertEquals(listOf(1f, 2f), windows[2].map { embedding -> embedding[0] })
    }

    @Test
    fun imageAndVideoReadingsAreAveraged() {
        val manifest = SceneModelManifest.parse(testManifestJson)
        val combined = FrameProbabilities.combine(
            mapOf("position" to floatArrayOf(0.2f, 0f, 0f), "rating" to floatArrayOf(1f, 0f)),
            mapOf("position" to floatArrayOf(0.8f, 0f, 0f)),
            manifest,
        )
        assertArrayEquals(floatArrayOf(0.5f, 0f, 0f), combined.getValue("position"), 1e-6f)
        assertArrayEquals(floatArrayOf(1f, 0f), combined.getValue("rating"), 1e-6f)
    }

    @Test
    fun confidentTagsKeepTheSureOnes() {
        val manifest = SceneModelManifest.parse(testManifestJson)
        val tags = FrameProbabilities.confidentTags(mapOf("position" to floatArrayOf(0.91f, 0.9f, 0.3f)), manifest, 0.5f)
        assertEquals(mapOf("Cowgirl" to 0.91), tags)
    }

    @Test
    fun scanStepsStayCheap() {
        assertEquals(6.0, ScanPlan.stepSeconds(300.0), 0.0)
        assertEquals(6.0, ScanPlan.stepSeconds(900.0), 0.0)
        assertEquals(12.0, ScanPlan.stepSeconds(1800.0), 0.0)
        assertEquals(48.0, ScanPlan.stepSeconds(7200.0), 0.0)
        assertEquals(150, ScanPlan.times(1800.0, ScanPlan.stepSeconds(1800.0)).size)
        assertEquals(50, ScanPlan.times(300.0, ScanPlan.stepSeconds(300.0)).size)
        assertEquals(listOf(0.0, 3.0, 6.0), ScanPlan.times(10.0, 3.0))
    }
}
