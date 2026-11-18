package com.playingwithclouds.veil.scenes

import com.playingwithclouds.veil.data.VideoSegment
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class SceneSegmentsTest {

    /** One-column frames from probabilities. */
    private fun frames(vararg values: Float): List<FloatArray> {
        return values.map { value -> floatArrayOf(value) }
    }

    @Test
    fun multiLabelUsesHysteresisAndClosesShortGaps() {
        // On at 0.7, held down to 0.45, a 2-frame dip closed, then off.
        val runs = SceneSegments.multiLabel(frames(0.1f, 0.7f, 0.45f, 0.5f, 0.1f, 0.1f, 0.8f, 0.9f, 0.2f, 0.1f, 0.1f), listOf("cowgirl"))
        assertEquals(1, runs.size)
        assertEquals(1, runs[0].start)
        assertEquals(7, runs[0].end)
    }

    @Test
    fun multiLabelDropsShortRuns() {
        assertTrue(SceneSegments.multiLabel(frames(0.1f, 0.9f, 0.9f, 0.1f, 0.1f, 0.1f), listOf("anal")).isEmpty())
    }

    @Test
    fun exclusiveSmoothsAndSkipsUnsureFrames() {
        val probabilities = List(10) { frame -> if (frame < 6) floatArrayOf(0.9f, 0.1f) else floatArrayOf(0.3f, 0.3f) }
        val runs = SceneSegments.exclusive(probabilities, listOf("sfw", "nudity"))
        assertEquals(1, runs.size)
        assertEquals("sfw", runs[0].label)
        assertEquals(0, runs[0].start)
    }

    @Test
    fun timedSegmentsNameTagsAndSpanTheStep() {
        val manifest = SceneModelManifest.parse(testManifestJson)
        val times = List(8) { frame -> frame * 3.0 }
        val position = List(8) { frame -> if (frame in 2..6) floatArrayOf(0.9f, 0.9f, 0f) else floatArrayOf(0f, 0f, 0f) }
        val segments = SceneSegments.timedSegments(times, 3.0, mapOf("position" to position), manifest)
        // Kneeling never becomes a segment; cowgirl runs from frame 2 to the end of frame 6.
        assertEquals(listOf(TimedSegment("Cowgirl", 6.0, 21.0)), segments)
    }

    @Test
    fun sectionsFoldShortPiecesAndNameTwoLabels() {
        val sections = SceneSections.sections(
            listOf(
                TimedSegment("Blowjob", 0.0, 120.0),
                TimedSegment("POV", 30.0, 125.0),
                TimedSegment("Kissing", 121.0, 124.0),
                TimedSegment("Cowgirl", 130.0, 300.0),
                TimedSegment("Cowgirl", 310.0, 400.0),
            ),
        )
        assertEquals(listOf("Blowjob", "Blowjob · POV", "Cowgirl"), sections.map { section -> section.title })
        assertTrue(sections.all { section -> section.length >= SceneSections.MIN_SECTION_SECONDS })
        assertEquals(400.0, sections.last().endSeconds, 0.0)
    }

    @Test
    fun chaptersLeaveTheUsersSegmentsAlone() {
        val chapters = SceneSections.chapters(
            listOf(
                VideoSegment("user", "My scene", 0.0, 100.0, "user"),
                VideoSegment("a", "Blowjob", 10.0, 90.0, "ai", "Blowjob"),
                VideoSegment("b", "Doggystyle", 200.0, 300.0, "ai", "Doggystyle"),
            ),
        )
        // The unrecognised stretch before Doggystyle folds into it, which is cut where the user's segment ends.
        assertEquals(listOf("My scene", "Doggystyle"), chapters.map { chapter -> chapter.title })
        assertEquals(100.0, chapters.last().seconds, 0.0)
    }
}
