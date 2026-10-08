package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
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
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
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

/** Galleries, entered through the site's category tiles. */
@Composable
fun GalleriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { GalleriesViewModel() }
    val categories by viewModel.categories.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Galleries", onBack = navigator::back) }) { padding ->
        LoadStateContent(categories, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { list ->
            LazyVerticalGrid(
                columns = GridCells.Adaptive(150.dp),
                modifier = Modifier.padding(padding),
                contentPadding = PaddingValues(VeilSpacing.gutter),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
            ) {
                items(list, key = { category -> category.id }) { category ->
                    CategoryTile(category, onClick = { navigator.openGalleryCategory(category.name) })
                }
            }
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
