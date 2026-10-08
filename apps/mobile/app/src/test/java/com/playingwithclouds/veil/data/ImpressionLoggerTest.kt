package com.playingwithclouds.veil.data

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

    private fun TestScope.logger(batchLimit: Int = 500): ImpressionLogger {
        return ImpressionLogger(this, { batch -> batches.add(batch) }, flushDelayMilliseconds = 3000, batchLimit = batchLimit)
    }

    @Test
    fun shownImpressionsAreSentAfterAQuietPeriod() = runTest {
        val logger = logger()
        logger.recordShown("scene:1", "newest", "feed", 0)
        logger.recordShown("scene:2", "newest", "feed", 1)
        advanceTimeBy(2999)
        runCurrent()
        assertTrue(batches.isEmpty())
        advanceTimeBy(2)
        runCurrent()
        assertEquals(1, batches.size)
        assertEquals(listOf("scene:1", "scene:2"), batches[0].map { impression -> impression.mediaId })
        assertTrue(batches[0].none { impression -> impression.clicked })
    }

    @Test
    fun aSceneIsShownOnlyOncePerSurface() = runTest {
        val logger = logger()
        logger.recordShown("scene:1", null, "feed", 0)
        logger.recordShown("scene:1", null, "feed", 0)
        logger.recordShown("scene:1", null, "row", 0)
        advanceUntilIdle()
        assertEquals(2, batches.single().size)
    }

    @Test
    fun resettingForgetsWhatWasShown() = runTest {
        val logger = logger()
        logger.recordShown("scene:1", null, "feed", 0)
        advanceUntilIdle()
        logger.resetShown()
        logger.recordShown("scene:1", null, "feed", 0)
        advanceUntilIdle()
        assertEquals(2, batches.size)
    }

    @Test
    fun aClickIsSentRightAway() = runTest {
        val logger = logger()
        logger.recordShown("scene:1", "affinity:tag", "feed", 4)
        logger.recordClicked("scene:1", "affinity:tag", "feed", 4)
        runCurrent()
        assertEquals(1, batches.size)
        assertEquals(listOf(false, true), batches[0].map { impression -> impression.clicked })
    }

    @Test
    fun aFullBatchIsSentWithoutWaiting() = runTest {
        val logger = logger(batchLimit = 2)
        logger.recordShown("scene:1", null, "feed", 0)
        logger.recordShown("scene:2", null, "feed", 1)
        runCurrent()
        assertEquals(1, batches.size)
    }

    @Test
    fun impressionsWithoutMediaIdAreDropped() = runTest {
        val logger = logger()
        logger.recordShown("", null, "feed", 0)
        advanceUntilIdle()
        assertTrue(batches.isEmpty())
    }

    @Test
    fun failedSendsDoNotBreakTheLogger() = runTest {
        var attempts = 0
        val failing = ImpressionLogger(this, { attempts++; error("offline") }, flushDelayMilliseconds = 1000)
        failing.recordClicked("scene:1", null, "feed", 0)
        runCurrent()
        failing.recordClicked("scene:2", null, "feed", 1)
        runCurrent()
        assertEquals(2, attempts)
    }
}
