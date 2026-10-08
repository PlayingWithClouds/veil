package com.playingwithclouds.veil.ui.scene

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.ResumePoint
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.theme.VeilColors

private const val PLAYER_ASPECT_RATIO = 16f / 9f
private const val MILLISECONDS_PER_SECOND = 1000L

/** A scene's watch page: player pinned on top, details and reactions below, fullscreen on demand. */
@Composable
fun SceneScreen(sceneId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "scene-$sceneId") { SceneViewModel(sceneId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val player = rememberScenePlayer()
    val fullscreen = rememberFullscreenState()
    val snackbarHostState = remember { SnackbarHostState() }
    val detail = state.detail
    var showingAlternates by remember { mutableStateOf(false) }
    val findAlternates = {
        viewModel.findAlike()
        showingAlternates = true
    }

    if (showingAlternates) {
        AlternatesSheet(viewModel, onDismiss = { showingAlternates = false })
    }
    LaunchedEffect(state.active) {
        val active = state.active ?: return@LaunchedEffect
        player.playStream(active.playable, startPosition(player, state.resume))
    }
    LaunchedEffect(state.message) {
        val note = state.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    PlaybackReporting(player, onError = viewModel::onPlayerError, onSaveProgress = viewModel::saveProgress)
    DisposableEffect(Unit) { onDispose { viewModel.refreshProgressStore() } }

    if (detail is LoadState.Failed) {
        Scaffold(topBar = { VeilTopBar("Video", onBack = navigator::back) }) { padding ->
            FailedMessage(detail.message, onRetry = viewModel::load, modifier = Modifier.padding(padding))
        }
        return
    }
    Scaffold(
        topBar = {
            if (!fullscreen.isFullscreen) {
                VeilTopBar(title = (detail as? LoadState.Loaded)?.value?.title.orEmpty(), onBack = navigator::back)
            }
        },
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            PlayerArea(state, player, fullscreen, onFindAlternates = findAlternates)
            if (!fullscreen.isFullscreen) {
                SceneContent(state, viewModel, player, navigator, onFindAlternates = findAlternates)
            }
        }
    }
}

/** Where playback starts: the current position when switching sources, else the saved resume point. */
private fun startPosition(player: ExoPlayer, resume: ResumePoint?): Long {
    if (player.mediaItemCount > 0) {
        return player.currentPosition
    }
    if (resume == null || resume.completed) {
        return 0
    }
    val duration = resume.durationSeconds
    if (duration != null && resume.progressSeconds >= duration * SceneRepository.COMPLETE_FRACTION) {
        return 0
    }
    return resume.progressSeconds * MILLISECONDS_PER_SECOND
}

/** The black player box: the video when a source plays, else the poster with a spinner or the problem. */
@Composable
private fun PlayerArea(state: SceneState, player: ExoPlayer, fullscreen: FullscreenState, onFindAlternates: () -> Unit) {
    var boxModifier = Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO)
    if (fullscreen.isFullscreen) {
        boxModifier = Modifier.fillMaxSize()
    }
    Box(boxModifier.background(Color.Black), contentAlignment = Alignment.Center) {
        val active = state.active
        if (active != null) {
            ScenePlayerView(player, onToggleFullscreen = fullscreen::toggle, modifier = Modifier.fillMaxSize())
        }
        if (active == null) {
            PlayerPlaceholder(state, onFindAlternates)
        }
        val problem = state.playbackProblem
        if (problem != null) {
            Text(
                problem,
                modifier = Modifier.align(Alignment.BottomCenter).background(Color.Black.copy(alpha = 0.7f)).padding(8.dp),
                color = VeilColors.error,
                style = MaterialTheme.typography.bodySmall,
                textAlign = TextAlign.Center,
            )
        }
    }
}

/** Shown until a source plays: the poster, and a spinner while sources are being found. */
@Composable
private fun PlayerPlaceholder(state: SceneState, onFindAlternates: () -> Unit) {
    val detail = (state.detail as? LoadState.Loaded)?.value
    if (detail == null) {
        LoadingIndicator()
        return
    }
    RemoteImage(detail.posterPath, Modifier.fillMaxSize(), ContentScale.Crop)
    Box(Modifier.fillMaxSize().background(Color.Black.copy(alpha = 0.5f)), contentAlignment = Alignment.Center) {
        val searching = state.fetchingDetail || state.resolvingStreamId != null
        if (searching) {
            Spinner()
            return@Box
        }
        if (state.streams.isEmpty()) {
            Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("No playable source found", color = Color.White, style = MaterialTheme.typography.titleSmall)
                PrimaryButton("Find other sources", onClick = onFindAlternates, icon = VeilIcons.Alternates)
            }
        }
    }
}
