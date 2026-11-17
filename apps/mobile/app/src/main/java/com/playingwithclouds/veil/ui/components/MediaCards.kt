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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.AudioSummary
import com.playingwithclouds.veil.data.CamSummary
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.StorySummary
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatCount
import com.playingwithclouds.veil.util.formatReleaseDate
import com.playingwithclouds.veil.util.formatTimeLeft

private const val LANDSCAPE_RATIO = 16f / 9f
private const val PORTRAIT_RATIO = 2f / 3f

/** Size of the creator's avatar beside a video's title. */
private val CreatorAvatarSize = 32.dp

/** Height of the resume bar along a cover's bottom edge. */
private val ResumeBarHeight = 3.dp

/**
 * A video as a landscape card: cover with runtime (or time left) and resume bar, then the
 * creator's avatar beside the title, who it's from and when, and a [note] in the accent (why it's
 * recommended). While [previewing], its preview clip plays muted over the cover. [edgeToEdge] runs
 * the cover across the full width without rounded corners and gives the caption the gutter, for a
 * full-width feed. A long press calls [onLongClick], e.g. for a sheet of actions.
 */
@Composable
fun VideoCard(
    video: VideoSummary,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    note: String? = null,
    previewing: Boolean = false,
    edgeToEdge: Boolean = false,
    onLongClick: (() -> Unit)? = null,
) {
    var coverShape: Shape = VeilShapes.card
    var captionModifier = Modifier.padding(top = VeilSpacing.medium, bottom = VeilSpacing.small)
    if (edgeToEdge) {
        coverShape = RectangleShape
        captionModifier = Modifier.padding(start = VeilSpacing.gutter, top = VeilSpacing.medium, end = VeilSpacing.gutter, bottom = VeilSpacing.large)
    }
    Column(modifier.fillMaxWidth().pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(coverShape)) {
            RemoteImage(video.coverUrl, Modifier.matchParentSize())
            val previewUrl = video.previewUrl
            if (previewing && previewUrl != null) {
                ScenePreview(previewUrl, Modifier.matchParentSize())
            }
            val runtime = runtimeLabel(video)
            if (runtime != null) {
                Badge(runtime, Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
            val watched = watchedFraction(video)
            if (watched != null) {
                ProgressBar(watched, Modifier.align(Alignment.BottomStart), trackColor = VeilColors.imageTrack, height = ResumeBarHeight)
            }
        }
        Row(captionModifier, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            Avatar(video.avatarUrl, Modifier.size(CreatorAvatarSize))
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.hairline)) {
                CardTitle(video.title)
                CardMeta(listOfNotNull(video.byline, formatReleaseDate(video.date)).joinToString(" · "))
                if (note != null) {
                    Text(note, style = VeilType.cardMeta, color = VeilColors.accent, maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
            }
        }
    }
}

/** The cover badge: time left while a video is part-watched, else its runtime; null when unknown. */
private fun runtimeLabel(video: VideoSummary): String? {
    val duration = video.durationSeconds
    if (duration == null || duration <= 0) {
        return null
    }
    val progress = video.progressSeconds
    if (progress != null && progress > 0 && progress < duration) {
        return formatTimeLeft(duration - progress)
    }
    return formatClock(duration)
}

/** How much of a part-watched video is watched, 0..1; null when it isn't being watched. */
private fun watchedFraction(video: VideoSummary): Float? {
    val duration = video.durationSeconds
    val progress = video.progressSeconds
    if (duration == null || duration <= 0 || progress == null || progress <= 0) {
        return null
    }
    return (progress / duration).toFloat().coerceIn(0f, 1f)
}

/**
 * A cam room: its live picture with LIVE and the audience (or REC while it's being recorded), then
 * its name and site.
 */
@Composable
fun CamCard(cam: CamSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(VeilShapes.card)) {
            RemoteImage(cam.pictureUrl, Modifier.matchParentSize())
            val badgeModifier = Modifier.align(Alignment.TopStart).padding(VeilSpacing.small)
            val viewers = cam.viewers
            if (cam.recording) {
                Badge("REC", badgeModifier, VeilColors.error, VeilColors.content)
            } else if (viewers != null) {
                Badge("LIVE · ${formatCount(viewers)}", badgeModifier, VeilColors.live, VeilColors.onLive)
            }
        }
        Column(Modifier.padding(top = VeilSpacing.small, bottom = VeilSpacing.small)) {
            CardTitle(cam.name, maxLines = 1)
            CardMeta(cam.site)
        }
    }
}

