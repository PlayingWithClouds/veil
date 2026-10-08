package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.SceneProgress
import com.playingwithclouds.veil.data.SceneSites
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.StudioSummary
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatReleaseDate
import com.playingwithclouds.veil.util.formatTimeLeft
import com.playingwithclouds.veil.util.formatVideoCount

private const val LANDSCAPE_RATIO = 16f / 9f

/**
 * A scene as a landscape card: poster with runtime (or time left) and resume bar, then title,
 * byline and, on recommendations, why it was picked in the accent. A long press opens
 * [onLongClick], typically the quick actions sheet.
 */
@Composable
fun SceneCard(
    scene: SceneSummary,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    isNew: Boolean = false,
    reason: String? = null,
    onLongClick: (() -> Unit)? = null,
) {
    val progress by WatchProgressStore.progress.collectAsStateWithLifecycle()
    val sceneProgress = progress[scene.id]
    Column(modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(VeilShapes.card)) {
            RemoteImage(scene.posterPath, Modifier.matchParentSize())
            val runtime = runtimeLabel(scene.durationSeconds, sceneProgress)
            if (runtime != null) {
                Badge(runtime, Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
            if (isNew) {
                Badge("NEW", Modifier.align(Alignment.TopStart).padding(VeilSpacing.small), VeilColors.accent, VeilColors.onAccent)
            }
            val resumeFraction = sceneProgress?.fraction
            if (resumeFraction != null) {
                ProgressBar(
                    progress = resumeFraction,
                    modifier = Modifier.align(Alignment.BottomStart),
                    trackColor = VeilColors.imageTrack,
                    height = 3.dp,
                )
            }
        }
        Column(
            Modifier.padding(horizontal = VeilSpacing.extraSmall, vertical = VeilSpacing.small),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.hairline),
        ) {
            Text(
                scene.title,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                color = VeilColors.content,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            SceneCaption(scene)
            if (!reason.isNullOrBlank()) {
                Text(reason, style = MaterialTheme.typography.labelMedium, color = VeilColors.accent, maxLines = 1, overflow = TextOverflow.Ellipsis)
            }
        }
    }
}

/** The poster badge: time left while the scene is part-watched, else its runtime; null when unknown. */
private fun runtimeLabel(durationSeconds: Int?, progress: SceneProgress?): String? {
    var total = durationSeconds
    if (progress?.durationSeconds != null) {
        total = progress.durationSeconds
    }
    if (total == null || total <= 0) {
        return null
    }
    if (progress != null && progress.progressSeconds in 1 until total) {
        return formatTimeLeft((total - progress.progressSeconds).toDouble())
    }
    return formatClock(total.toDouble())
}

/** The muted line under a scene title: channel or performers (else the site), then the release date. */
@Composable
private fun SceneCaption(scene: SceneSummary) {
    var source = scene.byline
    if (source == null) {
        source = SceneSites.hostOf(scene.sourceUrl)
    }
    val parts = listOfNotNull(source, formatReleaseDate(scene.date))
    if (parts.isEmpty()) {
        return
    }
    Text(
        parts.joinToString(" · "),
        style = MaterialTheme.typography.bodySmall,
        color = VeilColors.contentMuted,
        maxLines = 1,
        overflow = TextOverflow.Ellipsis,
    )
}

/** A gallery as a poster card with its image count. */
@Composable
fun GalleryCard(gallery: GallerySummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(3f / 4f).clip(VeilShapes.card)) {
            RemoteImage(gallery.coverPath, Modifier.matchParentSize())
            if (gallery.imageCount > 0) {
                Badge("${gallery.imageCount} photos", Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
        }
        Text(
            gallery.title,
            modifier = Modifier.padding(horizontal = VeilSpacing.extraSmall, vertical = VeilSpacing.small),
            style = MaterialTheme.typography.bodyMedium,
            color = VeilColors.content,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

/** A round photo with a name below: performers, studios and followed things. */
@Composable
fun AvatarCard(
    name: String,
    imagePath: String?,
    subtitle: String?,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    isFavorite: Boolean = false,
) {
    Column(
        modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(onClick).padding(VeilSpacing.small),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        Box {
            Avatar(imagePath, Modifier.fillMaxWidth(0.8f).aspectRatio(1f))
            if (isFavorite) {
                Icon(
                    VeilIcons.HeartFilled,
                    contentDescription = "Favorite",
                    tint = VeilColors.error,
                    modifier = Modifier.align(Alignment.TopEnd).size(18.dp),
                )
            }
        }
        Text(
            name,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium,
            color = VeilColors.content,
            textAlign = TextAlign.Center,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        if (subtitle != null) {
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
    }
}

/** A performer in the index grid. */
@Composable
fun PerformerCard(performer: PerformerSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    AvatarCard(
        name = performer.name,
        imagePath = performer.imagePath,
        subtitle = formatVideoCount(performer.sceneCount),
        onClick = onClick,
        modifier = modifier,
        isFavorite = performer.favorite,
    )
}

/** A studio in the index grid. */
@Composable
fun StudioCard(studio: StudioSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    AvatarCard(
        name = studio.name,
        imagePath = studio.imagePath,
        subtitle = formatVideoCount(studio.sceneCount),
        onClick = onClick,
        modifier = modifier,
    )
}

/** A collection with its cover and size. */
@Composable
fun CollectionCard(collection: CollectionSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(onClick)) {
        RemoteImage(
            collection.coverPath,
            Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(VeilShapes.card),
        )
        Column(Modifier.padding(horizontal = VeilSpacing.extraSmall, vertical = VeilSpacing.small)) {
            Text(
                collection.name,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                color = VeilColors.content,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                "${collection.itemCount} items",
                style = MaterialTheme.typography.bodySmall,
                color = VeilColors.contentMuted,
            )
        }
    }
}

/** A round image, or a person silhouette when there is none. */
@Composable
fun Avatar(imagePath: String?, modifier: Modifier = Modifier) {
    Box(modifier.clip(VeilShapes.capsule).background(VeilColors.surfaceHigh), contentAlignment = Alignment.Center) {
        if (imagePath == null) {
            Icon(VeilIcons.Performer, contentDescription = null, tint = VeilColors.contentFaint)
            return@Box
        }
        RemoteImage(imagePath, Modifier.matchParentSize())
    }
}

/** A row of text with a trailing value, for info lists. */
@Composable
fun InfoRow(label: String, value: String) {
    Row(
        Modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, color = VeilColors.contentMuted, style = MaterialTheme.typography.bodyMedium)
        Text(
            value,
            style = MaterialTheme.typography.bodyMedium,
            color = VeilColors.content,
            textAlign = TextAlign.End,
            modifier = Modifier.padding(start = VeilSpacing.large),
        )
    }
}
