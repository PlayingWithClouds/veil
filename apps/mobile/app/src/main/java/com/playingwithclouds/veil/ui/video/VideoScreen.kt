package com.playingwithclouds.veil.ui.video

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.SnackbarResult
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextAlign
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** The player box's shape on the page; fullscreen and picture-in-picture fill the screen instead. */
private const val PLAYER_ASPECT_RATIO = 16f / 9f

/**
 * A video's page: the player under the top bar, then what the video is, what to do with it and Up
 * next. Fullscreen and picture-in-picture show only the player. Watching is logged for as long as
 * the page is open.
 */
@Composable
fun VideoScreen(videoId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = videoId) { VideoViewModel(videoId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val player = rememberVideoPlayer()
    val fullscreen = rememberFullscreenState(player)
    val pictureInPicture = rememberPictureInPicture(player)
    val options = rememberPlayerOptions(player)
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(state.active) {
        val active = state.active ?: return@LaunchedEffect
        // Switching sources keeps the place; the first one starts at the resume point.
        var startAt = viewModel.startPositionMilliseconds()
        if (player.mediaItemCount > 0) {
            startAt = player.currentPosition
        }
        player.playStream(active.playable, startAt)
    }
    PlaybackErrors(player, onError = viewModel::onPlayerError)
    WatchTracking(player, onEvent = viewModel::onWatchEvent)
    VideoSnackbars(state, viewModel, snackbarHostState)

    val detail = state.detail
    if (detail is LoadState.Failed) {
        Scaffold(topBar = { VeilTopBar("Video", onBack = navigator::back) }) { padding ->
            FailedMessage(detail.message, onRetry = viewModel::load, modifier = Modifier.padding(padding))
        }
        return
    }
    val immersive = fullscreen.isFullscreen || pictureInPicture.isActive
    val title = (detail as? LoadState.Loaded)?.value?.title.orEmpty()
    ShowAmbientGlow((detail as? LoadState.Loaded)?.value?.summary?.coverUrl)
    Scaffold(
        topBar = {
            if (!immersive) {
                VeilTopBar(title, onBack = navigator::back)
            }
        },
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            val extras = PlayerExtras(
                markers = markersInPlayOrder((detail as? LoadState.Loaded)?.value?.segments.orEmpty()),
                onEnterPictureInPicture = pictureInPicture::enterPictureInPicture,
            )
            var playerModifier = Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO)
            if (immersive) {
                playerModifier = Modifier.fillMaxSize()
            }
            Box(playerModifier.background(VeilColors.canvas), contentAlignment = Alignment.Center) {
                PlayerArea(state, player, title, fullscreen, pictureInPicture, options, extras, navigator)
            }
            if (!immersive) {
                VideoContent(state, viewModel, player, navigator)
            }
        }
    }
}

/** The video once a source plays, else the cover with a spinner or why nothing plays. */
@Composable
private fun PlayerArea(
    state: VideoState,
    player: ExoPlayer,
    title: String,
    fullscreen: FullscreenState,
    pictureInPicture: PictureInPictureState,
    options: PlayerOptions,
    extras: PlayerExtras,
    navigator: AppNavigator,
) {
    val active = state.active
    if (active != null) {
        VideoPlayerView(
            player,
            stream = active.playable,
            title = title,
            fullscreen = fullscreen,
            pictureInPicture = pictureInPicture,
            options = options,
            extras = extras,
            related = state.related,
            onOpenVideo = { video -> navigator.openVideo(video.id) },
            modifier = Modifier.fillMaxSize(),
        )
        return
    }
    val detail = (state.detail as? LoadState.Loaded)?.value
    if (detail == null) {
        LoadingIndicator()
        return
    }
    RemoteImage(detail.summary.coverUrl, Modifier.fillMaxSize(), ContentScale.Crop)
    Box(Modifier.fillMaxSize().background(VeilColors.scrim), contentAlignment = Alignment.Center) {
        val problem = state.problem
        if (problem != null && !state.resolving) {
            Text(
                problem,
                Modifier.padding(VeilSpacing.large),
                color = VeilColors.content,
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
            )
            return@Box
        }
        Spinner()
    }
}

/** Shows the page's notes, and Undo for a thumbs down before it's logged for good. */
@Composable
private fun VideoSnackbars(state: VideoState, viewModel: VideoViewModel, snackbarHostState: SnackbarHostState) {
    LaunchedEffect(state.message) {
        val note = state.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    LaunchedEffect(state.pendingDislike) {
        if (!state.pendingDislike) {
            return@LaunchedEffect
        }
        val result = snackbarHostState.showSnackbar(
            "You won't be recommended this video again.",
            actionLabel = "Undo",
            duration = SnackbarDuration.Short,
        )
        if (result == SnackbarResult.ActionPerformed) {
            viewModel.undoDislike()
            return@LaunchedEffect
        }
        viewModel.settleDislike()
    }
}
