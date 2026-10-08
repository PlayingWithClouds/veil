package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.LazyGridLayoutInfo
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SceneQuickActionsSheet
import com.playingwithclouds.veil.ui.components.SceneShelf
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.SiteBadges
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.collectLatest

/** Header rows before the feed in the grid; impressions and previews count feed positions only. */
private const val HEADER_ROWS = 4

/** How long the feed has to rest before the card in focus starts its preview. */
private const val PREVIEW_SETTLE_MILLISECONDS = 250L

/**
 * The Home tab: site badges, the continue-watching shelf and the personalized feed, with
 * impression logging and a preview playing on the card in focus. Pull down to re-rank. The large
 * header scrolls away and a slim title bar takes its place; a long press on a scene opens its
 * quick actions.
 */
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val gridState = rememberLazyGridState()
    val headerScrolledAway by remember { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }
    var quickActionsFor by remember { mutableStateOf<SceneSummary?>(null) }

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.refreshProgress() }

    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            VideoFeed(viewModel, gridState, navigator, onLongPress = { scene -> quickActionsFor = scene })
            PinnedTitleBar("Home", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
        }
    }

    val selected = quickActionsFor
    if (selected != null) {
        SceneQuickActionsSheet(
            scene = selected,
            onDismiss = { quickActionsFor = null },
            onNotInterested = { viewModel.dropFromFeed(selected.id) },
        )
    }
}

/** The continue-watching shelf and the recommendation feed, logging impressions and previewing the card in focus. */
@Composable
private fun VideoFeed(
    viewModel: HomeViewModel,
    gridState: LazyGridState,
    navigator: AppNavigator,
    onLongPress: (SceneSummary) -> Unit,
) {
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val continueWatching by viewModel.continueWatching.collectAsStateWithLifecycle()
    var previewSceneId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(gridState) {
        snapshotFlow { gridState.layoutInfo }.collect { layoutInfo ->
            for (index in feedIndicesShown(layoutInfo)) {
                val item = state.items.getOrNull(index) ?: continue
                viewModel.recordShown(item, index)
            }
        }
    }
    LaunchedEffect(gridState) {
        snapshotFlow { feedIndexInFocus(gridState.layoutInfo) }.collectLatest { index ->
            delay(PREVIEW_SETTLE_MILLISECONDS)
            previewSceneId = index?.let { position -> state.items.getOrNull(position)?.scene?.id }
        }
    }

    PagedGrid(
        paged = viewModel.feed,
        keyOf = { item -> item.scene.id },
        emptyText = "Nothing to recommend yet. Search for something to get started.",
        gridState = gridState,
        header = {
            fullWidthItem("header") {
                LargeHeader("Home") { SettingsMenuButton(navigator) }
            }
            fullWidthItem("sources") {
                val sites by viewModel.siteFilter.sites.collectAsStateWithLifecycle()
                val selected by viewModel.siteFilter.selected.collectAsStateWithLifecycle()
                SiteBadges(sites, selected, onToggle = viewModel.siteFilter::toggle, onAll = viewModel.siteFilter::clear)
            }
            fullWidthItem("continue") {
                SceneShelf(
                    title = "Continue watching",
                    scenes = continueWatching,
                    onOpen = { scene -> navigator.openScene(scene.id) },
                    onSeeAll = navigator::openHistory,
                    onLongPress = onLongPress,
                )
            }
            fullWidthItem("for-you") {
                if (continueWatching.isNotEmpty()) {
                    SectionHeading("For you")
                }
            }
        },
    ) { position, item ->
        SceneCard(
            scene = item.scene,
            onClick = {
                viewModel.recordClicked(item, position)
                navigator.openScene(item.scene.id)
            },
            onLongClick = { onLongPress(item.scene) },
            previewing = item.scene.id == previewSceneId,
        )
    }
}

/** Feed positions (not grid indices) of the cards on screen enough to count as shown. */
private fun feedIndicesShown(layoutInfo: LazyGridLayoutInfo): List<Int> {
    return VisibleImpressions.shownIndices(feedExtents(layoutInfo), layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}

/** Feed position of the card whose preview should play, or null. */
private fun feedIndexInFocus(layoutInfo: LazyGridLayoutInfo): Int? {
    val extents = feedExtents(layoutInfo).filter { extent -> extent.index >= 0 }
    return VisibleImpressions.previewIndex(extents, layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}

/** The laid-out grid items as feed positions; header rows get negative positions. */
private fun feedExtents(layoutInfo: LazyGridLayoutInfo): List<ItemExtent> {
    return layoutInfo.visibleItemsInfo.map { item -> ItemExtent(item.index - HEADER_ROWS, item.offset.y, item.size.height) }
}
