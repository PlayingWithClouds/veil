package com.playingwithclouds.veil.ui.cams

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.repeatOnLifecycle
import com.playingwithclouds.veil.data.CamModel
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.glassControl
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatClock
import kotlinx.coroutines.delay

/** Cam sites send 4:3 room snapshots. */
private const val ROOM_THUMBNAIL_RATIO = 4f / 3f

/** Adaptive cell width of the room grid: two columns on phones. */
val CamCellWidth = 160.dp

/** How often running recordings and followed rooms are checked again while a cam screen is shown. */
const val CAM_POLL_MILLISECONDS = 10_000L

/** Runs [action] now and every [intervalMilliseconds] while the screen is started (in front). */
@Composable
fun PollWhileStarted(intervalMilliseconds: Long, action: () -> Unit) {
    val lifecycleOwner = LocalLifecycleOwner.current
    val currentAction by rememberUpdatedState(action)
    LaunchedEffect(lifecycleOwner) {
        lifecycleOwner.repeatOnLifecycle(Lifecycle.State.STARTED) {
            while (true) {
                currentAction()
                delay(intervalMilliseconds)
            }
        }
    }
}

/**
 * A live room as a card: its snapshot with the viewer count, REC while the backend records it and
 * AUTO while it is recorded automatically, then the name and the room's subject.
 */
@Composable
fun CamCard(room: CamModel, recording: Boolean, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(onClick)) {
        Box(Modifier.fillMaxWidth().aspectRatio(ROOM_THUMBNAIL_RATIO).clip(VeilShapes.card)) {
            RemoteImage(room.thumbnailUrl, Modifier.matchParentSize())
            Badge("${formatViewers(room.viewers)} watching", Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            Row(Modifier.align(Alignment.TopStart).padding(VeilSpacing.small), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                if (recording) {
                    Badge("REC", background = VeilColors.error)
                }
                if (room.autoRecord) {
                    Badge("AUTO")
                }
            }
        }
        Column(Modifier.padding(horizontal = VeilSpacing.extraSmall, vertical = VeilSpacing.small)) {
            Text(room.displayName, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
            val subject = room.subject
            if (!subject.isNullOrBlank()) {
                Text(subject, style = VeilType.cardMeta, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
            }
        }
    }
}

/** Width of a room in an avatar row, wider than the photo so names fit. */
private val CamAvatarCellWidth = 76.dp

/** Diameter of a room's photo in an avatar row. */
private val CamAvatarSize = 64.dp

/** A room's round photo, ringed in red while it is live. */
@Composable
fun CamAvatar(thumbnailUrl: String?, live: Boolean, modifier: Modifier = Modifier) {
    var ringed = modifier.size(CamAvatarSize)
    if (live) {
        ringed = ringed.border(LiveRingWidth, VeilColors.error, VeilShapes.capsule).padding(LiveRingWidth + VeilSpacing.hairline)
    }
    Avatar(thumbnailUrl, ringed)
}

private val LiveRingWidth = 2.dp

/** Followed rooms as round photos, live ones ringed, running edge to edge; nothing when none are followed. */
@Composable
fun FollowedCamsRow(followed: List<CamModel>, onOpen: (CamModel) -> Unit, onSeeAll: (() -> Unit)? = null) {
    if (followed.isEmpty()) {
        return
    }
    SectionHeading("Followed cams", onClick = onSeeAll)
    LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        items(followed, key = { model -> model.id }) { model ->
            Column(
                Modifier.width(CamAvatarCellWidth).pressClickable { onOpen(model) },
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            ) {
                CamAvatar(model.thumbnailUrl, live = model.online)
                Text(
                    model.displayName,
                    style = MaterialTheme.typography.labelMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    textAlign = TextAlign.Center,
                )
            }
        }
    }
}

/** The running recordings: photo, name, time so far and quality, with a stop button each. */
@Composable
fun RecordingStrip(recordings: List<CamRecording>, onOpen: (CamRecording) -> Unit, onStop: (CamRecording) -> Unit) {
    if (recordings.isEmpty()) {
        return
    }
    SectionHeading("Recording")
    Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        for (recording in recordings) {
            RecordingRow(recording, onOpen = { onOpen(recording) }, onStop = { onStop(recording) })
        }
    }
}

/** One running recording. */
@Composable
private fun RecordingRow(recording: CamRecording, onOpen: () -> Unit, onStop: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(onOpen).glassControl(VeilShapes.panel).padding(VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Avatar(recording.thumbnailUrl, Modifier.size(40.dp))
        Column(Modifier.weight(1f)) {
            Text(recording.displayName, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(RecordingDotSize).clip(VeilShapes.capsule).background(VeilColors.error))
                Text(recordingFacts(recording), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
            }
        }
        IconTap(VeilIcons.Stop, contentDescription = "Stop recording", onClick = onStop, tint = VeilColors.error)
    }
}

private val RecordingDotSize = 8.dp

/** "12:34 · 720p": how long it runs and the rendition. */
fun recordingFacts(recording: CamRecording): String {
    val facts = mutableListOf(formatClock(recording.durationSeconds))
    val quality = recording.quality
    if (quality != null) {
        facts.add(quality)
    }
    return facts.joinToString(" · ")
}

/** The live dot with "LIVE" or "Offline", for a room's page. */
@Composable
fun OnlineLabel(online: Boolean) {
    if (online) {
        Badge("LIVE", background = VeilColors.error)
        return
    }
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall), verticalAlignment = Alignment.CenterVertically) {
        Icon(VeilIcons.Cams, contentDescription = null, tint = VeilColors.contentMuted, modifier = Modifier.size(14.dp))
        Text("Offline", style = MaterialTheme.typography.labelMedium, color = VeilColors.contentMuted)
    }
}
