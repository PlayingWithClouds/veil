package com.playingwithclouds.veil.ui.video

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import com.playingwithclouds.veil.data.PlayQueue
import com.playingwithclouds.veil.data.RepeatMode
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.launch

/** Width of a queue row's cover. */
private val QueueCoverWidth = 96.dp

private const val COVER_ASPECT_RATIO = 16f / 9f

/** Size of the bar's icons. */
private val BarIconSize = 20.dp

/**
 * The queue as a music app shows it, in a sheet over the page (and over fullscreen): shuffle and
 * repeat, Compilation over the queue, then the video playing (on top when it isn't queued) and the
 * queue in the order it plays, the playing one highlighted. A tap plays a video; × takes it out;
 * the handle drags it elsewhere (not while shuffled, whose order is the phone's own).
 */
@Composable
fun QueueSheet(
    playing: VideoSummary?,
    playingId: String,
    onPlay: (String) -> Unit,
    onCompilation: (() -> Unit)?,
    onDismiss: () -> Unit,
) {
    val scope = rememberCoroutineScope()
    val contents by PlayQueue.contents.collectAsState()
    val shuffle by PlayQueue.shuffle.collectAsState()
    val repeat by PlayQueue.repeat.collectAsState()
    val shuffledIds by PlayQueue.shuffledIds.collectAsState()
    LaunchedEffect(Unit) { runCatching { PlayQueue.refresh() } }
    // Read so the order recomposes when the queue or the shuffled order changes.
    val order = remember(contents, shuffle, shuffledIds) { PlayQueue.playOrder() }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(Modifier.verticalScroll(rememberScrollState()).padding(bottom = VeilSpacing.huge)) {
            QueueHeader(order.size, shuffle, repeat, playingId, onCompilation)
            if (playing != null && order.none { video -> video.id == playingId }) {
                QueueRow(playing, highlighted = true, dragHandle = null, onClick = {}, onRemove = null)
            }
            if (order.isEmpty()) {
                Text(
                    "The queue is empty. Long press a video and pick Play next to queue it.",
                    Modifier.gutterPadding().padding(vertical = VeilSpacing.large),
                    style = MaterialTheme.typography.bodyMedium,
                    color = VeilColors.contentMuted,
                )
            }
            ReorderableQueue(
                videos = order,
                playingId = playingId,
                reorderable = !shuffle,
                onPlay = onPlay,
                onRemove = { videoId -> scope.launch { runCatching { PlayQueue.remove(videoId) } } },
                onMove = { videoId, position -> scope.launch { runCatching { PlayQueue.move(videoId, position) } } },
            )
        }
    }
}

/** "Queue · 5", then shuffle, repeat and Compilation as chips. */
@Composable
private fun QueueHeader(count: Int, shuffle: Boolean, repeat: RepeatMode, playingId: String, onCompilation: (() -> Unit)?) {
    Column(Modifier.gutterPadding().padding(bottom = VeilSpacing.medium), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text("Queue · $count", style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            FeedChip("Shuffle", onClick = { PlayQueue.setShuffle(!shuffle, playingId) }, selected = shuffle, icon = VeilIcons.Random)
            FeedChip(repeat.label, onClick = PlayQueue::cycleRepeat, selected = repeat != RepeatMode.OFF, icon = repeatIcon(repeat))
            if (onCompilation != null) {
                FeedChip("Compilation", onClick = onCompilation, icon = VeilIcons.Scenes)
            }
        }
    }
}

/** The repeat button's icon: a plain loop, or the one with a 1 on repeat one. */
fun repeatIcon(repeat: RepeatMode): ImageVector {
    if (repeat == RepeatMode.ONE) {
        return VeilIcons.RepeatOne
    }
    return VeilIcons.Repeat
}

/**
 * The queue's rows; with [reorderable], each row's handle drags it up or down, the others making
 * room as it passes their middle, and the drop moves it in the queue.
 */
@Composable
private fun ReorderableQueue(
    videos: List<VideoSummary>,
    playingId: String,
    reorderable: Boolean,
    onPlay: (String) -> Unit,
    onRemove: (String) -> Unit,
    onMove: (String, Int) -> Unit,
) {
    var shown by remember(videos) { mutableStateOf(videos) }
    var draggingId by remember { mutableStateOf<String?>(null) }
    var dragOffset by remember { mutableFloatStateOf(0f) }
    var rowHeight by remember { mutableIntStateOf(0) }
    Column {
        for (video in shown) {
            key(video.id) {
                val dragged = draggingId == video.id
                var rowModifier = Modifier.onSizeChanged { size -> rowHeight = size.height }
                if (dragged) {
                    rowModifier = rowModifier.zIndex(1f).graphicsLayer { translationY = dragOffset }
                }
                var handle: Modifier? = null
                if (reorderable) {
                    handle = Modifier.pointerInput(video.id) {
                        detectDragGestures(
                            onDragStart = {
                                draggingId = video.id
                                dragOffset = 0f
                            },
                            onDrag = { change, amount ->
                                change.consume()
                                val index = shown.indexOfFirst { item -> item.id == video.id }
                                val step = reorderStep(index, dragOffset + amount.y, rowHeight.toFloat(), shown.size - 1)
                                if (step.index != index) {
                                    shown = shown.toMutableList().apply { add(step.index, removeAt(index)) }
                                }
                                dragOffset = step.offset
                            },
                            onDragEnd = {
                                onMove(video.id, shown.indexOfFirst { item -> item.id == video.id })
                                draggingId = null
                                dragOffset = 0f
                            },
                            onDragCancel = {
                                draggingId = null
                                dragOffset = 0f
                            },
                        )
                    }
                }
                QueueRow(
                    video,
                    highlighted = video.id == playingId,
                    dragHandle = handle,
                    onClick = { onPlay(video.id) },
                    onRemove = { onRemove(video.id) },
                    modifier = rowModifier,
                )
            }
        }
    }
}

