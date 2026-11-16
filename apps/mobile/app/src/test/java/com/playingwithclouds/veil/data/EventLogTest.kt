package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.EventKind
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
class EventLogTest {

    private val batches = mutableListOf<List<LoggedEvent>>()

    private fun TestScope.log(batchLimit: Int = 200): EventLog {
        return EventLog(this, { batch -> batches.add(batch) }, flushDelayMilliseconds = 3000, batchLimit = batchLimit)
    }

    private fun progress(id: String): LoggedEvent {
        return LoggedEvent(EventKind.progress, RecordKind.video, id, mapOf("position" to 12.0))
    }

    @Test
    fun eventsAreSentAfterAQuietPeriod() = runTest {
        val log = log()
        log.log(progress("video:1"))
        log.log(progress("video:2"))
        advanceTimeBy(2999)
        runCurrent()
        assertTrue(batches.isEmpty())
        advanceTimeBy(2)
        runCurrent()
        assertEquals(listOf("video:1", "video:2"), batches.single().map { event -> event.subjectId })
    }

    @Test
    fun aFlushSendsAtOnceWithWhatWasQueued() = runTest {
        val log = log()
        log.log(progress("video:1"))
        log.log(progress("video:2"), flush = true)
        runCurrent()
        assertEquals(2, batches.single().size)
    }

    @Test
    fun aFullBatchIsSentWithoutWaiting() = runTest {
        val log = log(batchLimit = 2)
        log.log(progress("video:1"))
        log.log(progress("video:2"))
        runCurrent()
        assertEquals(1, batches.size)
    }

    @Test
    fun nothingIsQueuedWhileDisabled() = runTest {
        var enabled = false
        val log = EventLog(this, { batch -> batches.add(batch) }, enabled = { enabled }, flushDelayMilliseconds = 3000)
        log.log(progress("video:1"), flush = true)
        advanceUntilIdle()
        assertTrue(batches.isEmpty())
        enabled = true
        log.log(progress("video:2"))
        advanceUntilIdle()
        assertEquals(1, batches.size)
    }

    @Test
    fun eventsWithoutAnIdAreDropped() = runTest {
        val log = log()
        log.log(progress(""))
        advanceUntilIdle()
        assertTrue(batches.isEmpty())
    }

    @Test
    fun searchesAreLoggedWithoutASubject() = runTest {
        val log = log()
        log.log(LoggedEvent(EventKind.search, null, null, mapOf("text" to "redhead")), flush = true)
        runCurrent()
        assertEquals(listOf(EventKind.search), batches.single().map { event -> event.kind })
    }

    @Test
    fun failedSendsDoNotBreakTheLog() = runTest {
        var attempts = 0
        val failing = EventLog(this, { attempts++; error("offline") }, flushDelayMilliseconds = 1000)
        failing.log(progress("video:1"), flush = true)
        runCurrent()
        failing.log(progress("video:2"), flush = true)
        runCurrent()
        assertEquals(2, attempts)
    }
}
