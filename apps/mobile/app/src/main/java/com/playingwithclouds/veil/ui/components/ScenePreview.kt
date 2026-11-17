package com.playingwithclouds.veil.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.media3.common.C
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.SeekParameters
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import com.playingwithclouds.veil.api.ServerSettings

/** Loudness of a preview's sound with headphones in: present, not startling. */
private const val HEADPHONE_VOLUME = 0.6f

/**
 * A scene's preview clip playing on a loop, fading in over the poster once its first frame is
 * drawn; when the clip cannot be loaded nothing is drawn and the poster stays. It's silent, except
 * with headphones in when [sound] allows (one preview on screen at a time should). While
 * [scrubFraction] is set the clip holds still at that share of its
 * length (a finger dragged across the card), and plays on once it's null. The player lives only
 * while this is composed, so at most the previewing card holds one.
 */
@Composable
fun ScenePreview(
    previewVideo: String,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
    scrubFraction: Float? = null,
    sound: Boolean = true,
) {
    val context = LocalContext.current
    val headphones = rememberHeadphonesConnected()
    var firstFrameShown by remember(previewVideo) { mutableStateOf(false) }
    var failed by remember(previewVideo) { mutableStateOf(false) }
    val alpha by animateFloatAsState(if (firstFrameShown) 1f else 0f, label = "preview")
    val player = remember(previewVideo) {
        ExoPlayer.Builder(context).build().apply {
            volume = 0f
            repeatMode = Player.REPEAT_MODE_ONE
            setSeekParameters(SeekParameters.CLOSEST_SYNC)
        }
    }
    DisposableEffect(player) {
        val url = ServerSettings.imageUrl(previewVideo)
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                firstFrameShown = true
            }

            override fun onPlayerError(error: PlaybackException) {
                failed = true
            }
        }
        player.addListener(listener)
        if (url != null) {
            player.setMediaItem(MediaItem.fromUri(url))
            player.prepare()
            player.play()
        }
        onDispose {
            player.removeListener(listener)
            player.release()
        }
    }
    LaunchedEffect(player, headphones, sound) {
        player.volume = previewVolume(headphones && sound)
    }
    LaunchedEffect(player, scrubFraction) {
        holdOrPlay(player, scrubFraction)
    }
    if (failed) {
        return
    }
    ContentFrame(
        player = player,
        modifier = modifier.alpha(alpha),
        surfaceType = SURFACE_TYPE_TEXTURE_VIEW,
        contentScale = contentScale,
    )
}

/** A preview's volume: audible only when it may be heard (headphones in, sound allowed). */
private fun previewVolume(audible: Boolean): Float {
    if (audible) {
        return HEADPHONE_VOLUME
    }
    return 0f
}

/** Holds the clip at [scrubFraction] of its length, or plays it when there's none. */
private fun holdOrPlay(player: ExoPlayer, scrubFraction: Float?) {
    if (scrubFraction == null) {
        player.play()
        return
    }
    player.pause()
    val duration = player.duration
    if (duration != C.TIME_UNSET && duration > 0) {
        player.seekTo((duration * scrubFraction.coerceIn(0f, 1f)).toLong())
    }
}
