package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.Bookmark
import androidx.compose.material.icons.filled.BookmarkBorder
import androidx.compose.material.icons.filled.CreateNewFolder
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.automirrored.filled.ManageSearch
import androidx.compose.material.icons.filled.ThumbDown
import androidx.compose.material.icons.filled.ThumbDownOffAlt
import androidx.compose.material.icons.filled.ThumbUp
import androidx.compose.material.icons.filled.ThumbUpOffAlt
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.SceneDetail
import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.data.StreamGroup
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.data.groupStreams
import com.playingwithclouds.veil.data.qualityLabel
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SectionTitle
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatCount
import com.playingwithclouds.veil.util.formatVideoCount
import com.playingwithclouds.veil.util.formatDuration
import com.playingwithclouds.veil.util.formatReleaseDate

private const val MILLISECONDS_PER_SECOND = 1000.0

/** Everything below the player: title, reactions, sources, people, tags, markers and related scenes. */
@Composable
fun SceneContent(
    state: SceneState,
    viewModel: SceneViewModel,
    player: ExoPlayer,
    navigator: AppNavigator,
    onFindAlternates: () -> Unit,
) {
    val detail = (state.detail as? LoadState.Loaded)?.value ?: return
    var dialog by remember { mutableStateOf<SceneDialog?>(null) }
    SceneDialogs(dialog, detail, viewModel, player, onDismiss = { dialog = null })

    LazyColumn(contentPadding = PaddingValues(bottom = 16.dp)) {
        item { SceneTitleBlock(detail, state) }
        item { ReactionRow(state, viewModel, onOpen = { chosen -> dialog = chosen }, onFindAlternates = onFindAlternates) }
        if (state.streams.isNotEmpty()) {
            item { SourcePicker(groupStreams(state.streams), state, viewModel) }
        }
        item { StudioLine(detail, state, navigator) }
        item { PerformerRow(detail.performers, navigator) }
        item { TagChips(detail.tags, onClick = { tag -> navigator.openTag(tag.id) }) }
        item { DetailsText(detail.details) }
        item { MarkerSection(state.markers, player, viewModel, onAdd = { dialog = SceneDialog.Marker }) }
        if (state.related.isNotEmpty()) {
            item { SectionTitle("Related") }
            items(state.related, key = { scene -> scene.id }) { scene ->
                SceneCard(scene, onClick = { navigator.openScene(scene.id) }, modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp))
            }
        }
    }
}

