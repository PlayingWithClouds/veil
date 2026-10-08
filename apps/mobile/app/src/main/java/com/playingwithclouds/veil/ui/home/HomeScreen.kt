package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.AssistChip
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
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
import com.playingwithclouds.veil.ui.components.ScrollingTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem

/** How many tag chips the topic row offers. */
private const val TOPIC_COUNT = 12

/** Header rows before the feed in the grid; impressions count feed positions only. */
private const val HEADER_ROWS = 3

/**
 * The Home tab: topic chips from the feed, the continue-watching shelf, then the personalized
 * feed, with impression logging and refresh. The top bar slides away while scrolling down.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val gridState = rememberLazyGridState()
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val continueWatching by viewModel.continueWatching.collectAsStateWithLifecycle()
    val topics = remember(state.items) { topTags(state.items) }
    val scrollBehavior = TopAppBarDefaults.enterAlwaysScrollBehavior()

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

    Scaffold(
        modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection),
        topBar = {
            ScrollingTopBar(
                title = "Veil",
                scrollBehavior = scrollBehavior,
                actions = {
                    IconButton(onClick = { viewModel.feed.refresh() }) {
                        Icon(Icons.Filled.Refresh, contentDescription = "Refresh")
                    }
                    SettingsMenuButton(navigator)
                },
            )
        },
    ) { padding ->
        PagedGrid(
            paged = viewModel.feed,
            keyOf = { item -> item.scene.id },
            emptyText = "Nothing to recommend yet. Search for something to get started.",
            modifier = Modifier.padding(padding),
            gridState = gridState,
            header = {
                fullWidthItem("topics") { TopicChips(topics, onOpen = { tag -> navigator.openTag(tag.id) }) }
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

/** A scrolling row of tag chips; each opens its tag. */
@Composable
private fun TopicChips(tags: List<EntityRef>, onOpen: (EntityRef) -> Unit) {
    if (tags.isEmpty()) {
        return
    }
    LazyRow(contentPadding = PaddingValues(horizontal = 4.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        items(tags, key = { tag -> tag.id }) { tag ->
            AssistChip(onClick = { onOpen(tag) }, label = { Text(tag.name) })
        }
    }
}

/** The tags most common across the loaded feed, most frequent first. */
private fun topTags(items: List<RecommendedScene>): List<EntityRef> {
    val tags = items.flatMap { item -> item.scene.tags }
    val counts = tags.groupingBy { tag -> tag.id }.eachCount()
    return tags.distinctBy { tag -> tag.id }
        .sortedByDescending { tag -> counts[tag.id] }
        .take(TOPIC_COUNT)
}
