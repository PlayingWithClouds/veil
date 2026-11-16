package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ImpressionLoggerTest {

    private val batches = mutableListOf<List<LoggedEvent>>()

    private val video = RecordKind.video

    private fun TestScope.logger(): ImpressionLogger {
        return ImpressionLogger(EventLog(this, { batch -> batches.add(batch) }, flushDelayMilliseconds = 3000))
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
        assertEquals(listOf(ImpressionSurface.CONTINUE), batches[1].map { event -> event.data["surface"] })
    }

    @Test
    fun aChoiceIsSentRightAwayWithItsPosition() = runTest {
        val logger = logger()
        logger.recordShown("video:1", video, "home", 4)
        logger.recordClicked("video:1", video, "home", 4)
        runCurrent()
        val batch = batches.single()
        assertEquals(listOf(EventKind.impression, EventKind.choice), batch.map { event -> event.kind })
        assertEquals(listOf(4, 4), batch.map { event -> event.data["position"] })
    }
}
