package com.playingwithclouds.veil.ui.compilation

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import androidx.media3.ui.compose.state.rememberProgressStateWithTickInterval
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.ui.video.clippedMediaSource
import com.playingwithclouds.veil.ui.video.rememberVideoPlayer
import com.playingwithclouds.veil.util.formatClock

private const val PLAYER_ASPECT_RATIO = 16f / 9f
private const val PROGRESS_TICK_MILLISECONDS = 500L
private const val MILLISECONDS_PER_SECOND = 1000.0

/** The small spinner beside the buttons while more scenes are prepared. */
private val LoadingSpinnerSize = 20.dp

/**
 * The compilation mode: plays only the scenes the user likes, one stretch after another and across
 * videos (the one it was opened from, the queue, then the best-matching videos), jumping by itself.
 * Tap the picture to pause; Next scene skips; Open the video goes to the whole video at that point.
 */
@Composable
fun CompilationScreen(firstVideoId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "compilation:$firstVideoId") { CompilationViewModel(firstVideoId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val player = rememberVideoPlayer()
    var playingIndex by remember { mutableIntStateOf(0) }
    FeedPlaylist(player, state.clips) { index ->
        playingIndex = index
        viewModel.onClipPlaying(index)
    }
    val playing = state.clips.getOrNull(playingIndex)
    ShowAmbientGlow(playing?.video?.coverUrl)
    DetailScreen("Compilation", onBack = navigator::back) {
        Column(Modifier.padding(top = LocalTopBarInset.current), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            Box(Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO).background(VeilColors.canvas), contentAlignment = Alignment.Center) {
                if (state.clips.isEmpty()) {
                    CompilationPlaceholder(state)
                } else {
                    PlayerPicture(player)
                }
            }
            if (playing != null) {
                ClipCaption(playing, player, playingIndex, state, navigator)
            }
        }
    }
}

/** Why nothing plays yet: a spinner while preparing, else the problem. */
@Composable
private fun CompilationPlaceholder(state: CompilationState) {
    val problem = state.problem
    if (problem != null) {
        EmptyMessage(problem)
        return
    }
    LoadingIndicator()
}

/** The picture; a tap pauses or plays. */
@Composable
private fun PlayerPicture(player: ExoPlayer) {
    Box(Modifier.fillMaxSize().pressClickable { player.playWhenReady = !player.playWhenReady }) {
        ContentFrame(player = player, modifier = Modifier.fillMaxSize(), surfaceType = SURFACE_TYPE_TEXTURE_VIEW)
    }
}

/** The playing clip: its video, its scenes and where it is, the clip's progress, Next scene and Open the video. */
@Composable
private fun ClipCaption(playing: ReadyClip, player: ExoPlayer, index: Int, state: CompilationState, navigator: AppNavigator) {
    val progress = rememberProgressStateWithTickInterval(player, PROGRESS_TICK_MILLISECONDS)
    val clipLength = playing.clip.endSeconds - playing.clip.startSeconds
    val playedSeconds = progress.currentPositionMs / MILLISECONDS_PER_SECOND
    Column(Modifier.padding(horizontal = VeilSpacing.gutter), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        ProgressBar((playedSeconds / clipLength).toFloat())
        Text(playing.video.title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2)
        var place = "Scene ${index + 1} of ${state.clips.size}"
        if (!state.finished) {
            place = "$place+"
        }
        val scenes = playing.clip.tags.joinToString(" · ")
        Text("$scenes · ${formatClock(playing.clip.startSeconds)} · $place", style = VeilType.cardMeta, color = VeilColors.contentMuted)
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
            SecondaryButton("Next scene", onClick = { player.seekToNextMediaItem() }, enabled = player.hasNextMediaItem())
            PrimaryButton("Open the video", onClick = {
                player.pause()
                navigator.openVideoAt(playing.video.id, playing.clip.startSeconds + playedSeconds)
            })
            if (state.loading) {
                Spinner(size = LoadingSpinnerSize)
            }
        }
        Text("Playing the scenes you like: ${state.likedTags.sorted().joinToString(", ")}", style = MaterialTheme.typography.bodySmall, color = VeilColors.contentFaint)
    }
}

/**
 * Keeps [player]'s playlist in step with [clips] (each a stretch of its stream, appended as it's
 * prepared; the first starts playback) and reports the index playing.
 */
@Composable
internal fun FeedPlaylist(player: ExoPlayer, clips: List<ReadyClip>, onPlaying: (Int) -> Unit) {
    LaunchedEffect(clips.size) {
        val first = player.mediaItemCount
        for (clip in clips.drop(first)) {
            player.addMediaSource(clippedMediaSource(clip.stream, clip.clip.startSeconds, clip.clip.endSeconds))
        }
        if (first == 0 && clips.isNotEmpty()) {
            player.prepare()
            player.playWhenReady = true
        }
    }
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onMediaItemTransition(mediaItem: MediaItem?, reason: Int) {
                onPlaying(player.currentMediaItemIndex)
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
    }
}
