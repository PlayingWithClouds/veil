package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyGridScope
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.size
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.material3.pulltorefresh.PullToRefreshDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshState
import androidx.compose.material3.pulltorefresh.rememberPullToRefreshState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.paging.PagedList

/** How many items before the end the next page starts loading. */
private const val PREFETCH_DISTANCE = 6

/** Adaptive cell width for landscape scene cards: one column on phones, more on tablets. */
val SceneCellWidth = 300.dp

/** Adaptive cell width for round avatar cards. */
val AvatarCellWidth = 112.dp

/** Adaptive cell width for poster cards (galleries). */
val PosterCellWidth = 150.dp

/**
 * An endless grid fed by a [PagedList]: loads the next page near the end, shows the empty and
 * failed states, and supports pull-to-refresh.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun <T> PagedGrid(
    paged: PagedList<T>,
    keyOf: (T) -> Any,
    emptyText: String,
    modifier: Modifier = Modifier,
    cellWidth: Dp = SceneCellWidth,
    gridState: LazyGridState = rememberLazyGridState(),
    contentPadding: PaddingValues = PaddingValues(
        start = VeilSpacing.gutter,
        top = VeilSpacing.small,
        end = VeilSpacing.gutter,
        bottom = VeilSpacing.large + LocalFloatingBarInset.current,
    ),
    header: (LazyGridScope.() -> Unit)? = null,
    itemContent: @Composable (index: Int, item: T) -> Unit,
) {
    val state by paged.state.collectAsStateWithLifecycle()
    LoadMoreEffect(gridState, state.items.size) { paged.loadMore() }
    val refreshState = rememberPullToRefreshState()
    PullToRefreshBox(
        isRefreshing = state.isRefreshing,
        onRefresh = { paged.refresh() },
        modifier = modifier.fillMaxSize(),
        state = refreshState,
        indicator = { RefreshIndicator(refreshState, state.isRefreshing, Modifier.align(Alignment.TopCenter)) },
    ) {
        LazyVerticalGrid(
            columns = GridCells.Adaptive(cellWidth),
            state = gridState,
            contentPadding = contentPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            modifier = Modifier.fillMaxSize(),
        ) {
            header?.invoke(this)
            itemsIndexed(state.items, key = { _, item -> keyOf(item) }) { index, item ->
                Box(Modifier.animateItem()) { itemContent(index, item) }
            }
            item(span = { GridItemSpan(maxLineSpan) }) {
                PagedFooter(
                    isLoading = state.isLoading && !state.isRefreshing,
                    error = state.error,
                    isEmpty = state.hasLoaded && state.items.isEmpty() && !state.isLoading,
                    emptyText = emptyText,
                    onRetry = { paged.loadMore() },
                )
            }
        }
    }
}

/** The pull-to-refresh spinner in a dark circle: follows the pull, then holds while refreshing. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RefreshIndicator(refreshState: PullToRefreshState, isRefreshing: Boolean, modifier: Modifier) {
    val threshold = with(LocalDensity.current) { PullToRefreshDefaults.PositionalThreshold.toPx() }
    Box(
        modifier
            .graphicsLayer {
                var fraction = refreshState.distanceFraction.coerceIn(0f, 1f)
                if (isRefreshing) {
                    fraction = 1f
                }
                translationY = fraction * threshold - size.height
                alpha = fraction
            }
            .size(40.dp)
            .clip(VeilShapes.capsule)
            .background(VeilColors.surfaceHigh),
        contentAlignment = Alignment.Center,
    ) {
        // Only spin while shown, so an idle grid does not draw frames.
        val shown by remember { derivedStateOf { refreshState.distanceFraction > 0f } }
        if (shown || isRefreshing) {
            Spinner(size = 20.dp)
        }
    }
}

/** Starts the next page when the grid is scrolled near its last item. */
@Composable
private fun LoadMoreEffect(gridState: LazyGridState, itemCount: Int, loadMore: () -> Unit) {
    LaunchedEffect(gridState, itemCount) {
        snapshotFlow { gridState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0 }
            .collect { lastVisibleIndex ->
                if (lastVisibleIndex >= itemCount - PREFETCH_DISTANCE) {
                    loadMore()
                }
            }
    }
}

/** The line after the last item: a spinner while loading, the empty note, or the failure with retry. */
@Composable
private fun PagedFooter(isLoading: Boolean, error: String?, isEmpty: Boolean, emptyText: String, onRetry: () -> Unit) {
    if (isLoading) {
        Box(Modifier.fillMaxWidth().padding(VeilSpacing.large), contentAlignment = Alignment.Center) {
            Spinner()
        }
        return
    }
    if (error != null) {
        Box(Modifier.fillMaxWidth().padding(VeilSpacing.large), contentAlignment = Alignment.Center) {
            TextAction("$error. Tap to retry", onClick = onRetry, color = VeilColors.error)
        }
        return
    }
    if (isEmpty) {
        EmptyMessage(emptyText)
    }
}

/** A grid item that spans every column: headers and section rows, stacked top to bottom. */
fun LazyGridScope.fullWidthItem(key: Any? = null, content: @Composable () -> Unit) {
    item(key = key, span = { GridItemSpan(maxLineSpan) }) { Column { content() } }
}
