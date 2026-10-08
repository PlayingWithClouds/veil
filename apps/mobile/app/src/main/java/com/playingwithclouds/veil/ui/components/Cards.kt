package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.StudioSummary
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatCount
import com.playingwithclouds.veil.util.formatVideoCount
import com.playingwithclouds.veil.util.formatReleaseDate

private const val LANDSCAPE_RATIO = 16f / 9f

/** A scene as a landscape card: poster with runtime and resume bar, then title and byline. */
@Composable
fun SceneCard(
    scene: SceneSummary,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    isNew: Boolean = false,
) {
    val progress by WatchProgressStore.progress.collectAsStateWithLifecycle()
    val resumeFraction = progress[scene.id]?.fraction
    Column(modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(RoundedCornerShape(16.dp))) {
            RemoteImage(scene.posterPath, Modifier.matchParentSize())
            val duration = scene.durationSeconds
            if (duration != null && duration > 0) {
                Badge(formatClock(duration.toDouble()), Modifier.align(Alignment.BottomEnd).padding(6.dp))
            }
            if (isNew) {
                Badge("NEW", Modifier.align(Alignment.TopStart).padding(6.dp), VeilColors.accent, Color.Black)
            }
            if (resumeFraction != null) {
                LinearProgressIndicator(
                    progress = { resumeFraction },
                    modifier = Modifier.align(Alignment.BottomStart).fillMaxWidth().height(3.dp),
                    color = VeilColors.accent,
                    trackColor = Color.Black.copy(alpha = 0.5f),
                )
            }
        }
        Column(Modifier.padding(horizontal = 4.dp, vertical = 8.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(
                scene.title,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            SceneCaption(scene)
        }
    }
}

/** The muted line under a scene title: channel or performers, then the release date. */
@Composable
private fun SceneCaption(scene: SceneSummary) {
    val parts = listOfNotNull(scene.byline, formatReleaseDate(scene.date))
    if (parts.isEmpty()) {
        return
    }
    Text(
        parts.joinToString(" · "),
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        maxLines = 1,
        overflow = TextOverflow.Ellipsis,
    )
}

/** A small rounded label over an image. */
@Composable
fun Badge(
    text: String,
    modifier: Modifier = Modifier,
    background: Color = Color.Black.copy(alpha = 0.75f),
    foreground: Color = Color.White,
) {
    Text(
        text,
        modifier = modifier.clip(RoundedCornerShape(4.dp)).background(background).padding(horizontal = 6.dp, vertical = 2.dp),
        color = foreground,
        fontSize = 11.sp,
        fontWeight = FontWeight.SemiBold,
    )
}

/** A gallery as a poster card with its image count. */
@Composable
fun GalleryCard(gallery: GallerySummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(3f / 4f).clip(RoundedCornerShape(16.dp))) {
            RemoteImage(gallery.coverPath, Modifier.matchParentSize())
            if (gallery.imageCount > 0) {
                Badge("${gallery.imageCount} photos", Modifier.align(Alignment.BottomEnd).padding(6.dp))
            }
        }
        Text(
            gallery.title,
            modifier = Modifier.padding(horizontal = 4.dp, vertical = 8.dp),
            style = MaterialTheme.typography.bodyMedium,
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
        modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).pressClickable(onClick).padding(8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Box {
            Avatar(imagePath, Modifier.fillMaxWidth(0.8f).aspectRatio(1f))
            if (isFavorite) {
                Icon(
                    Icons.Filled.Favorite,
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
            textAlign = TextAlign.Center,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        if (subtitle != null) {
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
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
    Column(modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).pressClickable(onClick)) {
        RemoteImage(
            collection.coverPath,
            Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(RoundedCornerShape(16.dp)),
        )
        Column(Modifier.padding(horizontal = 4.dp, vertical = 8.dp)) {
            Text(
                collection.name,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                "${collection.itemCount} items",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

/** A round image, or a person silhouette when there is none. */
@Composable
fun Avatar(imagePath: String?, modifier: Modifier = Modifier) {
    Box(modifier.clip(CircleShape).background(VeilColors.surfaceHigh), contentAlignment = Alignment.Center) {
        if (imagePath == null) {
            Icon(Icons.Filled.Person, contentDescription = null, tint = VeilColors.contentFaint)
            return@Box
        }
        RemoteImage(imagePath, Modifier.matchParentSize())
    }
}

/** A small section heading. */
@Composable
fun SectionTitle(text: String, modifier: Modifier = Modifier) {
    Text(
        text,
        modifier = modifier.padding(horizontal = 16.dp, vertical = 8.dp),
        style = MaterialTheme.typography.titleSmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}

/** A row of text with a trailing value, for info lists. */
@Composable
fun InfoRow(label: String, value: String) {
    Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = MaterialTheme.colorScheme.onSurfaceVariant, style = MaterialTheme.typography.bodyMedium)
        Text(value, style = MaterialTheme.typography.bodyMedium, textAlign = TextAlign.End, modifier = Modifier.padding(start = 16.dp))
    }
}
