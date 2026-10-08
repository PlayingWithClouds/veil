package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SceneShelf
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.ShelfHeading
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.PillRow
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons

/** How many tag chips the topic row offers. */
private const val TOPIC_COUNT = 12

/** Header rows before the feed in the grid; impressions count feed positions only. */
private const val HEADER_ROWS = 4

/**
 * The Home tab: topic chips from the feed, the continue-watching shelf, then the personalized
 * feed, with impression logging and refresh. The large header scrolls away with the feed.
 */
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val gridState = rememberLazyGridState()
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val continueWatching by viewModel.continueWatching.collectAsStateWithLifecycle()
    val topics = remember(state.items) { topTags(state.items) }

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.refreshProgress() }
    LaunchedEffect(gridState) {
        snapshotFlow { gridState.layoutInfo }.collect { layoutInfo ->
            val extents = layoutInfo.visibleItemsInfo.map { item ->
                ItemExtent(item.index - HEADER_ROWS, item.offset.y, item.size.height)
            }
            val shown = VisibleImpressions.shownIndices(extents, layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
            for (index in shown) {
                val item = state.items.getOrNull(index) ?: continue
                viewModel.recordShown(item, index)
            }
        }
    }

    Scaffold { padding ->
        PagedGrid(
            paged = viewModel.feed,
            keyOf = { item -> item.scene.id },
            emptyText = "Nothing to recommend yet. Search for something to get started.",
            modifier = Modifier.padding(padding),
            gridState = gridState,
            header = {
                fullWidthItem("header") {
                    LargeHeader("Home") {
                        RoundIconButton(VeilIcons.Refresh, contentDescription = "Refresh", onClick = { viewModel.feed.refresh() })
                        SettingsMenuButton(navigator)
                    }
                }
                fullWidthItem("topics") { TopicPills(topics, onOpen = { tag -> navigator.openTag(tag.id) }) }
                fullWidthItem("continue") {
                    SceneShelf(
                        title = "Continue watching",
                        scenes = continueWatching,
                        onOpen = { scene -> navigator.openScene(scene.id) },
                        onSeeAll = navigator::openHistory,
                        modifier = Modifier.padding(bottom = 8.dp),
                    )
                }
                fullWidthItem("for-you") {
                    if (continueWatching.isNotEmpty()) {
                        ShelfHeading("For you")
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
            )
        }
    }
}

/** A scrolling row of tag pills; each opens its tag. */
@Composable
private fun TopicPills(tags: List<EntityRef>, onOpen: (EntityRef) -> Unit) {
    if (tags.isEmpty()) {
        return
    }
    PillRow(tags, labelOf = { tag -> tag.name }, onClick = onOpen, contentPadding = PaddingValues(horizontal = 4.dp, vertical = 4.dp))
}

/** The tags most common across the loaded feed, most frequent first. */
private fun topTags(items: List<RecommendedScene>): List<EntityRef> {
    val tags = items.flatMap { item -> item.scene.tags }
    val counts = tags.groupingBy { tag -> tag.id }.eachCount()
    return tags.distinctBy { tag -> tag.id }
        .sortedByDescending { tag -> counts[tag.id] }
        .take(TOPIC_COUNT)
}
