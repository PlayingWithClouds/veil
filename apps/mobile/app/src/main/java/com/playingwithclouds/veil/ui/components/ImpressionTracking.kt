package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.lazy.LazyListLayoutInfo
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.snapshotFlow
import com.playingwithclouds.veil.ui.home.ItemExtent
import com.playingwithclouds.veil.ui.home.VisibleImpressions

/**
 * Calls [onShown] with the position of every item that comes at least half into view in a lazy
 * list or row. [keys] are the list keys of the tracked items in order, so a list holding other
 * items too (headings, other rows) tracks only these; an item's position is its index in [keys].
 * [onShown] may be called again for the same item; the impression logger counts it once.
 */
@Composable
fun TrackShown(state: LazyListState, keys: List<Any>, onShown: (position: Int) -> Unit) {
    val currentOnShown by rememberUpdatedState(onShown)
    LaunchedEffect(state, keys) {
        val positionByKey = keys.withIndex().associate { (position, key) -> key to position }
        snapshotFlow { shownPositions(state.layoutInfo, positionByKey) }.collect { positions ->
            for (position in positions) {
                currentOnShown(position)
            }
        }
    }
}

/** The positions of the tracked items laid out and mostly inside the viewport. */
private fun shownPositions(layoutInfo: LazyListLayoutInfo, positionByKey: Map<Any, Int>): List<Int> {
    val extents = layoutInfo.visibleItemsInfo.mapNotNull { item ->
        val position = positionByKey[item.key] ?: return@mapNotNull null
        ItemExtent(position, item.offset, item.size)
    }
    return VisibleImpressions.shownIndices(extents, layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}
