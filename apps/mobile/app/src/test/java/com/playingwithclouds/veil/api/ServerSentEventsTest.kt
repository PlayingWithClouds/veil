package com.playingwithclouds.veil.api

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ServerSentEventsTest {

    /** Feeds lines to a parser and returns the events it completes. */
    private fun parse(vararg lines: String): List<ServerSentEvent> {
        val parser = ServerSentEventParser()
        return lines.mapNotNull { line -> parser.feed(line) }
    }

    @Test
    fun eventIsCompletedByABlankLine() {
        val events = parse("event: result", "data: {\"a\":1}", "")
        assertEquals(listOf(ServerSentEvent("result", "{\"a\":1}")), events)
    }

    @Test
    fun commentsAreIgnored() {
        assertTrue(parse(": connected", "").isEmpty())
    }

    @Test
    fun consecutiveEventsDoNotLeakIntoEachOther() {
        val events = parse("event: result", "data: one", "", "event: done", "data: {}", "")
        assertEquals(listOf(ServerSentEvent("result", "one"), ServerSentEvent("done", "{}")), events)
    }

    @Test
    fun multipleDataLinesAreJoined() {
        val events = parse("event: result", "data: a", "data: b", "")
        assertEquals("a\nb", events.single().data)
    }

    @Test
    fun parsesAStoredLibraryMatch() {
        val hit = LiveSearchEvents.parseResult(
            """{"source":"db","id":"scene:abc","item":{"title":"T","media_type":"scene","source_url":"https://x/1","external_id":"1","date":"2024-01-02","poster_path":"/api/blob/p.jpg","preview_images":["a","b"]}}""",
        )!!
        assertEquals("db", hit.source)
        assertEquals("scene:abc", hit.recordId)
        assertEquals("T", hit.title)
        assertEquals("scene", hit.mediaType)
        assertEquals("2024-01-02", hit.date)
        assertEquals("/api/blob/p.jpg", hit.posterPath)
        assertEquals(listOf("a", "b"), hit.previewImages)
        assertNull(hit.previewVideo)
    }

    @Test
    fun parsesAPluginResultWithoutRecordId() {
        val hit = LiveSearchEvents.parseResult(
            """{"source":"eporner","item":{"title":"T","media_type":"gallery","source_url":"u","external_id":"e"}}""",
        )!!
        assertNull(hit.recordId)
        assertEquals("eporner", hit.source)
        assertTrue(hit.previewImages.isEmpty())
    }

    @Test
    fun malformedPayloadsAreSkipped() {
        assertNull(LiveSearchEvents.parseResult("not json"))
        assertNull(LiveSearchEvents.parseResult("""{"source":"x"}"""))
    }
}
