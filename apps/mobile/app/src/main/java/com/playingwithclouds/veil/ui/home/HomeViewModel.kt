package com.playingwithclouds.veil.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.paging.PagedList
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

    init {
        feed.loadMore()
        refreshProgress()
    }

    /** Reloads the resume positions, e.g. after coming back from a scene. */
    fun refreshProgress() {
        viewModelScope.launch { WatchProgressStore.refresh() }
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
    }
}
