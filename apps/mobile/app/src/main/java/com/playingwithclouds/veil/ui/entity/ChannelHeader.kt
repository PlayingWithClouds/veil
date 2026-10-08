package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.background
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.InfoRow
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.SegmentedControl
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException

/** Height of the banner behind a channel's name. */
private val BannerHeight = 120.dp

/** Blur of the banner image, which is a photo or poster rather than real banner art. */
private val BannerBlur = 18.dp

/** Size of the channel avatar. */
private val ChannelAvatarSize = 80.dp

/** How a channel's filmography is ordered. */
enum class ChannelSort(val label: String, val backendName: String) {
    NEWEST("Newest", "date"),
    TOP_RATED("Top rated", "rating"),
}

/** The newest/top-rated switch above a channel's videos. */
@Composable
fun ChannelSortBar(selected: ChannelSort, onSelect: (ChannelSort) -> Unit) {
    SegmentedControl(
        labels = ChannelSort.entries.map { sort -> sort.label },
        selectedIndex = selected.ordinal,
        onSelect = { index -> onSelect(ChannelSort.entries[index]) },
        modifier = Modifier.padding(vertical = VeilSpacing.small),
    )
}

/**
 * The top of a performer or studio page laid out as a channel: a banner (the entity's own photo
 * or a poster, blurred), avatar, name, video count with a badge for videos found since the last
 * look, a bio teaser that opens the full facts, then the [actions] (follow bell first). Draws no
 * side padding; it sits in a grid that already pads with the gutter.
 */
@Composable
fun ChannelHeader(
    entityId: String,
    name: String,
    imagePath: String?,
    bannerPath: String?,
    videoCount: Int,
    facts: List<Pair<String, String?>>,
    details: String?,
    actions: @Composable () -> Unit,
) {
    var bioOpen by remember { mutableStateOf(false) }
    Column(Modifier.fillMaxWidth().padding(top = VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Banner(bannerPath, imagePath)
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large), verticalAlignment = Alignment.CenterVertically) {
            Avatar(imagePath, Modifier.size(ChannelAvatarSize))
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                Text(name, style = MaterialTheme.typography.headlineMedium, maxLines = 2, overflow = TextOverflow.Ellipsis)
                Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
                    Text("$videoCount videos", style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
                    NewScenesBadge(entityId)
                }
            }
        }
        BioTeaser(facts, details, onClick = { bioOpen = true })
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
            actions()
        }
    }
    if (bioOpen) {
        BioSheet(name, facts, details, onDismiss = { bioOpen = false })
    }
}

/** The rounded banner: [bannerPath] (else [imagePath]) blurred. */
@Composable
private fun Banner(bannerPath: String?, imagePath: String?) {
    var source = bannerPath
    if (source == null) {
        source = imagePath
    }
    Box(Modifier.fillMaxWidth().height(BannerHeight).clip(VeilShapes.card).background(VeilColors.surfaceHigh)) {
        if (source != null) {
            RemoteImage(source, Modifier.matchParentSize().blur(BannerBlur))
        }
    }
}

/** "N new" in accent while the entity is followed and has videos found since the last look. */
@Composable
private fun NewScenesBadge(entityId: String) {
    var newCount by remember(entityId) { mutableStateOf(0) }
    LaunchedEffect(entityId) {
        try {
            newCount = EntityRepository.subscriptionFor(entityId)?.newCount ?: 0
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            newCount = 0
        }
    }
    if (newCount > 0) {
        Badge("$newCount new", background = VeilColors.accent, foreground = VeilColors.onAccent)
    }
}

/** Two lines of the bio with a chevron; hidden when there is nothing to tell. */
@Composable
private fun BioTeaser(facts: List<Pair<String, String?>>, details: String?, onClick: () -> Unit) {
    var teaser = details
    if (teaser.isNullOrBlank()) {
        teaser = facts.firstOrNull { fact -> !fact.second.isNullOrBlank() }?.let { fact -> "${fact.first}: ${fact.second}" }
    }
    if (teaser.isNullOrBlank()) {
        return
    }
    Row(
        Modifier.fillMaxWidth().pressClickable(onClick),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(teaser, Modifier.weight(1f), style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted, maxLines = 2, overflow = TextOverflow.Ellipsis)
        Icon(VeilIcons.Chevron, contentDescription = "About", tint = VeilColors.contentMuted, modifier = Modifier.size(18.dp))
    }
}

/** The full bio and facts in a bottom sheet. */
@Composable
private fun BioSheet(name: String, facts: List<Pair<String, String?>>, details: String?, onDismiss: () -> Unit) {
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Text(
            name,
            modifier = Modifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
            style = MaterialTheme.typography.titleLarge,
            color = VeilColors.content,
        )
        if (!details.isNullOrBlank()) {
            Text(
                details,
                modifier = Modifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.contentMuted,
            )
        }
        for ((label, value) in facts) {
            if (!value.isNullOrBlank()) {
                InfoRow(label, value)
            }
        }
        Box(Modifier.height(VeilSpacing.large))
    }
}
