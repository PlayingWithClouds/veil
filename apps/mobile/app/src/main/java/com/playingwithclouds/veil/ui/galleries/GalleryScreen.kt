package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GalleryDetail
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.ImageViewer
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** One gallery; a stub gallery has its page fetched from the origin site on first open. */
class GalleryViewModel(private val galleryId: String) : ViewModel() {

    private val mutableGallery = MutableStateFlow<LoadState<GalleryDetail>>(LoadState.Loading)

    /** The gallery with its images. */
    val gallery: StateFlow<LoadState<GalleryDetail>> = mutableGallery

    init {
        load()
    }

    /** Loads the stored gallery, then fetches its images while it has none. */
    fun load() {
        mutableGallery.value = LoadState.Loading
        viewModelScope.launch {
            try {
                var detail = requireNotNull(GalleryRepository.gallery(galleryId)) { "Gallery not found" }
                mutableGallery.value = LoadState.Loaded(detail)
                if (detail.images.isEmpty()) {
                    detail = GalleryRepository.ensureImages(galleryId) ?: detail
                    mutableGallery.value = LoadState.Loaded(detail)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableGallery.value = LoadState.Failed(error.displayMessage())
            }
        }
    }
}

/** A gallery's page: the images as a grid that opens in a zoomable viewer. */
@Composable
fun GalleryScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "gallery-$id") { GalleryViewModel(id) }
    val gallery by viewModel.gallery.collectAsStateWithLifecycle()
    var viewerIndex by remember { mutableStateOf<Int?>(null) }
    val title = (gallery as? LoadState.Loaded)?.value?.title.orEmpty()

    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }) { padding ->
        LoadStateContent(gallery, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { detail ->
            val viewing = viewerIndex
            if (viewing != null) {
                ImageViewer(detail.images.map { image -> image.filePath }, viewing, onDismiss = { viewerIndex = null })
            }
            ImageGrid(detail, navigator, Modifier.padding(padding), onOpenImage = { index -> viewerIndex = index })
        }
    }
}

/** The images in a grid, below the gallery's tags and description. */
@Composable
private fun ImageGrid(detail: GalleryDetail, navigator: AppNavigator, modifier: Modifier, onOpenImage: (Int) -> Unit) {
    LazyVerticalGrid(
        columns = GridCells.Adaptive(110.dp),
        modifier = modifier,
        contentPadding = PaddingValues(8.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        fullWidthItem("about") {
            if (!detail.details.isNullOrBlank()) {
                Text(detail.details, Modifier.padding(8.dp), color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            TagChips(detail.tags, onClick = { tag -> navigator.openTag(tag.id) })
        }
        if (detail.images.isEmpty()) {
            fullWidthItem("empty") { EmptyMessage("This gallery has no images yet.") }
        }
        itemsIndexed(detail.images, key = { _, image -> image.id }) { index, image ->
            RemoteImage(
                image.filePath,
                Modifier.fillMaxWidth().aspectRatio(1f).clip(RoundedCornerShape(8.dp)).clickable { onOpenImage(index) },
            )
        }
    }
}
