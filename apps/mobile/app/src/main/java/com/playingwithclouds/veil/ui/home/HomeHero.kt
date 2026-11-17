package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.CamSummary
import com.playingwithclouds.veil.data.RecommendedVideo
import com.playingwithclouds.veil.ui.cams.CamAvatar
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.ScenePreview
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType

/** The hero's shape: taller than a video, so it fills the phone's first screen. */
private const val HERO_RATIO = 4f / 5f

/** Share of the hero the fade into the page covers, from the bottom. */
private const val HERO_FADE_START = 0.45f

/** Width of a live room in the faces row. */
private val LiveFaceWidth = 72.dp

/** The fade from the moving picture into the page under the hero's text. */
private val HeroFade = Brush.verticalGradient(
    HERO_FADE_START to Color.Transparent,
    1f to VeilColors.canvas,
)

/**
 * The first thing Home shows: one recommendation edge to edge under the see-through top bar, its
 * preview playing on a loop (silent, with sound on headphones), then why it's picked, its title and
 * who it's from, and Play. Motion and anticipation are what pull people in (Berridge's "wanting"),
 * and the hero changes on every open so it never goes stale.
 */
@Composable
fun HomeHero(pick: RecommendedVideo, onOpen: () -> Unit) {
    val video = pick.video
    Box(Modifier.fillMaxWidth().aspectRatio(HERO_RATIO).pressClickable(onOpen)) {
        RemoteImage(video.coverUrl, Modifier.matchParentSize())
        val previewUrl = video.previewUrl
        if (previewUrl != null) {
            ScenePreview(previewUrl, Modifier.matchParentSize())
        }
        Box(Modifier.matchParentSize().background(HeroFade))
        Column(
            Modifier.align(Alignment.BottomStart).gutterPadding().padding(bottom = VeilSpacing.large),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        ) {
            val reason = pick.reason
            if (reason != null) {
                Text(reason, style = VeilType.cardMeta, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
            }
            Text(video.title, style = MaterialTheme.typography.headlineSmall, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val byline = video.byline
            if (byline != null) {
                Text(byline, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted, maxLines = 1)
            }
            Row(Modifier.padding(top = VeilSpacing.small)) {
                PrimaryButton("Play", onClick = onOpen, icon = VeilIcons.Play)
            }
        }
    }
}

/**
 * Who's live now as round faces ringed in pink, right under the hero: live rooms are the real
 * uncertainty of what's happening now, without any made-up urgency. Hidden while nobody's live.
 */
@Composable
fun LiveFacesRow(rooms: List<CamSummary>, onOpen: (CamSummary) -> Unit) {
    if (rooms.isEmpty()) {
        return
    }
    Column(Modifier.fillMaxWidth()) {
        SectionHeading("Live now", modifier = Modifier.gutterPadding())
        LazyRow(
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        ) {
            items(rooms, key = { room -> room.id }) { room ->
                Column(
                    Modifier.width(LiveFaceWidth).pressClickable { onOpen(room) },
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
                ) {
                    CamAvatar(room.pictureUrl, live = true)
                    Text(
                        room.name,
                        style = VeilType.cardMeta,
                        color = VeilColors.content,
                        textAlign = TextAlign.Center,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }
        }
    }
}
