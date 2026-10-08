package com.playingwithclouds.veil.ui.galleries

import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.api.LiveSearch
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.LiveResultItem
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.toList

/** A category's galleries, streamed from the gallery site page by page. */
class GalleryCategoryViewModel(private val category: String) : ViewModel() {

    /** The category feed; it ends when a page brings nothing new. */
    val galleries = PagedList(viewModelScope, PAGE_SIZE, { gallery: GallerySummary -> gallery.id }, { page -> page.isEmpty() }) { offset ->
        LiveSearch.results(category.lowercase(), listOf(GalleryRepository.GALLERY_PLUGIN), offset, PAGE_SIZE)
            .toList()
            .mapNotNull { hit -> SearchRepository.toLiveResult(hit) }
            .filterIsInstance<LiveResultItem.GalleryResult>()
            .map { result -> result.gallery }
    }

    init {
        galleries.loadMore()
    }

    companion object {
        private const val PAGE_SIZE = 24
    }
}

/** The galleries of one category. */
@Composable
fun GalleryCategoryScreen(name: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "gallery-category-$name") { GalleryCategoryViewModel(name) }
    Scaffold(topBar = { VeilTopBar(name, onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.galleries,
            keyOf = { gallery -> gallery.id },
            emptyText = "No galleries found.",
            modifier = Modifier.padding(padding),
            cellWidth = PosterCellWidth,
        ) { _, gallery ->
            GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
        }
    }
}
