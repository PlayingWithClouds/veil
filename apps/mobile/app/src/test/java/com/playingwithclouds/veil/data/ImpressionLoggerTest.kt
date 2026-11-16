package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.RecordKind
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ImpressionLoggerTest {

    private val batches = mutableListOf<List<Impression>>()

    private val video = RecordKind.video

    private fun TestScope.logger(batchLimit: Int = 200): ImpressionLogger {
        return ImpressionLogger(this, { batch -> batches.add(batch) }, flushDelayMilliseconds = 3000, batchLimit = batchLimit)
    }

    @Test
    fun shownImpressionsAreSentAfterAQuietPeriod() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, "home", 0)
        logger.recordShown("video:2", video, "home", 1)
        advanceTimeBy(2999)
        runCurrent()
        assertTrue(batches.isEmpty())
        advanceTimeBy(2)
        runCurrent()
        assertEquals(1, batches.size)
        assertEquals(listOf("video:1", "video:2"), batches[0].map { impression -> impression.subjectId })
        assertTrue(batches[0].none { impression -> impression.clicked })
    }

    @Test
    fun nothingIsQueuedWhileDisabled() = runTest {
        var enabled = false
        val logger = ImpressionLogger(this, { batch -> batches.add(batch) }, enabled = { enabled }, flushDelayMilliseconds = 3000)
        logger.recordShown("video:1", video, "home", 0)
        logger.recordClicked("video:1", video, "home", 0)
        advanceUntilIdle()
        assertTrue(batches.isEmpty())
        enabled = true
        logger.recordShown("video:2", video, "home", 1)
        advanceUntilIdle()
        assertEquals(1, batches.size)
    }

    @Test
    fun aRecordIsShownOnlyOncePerSurface() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, "home", 0)
        logger.recordShown("video:1", video, "home", 0)
        logger.recordShown("video:1", video, "continue", 0)
        advanceUntilIdle()
        assertEquals(2, batches.single().size)
    }

    @Test
    fun recordsOfDifferentKindsWithOneIdAreShownApart() = runTest {
        val logger = logger()
        logger.recordShown("1", video, "home", 0)
        logger.recordShown("1", RecordKind.story, "home", 1)
        advanceUntilIdle()
        assertEquals(2, batches.single().size)
    }

    @Test
    fun resettingForgetsWhatWasShown() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, "home", 0)
        advanceUntilIdle()
        logger.resetShown()
        logger.recordShown("video:1", video, "home", 0)
        advanceUntilIdle()
        assertEquals(2, batches.size)
    }

    @Test
    fun resettingASurfaceKeepsTheOthers() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, ImpressionSurface.HOME, 0)
        logger.recordShown("video:1", video, ImpressionSurface.CONTINUE, 0)
        advanceUntilIdle()
        logger.resetShown(ImpressionSurface.CONTINUE)
        logger.recordShown("video:1", video, ImpressionSurface.HOME, 0)
        logger.recordShown("video:1", video, ImpressionSurface.CONTINUE, 0)
        advanceUntilIdle()
        assertEquals(listOf(ImpressionSurface.CONTINUE), batches[1].map { impression -> impression.surface })
    }

    @Test
    fun everyImpressionCarriesItsPosition() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, ImpressionSurface.HOME, 7)
        logger.recordClicked("video:1", video, ImpressionSurface.HOME, 7)
        runCurrent()
        assertEquals(listOf(7, 7), batches.single().map { impression -> impression.position })
    }

    @Test
    fun aClickIsSentRightAway() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, "home", 4)
        logger.recordClicked("video:1", video, "home", 4)
        runCurrent()
        assertEquals(1, batches.size)
        assertEquals(listOf(false, true), batches[0].map { impression -> impression.clicked })
    }

    @Test
    fun aFullBatchIsSentWithoutWaiting() = runTest {
        val logger = logger(batchLimit = 2)
        logger.recordShown("video:1", video, "home", 0)
        logger.recordShown("video:2", video, "home", 1)
        runCurrent()
        assertEquals(1, batches.size)
    }

    @Test
    fun impressionsWithoutAnIdAreDropped() = runTest {
        val logger = logger()
        logger.recordShown("", video, "home", 0)
        advanceUntilIdle()
        assertTrue(batches.isEmpty())
    }

    @Test
    fun failedSendsDoNotBreakTheLogger() = runTest {
        var attempts = 0
        val failing = ImpressionLogger(this, { attempts++; error("offline") }, flushDelayMilliseconds = 1000)
        failing.recordClicked("video:1", video, "home", 0)
        runCurrent()
        failing.recordClicked("video:2", video, "home", 1)
        runCurrent()
        assertEquals(2, attempts)
    }
}
