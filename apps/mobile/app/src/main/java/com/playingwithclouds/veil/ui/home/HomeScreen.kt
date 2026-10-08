package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.Modifier
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.VeilTopBar

/** The Home tab: the personalized feed, with impression logging and refresh. */
@Composable
fun HomeScreen(navigator: AppNavigator, onMenu: () -> Unit) {
    val viewModel = viewModel { HomeViewModel() }
    val gridState = rememberLazyGridState()
    val state by viewModel.feed.state.collectAsStateWithLifecycle()

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.refreshProgress() }
    LaunchedEffect(gridState) {
        snapshotFlow { gridState.layoutInfo }.collect { layoutInfo ->
            val extents = layoutInfo.visibleItemsInfo.map { item ->
                ItemExtent(item.index, item.offset.y, item.size.height)
            }
            val shown = VisibleImpressions.shownIndices(extents, layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
            for (index in shown) {
                val item = state.items.getOrNull(index) ?: continue
                viewModel.recordShown(item, index)
            }
        }
    }

    Scaffold(
        topBar = {
            VeilTopBar(
                title = "Veil",
                onMenu = onMenu,
                actions = {
                    IconButton(onClick = { viewModel.feed.refresh() }) {
                        Icon(Icons.Filled.Refresh, contentDescription = "Refresh")
                    }
                    IconButton(onClick = { navigator.openSearch() }) {
                        Icon(Icons.Filled.Search, contentDescription = "Search")
                    }
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
