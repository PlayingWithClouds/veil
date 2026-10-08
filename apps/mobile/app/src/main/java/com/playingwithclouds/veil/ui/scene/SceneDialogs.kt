package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.AlikeCandidate
import com.playingwithclouds.veil.data.SceneDetail
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SectionTitle
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatDuration

private const val MILLISECONDS_PER_SECOND = 1000.0

/** Which dialog the scene page has open. */
sealed interface SceneDialog {
    data object Collections : SceneDialog
    data object Block : SceneDialog
    data object Marker : SceneDialog
}

/** Shows the open dialog, if any. */
@Composable
fun SceneDialogs(dialog: SceneDialog?, detail: SceneDetail, viewModel: SceneViewModel, player: ExoPlayer, onDismiss: () -> Unit) {
    when (dialog) {
        SceneDialog.Collections -> CollectionPickerDialog(viewModel, onDismiss)
        SceneDialog.Block -> BlockDialog(detail, viewModel, onDismiss)
        SceneDialog.Marker -> MarkerDialog(
            seconds = player.currentPosition / MILLISECONDS_PER_SECOND,
            onSave = { seconds, tagName, label ->
                viewModel.addMarker(seconds, tagName, label)
                onDismiss()
            },
            onDismiss = onDismiss,
        )
        null -> Unit
    }
}

/** Ticks the user's collections the scene belongs to; a new one can be created on the spot. */
@Composable
private fun CollectionPickerDialog(viewModel: SceneViewModel, onDismiss: () -> Unit) {
    val collections by viewModel.userCollections.collectAsStateWithLifecycle()
    val memberOf by viewModel.collectionIds.collectAsStateWithLifecycle()
    var creating by remember { mutableStateOf(false) }

    if (creating) {
        TextInputDialog(
            title = "New collection",
            label = "Name",
            confirmLabel = "Create",
            onConfirm = { name ->
                creating = false
                viewModel.createCollectionWithScene(name)
            },
            onDismiss = { creating = false },
        )
        return
    }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Add to collection") },
        text = {
            if (collections.isEmpty()) {
                Text("You have no collections yet.")
                return@AlertDialog
            }
            LazyColumn {
                items(collections, key = { collection -> collection.id }) { collection ->
                    val included = collection.id in memberOf
                    Row(
                        Modifier.fillMaxWidth().clickable { viewModel.setInCollection(collection, !included) },
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Checkbox(checked = included, onCheckedChange = { checked -> viewModel.setInCollection(collection, checked) })
                        Text(collection.name, maxLines = 1, overflow = TextOverflow.Ellipsis)
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("Done") } },
        dismissButton = { TextButton(onClick = { creating = true }) { Text("New collection") } },
    )
}

/** Lists the scene's studio, performers and tags; each can be blocked from recommendations. */
@Composable
private fun BlockDialog(detail: SceneDetail, viewModel: SceneViewModel, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Block from recommendations") },
        text = {
            LazyColumn {
                val studio = detail.studio
                if (studio != null) {
                    item { BlockRow("Studio", studio.name) { viewModel.block("studio", studio.id, studio.name) } }
                }
                items(detail.performers, key = { performer -> "performer-${performer.id}" }) { performer ->
                    BlockRow("Performer", performer.name) { viewModel.block("performer", performer.id, performer.name) }
                }
                items(detail.tags, key = { tag -> "tag-${tag.id}" }) { tag ->
                    BlockRow("Tag", tag.name) { viewModel.block("tag", tag.id, tag.name) }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("Done") } },
    )
}

/** One blockable entity with its kind and a block button. */
@Composable
private fun BlockRow(kind: String, name: String, onBlock: () -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Column(Modifier.weight(1f)) {
            Text(name, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(kind, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
        TextButton(onClick = onBlock) { Text("Block") }
    }
}

/** Asks for a tag (e.g. a position) or a note to pin at the current playback time. */
@Composable
private fun MarkerDialog(seconds: Double, onSave: (Double, String?, String?) -> Unit, onDismiss: () -> Unit) {
    var tagName by remember { mutableStateOf("") }
    var label by remember { mutableStateOf("") }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Marker at ${formatClock(seconds)}") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(tagName, { text -> tagName = text }, label = { Text("Tag (e.g. a position)") }, singleLine = true)
                OutlinedTextField(label, { text -> label = text }, label = { Text("Note") }, singleLine = true)
            }
        },
        confirmButton = {
            TextButton(
                enabled = tagName.isNotBlank() || label.isNotBlank(),
                onClick = { onSave(seconds, tagName.trim().ifEmpty { null }, label.trim().ifEmpty { null }) },
            ) { Text("Save") }
        },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}

/** Copies of the scene on other sites, best match first; choosing one attaches its source. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AlternatesSheet(viewModel: SceneViewModel, onDismiss: () -> Unit) {
    val alike by viewModel.alike.collectAsStateWithLifecycle()
    ModalBottomSheet(onDismissRequest = onDismiss) {
        SectionTitle("Other sources for this video")
        if (alike.isLoading) {
            CircularProgressIndicator(Modifier.padding(24.dp).size(32.dp))
            return@ModalBottomSheet
        }
        if (alike.candidates.isEmpty()) {
            EmptyMessage("No matching videos found on other sites.")
            return@ModalBottomSheet
        }
        LazyColumn {
            items(alike.candidates, key = { candidate -> candidate.sourceUrl }) { candidate ->
                AlternateRow(candidate, attaching = alike.attachingUrl == candidate.sourceUrl) {
                    viewModel.attachAlike(candidate)
                    onDismiss()
                }
            }
        }
    }
}

/** A candidate with poster, title, site, runtime and match percentage. */
@Composable
private fun AlternateRow(candidate: AlikeCandidate, attaching: Boolean, onChoose: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().clickable(enabled = !attaching, onClick = onChoose).padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(candidate.posterUrl, Modifier.width(120.dp).size(width = 120.dp, height = 68.dp))
        Column(Modifier.weight(1f)) {
            Text(candidate.title, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val facts = listOfNotNull(candidate.plugin, formatDuration(candidate.durationSeconds), "${(candidate.matchScore * PERCENT).toInt()}% match")
            Text(facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
        if (attaching) {
            CircularProgressIndicator(Modifier.size(24.dp))
        }
    }
}

private const val PERCENT = 100
