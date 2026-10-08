package com.playingwithclouds.veil.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** The recommendation feed of the Home tab. */
class HomeViewModel : ViewModel() {

    private var firstPageLoaded = false

    /**
     * The ranked feed. The first load re-uses a recent ranking; every later load of the first page
     * is a user refresh and asks the engine to re-rank.
     */
    val feed = PagedList<RecommendedScene>(viewModelScope, PAGE_SIZE, { item -> item.scene.id }) { offset ->
        val forceRanking = offset == 0 && firstPageLoaded
        val page = FeedRepository.recommendations(PAGE_SIZE, offset, forceRanking)
        if (offset == 0) {
            firstPageLoaded = true
            AppServices.impressions.resetShown()
        }
        page
    }

    private val mutableContinueWatching = MutableStateFlow<List<SceneSummary>>(emptyList())

    /** Started but unfinished scenes, most recent first. */
    val continueWatching: StateFlow<List<SceneSummary>> = mutableContinueWatching

    init {
        feed.loadMore()
        refreshProgress()
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
