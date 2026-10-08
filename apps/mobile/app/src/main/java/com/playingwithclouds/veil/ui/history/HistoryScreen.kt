package com.playingwithclouds.veil.ui.history

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.WatchHistoryEntry
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.launch

/** The watch history, most recently watched first. */
class HistoryViewModel : ViewModel() {

    /** Watched scenes; rows whose scene is gone are skipped. */
    val history = PagedList<WatchHistoryEntry>(viewModelScope, PAGE_SIZE, { entry -> entry.mediaId }) { offset ->
        FeedRepository.watchHistory(PAGE_SIZE, offset).filter { entry -> entry.scene != null }
    }

    init {
        history.loadMore()
    }

    /** Removes a scene from the history. */
    fun remove(entry: WatchHistoryEntry) {
        history.update { entries -> entries.filter { other -> other.mediaId != entry.mediaId } }
        viewModelScope.launch {
            runCatching { FeedRepository.deleteWatchHistory(entry.mediaId) }
            WatchProgressStore.refresh()
        }
    }

    companion object {
        private const val PAGE_SIZE = 40
    }
}

/** Everything watched, most recent first. */
@Composable
fun HistoryScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HistoryViewModel() }
    LaunchedEffect(Unit) { WatchProgressStore.refresh() }
    Scaffold(topBar = { VeilTopBar("History", onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.history,
            keyOf = { entry -> entry.mediaId },
            emptyText = "Nothing watched yet.",
            modifier = Modifier.padding(padding),
        ) { _, entry ->
            val scene = entry.scene ?: return@PagedGrid
            Box {
                SceneCard(scene, onClick = { navigator.openScene(scene.id) })
                IconButton(onClick = { viewModel.remove(entry) }, modifier = Modifier.align(Alignment.TopEnd)) {
                    Icon(Icons.Filled.Close, contentDescription = "Remove from history")
                }
            }
        }
    }
}
