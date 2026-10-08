package com.playingwithclouds.veil.ui.search

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.LiveResultItem
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.data.SearchSuggestion
import com.playingwithclouds.veil.data.StudioSummary
import com.playingwithclouds.veil.data.SubscriptionRepository
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Which kind of result the search page lists. */
enum class SearchScope(val label: String) {
    ALL("All"),
    SCENES("Videos"),
    GALLERIES("Galleries"),
    PERFORMERS("Performers"),
    STUDIOS("Studios"),
}

/** Everything the search page shows. */
data class SearchState(
    val query: String = "",
    /** The query the results belong to; null while the user is still typing. */
    val submitted: String? = null,
    val suggestions: List<SearchSuggestion> = emptyList(),
    val scope: SearchScope = SearchScope.ALL,
    val sites: List<SearchSite> = emptyList(),
    val selectedSites: Set<String> = emptySet(),
    val scenes: List<SceneSummary> = emptyList(),
    val galleries: List<GallerySummary> = emptyList(),
    val performers: List<PerformerSummary> = emptyList(),
    val studios: List<StudioSummary> = emptyList(),
    /** Library matches are loading. */
    val isLoading: Boolean = false,
    /** Source sites are still answering. */
    val isSearchingSites: Boolean = false,
    val error: String? = null,
    /** The submitted search is already followed. */
    val isFollowed: Boolean = false,
)

/** Library search plus live search on the source sites, with suggestions while typing. */
class SearchViewModel(initialQuery: String) : ViewModel() {

    private val mutableState = MutableStateFlow(SearchState(query = initialQuery))
    private var suggestionJob: Job? = null
    private var searchJob: Job? = null

    /** The page state. */
    val state: StateFlow<SearchState> = mutableState

    init {
        viewModelScope.launch {
            val sites = runCatching { SearchRepository.searchSites() }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(sites = sites) }
        }
        if (initialQuery.isNotBlank()) {
            submit(initialQuery)
        }
        if (initialQuery.isBlank()) {
            loadSuggestions("")
        }
    }

    /** Takes the text of the search box and refreshes the suggestions after a short pause. */
    fun onQueryChange(text: String) {
        mutableState.update { current -> current.copy(query = text, submitted = null) }
        suggestionJob?.cancel()
        suggestionJob = viewModelScope.launch {
            delay(SUGGESTION_DEBOUNCE_MILLISECONDS)
            loadSuggestions(text)
        }
    }

    /** Fetches suggestions for the text; a failure leaves the list empty. */
    private fun loadSuggestions(text: String) {
        viewModelScope.launch {
            val suggestions = runCatching { SearchRepository.suggestions(text.trim(), SUGGESTION_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(suggestions = suggestions) }
        }
    }

    /** Runs the search for the text and remembers it as a recent one. */
    fun submit(text: String) {
        val query = text.trim()
        if (query.isEmpty()) {
            return
        }
        mutableState.update { current -> current.copy(query = query, submitted = query, error = null, isFollowed = false) }
        viewModelScope.launch { runCatching { SearchRepository.recordSearch(query) } }
        runSearch(query)
    }

    /** Removes a recent search from the suggestions. */
    fun forget(suggestion: SearchSuggestion) {
        mutableState.update { current -> current.copy(suggestions = current.suggestions - suggestion) }
        viewModelScope.launch { runCatching { SearchRepository.forgetSearch(suggestion.text) } }
    }

    /** Narrows the results to one kind. */
    fun setScope(scope: SearchScope) {
        mutableState.update { current -> current.copy(scope = scope) }
    }

    /** Includes or excludes a site and searches again. */
    fun toggleSite(name: String) {
        mutableState.update { current ->
            val selected = current.selectedSites.toMutableSet()
            if (!selected.add(name)) {
                selected.remove(name)
            }
            current.copy(selectedSites = selected)
        }
        val query = mutableState.value.submitted ?: return
        runSearch(query)
    }

    /** Follows the submitted search, so the backend re-runs it on a schedule. */
    fun followSearch() {
        val current = mutableState.value
        val query = current.submitted ?: return
        viewModelScope.launch {
            try {
                SubscriptionRepository.subscribeSearch(query, current.selectedSites.toList())
                mutableState.update { state -> state.copy(isFollowed = true) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { state -> state.copy(error = error.displayMessage()) }
            }
        }
    }

    /** Loads the library matches, then streams the live results of the sites. */
    private fun runSearch(query: String) {
        searchJob?.cancel()
        val sources = mutableState.value.selectedSites.toList()
        mutableState.update { current ->
            current.copy(
                scenes = emptyList(),
                galleries = emptyList(),
                performers = emptyList(),
                studios = emptyList(),
                isLoading = true,
                isSearchingSites = false,
            )
        }
        searchJob = viewModelScope.launch {
            try {
                loadLibraryMatches(query, sources)
                streamLiveResults(query, sources)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(error = error.displayMessage()) }
            }
            mutableState.update { current -> current.copy(isLoading = false, isSearchingSites = false) }
        }
    }

    /** Fetches stored scenes, galleries, performers and studios for the query in parallel. */
    private suspend fun loadLibraryMatches(query: String, sources: List<String>) {
        coroutineScope {
            val scenes = async { runCatching { SearchRepository.searchScenes(query, sources, SCENE_LIMIT, 0) }.getOrDefault(emptyList()) }
            val galleries = async { runCatching { SearchRepository.searchGalleries(query, sources, GALLERY_LIMIT, 0) }.getOrDefault(emptyList()) }
            val performers = async { runCatching { SearchRepository.searchPerformers(query, ENTITY_LIMIT) }.getOrDefault(emptyList()) }
            val studios = async { runCatching { SearchRepository.searchStudios(query, ENTITY_LIMIT) }.getOrDefault(emptyList()) }
            awaitAll(scenes, galleries, performers, studios)
            mutableState.update { current ->
                current.copy(
                    scenes = scenes.await(),
                    galleries = galleries.await(),
                    performers = performers.await(),
                    studios = studios.await(),
                    isLoading = false,
                    isSearchingSites = true,
                )
            }
        }
    }

    /** Adds results from the source sites as they arrive, skipping what is already listed. */
    private suspend fun streamLiveResults(query: String, sources: List<String>) {
        SearchRepository.liveSearch(query, sources)
            .catch { }
            .collect { hit ->
                val result = SearchRepository.toLiveResult(hit) ?: return@collect
                mutableState.update { current -> current.withLiveResult(result) }
            }
    }

    companion object {
        private const val SUGGESTION_DEBOUNCE_MILLISECONDS = 200L
        private const val SUGGESTION_LIMIT = 12
        private const val SCENE_LIMIT = 40
        private const val GALLERY_LIMIT = 30
        private const val ENTITY_LIMIT = 20
    }
}

/** The state with a live result appended, unless the library already listed it. */
fun SearchState.withLiveResult(result: LiveResultItem): SearchState {
    when (result) {
        is LiveResultItem.SceneResult -> {
            if (scenes.any { scene -> scene.id == result.scene.id }) {
                return this
            }
            return copy(scenes = scenes + result.scene)
        }
        is LiveResultItem.GalleryResult -> {
            if (galleries.any { gallery -> gallery.id == result.gallery.id }) {
                return this
            }
            return copy(galleries = galleries + result.gallery)
        }
    }
}
