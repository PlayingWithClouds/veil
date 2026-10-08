package com.playingwithclouds.veil.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import com.playingwithclouds.veil.api.ServerSettings

/**
 * A scene's preview clip playing muted on a loop, fading in over the poster once its first frame
 * is drawn. The player lives only while this is composed, so at most the previewing card holds one.
 */
@Composable
fun ScenePreview(previewVideo: String, modifier: Modifier = Modifier, contentScale: ContentScale = ContentScale.Crop) {
    val context = LocalContext.current
    var firstFrameShown by remember(previewVideo) { mutableStateOf(false) }
    val alpha by animateFloatAsState(if (firstFrameShown) 1f else 0f, label = "preview")
    val player = remember(previewVideo) {
        ExoPlayer.Builder(context).build().apply {
            volume = 0f
            repeatMode = Player.REPEAT_MODE_ONE
        }
    }
    DisposableEffect(player) {
        val url = ServerSettings.imageUrl(previewVideo)
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                firstFrameShown = true
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
    ContentFrame(
        player = player,
        modifier = modifier.alpha(alpha),
        surfaceType = SURFACE_TYPE_TEXTURE_VIEW,
        contentScale = contentScale,
    )
}
