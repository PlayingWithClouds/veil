package com.playingwithclouds.veil.ui.search

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.SceneLabelWords
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.SearchBatch
import com.playingwithclouds.veil.data.SearchFilter
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SiteChoice
import com.playingwithclouds.veil.data.SiteStyle
import com.playingwithclouds.veil.data.SiteStyles
import com.playingwithclouds.veil.data.TagSummary
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
    val filter: SearchFilter = SearchFilter(),
    /** The sites the filter offers, and their styles. */
    val siteChoices: List<SiteChoice> = emptyList(),
    val siteStyles: List<SiteStyle> = emptyList(),
    /** The tags most common among the library's videos matching the filter, offered for hiding. */
    val tagCounts: List<Pair<TagSummary, Int>> = emptyList(),
    /** The POV tag of the filter's Camera group, once looked up. */
    val povTag: TagSummary? = null,
    /** The scene labels in the library, most common first, for "What's in it" and typed words. */
    val sceneLabels: List<Pair<TagSummary, Int>> = emptyList(),
    /** The library's videos with the filter's scene labels (and the text), most matching first. */
    val sceneVideos: List<VideoSummary> = emptyList(),
    /** While the library's scene videos are read. */
    val loadingScenes: Boolean = false,
)

/** Tags offered for hiding. */
private const val TAG_COUNT_LIMIT = 24

/** Scene labels offered in the filter. */
private const val SCENE_LABEL_LIMIT = 40

/** Library videos shown for scene labels. */
private const val SCENE_VIDEO_LIMIT = 60

/**
 * Searches the sites a page at a time, as the web's search page does: each site's results join
 * the list as it answers; scrolling to the end asks for the next page until every site ran out.
 * Scene labels (picked in "What's in it", or typed: "pov blowjob") also list the library's scanned
 * videos that have them; only the other words go to the sites, and none are left when every word
 * named a label.
 */
class SearchViewModel(initialQuery: String) : ViewModel() {

    private val mutableState = MutableStateFlow(SearchState(text = initialQuery))
    val state: StateFlow<SearchState> = mutableState.asStateFlow()

    private var pageJob: Job? = null
    private var sceneJob: Job? = null
    private var page = 0
    private val exhaustedSites = mutableSetOf<String>()

    init {
        viewModelScope.launch {
            val recent = runCatching { SearchRepository.recent(RECENT_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(recent = recent) }
        }
        viewModelScope.launch {
            val choices = runCatching { SearchRepository.siteChoices() }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(siteChoices = choices, siteStyles = SiteStyles.of(choices)) }
        }
        viewModelScope.launch {
            val povTag = runCatching { SearchRepository.povTag() }.getOrNull()
            mutableState.update { state -> state.copy(povTag = povTag) }
        }
        loadTagCounts()
        // The labels are read before the first search, so typed label words are recognised.
        viewModelScope.launch {
            val labels = runCatching { SceneRepository.labelCounts(SCENE_LABEL_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(sceneLabels = labels) }
            if (initialQuery.isNotBlank()) {
                submit()
            }
        }
    }

    /** Takes a changed filter and searches again with it when a search is shown or scene labels are asked for. */
    fun setFilter(filter: SearchFilter) {
        mutableState.update { state -> state.copy(filter = filter) }
        loadTagCounts()
        val query = state.value.query
        if (query == "" && filter.sceneLabels.isEmpty()) {
            mutableState.update { state -> state.copy(query = null, sceneVideos = emptyList()) }
            return
        }
        if (query != null || filter.sceneLabels.isNotEmpty()) {
            submit()
        }
    }

    /** The tags offered for hiding, under the current filter. */
    private fun loadTagCounts() {
        val filter = state.value.filter
        viewModelScope.launch {
            val counts = runCatching { SearchRepository.tagCounts(filter, TAG_COUNT_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(tagCounts = counts.filter { (tag, _) -> tag.id !in state.filter.hiddenTags }) }
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

    /**
     * Starts a search for the box's text, replacing the one before. Words naming scene labels join
     * the filter; the rest is asked of the sites. With labels and no other words only the library
     * is listed.
     */
    fun submit() {
        val typed = state.value.text.trim()
        val words = SceneLabelWords.match(typed, state.value.sceneLabels.map { (label, _) -> label })
        var filter = state.value.filter
        for (label in words.labels) {
            if (label.id !in filter.sceneLabels) {
                filter = filter.toggledSceneLabel(label)
            }
        }
        val query = words.text
        if (query.isEmpty() && filter.sceneLabels.isEmpty()) {
            return
        }
        pageJob?.cancel()
        page = 0
        exhaustedSites.clear()
        AppServices.impressions.resetShown(ImpressionSurface.SEARCH)
        if (typed.isNotEmpty()) {
            SearchRepository.logSearch(typed)
        }
        mutableState.update { state ->
            state.copy(text = query, query = query, filter = filter, videos = emptyList(), performers = emptyList(), done = query.isEmpty(), error = null)
        }
        loadSceneVideos()
        if (query.isNotEmpty()) {
            askPage()
        }
    }

    /** Lists the library's videos with the filter's scene labels and the query's words. */
    private fun loadSceneVideos() {
        val current = state.value
        sceneJob?.cancel()
        if (current.filter.sceneLabels.isEmpty()) {
            mutableState.update { state -> state.copy(sceneVideos = emptyList(), loadingScenes = false) }
            return
        }
        mutableState.update { state -> state.copy(loadingScenes = true) }
        sceneJob = viewModelScope.launch {
            val videos = runCatching { SearchRepository.sceneVideos(current.query.orEmpty(), current.filter, SCENE_VIDEO_LIMIT) }.getOrDefault(emptyList())
            mutableState.update { state -> state.copy(sceneVideos = videos, loadingScenes = false) }
        }
    }

    /**
     * Opens a result: at its first scene with the filter's labels when any are asked for, else
     * from where it would start anyway ([onOpen] gets null).
     */
    fun openResult(videoId: String, onOpen: (Double?) -> Unit) {
        val labels = state.value.filter.sceneLabels.values
        if (labels.isEmpty()) {
            onOpen(null)
            return
        }
        viewModelScope.launch {
            onOpen(runCatching { SceneRepository.firstSceneStart(videoId, labels) }.getOrNull())
        }
    }

    /** Asks for the next page, unless one is being asked or every site ran out. */
    fun loadMore() {
        val current = state.value
        if (current.query.isNullOrEmpty() || current.searching || current.done) {
            return
        }
        page++
        askPage()
    }

    /** Asks every site for the current page, adding each answer as it comes. */
    private fun askPage() {
        val query = state.value.query ?: return
        val filter = state.value.filter
        val asked = page
        var found = 0
        mutableState.update { state -> state.copy(searching = true, sitesAsked = 0, sitesAnswered = 0) }
        pageJob = viewModelScope.launch {
            try {
                SearchRepository.page(query, asked, filter).collect { batch ->
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
