package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GalleryCategory
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** The gallery site's category index. */
class GalleriesViewModel : ViewModel() {

    private val mutableCategories = MutableStateFlow<LoadState<List<GalleryCategory>>>(LoadState.Loading)

    /** The categories. */
    val categories: StateFlow<LoadState<List<GalleryCategory>>> = mutableCategories

    init {
        load()
    }

    /** Loads the categories. */
    fun load() {
        viewModelScope.loadInto(mutableCategories) { GalleryRepository.categories() }
    }
}

/**
 * The Galleries tab: the gallery site's categories as image tiles, each opening its galleries. The
 * large header scrolls away and a slim title bar takes its place.
 */
@Composable
fun GalleriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { GalleriesViewModel() }
    val categories by viewModel.categories.collectAsStateWithLifecycle()
    val gridState = rememberLazyGridState()
    val headerScrolledAway by remember { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }
    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            LoadStateContent(categories, onRetry = viewModel::load) { list ->
                LazyVerticalGrid(
                    columns = GridCells.Adaptive(150.dp),
                    state = gridState,
                    contentPadding = PaddingValues(
                        start = VeilSpacing.gutter,
                        end = VeilSpacing.gutter,
                        bottom = VeilSpacing.large + LocalFloatingBarInset.current,
                    ),
                    horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
                    verticalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
                ) {
                    item(key = "header", span = { GridItemSpan(maxLineSpan) }) {
                        LargeHeader("Galleries") { SettingsMenuButton(navigator) }
                    }
                    items(list, key = { category -> category.id }) { category ->
                        CategoryTile(category, onClick = { navigator.openGalleryCategory(category.name) })
                    }
                }
            }
            PinnedTitleBar("Galleries", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
        }
    }
}

/** A category's thumbnail with its name over it. */
@Composable
private fun CategoryTile(category: GalleryCategory, onClick: () -> Unit) {
    Box(
        Modifier.fillMaxWidth().aspectRatio(4f / 3f).pressClickable(onClick).clip(VeilShapes.card),
    ) {
        RemoteImage(category.poster, Modifier.matchParentSize())
        Box(Modifier.matchParentSize().background(Brush.verticalGradient(0.5f to Color.Transparent, 1f to VeilColors.imageLabel)))
        Text(
            category.name,
            modifier = Modifier.align(Alignment.BottomStart).padding(VeilSpacing.medium),
            color = VeilColors.content,
            style = MaterialTheme.typography.titleSmall,
        )
    }
}
