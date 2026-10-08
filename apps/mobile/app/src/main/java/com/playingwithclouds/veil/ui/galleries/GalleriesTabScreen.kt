package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.SiteFilter
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.SiteBadges
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.paging.PagedList

/** The newest galleries, narrowed to the sites picked in the badges. */
class GalleriesTabViewModel : ViewModel() {

    /** The site badges; a new pick reloads the list. */
    val siteFilter = SiteFilter(viewModelScope, offers = { site -> site.listsGalleries }) { galleries.refresh() }

    /** The galleries, newest first. */
    val galleries: PagedList<GallerySummary> = PagedList(viewModelScope, PAGE_SIZE, { gallery -> gallery.id }) { offset ->
        SearchRepository.searchGalleries("", siteFilter.selectedSources(), PAGE_SIZE, offset)
    }

    init {
        galleries.loadMore()
    }

    companion object {
        private const val PAGE_SIZE = 40
    }
}

/**
 * The Galleries tab: site badges, a way into the category index, then the newest galleries. The
 * large header scrolls away and a slim title bar takes its place.
 */
@Composable
fun GalleriesTabScreen(navigator: AppNavigator) {
    val viewModel = viewModel { GalleriesTabViewModel() }
    val gridState = rememberLazyGridState()
    val headerScrolledAway by remember { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }

    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            PagedGrid(
                paged = viewModel.galleries,
                keyOf = { gallery -> gallery.id },
                emptyText = "No galleries from these sites yet.",
                cellWidth = PosterCellWidth,
                gridState = gridState,
                header = {
                    fullWidthItem("header") {
                        LargeHeader("Galleries") { SettingsMenuButton(navigator) }
                    }
                    fullWidthItem("sources") {
                        val sites by viewModel.siteFilter.sites.collectAsStateWithLifecycle()
                        val selected by viewModel.siteFilter.selected.collectAsStateWithLifecycle()
                        SiteBadges(sites, selected, onToggle = viewModel.siteFilter::toggle, onAll = viewModel.siteFilter::clear)
                    }
                    fullWidthItem("categories") {
                        SectionHeading("Categories", onClick = navigator::openGalleries)
                    }
                },
            ) { _, gallery ->
                GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
            }
            PinnedTitleBar("Galleries", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
        }
    }
}
