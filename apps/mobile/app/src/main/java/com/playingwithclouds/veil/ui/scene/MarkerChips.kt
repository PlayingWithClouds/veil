package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Modifier
import androidx.media3.common.Player
import androidx.media3.ui.compose.state.rememberProgressStateWithTickInterval
import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Media-time step of the position the active chip follows. */
private const val CHIP_TICK_MILLISECONDS = 1_000L

private const val MILLISECONDS_PER_SECOND = 1000.0

/**
 * One chip per marker in play order, the one playback is in highlighted and scrolled into view; a
 * tap seeks the [player] there. The row is lazy and scrolls sideways, so it stays cheap with the
 * hundreds of markers a scene-detection model can produce.
 */
@Composable
fun MarkerChips(markers: List<SceneMarker>, player: Player, modifier: Modifier = Modifier) {
    if (markers.isEmpty()) {
        return
    }
    val ordered = markersInPlayOrder(markers)
    val progress = rememberProgressStateWithTickInterval(player, CHIP_TICK_MILLISECONDS)
    val activeIndex = activeMarkerIndex(ordered, progress.currentPositionMs / MILLISECONDS_PER_SECOND)
    val listState = rememberLazyListState()
    LaunchedEffect(activeIndex) {
        if (activeIndex >= 0) {
            listState.animateScrollToItem(activeIndex)
        }
    }
    LazyRow(
        modifier.fillMaxWidth(),
        state = listState,
        contentPadding = PaddingValues(horizontal = VeilSpacing.gutter),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        itemsIndexed(ordered, key = { _, marker -> marker.id }) { index, marker ->
            Pill(
                markerChipLabel(marker),
                onClick = { player.seekTo((marker.seconds * MILLISECONDS_PER_SECOND).toLong()) },
                selected = index == activeIndex,
            )
        }
    }
}
