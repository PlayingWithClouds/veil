package com.playingwithclouds.veil.ui.search

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.SearchBatch
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Recent searches offered while the box is empty. */
private const val RECENT_LIMIT = 12

/**
 * The search screen. [query] is the search running (null while only typing); [sitesAsked] and
 * [sitesAnswered] count the sites of the page being asked; [done] once every site ran out.
 */
data class SearchState(
    val text: String = "",
    val query: String? = null,
    val recent: List<String> = emptyList(),
    val videos: List<VideoSummary> = emptyList(),
    val performers: List<PerformerSummary> = emptyList(),
    val searching: Boolean = false,
    val sitesAsked: Int = 0,
    val sitesAnswered: Int = 0,
    val done: Boolean = false,
    val error: String? = null,
)

/**
 * Searches the sites a page at a time, as the web's search page does: each site's results join
 * the list as it answers; scrolling to the end asks for the next page until every site ran out.
 */
class SearchViewModel(initialQuery: String) : ViewModel() {

    private val mutableState = MutableStateFlow(SearchState(text = initialQuery))
    val state: StateFlow<SearchState> = mutableState.asStateFlow()

    private var pageJob: Job? = null
    private var page = 0
    private val exhaustedSites = mutableSetOf<String>()

    init {
        viewModelScope.launch {
            val recent = runCatching { SearchRepository.recent(RECENT_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(recent = recent) }
        }
        if (initialQuery.isNotBlank()) {
            submit()
        }
    }

    /** Takes the box's text. */
    fun onTextChange(text: String) {
        mutableState.update { state -> state.copy(text = text) }
    }

    /** Searches for a recent search again. */
    fun searchFor(text: String) {
        onTextChange(text)
        submit()
    }

    /** Starts a search for the box's text, replacing the one before. */
    fun submit() {
        val query = state.value.text.trim()
        if (query.isEmpty()) {
            return
        }
        pageJob?.cancel()
        page = 0
        exhaustedSites.clear()
        AppServices.impressions.resetShown(ImpressionSurface.SEARCH)
        SearchRepository.logSearch(query)
        mutableState.update { state ->
            state.copy(query = query, videos = emptyList(), performers = emptyList(), done = false, error = null)
        }
        askPage()
    }

    /** Asks for the next page, unless one is being asked or every site ran out. */
    fun loadMore() {
        val current = state.value
        if (current.query == null || current.searching || current.done) {
            return
        }
        page++
        askPage()
    }

    /** Asks every site for the current page, adding each answer as it comes. */
    private fun askPage() {
        val query = state.value.query ?: return
        val asked = page
        var found = 0
        mutableState.update { state -> state.copy(searching = true, sitesAsked = 0, sitesAnswered = 0) }
        pageJob = viewModelScope.launch {
            try {
                SearchRepository.page(query, asked).collect { batch ->
                    found += batch.videos.size
                    add(batch)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { state -> state.copy(error = error.displayMessage()) }
            }
            val allOut = exhaustedSites.size >= state.value.sitesAsked
            mutableState.update { state -> state.copy(searching = false, done = found == 0 || allOut) }
        }
    }

    /** Adds one site's answer: new videos at the end, new performers in the row. */
    private fun add(batch: SearchBatch) {
        if (batch.exhausted) {
            exhaustedSites.add(batch.plugin)
        }
        mutableState.update { state ->
            val knownVideos = state.videos.map { video -> video.id }.toSet()
            val knownPerformers = state.performers.map { performer -> performer.id }.toSet()
            var answered = state.sitesAnswered
            if (batch.final) {
                answered++
            }
            state.copy(
                videos = state.videos + batch.videos.filter { video -> video.id !in knownVideos }.distinctBy { video -> video.id },
                performers = state.performers + batch.performers.filter { performer -> performer.id !in knownPerformers }.distinctBy { performer -> performer.id },
                sitesAsked = batch.sites,
                sitesAnswered = answered,
            )
        }
    }
}
