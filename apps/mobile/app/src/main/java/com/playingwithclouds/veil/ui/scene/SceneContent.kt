package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
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
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.SceneDetail
import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.data.StreamGroup
import com.playingwithclouds.veil.data.StreamOption
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
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.PillRow
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
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
        Text(detail.title, style = MaterialTheme.typography.headlineSmall)
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
    LazyRow(contentPadding = PaddingValues(horizontal = 8.dp), horizontalArrangement = Arrangement.spacedBy(2.dp)) {
        item {
            ActionButton(likeIcon(state.verdict == Verdict.UP), "Like", active = state.verdict == Verdict.UP) { viewModel.react(Verdict.UP) }
        }
        item {
            ActionButton(dislikeIcon(state.verdict == Verdict.DOWN), "Dislike", active = state.verdict == Verdict.DOWN) { viewModel.react(Verdict.DOWN) }
        }
        item {
            ActionButton(VeilIcons.HeartFilled, "O ${state.oCount}", onLongClick = viewModel::decrementOCount) { viewModel.incrementOCount() }
        }
        item {
            ActionButton(watchlistIcon(state.onWatchlist), "Watchlist", active = state.onWatchlist) { viewModel.toggleWatchlist() }
        }
        item {
            ActionButton(VeilIcons.CollectionAdd, "Collect") {
                viewModel.loadUserCollections()
                onOpen(SceneDialog.Collections)
            }
        }
        if (best != null) {
            item { ActionButton(VeilIcons.Download, "Download") { viewModel.download(best) } }
        }
        item { ActionButton(VeilIcons.Alternates, "Alternates") { onFindAlternates() } }
        item { ActionButton(VeilIcons.Block, "Block") { onOpen(SceneDialog.Block) } }
    }
}

/** The like icon, filled when chosen. */
private fun likeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return VeilIcons.LikeFilled
    }
    return VeilIcons.Like
}

/** The dislike icon, filled when chosen. */
private fun dislikeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return VeilIcons.DislikeFilled
    }
    return VeilIcons.Dislike
}

/** The watchlist icon, filled when the scene is saved. */
private fun watchlistIcon(saved: Boolean): ImageVector {
    if (saved) {
        return VeilIcons.BookmarkFilled
    }
    return VeilIcons.Bookmark
}

/**
 * A round icon button over a caption, white with a dark icon while [active]; a long press can
 * trigger a second action.
 */
@Composable
private fun ActionButton(
    icon: ImageVector,
    label: String,
    active: Boolean = false,
    onLongClick: (() -> Unit)? = null,
    onClick: () -> Unit,
) {
    var background = VeilColors.surfaceHigh
    var tint = VeilColors.content
    if (active) {
        background = VeilColors.content
        tint = VeilColors.canvas
    }
    Column(
        Modifier.width(68.dp).pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick).padding(vertical = 6.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Box(Modifier.size(48.dp).clip(CircleShape).background(background), contentAlignment = Alignment.Center) {
            Icon(icon, contentDescription = label, tint = tint, modifier = Modifier.size(22.dp))
        }
        Text(label, style = MaterialTheme.typography.labelSmall, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
    }
}

/** Sources grouped by site, with a pill per quality; the playing one is selected. */
@Composable
private fun SourcePicker(groups: List<StreamGroup>, state: SceneState, viewModel: SceneViewModel) {
    Column(Modifier.padding(vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        SectionTitle("Sources")
        for (group in groups) {
            Text(group.provider, Modifier.padding(horizontal = 16.dp), style = MaterialTheme.typography.labelMedium, color = VeilColors.contentMuted)
            PillRow(
                group.streams,
                labelOf = { stream -> sourceLabel(stream) },
                onClick = { stream -> viewModel.play(stream) },
                isSelected = { stream -> state.active?.option?.id == stream.id },
                contentPadding = PaddingValues(horizontal = 16.dp),
            )
        }
    }
}

/** A source's quality, ticked when the backend verified it plays. */
private fun sourceLabel(stream: StreamOption): String {
    var label = stream.qualityLabel()
    if (stream.verified) {
        label += " ✓"
    }
    return label
}

/** The studio or channel of the scene, else the site it came from. */
@Composable
private fun StudioLine(detail: SceneDetail, state: SceneState, navigator: AppNavigator) {
    val studio = detail.studio
    if (studio == null) {
        return
    }
    Row(
        Modifier.fillMaxWidth().pressClickable { navigator.openStudio(studio.id) }.padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Avatar(studio.imagePath, Modifier.size(44.dp))
        Column {
            Text(studio.name, style = MaterialTheme.typography.titleSmall)
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
                Modifier.width(76.dp).pressClickable { navigator.openPerformer(performer.id) },
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Avatar(performer.imagePath, Modifier.size(64.dp))
                Text(performer.name, style = MaterialTheme.typography.labelMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
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
            TextAction("Add at current time", onClick = onAdd, icon = VeilIcons.Plus, modifier = Modifier.padding(end = 8.dp))
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
                    IconTap(VeilIcons.Delete, contentDescription = "Delete marker", onClick = { viewModel.deleteMarker(marker) })
                }
            }
        }
    }
}
