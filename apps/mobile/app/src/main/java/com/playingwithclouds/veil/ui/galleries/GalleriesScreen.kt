package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** Galleries per page of the feed. */
private const val FEED_PAGE = 24

/** Galleries in each row above the feed. */
private const val ROW_LIMIT = 20

/** Width of a gallery in the rows above the feed. */
private val RowCardWidth = 150.dp

/** Where the feed is: recommended galleries first, then the rest of the library. */
private enum class FeedStage { RANKED, REST }

/** The galleries overview: rows of recently viewed and favourites, and the endless feed. */
class GalleriesViewModel : ViewModel() {

    private val recentState = MutableStateFlow(emptyList<GallerySummary>())
    private val favoritesState = MutableStateFlow(emptyList<GallerySummary>())
    private var stage = FeedStage.RANKED
    private var stageOffset = 0

    val recent: StateFlow<List<GallerySummary>> = recentState.asStateFlow()
    val favorites: StateFlow<List<GallerySummary>> = favoritesState.asStateFlow()

    /** Recommended galleries, then the rest of the library newest first. */
    val feed = PagedList<GallerySummary>(viewModelScope, FEED_PAGE, { gallery -> gallery.id }) { offset ->
        if (offset == 0) {
            stage = FeedStage.RANKED
            stageOffset = 0
        }
        nextFeedPage()
    }

    init {
        feed.loadMore()
        viewModelScope.launch { recentState.value = runCatching { GalleryRepository.recentlyViewed(ROW_LIMIT) }.getOrDefault(emptyList()) }
        viewModelScope.launch { favoritesState.value = runCatching { GalleryRepository.favorites(ROW_LIMIT) }.getOrDefault(emptyList()) }
    }

    /** The next page of the feed; when the recommendations run short, the library's rest fills it up. */
    private suspend fun nextFeedPage(): List<GallerySummary> {
        if (stage == FeedStage.RANKED) {
            val ranked = GalleryRepository.recommended(FEED_PAGE, stageOffset)
            stageOffset += FEED_PAGE
            if (ranked.size == FEED_PAGE) {
                return ranked.distinctBy { gallery -> gallery.id }
            }
            stage = FeedStage.REST
            stageOffset = 0
            return (ranked + nextRestPage()).distinctBy { gallery -> gallery.id }
        }
        return nextRestPage()
    }

    /** The next page of the library's galleries, newest first. */
    private suspend fun nextRestPage(): List<GallerySummary> {
        val page = GalleryRepository.listing(FEED_PAGE, stageOffset)
        stageOffset += FEED_PAGE
        return page
    }
}

/** Galleries: recently viewed and favourites as rows, then more for you, endlessly. */
@Composable
fun GalleriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { GalleriesViewModel() }
    val recent by viewModel.recent.collectAsStateWithLifecycle()
    val favorites by viewModel.favorites.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Galleries", onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.feed,
            keyOf = { gallery -> gallery.id },
            emptyText = "No galleries yet. They come in from the gallery sites.",
            modifier = Modifier.padding(padding),
            cellWidth = PosterCellWidth,
            header = {
                fullWidthItem("recent") {
                    Shelf("Recently viewed", recent, key = { gallery -> gallery.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { gallery, _ ->
                        GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                    }
                }
                fullWidthItem("favorites") {
                    Shelf("Favourites", favorites, key = { gallery -> gallery.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { gallery, _ ->
                        GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                    }
                }
                fullWidthItem("feed-heading") { SectionHeading("More for you") }
            },
        ) { _, gallery ->
            GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
        }
    }
}
