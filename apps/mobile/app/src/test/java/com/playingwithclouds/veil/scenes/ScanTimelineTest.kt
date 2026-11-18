package com.playingwithclouds.veil.scenes

import com.playingwithclouds.veil.data.VideoSegment
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ScanTimelineTest {

    private val manifest = SceneModelManifest.parse(testManifestJson)
    private val times = List(20) { index -> index * 2.0 }

    /** Frame [index]'s image heads: cowgirl on from frame 5 to 14. */
    private fun frame(index: Int): Map<String, FloatArray> {
        var cowgirl = 0.1f
        if (index in 5..14) {
            cowgirl = 0.9f
        }
        return mapOf("position" to floatArrayOf(cowgirl, 0.1f, 0.1f))
    }

    @Test
    fun scanStartsWhereThePlayerIsThenWraps() {
        assertEquals(listOf(3, 4, 0, 1, 2), ScanTimeline.order(5, 3))
        assertEquals(listOf(0, 1, 2), ScanTimeline.order(3, 0))
        assertEquals(5, ScanTimeline.indexAt(times, 11.0))
        assertEquals(0, ScanTimeline.indexAt(times, -1.0))
    }

    @Test
    fun windowsMatchAScanFromTheStart() {
        val timeline = ScanTimeline(times.take(5), 2.0, 3)
        val embeddings = List(5) { index -> floatArrayOf(index.toFloat()) }
        for (index in ScanTimeline.order(5, 3)) {
            timeline.add(index, embeddings[index], frame(index))
        }
        val expected = FrameProbabilities.windows(embeddings, 3)
        for (index in 0 until 5) {
            assertEquals(expected[index].map { values -> values.toList() }, timeline.windowOf(index).map { values -> values.toList() })
        }
    }

    @Test
    fun aWindowIsReadOnceAllItsFramesAreIn() {
        val timeline = ScanTimeline(times.take(6), 2.0, 3)
        for (index in listOf(3, 4, 5)) {
            timeline.add(index, floatArrayOf(index.toFloat()), frame(index))
        }
        assertEquals(listOf(5), timeline.pendingWindows())
        timeline.setVideo(5, frame(5))
        for (index in listOf(0, 1, 2)) {
            timeline.add(index, floatArrayOf(index.toFloat()), frame(index))
        }
        assertEquals(listOf(0, 1, 2, 3, 4), timeline.pendingWindows())
        assertTrue(ScanTimeline(times, 2.0, 0).pendingWindows().isEmpty())
    }

    @Test
    fun partialScansShowWhatTheirFramesHold() {
        val timeline = ScanTimeline(times, 2.0, 0)
        for (index in 10 until 20) {
            timeline.add(index, floatArrayOf(0f), frame(index))
        }
        assertEquals(listOf(10 until 20), timeline.scannedRuns())
        assertEquals(listOf(TimedSegment("Cowgirl", 20.0, 30.0)), timeline.segments(manifest))
        assertEquals(0.5f, timeline.progress)
    }

    @Test
    fun aCompleteScanEqualsOneFromTheStart() {
        val timeline = ScanTimeline(times, 2.0, 0)
        for (index in ScanTimeline.order(times.size, 10)) {
            timeline.add(index, floatArrayOf(0f), frame(index))
        }
        val fromStart = SceneSegments.timedSegments(times, 2.0, byHead(times.indices.map(::frame)), manifest)
        assertTrue(timeline.complete)
        assertEquals(fromStart, timeline.segments(manifest))
        assertEquals(listOf(TimedSegment("Cowgirl", 10.0, 30.0)), fromStart)
    }

    @Test
    fun theScanReplacesStoredAiSegmentsOnceItFindsSomething() {
        val stored = listOf(
            VideoSegment("user", "Intro", 0.0, 10.0, "user"),
            VideoSegment("old", "Blowjob", 20.0, 80.0, SceneSections.ORIGIN_AI, "Blowjob"),
        )
        assertEquals(stored, SceneSections.withScanned(stored, emptyList(), done = false))
        val found = listOf(TimedSegment("Cowgirl", 30.0, 90.0))
        val merged = SceneSections.withScanned(stored, found, done = false)
        assertEquals(listOf("Intro", "Cowgirl"), merged.map { segment -> segment.title })
        assertEquals(listOf("Intro"), SceneSections.withScanned(stored, emptyList(), done = true).map { segment -> segment.title })
    }

    @Test
    fun theOpenVideoIsScannedUnlessSomethingStandsAgainstIt() {
        assertEquals(OpenScanDecision.SCAN, OpenScanPolicy.decide(true, false, wifiOnly = true, metered = false, localCopy = false, powerSave = false))
        assertEquals(OpenScanDecision.SCAN, OpenScanPolicy.decide(true, false, wifiOnly = false, metered = true, localCopy = false, powerSave = false))
        assertEquals(OpenScanDecision.SCAN, OpenScanPolicy.decide(true, false, wifiOnly = true, metered = true, localCopy = true, powerSave = false))
        assertEquals(OpenScanDecision.WAITING_FOR_WIFI, OpenScanPolicy.decide(true, false, wifiOnly = true, metered = true, localCopy = false, powerSave = false))
        assertEquals(OpenScanDecision.BATTERY_SAVER, OpenScanPolicy.decide(true, false, wifiOnly = false, metered = false, localCopy = false, powerSave = true))
        assertEquals(OpenScanDecision.ALREADY_SCANNED, OpenScanPolicy.decide(true, true, wifiOnly = false, metered = false, localCopy = false, powerSave = false))
        assertEquals(OpenScanDecision.NO_MODEL, OpenScanPolicy.decide(false, false, wifiOnly = false, metered = false, localCopy = false, powerSave = false))
    }
}
