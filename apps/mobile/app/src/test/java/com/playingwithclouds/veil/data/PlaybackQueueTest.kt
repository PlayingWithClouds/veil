package com.playingwithclouds.veil.data

import kotlin.random.Random
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

class PlaybackQueueTest {

    private fun entry(id: String): QueueEntry {
        return QueueEntry(id, "Scene $id", null)
    }

    private fun ids(entries: List<QueueEntry>): List<String> {
        return entries.map { queued -> queued.sceneId }
    }

    @Before
    fun emptyQueue() {
        PlaybackQueue.clear()
        PlaybackQueue.setShuffle(false)
    }

    @Test
    fun playNextGoesFirstAndMovesAnExistingEntryUp() {
        val queue = listOf(entry("a"), entry("b"))
        assertEquals(listOf("c", "a", "b"), ids(queueWithPlayNext(queue, entry("c"))))
        assertEquals(listOf("b", "a"), ids(queueWithPlayNext(queue, entry("b"))))
    }

    @Test
    fun addAppendsAndKeepsPlaceOfAnExistingEntry() {
        val queue = listOf(entry("a"), entry("b"))
        assertEquals(listOf("a", "b", "c"), ids(queueWithAdded(queue, entry("c"))))
        assertEquals(listOf("a", "b"), ids(queueWithAdded(queue, entry("a"))))
    }

    @Test
    fun nextIndexIsFirstUnlessShufflingAndMinusOneWhenEmpty() {
        assertEquals(-1, nextQueueIndex(0, shuffle = true, random = Random(1)))
        assertEquals(0, nextQueueIndex(5, shuffle = false, random = Random(1)))
        val picked = nextQueueIndex(5, shuffle = true, random = Random(1))
        assertEquals(true, picked in 0 until 5)
    }

    @Test
    fun playlistOrderDropsDuplicatesAndShufflesOnRequest() {
        val entries = listOf(entry("a"), entry("b"), entry("a"), entry("c"))
        assertEquals(listOf("a", "b", "c"), ids(playlistOrder(entries, shuffled = false, random = Random(1))))
        val shuffled = playlistOrder(entries, shuffled = true, random = Random(7))
        assertEquals(listOf("a", "b", "c"), ids(shuffled).sorted())
    }

    @Test
    fun takeNextRemovesWhatItReturns() {
        PlaybackQueue.add(entry("a"))
        PlaybackQueue.add(entry("b"))
        assertEquals("a", PlaybackQueue.takeNext()?.sceneId)
        assertEquals(listOf("b"), ids(PlaybackQueue.state.value.entries))
        assertEquals("b", PlaybackQueue.takeNext()?.sceneId)
        assertNull(PlaybackQueue.takeNext())
    }

    @Test
    fun startPlaylistReturnsTheFirstAndQueuesTheRest() {
        val first = PlaybackQueue.startPlaylist(listOf(entry("a"), entry("b"), entry("c")), shuffled = false)
        assertEquals("a", first?.sceneId)
        assertEquals(listOf("b", "c"), ids(PlaybackQueue.state.value.entries))
        assertEquals(false, PlaybackQueue.state.value.shuffle)
        assertNull(PlaybackQueue.startPlaylist(emptyList(), shuffled = true))
    }

    @Test
    fun shufflingPlaylistKeepsShuffleOn() {
        PlaybackQueue.startPlaylist(listOf(entry("a"), entry("b"), entry("c")), shuffled = true, random = Random(3))
        assertEquals(true, PlaybackQueue.state.value.shuffle)
        assertEquals(2, PlaybackQueue.state.value.entries.size)
    }
}
