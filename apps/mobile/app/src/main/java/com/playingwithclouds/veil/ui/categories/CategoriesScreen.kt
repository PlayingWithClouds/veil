package com.playingwithclouds.veil.ui.categories

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CategoryGroup
import com.playingwithclouds.veil.data.CategoryRepository
import com.playingwithclouds.veil.data.CategoryTile
import com.playingwithclouds.veil.data.VideoScope
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.paging.PagedState
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/** Columns of the category tiles. */
private const val TILE_COLUMNS = 3

/** Videos per page of a category or tag. */
private const val VIDEO_PAGE = 24

/** The list asks for more this close to its end. */
private const val LOAD_MORE_DISTANCE = 5

/** Our categories, grouped, with how many videos each holds. */
class CategoriesViewModel : ViewModel() {

    private val groupsState = MutableStateFlow<LoadState<List<CategoryGroup>>>(LoadState.Loading)
    val groups: StateFlow<LoadState<List<CategoryGroup>>> = groupsState.asStateFlow()

    init {
        load()
    }

    /** Loads the categories. */
    fun load() {
        viewModelScope.loadInto(groupsState) { CategoryRepository.groups() }
    }
}

/** Categories, as the web shows them: a heading per group, its categories as picture tiles; empty ones hide. */
@Composable
fun CategoriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { CategoriesViewModel() }
    val state by viewModel.groups.collectAsStateWithLifecycle()
    DetailScreen("Categories", onBack = navigator::back) {
        val barInset = LocalTopBarInset.current
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(top = barInset)) { groups ->
            val shown = groups.map { group -> group.copy(children = group.children.filter { tile -> tile.count > 0 }) }
                .filter { group -> group.children.isNotEmpty() || group.parent.count > 0 }
            if (shown.isEmpty()) {
                EmptyMessage("No category holds any videos yet.", Modifier.padding(top = barInset))
                return@LoadStateContent
            }
            LazyVerticalGrid(
                GridCells.Fixed(TILE_COLUMNS),
                contentPadding = PaddingValues(start = VeilSpacing.gutter, top = barInset, end = VeilSpacing.gutter, bottom = VeilSpacing.huge),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
            ) {
                for (group in shown) {
                    item(key = "group-${group.parent.id}", span = { GridItemSpan(maxLineSpan) }) {
                        SectionHeading(group.parent.name, onClick = { navigator.openCategory(group.parent.id, group.parent.name) })
                    }
                    items(group.children, key = { tile -> tile.id }) { tile ->
                        CategoryTileCard(tile, onClick = { navigator.openCategory(tile.id, tile.name) })
                    }
                }
            }
        }
    }
}

/** A category as a tile: its picture, name and how many videos it holds. */
@Composable
private fun CategoryTileCard(tile: CategoryTile, onClick: () -> Unit) {
    Column(Modifier.fillMaxWidth().pressClickable(onClick)) {
        RemoteImage(tile.pictureUrl, Modifier.fillMaxWidth().aspectRatio(1f).clip(VeilShapes.card))
        Text(tile.name, Modifier.padding(top = VeilSpacing.extraSmall), style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
        Text("${tile.count}", style = VeilType.cardMeta, color = VeilColors.contentMuted)
    }
}

/** The videos in a category or with a tag, newest first. */
class ScopedVideosViewModel(private val scope: VideoScope) : ViewModel() {

    val videos = PagedList<VideoSummary>(viewModelScope, VIDEO_PAGE, { video -> video.id }) { offset ->
        CategoryRepository.videos(scope, VIDEO_PAGE, offset)
    }

    init {
        videos.loadMore()
    }
}

/** A category's or tag's videos as full-width cards, growing as it's scrolled to its end. */
@Composable
fun ScopedVideosScreen(scope: VideoScope, title: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "videos-$scope") { ScopedVideosViewModel(scope) }
    val state by viewModel.videos.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    LaunchedEffect(listState) {
        snapshotFlow {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisible >= listState.layoutInfo.totalItemsCount - LOAD_MORE_DISTANCE
        }.collect { nearEnd ->
            if (nearEnd) {
                viewModel.videos.loadMore()
            }
        }
    }
    DetailScreen(title, onBack = navigator::back) {
        ScopedVideoList(state, listState, navigator)
    }
}

/** The scope's videos under the bar, or why there are none. */
@Composable
private fun ScopedVideoList(state: PagedState<VideoSummary>, listState: LazyListState, navigator: AppNavigator) {
    val barInset = LocalTopBarInset.current
    if (state.hasLoaded && state.items.isEmpty()) {
        EmptyMessage(state.error ?: "No videos here yet.", Modifier.padding(top = barInset))
        return
    }
    LazyColumn(state = listState, contentPadding = PaddingValues(top = barInset, bottom = VeilSpacing.huge)) {
        items(state.items, key = { video -> video.id }) { video ->
            VideoCard(video, onClick = { navigator.openVideo(video.id) }, edgeToEdge = true)
        }
        if (state.isLoading) {
            item("loading") { Row(Modifier.fillMaxWidth().padding(VeilSpacing.large), horizontalArrangement = Arrangement.Center) { Spinner() } }
        }
    }
}