/** Where a dragged row is after moving [offset] from its slot at [index]: the next slot once past half a row. */
data class ReorderStep(val index: Int, val offset: Float)

/** One drag step of a row of [rowHeight] at [index] among rows up to [lastIndex]. */
fun reorderStep(index: Int, offset: Float, rowHeight: Float, lastIndex: Int): ReorderStep {
    if (rowHeight <= 0f) {
        return ReorderStep(index, offset)
    }
    if (offset > rowHeight / 2 && index < lastIndex) {
        return ReorderStep(index + 1, offset - rowHeight)
    }
    if (offset < -rowHeight / 2 && index > 0) {
        return ReorderStep(index - 1, offset + rowHeight)
    }
    return ReorderStep(index, offset)
}

/** One video of the queue: handle, cover, title and byline, ×; the playing one on a lighter fill. */
@Composable
private fun QueueRow(
    video: VideoSummary,
    highlighted: Boolean,
    dragHandle: Modifier?,
    onClick: () -> Unit,
    onRemove: (() -> Unit)?,
    modifier: Modifier = Modifier,
) {
    var rowModifier = modifier.fillMaxWidth()
    if (highlighted) {
        rowModifier = rowModifier.background(VeilColors.surfaceHigh)
    }
    Row(
        rowModifier.pressClickable(onClick).padding(horizontal = VeilSpacing.small, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (dragHandle != null) {
            Box(dragHandle.size(BarIconSize * 2), contentAlignment = Alignment.Center) {
                Icon(VeilIcons.DragHandle, contentDescription = "Drag to reorder", tint = VeilColors.contentFaint, modifier = Modifier.size(BarIconSize))
            }
        }
        RemoteImage(video.coverUrl, Modifier.width(QueueCoverWidth).aspectRatio(COVER_ASPECT_RATIO).clip(VeilShapes.small))
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.hairline)) {
            if (highlighted) {
                Text("Playing", style = VeilType.cardMeta, color = VeilColors.content)
            }
            Text(video.title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val byline = video.byline
            if (byline != null) {
                Text(byline, style = VeilType.cardMeta, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
            }
        }
        if (onRemove != null) {
            IconTap(VeilIcons.Close, contentDescription = "Remove from the queue", onClick = onRemove)
        }
    }
}

/**
 * The queue's bar at the bottom of the video page, as a music app's: what plays next (or how many
 * are queued), with shuffle and repeat when on; a tap or a swipe up opens [QueueSheet]. Hidden
 * while the queue holds nothing but the video playing.
 */
@Composable
fun QueueBar(playingId: String, onOpen: () -> Unit, modifier: Modifier = Modifier) {
    val contents by PlayQueue.contents.collectAsState()
    val shuffle by PlayQueue.shuffle.collectAsState()
    val repeat by PlayQueue.repeat.collectAsState()
    val shuffledIds by PlayQueue.shuffledIds.collectAsState()
    val order = remember(contents, shuffle, shuffledIds) { PlayQueue.playOrder() }
    if (order.none { video -> video.id != playingId }) {
        return
    }
    val nextId = remember(order, repeat, playingId) { PlayQueue.next(playingId) }
    val next = order.firstOrNull { video -> video.id == nextId }
    var label = "Queue · ${order.size}"
    if (next != null && next.id != playingId) {
        label = "Next: ${next.title}"
    }
    Row(
        modifier
            .fillMaxWidth()
            .background(VeilColors.surface)
            .pressClickable(onOpen)
            .pointerInput(Unit) {
                detectDragGestures { change, amount ->
                    if (amount.y < 0) {
                        change.consume()
                        onOpen()
                    }
                }
            }
            .padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.medium),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(VeilIcons.Queue, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(BarIconSize))
        Text(label, Modifier.weight(1f), style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
        if (shuffle) {
            Icon(VeilIcons.Random, contentDescription = "Shuffle on", tint = VeilColors.contentMuted, modifier = Modifier.size(BarIconSize))
        }
        if (repeat != RepeatMode.OFF) {
            Icon(repeatIcon(repeat), contentDescription = repeat.label, tint = VeilColors.contentMuted, modifier = Modifier.size(BarIconSize))
        }
    }
}
