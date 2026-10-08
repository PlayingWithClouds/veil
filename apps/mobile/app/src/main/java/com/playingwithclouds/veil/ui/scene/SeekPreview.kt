package com.playingwithclouds.veil.ui.scene

import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.width
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.media3.common.C
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.SeekParameters
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.conflate

/** Width of the frame shown above the thumb while scrubbing. */
private val SeekPreviewWidth = 176.dp

/** Pause after each preview seek, so a fast drag asks for a few frames rather than every position. */
private const val SEEK_PREVIEW_INTERVAL_MILLISECONDS = 90L

/**
 * A muted second player on the same stream that shows the frame at a scrub position: lowest
 * bitrate, seeking to the nearest keyframe so frames come quickly. Prepared on first use, so pages
 * where nobody scrubs never load it.
 */
class SeekPreview(private val context: Context, private val stream: PlayableStream) {

    /** The preview player, once scrubbing started. */
    var player by mutableStateOf<ExoPlayer?>(null)
        private set

    /** Shows the frame at [positionMilliseconds], preparing the player first if needed. */
    fun show(positionMilliseconds: Long) {
        var current = player
        if (current == null) {
            current = createPlayer()
            player = current
        }
        current.seekTo(positionMilliseconds)
    }

    /** Frees the player and its decoder. */
    fun release() {
        player?.release()
        player = null
    }

    /** A paused, silent, lowest-quality player on the stream. */
    private fun createPlayer(): ExoPlayer {
        return ExoPlayer.Builder(context).build().apply {
            volume = 0f
            setSeekParameters(SeekParameters.CLOSEST_SYNC)
            trackSelectionParameters = trackSelectionParameters.buildUpon()
                .setForceLowestBitrate(true)
                .setTrackTypeDisabled(C.TRACK_TYPE_AUDIO, true)
                .build()
            setMediaSource(streamMediaSource(stream))
            playWhenReady = false
            prepare()
        }
    }
}

/** The seek preview of the player on screen; null where there is none. */
val LocalSeekPreview = staticCompositionLocalOf<SeekPreview?> { null }

/** A seek preview for [stream] that is released when it leaves the screen or the stream changes. */
@Composable
fun rememberSeekPreview(stream: PlayableStream): SeekPreview {
    val context = LocalContext.current
    val preview = remember(stream) { SeekPreview(context, stream) }
    DisposableEffect(preview) { onDispose { preview.release() } }
    return preview
}

/** The frame at [positionMilliseconds] in a small rounded 16:9 box, following the position as it changes. */
@Composable
fun SeekPreviewFrame(preview: SeekPreview, positionMilliseconds: Long, modifier: Modifier = Modifier) {
    val position by rememberUpdatedState(positionMilliseconds)
    LaunchedEffect(preview) {
        snapshotFlow { position }.conflate().collect { target ->
            preview.show(target)
            delay(SEEK_PREVIEW_INTERVAL_MILLISECONDS)
        }
    }
    Box(
        modifier
            .width(SeekPreviewWidth)
            .aspectRatio(16f / 9f)
            .clip(VeilShapes.small)
            .background(VeilColors.canvas)
            .border(1.dp, VeilColors.glassEdge, VeilShapes.small),
    ) {
        val player = preview.player
        if (player != null) {
            ContentFrame(
                player = player,
                modifier = Modifier.fillMaxSize(),
                surfaceType = SURFACE_TYPE_TEXTURE_VIEW,
                contentScale = ContentScale.Crop,
            )
        }
    }
}
