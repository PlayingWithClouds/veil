package com.playingwithclouds.veil.ui.components

import android.util.LruCache
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.produceState
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.platform.LocalContext
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.SeekParameters
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.PreviewPlan
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.scenes.PreviewMoments
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.video.clippedMediaSource
import kotlinx.coroutines.delay

/** How long each still of the flipbook shows. */
private const val FLIPBOOK_MILLISECONDS = 700L

/** Preview plans kept in memory, so a card scrolled back to doesn't read again. */
private val previewPlans = LruCache<String, PreviewPlan>(100)

/**
 * A card's preview while it's the one previewing. With the scene model's segments, it jumps
 * through 4–6 short moments across the video, chosen from the scenes the user likes most (seeking
 * in the stream, nothing is cut), even when the site has a preview clip. Else the site's clip;
 * else, without one, the site's stills across the video as a flipbook. Silent.
 */
@Composable
fun CardPreview(video: VideoSummary, modifier: Modifier = Modifier) {
    val modelState by SceneModelStore.state.collectAsState()
    val liked by SceneRepository.likedTags.collectAsState()
    val previewUrl = video.previewUrl
    val modelReady = modelState.ready != null
    val wantsPlan = modelReady || previewUrl == null
    LaunchedEffect(modelReady) {
        if (modelReady) {
            SceneRepository.ensureLikedTags()
        }
    }
    val plan by produceState<PreviewPlan?>(previewPlans.get(video.id), video.id, wantsPlan) {
        if (value != null || !wantsPlan) {
            return@produceState
        }
        val loaded = runCatching { SceneRepository.previewPlan(video.id) }.getOrNull() ?: return@produceState
        previewPlans.put(video.id, loaded)
        value = loaded
    }
    val moments = remember(plan, liked) { PreviewMoments.pick(plan?.segments.orEmpty(), liked) }
    if (moments.isNotEmpty()) {
        MomentPreview(video.id, moments, modifier)
        return
    }
    if (previewUrl != null) {
        ScenePreview(previewUrl, modifier, sound = false)
        return
    }
    val frames = plan?.frames.orEmpty()
    if (frames.isNotEmpty()) {
        Flipbook(frames, modifier)
    }
}

/** Plays [PreviewMoments.MOMENT_SECONDS] at each moment of the video's stream in turn, on a loop, fading in over the cover. */
@Composable
private fun MomentPreview(videoId: String, moments: List<Double>, modifier: Modifier) {
    val stream by produceState<PlayableStream?>(null, videoId) {
        value = runCatching { SceneRepository.playable(videoId, readPages = false) }.getOrNull()?.stream
    }
    val playable = stream ?: return
    val context = LocalContext.current
    var firstFrameShown by remember(videoId) { mutableStateOf(false) }
    val alpha by animateFloatAsState(if (firstFrameShown) 1f else 0f, label = "moments")
    val player = remember(videoId) {
        ExoPlayer.Builder(context).build().apply {
            volume = 0f
            repeatMode = Player.REPEAT_MODE_ALL
            setSeekParameters(SeekParameters.CLOSEST_SYNC)
        }
    }
    DisposableEffect(player, playable) {
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                firstFrameShown = true
            }
        }
        player.addListener(listener)
        for (moment in moments) {
            player.addMediaSource(clippedMediaSource(playable, moment, moment + PreviewMoments.MOMENT_SECONDS))
        }
        player.prepare()
        player.play()
        onDispose {
            player.removeListener(listener)
            player.release()
        }
    }
    ContentFrame(player = player, modifier = modifier.alpha(alpha), surfaceType = SURFACE_TYPE_TEXTURE_VIEW)
}

/** The site's stills across the video, one after another on a loop. */
@Composable
private fun Flipbook(frames: List<String>, modifier: Modifier) {
    var index by remember(frames) { mutableIntStateOf(0) }
    LaunchedEffect(frames) {
        while (true) {
            delay(FLIPBOOK_MILLISECONDS)
            index = (index + 1) % frames.size
        }
    }
    RemoteImage(frames[index], modifier)
}
