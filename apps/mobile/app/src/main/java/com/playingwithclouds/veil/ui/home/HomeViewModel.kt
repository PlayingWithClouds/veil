package com.playingwithclouds.veil.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.FilterPreset
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.SavedFilterRepository
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * The Home tab's feed: the recommendation ranking, or a plain listing once a tag is picked, narrowed
 * by the filter row (sites, runtime, tag, saved presets), with shelves to interleave.
 */
class HomeViewModel : ViewModel() {

    private var firstPageLoaded = false

    private var filterChanged = false

    private val mutableFilter = MutableStateFlow(HomeFilter())

    /** What the feed is narrowed to right now. */
    val filter: StateFlow<HomeFilter> = mutableFilter

    private val mutableSites = MutableStateFlow<List<SearchSite>>(emptyList())

    /** The sites that list scenes, for the site picker. */
    val sites: StateFlow<List<SearchSite>> = mutableSites

    private val mutablePresets = MutableStateFlow<List<FilterPreset>>(emptyList())

    /** The saved filter presets, shown as chips. */
    val presets: StateFlow<List<FilterPreset>> = mutablePresets

    private val mutableTagChips = MutableStateFlow<List<EntityRef>>(emptyList())

    /** The tags most present in the ranked feed, as content-type chips. */
    val tagChips: StateFlow<List<EntityRef>> = mutableTagChips

    private val mutableShelves = MutableStateFlow<List<HomeShelf>>(emptyList())

    /** The rows interleaved into the unfiltered feed. */
    val shelves: StateFlow<List<HomeShelf>> = mutableShelves

    /**
     * The feed. The first load re-uses a recent ranking; every later load of the first page is a
     * user refresh and asks the engine to re-rank, except when it only follows a filter change.
     */
    val feed: PagedList<RecommendedScene> = PagedList(viewModelScope, PAGE_SIZE, { item -> item.scene.id }) { offset -> loadPage(offset) }

    init {
        feed.loadMore()
        viewModelScope.launch { loadSites() }
        viewModelScope.launch { loadPresets() }
    }

    /** Reloads the resume positions shown on cards, e.g. after coming back from a scene. */
    fun refreshProgress() {
        viewModelScope.launch { WatchProgressStore.refresh() }
    }

    /** Replaces the filter and reloads the feed from the stored ranking. */
    fun applyFilter(newFilter: HomeFilter) {
        if (newFilter == mutableFilter.value) {
            return
        }
        mutableFilter.value = newFilter
        filterChanged = true
        feed.refresh()
    }

    /** Toggles one site in the filter. */
    fun toggleSource(name: String) {
        applyFilter(mutableFilter.value.toggleSource(name))
    }

    /** Stores the current filter as a preset chip. */
    fun saveCurrentAsPreset(name: String) {
        val current = mutableFilter.value
        viewModelScope.launch {
            try {
                val preset = SavedFilterRepository.create(name, current.tagId, current.sources.toList(), current.runtime)
                mutablePresets.update { presets -> listOf(preset) + presets }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                // The chip simply does not appear; the filter itself stays applied.
            }
        }
    }

    /** Deletes a preset. */
    fun deletePreset(preset: FilterPreset) {
        mutablePresets.update { presets -> presets.filter { saved -> saved.id != preset.id } }
        AppScope.launch { runCatching { SavedFilterRepository.delete(preset.id) } }
    }

    /** Removes a scene from the loaded feed, e.g. after the user marked it not interested. */
    fun dropFromFeed(sceneId: String) {
        feed.update { items -> items.filter { item -> item.scene.id != sceneId } }
        mutableShelves.update { shelves -> shelves.map { shelf -> withoutScene(shelf, sceneId) } }
    }

    /** Swipe-away: removes the scene and records a dislike, which the recommender weighs negatively. */
    fun notInterested(scene: SceneSummary) {
        dropFromFeed(scene.id)
        AppScope.launch { runCatching { SceneRepository.setVerdict(scene.id, Verdict.DOWN) } }
    }

