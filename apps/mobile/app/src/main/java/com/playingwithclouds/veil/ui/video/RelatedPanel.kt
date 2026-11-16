package com.playingwithclouds.veil.ui.video

import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.draggable
import androidx.compose.foundation.gestures.rememberDraggableState
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.displayCutoutPadding
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.components.TrackShown
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.glass
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlin.math.roundToInt

/** Width of a card in the panel; narrow enough that a landscape screen shows three and a peek. */
private val PanelCardWidth = 200.dp

/** How far the panel has to be dragged down to close. */
private val DismissDistance = 64.dp

/**
 * Related videos over the fullscreen video: a glass panel rising from the bottom with a row of
 * video cards. Dragging it down, tapping beside it or going back closes it.
 */
@Composable
fun BoxScope.RelatedPanel(
    visible: Boolean,
    related: List<VideoSummary>,
    onDismiss: () -> Unit,
    onOpenVideo: (VideoSummary) -> Unit,
) {
    BackHandler(enabled = visible, onBack = onDismiss)
    AnimatedVisibility(visible, enter = fadeIn(), exit = fadeOut(), modifier = Modifier.fillMaxSize()) {
        Box(
            Modifier
                .fillMaxSize()
                .clickable(interactionSource = remember { MutableInteractionSource() }, indication = null, onClick = onDismiss),
        )
    }
    AnimatedVisibility(
        visible,
        enter = slideInVertically { height -> height } + fadeIn(),
        exit = slideOutVertically { height -> height } + fadeOut(),
        modifier = Modifier.align(Alignment.BottomCenter),
    ) {
        PanelContent(related, onDismiss, onOpenVideo)
    }
}

/** The glass panel itself, following the finger while dragged down. */
@Composable
private fun PanelContent(related: List<VideoSummary>, onDismiss: () -> Unit, onOpenVideo: (VideoSummary) -> Unit) {
    var dragOffset by remember { mutableFloatStateOf(0f) }
    val dismissPixels = with(LocalDensity.current) { DismissDistance.toPx() }
    val dragState = rememberDraggableState { delta -> dragOffset = maxOf(0f, dragOffset + delta) }
    Column(
        Modifier
            .fillMaxWidth()
            .displayCutoutPadding()
            .padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.medium)
            .offset { IntOffset(0, dragOffset.roundToInt()) }
            .glass(VeilShapes.sheet)
            .draggable(
                dragState,
                Orientation.Vertical,
                onDragStopped = {
                    if (isPanelDismissSwipe(dragOffset, dismissPixels)) {
                        onDismiss()
                    }
                    dragOffset = 0f
                },
            )
            .padding(bottom = VeilSpacing.small),
    ) {
        // SectionHeading brings its own space above and below.
        SectionHeading("Related", Modifier.padding(horizontal = VeilSpacing.large))
        val rowState = rememberLazyListState()
        // The page's Up next list logs the same videos on this surface, at the same positions.
        TrackShown(rowState, related.map { video -> video.id }) { position ->
            AppServices.impressions.recordShown(related[position].id, RecordKind.video, ImpressionSurface.RELATED, position)
        }
        LazyRow(
            state = rowState,
            contentPadding = PaddingValues(horizontal = VeilSpacing.large),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        ) {
            itemsIndexed(related, key = { _, video -> video.id }) { position, video ->
                VideoCard(
                    video,
                    onClick = {
                        AppServices.impressions.recordClicked(video.id, RecordKind.video, ImpressionSurface.RELATED, position)
                        onOpenVideo(video)
                    },
                    modifier = Modifier.width(PanelCardWidth),
                )
            }
        }
    }
}
