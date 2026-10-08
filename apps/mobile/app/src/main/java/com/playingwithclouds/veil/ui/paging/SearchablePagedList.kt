package com.playingwithclouds.veil.ui.paging

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** A [PagedList] whose content follows a search box: typing reloads the list after a short pause. */
class SearchablePagedList<T>(
    private val scope: CoroutineScope,
    pageSize: Int,
    keyOf: (T) -> Any,
    private val debounceMilliseconds: Long = DEBOUNCE_MILLISECONDS,
    private val fetchPage: suspend (query: String, offset: Int) -> List<T>,
) {

    private val mutableQuery = MutableStateFlow("")
    private var debounceJob: Job? = null

    /** The text in the search box. */
    val query: StateFlow<String> = mutableQuery

    /** The list for the current query. */
    val paged = PagedList(scope, pageSize, keyOf) { offset -> fetchPage(mutableQuery.value.trim(), offset) }

    init {
        paged.loadMore()
    }

    /** Takes new text from the search box and reloads once typing pauses. */
    fun setQuery(text: String) {
        mutableQuery.value = text
        debounceJob?.cancel()
        debounceJob = scope.launch {
            delay(debounceMilliseconds)
            paged.refresh()
        }
    }

    companion object {
        const val DEBOUNCE_MILLISECONDS = 300L
    }
}
