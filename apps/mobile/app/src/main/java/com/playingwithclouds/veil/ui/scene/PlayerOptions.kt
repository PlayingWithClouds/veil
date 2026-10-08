package com.playingwithclouds.veil.ui.scene

import android.content.Context
import android.content.SharedPreferences
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import androidx.core.content.edit
import androidx.media3.common.C
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.SeekParameters
import kotlinx.coroutines.delay

/** How often an active A–B loop checks whether playback reached B. */
private const val LOOP_CHECK_INTERVAL_MILLISECONDS = 80L

private const val PREFERENCES = "player"
private const val KEY_ZOOM_TO_FILL = "zoom_to_fill"
private const val KEY_PIP_PRIVACY = "pip_privacy"

/**
 * What the viewer can switch on the player besides speed and quality: the A–B loop of this page
 * and the zoom-to-fill choice, which is remembered across scenes.
 */
@Stable
class PlayerOptions(private val preferences: SharedPreferences) {

    /** The loop between two marked positions; inactive until both are set. */
    var loop by mutableStateOf(AbLoop())
        private set

    /** Fills the player box by cropping the video's edges instead of letterboxing. */
    var zoomToFill by mutableStateOf(preferences.getBoolean(KEY_ZOOM_TO_FILL, false))
        private set

    /** What the picture-in-picture window shows of the video; remembered across scenes. */
    var pictureInPicturePrivacy by mutableStateOf(readPrivacy())
        private set

    /** Changes what the picture-in-picture window shows and remembers it. */
    fun updatePictureInPicturePrivacy(privacy: PictureInPicturePrivacy) {
        pictureInPicturePrivacy = privacy
        preferences.edit { putString(KEY_PIP_PRIVACY, privacy.name) }
    }

    /** The stored privacy choice; blurred until the viewer decides otherwise. */
    private fun readPrivacy(): PictureInPicturePrivacy {
        val stored = preferences.getString(KEY_PIP_PRIVACY, null)
        return PictureInPicturePrivacy.entries.firstOrNull { privacy -> privacy.name == stored } ?: PictureInPicturePrivacy.BLUR
    }

    /** Marks the player's current position as A, then B, then clears the loop. */
    fun advanceLoop(player: Player) {
        loop = advanceLoop(loop, player.currentPosition)
    }

    /** Drops the loop. */
    fun clearLoop() {
        loop = AbLoop()
    }

    /** Switches zoom to fill and remembers it. */
    fun updateZoomToFill(enabled: Boolean) {
        zoomToFill = enabled
        preferences.edit { putBoolean(KEY_ZOOM_TO_FILL, enabled) }
    }
}

/** The page's [PlayerOptions], jumping back to A whenever playback reaches B of an active loop. */
@Composable
fun rememberPlayerOptions(player: Player): PlayerOptions {
    val context = LocalContext.current
    val options = remember { PlayerOptions(context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)) }
    val loop = options.loop
    LaunchedEffect(loop, player) {
        if (!loop.isActive) {
            return@LaunchedEffect
        }
        while (true) {
            delay(LOOP_CHECK_INTERVAL_MILLISECONDS)
            val target = loopRestartTarget(loop, player.currentPosition)
            if (target != null) {
                player.seekTo(target)
            }
        }
    }
    return options
}

/**
 * Moves a paused player one frame ([direction] -1 or 1), pausing it first when it plays. Seeks
 * exactly, so the frame shown is the one asked for rather than the nearest keyframe.
 */
fun Player.stepFrame(direction: Int) {
    pause()
    if (this is ExoPlayer) {
        setSeekParameters(SeekParameters.EXACT)
    }
    val frameRate = videoFormatFrameRate()
    val delta = direction * frameDurationMilliseconds(frameRate)
    seekTo(clampedSeekTarget(currentPosition, delta, duration))
}

/** The frame rate of the playing video as the stream reports it; 0 when unknown. */
private fun Player.videoFormatFrameRate(): Float {
    for (group in currentTracks.groups) {
        if (group.type == C.TRACK_TYPE_VIDEO && group.isSelected) {
            return group.getTrackFormat(0).frameRate.coerceAtLeast(0f)
        }
    }
    return 0f
}
