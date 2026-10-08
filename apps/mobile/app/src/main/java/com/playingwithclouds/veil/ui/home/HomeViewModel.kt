package com.playingwithclouds.veil.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** What the Home tab lists. */
enum class HomeMode(val label: String) {
    VIDEOS("Videos"),
    GALLERIES("Galleries"),
}

/** The recommendation feed of the Home tab, or the newest galleries, narrowed to chosen sites. */
class HomeViewModel : ViewModel() {

    private var firstPageLoaded = false

    private var sourcesChanged = false

    private val mutableMode = MutableStateFlow(HomeMode.VIDEOS)

    /** Whether Home shows videos or galleries. */
    val mode: StateFlow<HomeMode> = mutableMode

    private val mutableSites = MutableStateFlow<List<SearchSite>>(emptyList())

    /** The sites that list content, offered as filter badges. */
    val sites: StateFlow<List<SearchSite>> = mutableSites

    private val mutableSelectedSources = MutableStateFlow<Set<String>>(emptySet())

    /** The plugin names the lists are narrowed to; empty means every site. */
    val selectedSources: StateFlow<Set<String>> = mutableSelectedSources

    /**
     * The ranked feed. The first load re-uses a recent ranking; every later load of the first page
     * is a user refresh and asks the engine to re-rank, except when it only follows a site change.
     */
    val feed = PagedList<RecommendedScene>(viewModelScope, PAGE_SIZE, { item -> item.scene.id }) { offset ->
        val forceRanking = offset == 0 && firstPageLoaded && !sourcesChanged
        val page = FeedRepository.recommendations(PAGE_SIZE, offset, forceRanking, mutableSelectedSources.value.toList())
        if (offset == 0) {
            firstPageLoaded = true
            sourcesChanged = false
            AppServices.impressions.resetShown()
        }
        page
    }

    /** The newest galleries of the chosen sites. */
    val galleries = PagedList<GallerySummary>(viewModelScope, PAGE_SIZE, { gallery -> gallery.id }) { offset ->
        SearchRepository.searchGalleries("", mutableSelectedSources.value.toList(), PAGE_SIZE, offset)
    }

    private val mutableContinueWatching = MutableStateFlow<List<SceneSummary>>(emptyList())

    /** Started but unfinished scenes, most recent first. */
    val continueWatching: StateFlow<List<SceneSummary>> = mutableContinueWatching

    init {
        feed.loadMore()
        refreshProgress()
        viewModelScope.launch { loadSites() }
    }

    /** Switches between videos and galleries; site choices that the new mode has no use for are dropped. */
    fun selectMode(mode: HomeMode) {
        if (mode == mutableMode.value) {
            return
        }
        mutableMode.value = mode
        val offered = sitesFor(mode).map { site -> site.name }.toSet()
        mutableSelectedSources.value = mutableSelectedSources.value.intersect(offered)
        reloadCurrent()
    }

    /** Adds or removes a site from the filter and reloads. */
    fun toggleSource(name: String) {
        val current = mutableSelectedSources.value
        if (current.contains(name)) {
            mutableSelectedSources.value = current - name
        } else {
            mutableSelectedSources.value = current + name
        }
        reloadCurrent()
    }

    /** Clears the site filter. */
    fun clearSources() {
        if (mutableSelectedSources.value.isEmpty()) {
            return
        }
        mutableSelectedSources.value = emptySet()
        reloadCurrent()
    }

    /** The sites that list what [mode] shows. */
    fun sitesFor(mode: HomeMode): List<SearchSite> {
        if (mode == HomeMode.GALLERIES) {
            return mutableSites.value.filter { site -> site.listsGalleries }
        }
        return mutableSites.value.filter { site -> site.listsScenes }
    }

    /** Reloads the list of the current mode from the top. */
    private fun reloadCurrent() {
        if (mutableMode.value == HomeMode.GALLERIES) {
            galleries.refresh()
            return
        }
        sourcesChanged = true
        feed.refresh()
    }

    /** Loads the filterable sites; without them Home simply shows no badges. */
    private suspend fun loadSites() {
        try {
            mutableSites.value = SearchRepository.searchSites()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Badges are optional.
        }
    }

    /** Reloads the resume positions and the continue-watching shelf, e.g. after coming back from a scene. */
    fun refreshProgress() {
        viewModelScope.launch { WatchProgressStore.refresh() }
        viewModelScope.launch { loadContinueWatching() }
    }

    /** Fills the shelf from the watch history; a failure leaves it as it was. */
    private suspend fun loadContinueWatching() {
        try {
            val history = FeedRepository.watchHistory(HISTORY_SCAN, 0)
            val unfinished = history.filter { entry -> !entry.completed && entry.progressSeconds > 0 }
            mutableContinueWatching.value = unfinished.mapNotNull { entry -> entry.scene }.take(SHELF_SIZE)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // The shelf is optional; the feed below carries on.
        }
    }

    /** Removes a scene from the loaded feed, e.g. after the user marked it not interested. */
    fun dropFromFeed(sceneId: String) {
        feed.update { items -> items.filter { item -> item.scene.id != sceneId } }
        mutableContinueWatching.value = mutableContinueWatching.value.filter { scene -> scene.id != sceneId }
    }

    /** Logs that a recommendation was shown. */
    fun recordShown(item: RecommendedScene, position: Int) {
        AppServices.impressions.recordShown(item.scene.id, item.source, SURFACE, position)
    }

    /** Logs that a recommendation was opened. */
    fun recordClicked(item: RecommendedScene, position: Int) {
        AppServices.impressions.recordClicked(item.scene.id, item.source, SURFACE, position)
    }

    companion object {
        private const val PAGE_SIZE = 24
        private const val SURFACE = "feed"
        private const val HISTORY_SCAN = 40
        private const val SHELF_SIZE = 12
    }
}
