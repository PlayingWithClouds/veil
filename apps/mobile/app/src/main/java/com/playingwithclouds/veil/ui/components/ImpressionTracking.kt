package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.snapshotFlow
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.ui.home.ItemExtent
import com.playingwithclouds.veil.ui.home.VisibleImpressions

/**
 * Logs a "shown" impression for each scene of a vertical grid once at least half of its card is on
 * screen. A scene's position is its index in [sceneIds], the order the user sees; the grid's items
 * must be keyed by scene id. [sources] holds the candidate source of scenes a ranking served.
 * Impressions of [surface] are forgotten when the grid enters the composition, so coming back to a
 * screen logs what is shown again; set [forgetShown] to false for a second list on a surface that
 * is already logging.
 */
@Composable
fun LogShownScenes(gridState: LazyGridState, surface: String, sceneIds: List<String>, sources: Map<String, String> = emptyMap(), forgetShown: Boolean = true) {
    val latestIds = rememberUpdatedState(sceneIds)
    val latestSources = rememberUpdatedState(sources)
    LaunchedEffect(gridState, surface) {
        if (forgetShown) {
            AppServices.impressions.resetShown(surface)
        }
        snapshotFlow {
            val positionBySceneId = positionsOf(latestIds.value)
            val extents = gridState.layoutInfo.visibleItemsInfo.mapNotNull { cell ->
                positionBySceneId[cell.key]?.let { position -> ItemExtent(position, cell.offset.y, cell.size.height) }
            }
            VisibleImpressions.shownIndices(extents, gridState.layoutInfo.viewportStartOffset, gridState.layoutInfo.viewportEndOffset)
        }.collect { positions ->
            logShown(positions, latestIds.value, latestSources.value, surface)
        }
    }
}

/**
 * [LogShownScenes] for a column or row of scenes: a LazyColumn keyed by scene id, or a LazyRow
 * such as a shelf (its extents run along the scroll axis, whichever it is).
 */
@Composable
fun LogShownScenes(listState: LazyListState, surface: String, sceneIds: List<String>, sources: Map<String, String> = emptyMap(), forgetShown: Boolean = true) {
    val latestIds = rememberUpdatedState(sceneIds)
    val latestSources = rememberUpdatedState(sources)
    LaunchedEffect(listState, surface) {
        if (forgetShown) {
            AppServices.impressions.resetShown(surface)
        }
        snapshotFlow {
            val positionBySceneId = positionsOf(latestIds.value)
            val extents = listState.layoutInfo.visibleItemsInfo.mapNotNull { cell ->
                positionBySceneId[cell.key]?.let { position -> ItemExtent(position, cell.offset, cell.size) }
            }
            VisibleImpressions.shownIndices(extents, listState.layoutInfo.viewportStartOffset, listState.layoutInfo.viewportEndOffset)
        }.collect { positions ->
            logShown(positions, latestIds.value, latestSources.value, surface)
        }
    }
}

/** Logs that the scene at [position] of a list on [surface] was opened. */
fun logSceneClicked(surface: String, sceneId: String, position: Int, source: String? = null) {
    AppServices.impressions.recordClicked(sceneId, source, surface, position)
}

/** Position of each scene id in the list. */
private fun positionsOf(sceneIds: List<String>): Map<String, Int> {
    return sceneIds.withIndex().associate { (position, sceneId) -> sceneId to position }
}

/** Records the scenes at the given positions as shown. */
private fun logShown(positions: List<Int>, sceneIds: List<String>, sources: Map<String, String>, surface: String) {
    for (position in positions) {
        val sceneId = sceneIds.getOrNull(position) ?: continue
        AppServices.impressions.recordShown(sceneId, sources[sceneId], surface, position)
    }
}
