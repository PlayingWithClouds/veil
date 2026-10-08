package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
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
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilCheck
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.theme.VeilColors
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
    VeilDialog(
        title = "Add to collection",
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("New collection", onClick = { creating = true })
            PrimaryButton("Done", onClick = onDismiss)
        },
    ) {
        if (collections.isEmpty()) {
            Text("You have no collections yet.", color = VeilColors.contentMuted)
            return@VeilDialog
        }
        LazyColumn {
            items(collections, key = { collection -> collection.id }) { collection ->
                val included = collection.id in memberOf
                Row(
                    Modifier.fillMaxWidth().pressClickable { viewModel.setInCollection(collection, !included) }.padding(vertical = 10.dp),
                    horizontalArrangement = Arrangement.spacedBy(14.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    VeilCheck(checked = included)
                    Text(collection.name, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
            }
        }
    }
}

/** Lists the scene's studio, performers and tags; each can be blocked from recommendations. */
@Composable
private fun BlockDialog(detail: SceneDetail, viewModel: SceneViewModel, onDismiss: () -> Unit) {
    VeilDialog(
        title = "Block from recommendations",
        onDismissRequest = onDismiss,
        buttons = { PrimaryButton("Done", onClick = onDismiss) },
    ) {
        LazyColumn(verticalArrangement = Arrangement.spacedBy(4.dp)) {
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
    }
}

/** One blockable entity with its kind and a block button. */
@Composable
private fun BlockRow(kind: String, name: String, onBlock: () -> Unit) {
    Row(Modifier.fillMaxWidth().padding(vertical = 4.dp), verticalAlignment = Alignment.CenterVertically) {
        Column(Modifier.weight(1f)) {
            Text(name, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(kind, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        SecondaryButton("Block", onClick = onBlock)
    }
}

/** Asks for a tag (e.g. a position) or a note to pin at the current playback time. */
@Composable
private fun MarkerDialog(seconds: Double, onSave: (Double, String?, String?) -> Unit, onDismiss: () -> Unit) {
    var tagName by remember { mutableStateOf("") }
    var label by remember { mutableStateOf("") }
    VeilDialog(
        title = "Marker at ${formatClock(seconds)}",
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("Cancel", onClick = onDismiss)
            PrimaryButton(
                "Save",
                enabled = tagName.isNotBlank() || label.isNotBlank(),
                onClick = { onSave(seconds, tagName.trim().ifEmpty { null }, label.trim().ifEmpty { null }) },
            )
        },
    ) {
        VeilTextField(tagName, { text -> tagName = text }, label = "Tag", placeholder = "e.g. a position", modifier = Modifier.fillMaxWidth())
        VeilTextField(label, { text -> label = text }, label = "Note", modifier = Modifier.fillMaxWidth())
    }
}

/** Copies of the scene on other sites, best match first; choosing one attaches its source. */
@Composable
fun AlternatesSheet(viewModel: SceneViewModel, onDismiss: () -> Unit) {
    val alike by viewModel.alike.collectAsStateWithLifecycle()
    VeilBottomSheet(onDismissRequest = onDismiss) {
        SectionTitle("Other sources for this video")
        if (alike.isLoading) {
            Spinner(Modifier.padding(24.dp).align(Alignment.CenterHorizontally))
            return@VeilBottomSheet
        }
        if (alike.candidates.isEmpty()) {
            EmptyMessage("No matching videos found on other sites.")
            return@VeilBottomSheet
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
        Modifier.fillMaxWidth().pressClickable(enabled = !attaching, onClick = onChoose).padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(candidate.posterUrl, Modifier.size(width = 120.dp, height = 68.dp).clip(RoundedCornerShape(12.dp)))
        Column(Modifier.weight(1f)) {
            Text(candidate.title, style = MaterialTheme.typography.bodyMedium, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val facts = listOfNotNull(candidate.plugin, formatDuration(candidate.durationSeconds), "${(candidate.matchScore * PERCENT).toInt()}% match")
            Text(facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        if (attaching) {
            Spinner(size = 22.dp)
        }
    }
}

private const val PERCENT = 100