/** Title, release date, runtime and view count. */
@Composable
private fun SceneTitleBlock(detail: SceneDetail, state: SceneState) {
    Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(detail.title, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
        val facts = listOfNotNull(
            formatReleaseDate(detail.date),
            formatDuration(detail.durationSeconds),
            "${formatCount(detail.viewCount)} views",
            state.site?.name,
        )
        Text(facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        if (state.fetchingDetail) {
            Text("Fetching details from the site...", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

/** The row of reactions and actions: like, dislike, O-counter, watchlist, collections, download, alternates, block. */
@Composable
private fun ReactionRow(
    state: SceneState,
    viewModel: SceneViewModel,
    onOpen: (SceneDialog) -> Unit,
    onFindAlternates: () -> Unit,
) {
    val best = state.active?.option ?: state.streams.firstOrNull()
    LazyRow(contentPadding = PaddingValues(horizontal = 8.dp), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        item {
            ActionButton(likeIcon(state.verdict == Verdict.UP), "Like") { viewModel.react(Verdict.UP) }
        }
        item {
            ActionButton(dislikeIcon(state.verdict == Verdict.DOWN), "Dislike") { viewModel.react(Verdict.DOWN) }
        }
        item {
            ActionButton(Icons.Filled.Favorite, "O ${state.oCount}", onLongClick = viewModel::decrementOCount) { viewModel.incrementOCount() }
        }
        item {
            ActionButton(watchlistIcon(state.onWatchlist), "Watchlist") { viewModel.toggleWatchlist() }
        }
        item {
            ActionButton(Icons.Filled.CreateNewFolder, "Collect") {
                viewModel.loadUserCollections()
                onOpen(SceneDialog.Collections)
            }
        }
        if (best != null) {
            item { ActionButton(Icons.Filled.Download, "Download") { viewModel.download(best) } }
        }
        item { ActionButton(Icons.AutoMirrored.Filled.ManageSearch, "Alternates") { onFindAlternates() } }
        item { ActionButton(Icons.Filled.Block, "Block") { onOpen(SceneDialog.Block) } }
    }
}

/** The like icon, filled when chosen. */
private fun likeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return Icons.Filled.ThumbUp
    }
    return Icons.Filled.ThumbUpOffAlt
}

/** The dislike icon, filled when chosen. */
private fun dislikeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return Icons.Filled.ThumbDown
    }
    return Icons.Filled.ThumbDownOffAlt
}

/** The watchlist icon, filled when the scene is saved. */
private fun watchlistIcon(saved: Boolean): ImageVector {
    if (saved) {
        return Icons.Filled.Bookmark
    }
    return Icons.Filled.BookmarkBorder
}

/** An icon over a caption; a long press can trigger a second action. */
@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ActionButton(icon: ImageVector, label: String, onLongClick: (() -> Unit)? = null, onClick: () -> Unit) {
    Column(
        Modifier.width(72.dp).combinedClickable(onClick = onClick, onLongClick = onLongClick).padding(vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Icon(icon, contentDescription = label)
        Text(label, style = MaterialTheme.typography.labelSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
    }
}

/** Sources grouped by site, with a chip per quality; the playing one is selected. */
@Composable
private fun SourcePicker(groups: List<StreamGroup>, state: SceneState, viewModel: SceneViewModel) {
    Column(Modifier.padding(vertical = 8.dp)) {
        SectionTitle("Sources")
        for (group in groups) {
            Text(group.provider, Modifier.padding(horizontal = 16.dp), style = MaterialTheme.typography.labelMedium)
            LazyRow(contentPadding = PaddingValues(horizontal = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(group.streams, key = { stream -> stream.id }) { stream ->
                    var label = stream.qualityLabel()
                    if (stream.verified) {
                        label += " ✓"
                    }
                    FilterChip(
                        selected = state.active?.option?.id == stream.id,
                        onClick = { viewModel.play(stream) },
                        label = { Text(label) },
                    )
                }
            }
        }
    }
}

/** The studio or channel of the scene, else the site it came from. */
@Composable
private fun StudioLine(detail: SceneDetail, state: SceneState, navigator: AppNavigator) {
    val studio = detail.studio
    if (studio == null) {
        return
    }
    Row(
        Modifier.fillMaxWidth().clickable { navigator.openStudio(studio.id) }.padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Avatar(studio.imagePath, Modifier.size(40.dp))
        Column {
            Text(studio.name, style = MaterialTheme.typography.bodyLarge)
            Text(formatVideoCount(studio.sceneCount), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
        val site = state.site
        if (site?.iconUrl != null) {
            RemoteImage(site.iconUrl, Modifier.size(20.dp))
        }
    }
}

/** The credited performers as round photos. */
@Composable
private fun PerformerRow(performers: List<EntityRef>, navigator: AppNavigator) {
    if (performers.isEmpty()) {
        return
    }
    LazyRow(contentPadding = PaddingValues(horizontal = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        items(performers, key = { performer -> performer.id }) { performer ->
            Column(
                Modifier.width(72.dp).clickable { navigator.openPerformer(performer.id) },
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Avatar(performer.imagePath, Modifier.size(56.dp))
                Text(performer.name, style = MaterialTheme.typography.labelSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
            }
        }
    }
}

/** The scene description, collapsed to a few lines until tapped. */
@Composable
private fun DetailsText(details: String?) {
    if (details.isNullOrBlank()) {
        return
    }
    var expanded by remember { mutableStateOf(false) }
    var maxLines = COLLAPSED_DETAIL_LINES
    if (expanded) {
        maxLines = Int.MAX_VALUE
    }
    Text(
        details,
        modifier = Modifier.fillMaxWidth().clickable { expanded = !expanded }.padding(16.dp),
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        maxLines = maxLines,
        overflow = TextOverflow.Ellipsis,
    )
}

private const val COLLAPSED_DETAIL_LINES = 3

/** Markers of the scene: tap to jump, personal ones can be removed, and one can be added at the current time. */
@Composable
private fun MarkerSection(markers: List<SceneMarker>, player: ExoPlayer, viewModel: SceneViewModel, onAdd: () -> Unit) {
    Column {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween) {
            SectionTitle("Markers")
            TextButton(onClick = onAdd) {
                Icon(Icons.Filled.Add, contentDescription = null, modifier = Modifier.size(18.dp))
                Text("Add at current time")
            }
        }
        for (marker in markers) {
            Row(
                Modifier.fillMaxWidth().clickable { player.seekTo((marker.seconds * MILLISECONDS_PER_SECOND).toLong()) }.padding(horizontal = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(formatClock(marker.seconds), color = MaterialTheme.colorScheme.onSurfaceVariant)
                Text(marker.title, Modifier.weight(1f), maxLines = 1, overflow = TextOverflow.Ellipsis)
                if (marker.personal) {
                    IconButton(onClick = { viewModel.deleteMarker(marker) }) {
                        Icon(Icons.Filled.Delete, contentDescription = "Delete marker")
                    }
                }
            }
        }
    }
}
