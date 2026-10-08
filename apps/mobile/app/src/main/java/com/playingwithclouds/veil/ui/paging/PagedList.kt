package com.playingwithclouds.veil.ui.paging

import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** What a paged list shows right now. */
data class PagedState<T>(
    val items: List<T> = emptyList(),
    /** A page is on its way. */
    val isLoading: Boolean = false,
    /** The first page is being reloaded while the old items stay visible. */
    val isRefreshing: Boolean = false,
    /** The last page came back short, so there is nothing more to load. */
    val endReached: Boolean = false,
    val error: String? = null,
    /** Whether the first page has arrived (or failed) at least once. */
    val hasLoaded: Boolean = false,
)

/**
 * An endless list loaded page by page. Pages that overlap (feeds can shift between requests) are
 * deduplicated by [keyOf], and a response that arrives after a refresh is ignored.
 */
class PagedList<T>(
    private val scope: CoroutineScope,
    private val pageSize: Int,
    private val keyOf: (T) -> Any,
    private val isLastPage: (List<T>) -> Boolean = { page -> page.size < pageSize },
    private val fetchPage: suspend (offset: Int) -> List<T>,
) {

    private val mutableState = MutableStateFlow(PagedState<T>())
    private var generation = 0
    private var loadJob: Job? = null
    private var nextOffset = 0

    /** The list as loaded so far. */
    val state: StateFlow<PagedState<T>> = mutableState

    /** Starts over from the first page; the current items stay until it arrives. */
    fun refresh() {
        generation++
        loadJob?.cancel()
        nextOffset = 0
        mutableState.value = mutableState.value.copy(isLoading = false, isRefreshing = true, endReached = false, error = null)
        loadNextPage()
    }

    /** Loads the next page, unless one is on its way or the end was reached. */
    fun loadMore() {
        val current = mutableState.value
        if (current.isLoading || current.endReached) {
            return
        }
        loadNextPage()
    }

    /** Applies a change to the loaded items, e.g. removing a deleted entry. */
    fun update(transform: (List<T>) -> List<T>) {
        mutableState.value = mutableState.value.copy(items = transform(mutableState.value.items))
    }

    /** Fetches the page at the next offset and merges it in. */
    private fun loadNextPage() {
        val requestGeneration = generation
        val requestOffset = nextOffset
        mutableState.value = mutableState.value.copy(isLoading = true, error = null)
        loadJob = scope.launch {
            try {
                val page = fetchPage(requestOffset)
                if (requestGeneration == generation) {
                    applyPage(page, requestOffset)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                if (requestGeneration == generation) {
                    mutableState.value = mutableState.value.copy(
                        isLoading = false,
                        isRefreshing = false,
                        error = error.displayMessage(),
                        hasLoaded = true,
                    )
                }
            }
        }
    }

    /** Appends (or, for the first page, installs) a fetched page. */
    private fun applyPage(page: List<T>, offset: Int) {
        val current = mutableState.value
        var merged = page
        if (offset > 0) {
            val knownKeys = current.items.map(keyOf).toSet()
            merged = current.items + page.filter { item -> keyOf(item) !in knownKeys }
        }
        nextOffset = offset + pageSize
        mutableState.value = current.copy(
            items = merged,
            isLoading = false,
            isRefreshing = false,
            endReached = isLastPage(page),
            error = null,
            hasLoaded = true,
        )
    }
}
