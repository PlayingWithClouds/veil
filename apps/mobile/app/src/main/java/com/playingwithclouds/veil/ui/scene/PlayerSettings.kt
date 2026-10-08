package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.Dp
import androidx.media3.common.C
import androidx.media3.common.Player
import androidx.media3.common.TrackSelectionOverride
import androidx.media3.common.Tracks
import androidx.media3.ui.compose.state.rememberPlaybackSpeedState
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilMenu
import com.playingwithclouds.veil.ui.design.VeilMenuItem

/** Which list the settings menu shows. */
private enum class SettingsPage { MAIN, SPEED, QUALITY }

/**
 * A glass gear opening the playback settings: speed, and the video quality when the stream offers
 * several (HLS variants). [onOpenChange] reports the menu so the controls stay up while it is open.
 */
@Composable
fun PlayerSettingsButton(player: Player, buttonSize: Dp, onOpenChange: (Boolean) -> Unit, modifier: Modifier = Modifier) {
    var open by remember { mutableStateOf(false) }
    var page by remember { mutableStateOf(SettingsPage.MAIN) }
    val tracks = rememberPlayerTracks(player)
    val setOpen = { value: Boolean ->
        open = value
        page = SettingsPage.MAIN
        onOpenChange(value)
    }
    // A menu that vanishes with the controls must not keep them from hiding later.
    DisposableEffect(Unit) {
        onDispose {
            if (open) {
                onOpenChange(false)
            }
        }
    }
    Box(modifier) {
        RoundIconButton(VeilIcons.Settings, contentDescription = "Playback settings", onClick = { setOpen(true) }, size = buttonSize)
        VeilMenu(expanded = open, onDismissRequest = { setOpen(false) }) {
            when (page) {
                SettingsPage.MAIN -> MainSettings(player, tracks, onPage = { chosen -> page = chosen })
                SettingsPage.SPEED -> SpeedSettings(player, onDone = { setOpen(false) })
                SettingsPage.QUALITY -> QualitySettings(player, tracks, onDone = { setOpen(false) })
            }
        }
    }
}

/** The menu's first page: one entry per setting with its current value. */
@Composable
private fun MainSettings(player: Player, tracks: Tracks, onPage: (SettingsPage) -> Unit) {
    val speedState = rememberPlaybackSpeedState(player)
    VeilMenuItem("Speed · ${formatSpeed(speedState.playbackSpeed)}", onClick = { onPage(SettingsPage.SPEED) })
    if (videoQualities(tracks).size > 1) {
        VeilMenuItem("Quality · ${selectedQualityLabel(player, tracks)}", onClick = { onPage(SettingsPage.QUALITY) })
    }
}

/** The playback speeds, the current one ticked. */
@Composable
private fun SpeedSettings(player: Player, onDone: () -> Unit) {
    val speedState = rememberPlaybackSpeedState(player)
    for (speed in PLAYBACK_SPEEDS) {
        VeilMenuItem(
            formatSpeed(speed),
            onClick = {
                speedState.updatePlaybackSpeed(speed)
                onDone()
            },
            selected = speedState.playbackSpeed == speed,
        )
    }
}

/** Auto plus every video quality of the stream, the chosen one ticked. */
@Composable
private fun QualitySettings(player: Player, tracks: Tracks, onDone: () -> Unit) {
    val selectedHeight = selectedQualityHeight(player, tracks)
    VeilMenuItem(
        "Auto",
        onClick = {
            player.selectQuality(null)
            onDone()
        },
        selected = selectedHeight == null,
    )
    for (quality in videoQualities(tracks)) {
        VeilMenuItem(
            videoQualityLabel(quality.height),
            onClick = {
                player.selectQuality(quality)
                onDone()
            },
            selected = selectedHeight == quality.height,
        )
    }
}

/** The player's current tracks, updated as the stream's tracks become known or change. */
@Composable
private fun rememberPlayerTracks(player: Player): Tracks {
    var tracks by remember(player) { mutableStateOf(player.currentTracks) }
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onTracksChanged(changed: Tracks) {
                tracks = changed
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
    }
    return tracks
}

/** The selectable video qualities of [tracks], highest first. */
private fun videoQualities(tracks: Tracks): List<VideoQuality> {
    val found = mutableListOf<VideoQuality>()
    tracks.groups.forEachIndexed { groupIndex, group ->
        if (group.type == C.TRACK_TYPE_VIDEO) {
            found.addAll(qualitiesOf(group, groupIndex))
        }
    }
    return distinctQualities(found)
}

/** The supported tracks of one video group. */
private fun qualitiesOf(group: Tracks.Group, groupIndex: Int): List<VideoQuality> {
    return (0 until group.length)
        .filter { trackIndex -> group.isTrackSupported(trackIndex) }
        .map { trackIndex -> VideoQuality(group.getTrackFormat(trackIndex).height, groupIndex, trackIndex) }
}

/** Pins the video to [quality], or lets the player adapt again when it is null. */
private fun Player.selectQuality(quality: VideoQuality?) {
    val builder = trackSelectionParameters.buildUpon()
    if (quality == null) {
        trackSelectionParameters = builder.clearOverridesOfType(C.TRACK_TYPE_VIDEO).build()
        return
    }
    val group = currentTracks.groups[quality.groupIndex].mediaTrackGroup
    trackSelectionParameters = builder.setOverrideForType(TrackSelectionOverride(group, quality.trackIndex)).build()
}

/** The height the video is pinned to, or null on Auto (also when the pin belongs to a previous stream). */
private fun selectedQualityHeight(player: Player, tracks: Tracks): Int? {
    val override = player.trackSelectionParameters.overrides.values
        .firstOrNull { candidate -> candidate.type == C.TRACK_TYPE_VIDEO }
    if (override == null || override.trackIndices.isEmpty()) {
        return null
    }
    val current = tracks.groups.any { group -> group.mediaTrackGroup == override.mediaTrackGroup }
    if (!current) {
        return null
    }
    return override.mediaTrackGroup.getFormat(override.trackIndices.first()).height
}

/** The current quality for the menu: the pinned height, else Auto. */
private fun selectedQualityLabel(player: Player, tracks: Tracks): String {
    val height = selectedQualityHeight(player, tracks)
    if (height == null) {
        return "Auto"
    }
    return videoQualityLabel(height)
}
