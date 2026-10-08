package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.PlaybackQueue
import com.playingwithclouds.veil.data.QueueEntry
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Tallest the queue list grows before it scrolls. */
private val QueueListMaxHeight = 420.dp

/**
 * The scenes lined up to play after this one, with a shuffle switch (the next scene is picked at
 * random) and a way to clear. Tapping a scene plays it now through [onPlay].
 */
@Composable
fun QueueSheet(onPlay: (QueueEntry) -> Unit, onDismiss: () -> Unit) {
    val queue by PlaybackQueue.state.collectAsStateWithLifecycle()
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween) {
            SectionHeading("Up next", Modifier.gutterPadding())
            Row(Modifier.padding(end = VeilSpacing.small), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
                Pill("Shuffle", onClick = { PlaybackQueue.setShuffle(!queue.shuffle) }, selected = queue.shuffle, icon = VeilIcons.Random)
                if (queue.entries.isNotEmpty()) {
                    TextAction("Clear", onClick = PlaybackQueue::clear)
                }
            }
        }
        if (queue.entries.isEmpty()) {
            EmptyMessage("Nothing is queued. Use Play next or Add to queue on any video.")
            return@VeilBottomSheet
        }
        LazyColumn(Modifier.heightIn(max = QueueListMaxHeight)) {
            items(queue.entries, key = { entry -> entry.sceneId }) { entry ->
                QueueRow(entry, onPlay = { onPlay(entry) }, onRemove = { PlaybackQueue.remove(entry.sceneId) })
            }
        }
    }
}

/** A queued scene with its poster and title, a remove button on the right. */
@Composable
private fun QueueRow(entry: QueueEntry, onPlay: () -> Unit, onRemove: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(onPlay).padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(entry.posterPath, Modifier.size(width = 120.dp, height = 68.dp).clip(VeilShapes.card))
        Text(
            entry.title,
            modifier = Modifier.weight(1f),
            style = MaterialTheme.typography.bodyMedium,
            color = VeilColors.content,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        IconTap(VeilIcons.Close, contentDescription = "Remove from queue", onClick = onRemove)
    }
}