    /** Logs that a recommendation was shown; plain tag listings are not recommendations. */
    fun recordShown(item: RecommendedScene, position: Int) {
        if (item.source == LISTING_SOURCE) {
            return
        }
        AppServices.impressions.recordShown(item.scene.id, item.source, SURFACE, position)
    }

    /** Logs that a recommendation was opened. */
    fun recordClicked(item: RecommendedScene, position: Int) {
        if (item.source == LISTING_SOURCE) {
            return
        }
        AppServices.impressions.recordClicked(item.scene.id, item.source, SURFACE, position)
    }

    /** Fetches one page for the current filter and, for the first page, refreshes what hangs off it. */
    private suspend fun loadPage(offset: Int): List<RecommendedScene> {
        val current = mutableFilter.value
        val forceRanking = offset == 0 && firstPageLoaded && !filterChanged
        val page = fetchPage(current, offset, forceRanking)
        if (offset == 0) {
            firstPageLoaded = true
            filterChanged = false
            AppServices.impressions.resetShown()
            refreshAroundFirstPage(current, page)
        }
        return page
    }

    /** One page from the ranking, or from the tag listing once a tag is picked. */
    private suspend fun fetchPage(current: HomeFilter, offset: Int, forceRanking: Boolean): List<RecommendedScene> {
        val sources = current.sources.toList()
        if (current.tagId == null) {
            return FeedRepository.recommendations(PAGE_SIZE, offset, forceRanking, sources, current.runtime)
        }
        val scenes = FeedRepository.filteredScenes(PAGE_SIZE, offset, current.tagId, sources, current.runtime)
        return scenes.map { scene -> RecommendedScene(scene, LISTING_SOURCE, "") }
    }

    /** Derives the tag chips and shelves from an unfiltered first page; a filtered feed shows no shelves. */
    private fun refreshAroundFirstPage(current: HomeFilter, page: List<RecommendedScene>) {
        if (current.isActive) {
            mutableShelves.value = emptyList()
            return
        }
        val scenes = page.map { item -> item.scene }
        mutableTagChips.value = topTags(scenes, TAG_CHIP_COUNT)
        viewModelScope.launch { loadShelves(scenes) }
    }

    /** Builds the shelves from the recommender's rows and the feed's credits; a failure leaves none. */
    private suspend fun loadShelves(feedScenes: List<SceneSummary>) {
        try {
            val rows = FeedRepository.recommendedRows(ROW_LIMIT, ROW_SIZE)
            mutableShelves.value = buildShelves(rows, feedScenes)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            mutableShelves.value = buildShelves(emptyList(), feedScenes)
        }
    }

    /** Loads the site picker's sites; without them it simply offers none. */
    private suspend fun loadSites() {
        try {
            mutableSites.value = SearchRepository.searchSites().filter { site -> site.listsScenes }
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // The picker is optional.
        }
    }

    /** Loads the saved presets; without them no preset chips show. */
    private suspend fun loadPresets() {
        try {
            mutablePresets.value = SavedFilterRepository.list()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Presets are optional.
        }
    }

    /** A scene shelf without [sceneId]; other shelves unchanged. */
    private fun withoutScene(shelf: HomeShelf, sceneId: String): HomeShelf {
        if (shelf !is HomeShelf.Scenes) {
            return shelf
        }
        return shelf.copy(scenes = shelf.scenes.filter { scene -> scene.id != sceneId })
    }

    companion object {
        private const val PAGE_SIZE = 24
        private const val SURFACE = "feed"

        /** Source label of cards from the tag listing, which the recommender never served. */
        private const val LISTING_SOURCE = "listing"
        private const val TAG_CHIP_COUNT = 10
        private const val ROW_LIMIT = 4
        private const val ROW_SIZE = 10
    }
}