/** A gallery as a portrait cover with its photo count, then its title and who's in it. */
@Composable
fun GalleryCard(gallery: GallerySummary, onClick: () -> Unit, modifier: Modifier = Modifier, onLongClick: (() -> Unit)? = null) {
    Column(modifier.fillMaxWidth().pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(GALLERY_RATIO).clip(VeilShapes.card)) {
            RemoteImage(gallery.coverUrl, Modifier.matchParentSize())
            if (gallery.imageCount > 0) {
                Badge("${gallery.imageCount} photos", Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
        }
        Column(Modifier.padding(top = VeilSpacing.small, bottom = VeilSpacing.small)) {
            CardTitle(gallery.title)
            val byline = gallery.byline
            if (byline != null) {
                CardMeta(byline)
            }
        }
    }
}

/** Gallery covers are portrait photos. */
private const val GALLERY_RATIO = 3f / 4f

/** A story as a portrait cover with its chapter count, then its title and author. */
@Composable
fun StoryCard(story: StorySummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(PORTRAIT_RATIO).clip(VeilShapes.card)) {
            RemoteImage(story.coverUrl, Modifier.matchParentSize())
            if (story.chapterCount > 1) {
                Badge("${story.chapterCount} ch.", Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
        }
        Column(Modifier.padding(top = VeilSpacing.small, bottom = VeilSpacing.small)) {
            CardTitle(story.title)
            val author = story.author
            if (author != null) {
                CardMeta(author)
            }
        }
    }
}

/** An audio as a square cover with its runtime, then its title and voice. */
@Composable
fun AudioCard(audio: AudioSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(1f).clip(VeilShapes.card)) {
            RemoteImage(audio.coverUrl, Modifier.matchParentSize())
            val duration = audio.durationSeconds
            if (duration != null && duration > 0) {
                Badge(formatClock(duration), Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            }
        }
        Column(Modifier.padding(top = VeilSpacing.small, bottom = VeilSpacing.small)) {
            CardTitle(audio.title)
            val voice = audio.voice
            if (voice != null) {
                CardMeta(voice)
            }
        }
    }
}

/** A performer as a round picture with the name under it. */
@Composable
fun PerformerBubble(performer: PerformerSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(
        modifier.fillMaxWidth().pressClickable(onClick),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        Avatar(performer.imageUrl, Modifier.fillMaxWidth().aspectRatio(1f))
        Text(
            performer.name,
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium,
            color = VeilColors.content,
            textAlign = TextAlign.Center,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

/** A round image, or a person silhouette when there is none. */
@Composable
fun Avatar(imageUrl: String?, modifier: Modifier = Modifier) {
    Box(modifier.clip(VeilShapes.capsule).background(VeilColors.surfaceHigh), contentAlignment = Alignment.Center) {
        if (imageUrl == null) {
            Icon(VeilIcons.Performer, contentDescription = null, tint = VeilColors.contentFaint)
            return@Box
        }
        RemoteImage(imageUrl, Modifier.matchParentSize())
    }
}

/** A card's title, two lines at most. */
@Composable
private fun CardTitle(title: String, maxLines: Int = 2) {
    Text(title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = maxLines, overflow = TextOverflow.Ellipsis)
}

/** The muted line under a card's title; nothing when empty. */
@Composable
private fun CardMeta(text: String) {
    if (text.isEmpty()) {
        return
    }
    Text(text, style = VeilType.cardMeta, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
}
