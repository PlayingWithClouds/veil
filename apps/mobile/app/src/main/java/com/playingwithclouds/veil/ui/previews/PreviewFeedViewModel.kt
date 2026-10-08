package com.playingwithclouds.veil.ui.previews

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.DiscoveryRepository
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.TagFilter
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/**
 * The vertical preview feed: the recommendation ranking, narrowed to scenes that have a preview
 * clip and to those without the tags the user hides here.
 */
class PreviewFeedViewModel : ViewModel() {

    private val mutableTagFilter = MutableStateFlow(TagFilter())

    /** The tags hidden from the feed. Only [TagFilter.exclude] applies. */
    val tagFilter: StateFlow<TagFilter> = mutableTagFilter

    private var firstPageLoaded = false

    /** Whether the last fetched page was the final one, counted before scenes without previews drop out. */
    private var rankingExhausted = false

    /** Scenes with a preview clip, in ranking order. */
    val feed: PagedList<RecommendedScene> = PagedList(
        viewModelScope,
        PAGE_SIZE,
        { item -> item.scene.id },
        isLastPage = { _ -> rankingExhausted },
    ) { offset ->
        val refresh = offset == 0 && firstPageLoaded
        val page = DiscoveryRepository.previewFeed(PAGE_SIZE, offset, refresh, mutableTagFilter.value.excludeIds())
        firstPageLoaded = true
        rankingExhausted = page.size < PAGE_SIZE
        page.filter { item -> item.scene.previewVideo != null }
    }

    init {
        feed.loadMore()
    }

    /** Applies a new tag filter and reloads the feed from the top. */
    fun setTagFilter(filter: TagFilter) {
        mutableTagFilter.value = TagFilter(exclude = filter.exclude)
        AppServices.impressions.resetShown()
        feed.refresh()
    }

    /** Logs that a preview came on screen. */
    fun recordShown(item: RecommendedScene, position: Int) {
        AppServices.impressions.recordShown(item.scene.id, item.source, SURFACE, position)
    }

    /** Logs that a preview was opened. */
    fun recordClicked(item: RecommendedScene, position: Int) {
        AppServices.impressions.recordClicked(item.scene.id, item.source, SURFACE, position)
    }

    /** Removes a scene from the loaded feed, e.g. after the user marked it not interested. */
    fun drop(sceneId: String) {
        feed.update { items -> items.filter { item -> item.scene.id != sceneId } }
    }

    companion object {
        private const val PAGE_SIZE = 24
        private const val SURFACE = "preview-feed"
    }
}
