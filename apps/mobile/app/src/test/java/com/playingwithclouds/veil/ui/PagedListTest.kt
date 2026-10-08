package com.playingwithclouds.veil.ui

import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class PagedListTest {

    @Test
    fun pagesAreAppendedInOrder() = runTest {
        val requestedOffsets = mutableListOf<Int>()
        val list = PagedList<Int>(this, 3, { number -> number }) { offset ->
            requestedOffsets.add(offset)
            (offset until offset + 3).toList()
        }
        list.loadMore()
        advanceUntilIdle()
        list.loadMore()
        advanceUntilIdle()
        assertEquals((0 until 6).toList(), list.state.value.items)
        assertEquals(listOf(0, 3), requestedOffsets)
    }

    @Test
    fun aShortPageEndsTheList() = runTest {
        var requests = 0
        val list = PagedList<Int>(this, 3, { number -> number }) { requests++; listOf(1, 2) }
        list.loadMore()
        advanceUntilIdle()
        assertTrue(list.state.value.endReached)
        list.loadMore()
        advanceUntilIdle()
        assertEquals(1, requests)
    }

    @Test
    fun overlappingPagesAreDeduplicated() = runTest {
        val pages = listOf(listOf(1, 2, 3), listOf(3, 4, 5))
        var next = 0
        val list = PagedList<Int>(this, 3, { number -> number }) { pages[next++] }
        list.loadMore()
        advanceUntilIdle()
        list.loadMore()
        advanceUntilIdle()
        assertEquals(listOf(1, 2, 3, 4, 5), list.state.value.items)
    }

    @Test
    fun refreshKeepsOldItemsUntilTheNewPageArrives() = runTest {
        val gate = CompletableDeferred<List<Int>>()
        var call = 0
        val list = PagedList<Int>(this, 3, { number -> number }) {
            call++
            if (call == 1) listOf(1, 2, 3) else gate.await()
        }
        list.loadMore()
        advanceUntilIdle()
        list.refresh()
        runCurrent()
        assertEquals(listOf(1, 2, 3), list.state.value.items)
        assertTrue(list.state.value.isRefreshing)
        gate.complete(listOf(7, 8, 9))
        advanceUntilIdle()
        assertEquals(listOf(7, 8, 9), list.state.value.items)
        assertFalse(list.state.value.isRefreshing)
    }

    @Test
    fun aResponseThatArrivesAfterARefreshIsIgnored() = runTest {
        val slow = CompletableDeferred<List<Int>>()
        var call = 0
        val list = PagedList<Int>(this, 3, { number -> number }) {
            call++
            if (call == 1) slow.await() else listOf(4, 5, 6)
        }
        list.loadMore()
        runCurrent()
        list.refresh()
        advanceUntilIdle()
        slow.complete(listOf(1, 2, 3))
        advanceUntilIdle()
        assertEquals(listOf(4, 5, 6), list.state.value.items)
    }

    @Test
    fun failuresAreReportedAndCanBeRetried() = runTest {
        var fail = true
        val list = PagedList<Int>(this, 3, { number -> number }) {
            if (fail) error("offline")
            listOf(1, 2, 3)
        }
        list.loadMore()
        advanceUntilIdle()
        assertEquals("offline", list.state.value.error)
        assertTrue(list.state.value.hasLoaded)
        fail = false
        list.loadMore()
        advanceUntilIdle()
        assertNull(list.state.value.error)
        assertEquals(listOf(1, 2, 3), list.state.value.items)
    }

    @Test
    fun customEndRuleKeepsLoadingShortPages() = runTest {
        val list = PagedList<Int>(this, 5, { number -> number }, { page -> page.isEmpty() }) { offset ->
            if (offset == 0) listOf(1, 2) else emptyList()
        }
        list.loadMore()
        advanceUntilIdle()
        assertFalse(list.state.value.endReached)
        list.loadMore()
        advanceUntilIdle()
        assertTrue(list.state.value.endReached)
        assertNotNull(list.state.value.items)
    }
}
