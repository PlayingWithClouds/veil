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

    /** Cover tags known for record "1" only. */
    private class FakeCoverTags : CoverTagSource {
        val tags = mapOf("Cowgirl" to 0.8)

        override fun active(): Boolean = true

        override fun cached(recordId: String): Map<String, Double>? = tags.takeIf { recordId == "1" }

        override suspend fun await(recordId: String): Map<String, Double>? = cached(recordId)
    }

    @Test
    fun coverTagsRideAlongWhenRead() = runTest {
        val logger = ImpressionLogger(EventLog(this, { batch -> batches.add(batch) }, flushDelayMilliseconds = 3000), FakeCoverTags(), this)
        logger.recordShown("1", video, "home", 0)
        logger.recordShown("2", video, "home", 1)
        logger.recordClicked("1", video, "home", 0)
        advanceUntilIdle()
        val events = batches.flatten()
        val byKind = events.groupBy { event -> event.kind }
        assertEquals(mapOf("Cowgirl" to 0.8), byKind.getValue(EventKind.choice).single().data["coverTags"])
        val impressions = byKind.getValue(EventKind.impression).associateBy { event -> event.subjectId }
        assertEquals(mapOf("Cowgirl" to 0.8), impressions.getValue("1").data["coverTags"])
        assertEquals(null, impressions.getValue("2").data["coverTags"])
    }
}
