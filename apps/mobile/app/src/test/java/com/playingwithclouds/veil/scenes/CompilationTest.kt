package com.playingwithclouds.veil.scenes

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import kotlin.random.Random

class CompilationTest {

    @Test
    fun clipsJoinNearLikedStretchesAndDropTheRest() {
        val source = CompilationSource(
            "v1",
            listOf(
                TimedSegment("Blowjob", 10.0, 40.0),
                TimedSegment("Deepthroat", 44.0, 60.0),
                TimedSegment("Kissing", 70.0, 200.0),
                TimedSegment("blowjob", 300.0, 304.0),
                TimedSegment("Blowjob", 400.0, 800.0),
            ),
            durationSeconds = 500.0,
        )
        val clips = CompilationPlanner.clips(source, setOf("Blowjob", "Deepthroat"))
        assertEquals(2, clips.size)
        assertEquals(CompilationClip("v1", 9.0, 60.0, listOf("Blowjob", "Deepthroat")), clips[0])
        // Cut to the runtime.
        assertEquals(399.0, clips[1].startSeconds, 0.0)
        assertEquals(500.0, clips[1].endSeconds, 0.0)
    }

    @Test
    fun longClipsAreCut() {
        val source = CompilationSource("v1", listOf(TimedSegment("Anal", 0.0, 1000.0)), null)
        val clip = CompilationPlanner.clips(source, setOf("anal")).single()
        assertEquals(CompilationPlanner.MAX_CLIP_SECONDS, clip.endSeconds - clip.startSeconds, 0.0)
    }

    @Test
    fun planKeepsTheVideosOrder() {
        val first = CompilationSource("a", listOf(TimedSegment("Cowgirl", 100.0, 130.0)), null)
        val second = CompilationSource("b", listOf(TimedSegment("Cowgirl", 10.0, 40.0)), null)
        assertEquals(listOf("a", "b"), CompilationPlanner.plan(listOf(first, second), setOf("Cowgirl")).map { clip -> clip.videoId })
    }

    @Test
    fun optionsPickTheScenesAndTheShortestClip() {
        val source = CompilationSource(
            "v1",
            listOf(TimedSegment("Cowgirl", 0.0, 12.0), TimedSegment("Anal", 100.0, 140.0), TimedSegment("Kissing", 200.0, 260.0)),
            null,
        )
        val all = CompilationPlanner.plan(listOf(source), setOf("Cowgirl", "Anal"), CompilationOrder.QUEUE, CompilationPlanner.MIN_CLIP_SECONDS, Random(1))
        assertEquals(listOf(listOf("Cowgirl"), listOf("Anal")), all.map { clip -> clip.tags })
        val long = CompilationPlanner.plan(listOf(source), setOf("Cowgirl", "Anal"), CompilationOrder.QUEUE, 30.0, Random(1))
        assertEquals(listOf(listOf("Anal")), long.map { clip -> clip.tags })
    }

    @Test
    fun shuffleMixesTheClipsOfEveryVideo() {
        val sources = (0 until 6).map { index ->
            CompilationSource("v$index", listOf(TimedSegment("Cowgirl", 0.0, 20.0), TimedSegment("Cowgirl", 100.0, 120.0)), null)
        }
        val inOrder = CompilationPlanner.plan(sources, setOf("Cowgirl"), CompilationOrder.QUEUE, CompilationPlanner.MIN_CLIP_SECONDS, Random(3))
        val shuffled = CompilationPlanner.plan(sources, setOf("Cowgirl"), CompilationOrder.SHUFFLE, CompilationPlanner.MIN_CLIP_SECONDS, Random(3))
        assertEquals(inOrder.toSet(), shuffled.toSet())
        assertEquals(12, shuffled.size)
        assertTrue(inOrder != shuffled)
        assertEquals(listOf("v0", "v0", "v1"), inOrder.take(3).map { clip -> clip.videoId })
    }

    @Test
    fun previewMomentsPreferLikedAndSpreadOut() {
        val segments = listOf(
            TimedSegment("Cowgirl", 100.0, 200.0),
            TimedSegment("Cowgirl", 205.0, 220.0),
            TimedSegment("Kissing", 0.0, 30.0),
            TimedSegment("Doggystyle", 400.0, 460.0),
            TimedSegment("Facial", 600.0, 620.0),
        )
        val moments = PreviewMoments.pick(segments, setOf("cowgirl"))
        assertTrue(moments.size in PreviewMoments.MIN_MOMENTS..PreviewMoments.MAX_MOMENTS)
        assertEquals(moments.sorted(), moments)
        assertTrue(moments.contains(105.0))
        assertTrue(moments.zipWithNext().all { (left, right) -> right - left >= 15.0 })
    }

    @Test
    fun tooFewMomentsAreNoPreview() {
        assertTrue(PreviewMoments.pick(listOf(TimedSegment("Cowgirl", 0.0, 30.0)), setOf("Cowgirl")).isEmpty())
    }
}
