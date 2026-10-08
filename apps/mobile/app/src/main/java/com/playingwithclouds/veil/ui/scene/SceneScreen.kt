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
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextAlign
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.PlaybackQueue
import com.playingwithclouds.veil.data.QueueEntry
import com.playingwithclouds.veil.data.ResumePoint
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.SceneSummary
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
import com.playingwithclouds.veil.ui.theme.VeilSpacing

private const val PLAYER_ASPECT_RATIO = 16f / 9f
private const val MILLISECONDS_PER_SECOND = 1000L

/** A scene's watch page: player pinned on top, details and reactions below, fullscreen on demand. */
@Composable
fun SceneScreen(sceneId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "scene-$sceneId") { SceneViewModel(sceneId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val player = rememberScenePlayer()
    val fullscreen = rememberFullscreenState(player)
    val pictureInPicture = rememberPictureInPicture(player)
    val options = rememberPlayerOptions(player)
    val queue by PlaybackQueue.state.collectAsStateWithLifecycle()
    var showingQueue by remember { mutableStateOf(false) }
    val playQueued = { entry: QueueEntry ->
        PlaybackQueue.remove(entry.sceneId)
        navigator.replaceWithScene(entry.sceneId)
    }
    val playNextInQueue = {
        val next = PlaybackQueue.takeNext()
        if (next != null) {
            navigator.replaceWithScene(next.sceneId)
        }
    }
    PlayNextWhenEnded(player, enabled = queue.entries.isNotEmpty(), onEnded = playNextInQueue)
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
    if (showingQueue) {
        QueueSheet(
            onPlay = { entry ->
                showingQueue = false
                playQueued(entry)
            },
            onDismiss = { showingQueue = false },
        )
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
    PlaybackHeatReporting(player, loopStartMilliseconds = { options.loop.startMilliseconds }, onReport = viewModel::recordHeat)
    PlaybackReporting(player, onError = viewModel::onPlayerError, onSaveProgress = viewModel::saveProgress)
    DisposableEffect(Unit) { onDispose { viewModel.refreshProgressStore() } }

    if (detail is LoadState.Failed) {
        Scaffold(topBar = { VeilTopBar("Video", onBack = navigator::back) }) { padding ->
            FailedMessage(detail.message, onRetry = viewModel::load, modifier = Modifier.padding(padding))
        }
        return
    }
    val immersive = fullscreen.isFullscreen || pictureInPicture.isActive
    var onNext: (() -> Unit)? = null
    if (queue.entries.isNotEmpty()) {
        onNext = playNextInQueue
    }
    val extras = PlayerExtras(
        markers = state.markers,
        heatmap = state.heatmap,
        onNext = onNext,
        onEnterPictureInPicture = pictureInPicture::enterPictureInPicture,
    )
    Scaffold(
        topBar = {
            if (!immersive) {
                VeilTopBar(title = (detail as? LoadState.Loaded)?.value?.title.orEmpty(), onBack = navigator::back)
            }
        },
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            PlayerArea(
                state,
                player,
                fullscreen,
                pictureInPicture,
                options,
                extras,
                immersive,
                onFindAlternates = findAlternates,
                onOpenScene = { scene -> navigator.openScene(scene.id) },
            )
            if (!immersive) {
                SceneContent(
                    state,
                    viewModel,
                    player,
                    navigator,
                    onFindAlternates = findAlternates,
                    queuedCount = queue.entries.size,
                    onOpenQueue = { showingQueue = true },
                )
            }
        }
    }
}

/** Calls [onEnded] when the video plays to its end, while [enabled] (something is queued). */
@Composable
private fun PlayNextWhenEnded(player: ExoPlayer, enabled: Boolean, onEnded: () -> Unit) {
    val currentOnEnded by rememberUpdatedState(onEnded)
    DisposableEffect(player, enabled) {
        val listener = object : Player.Listener {
            override fun onPlaybackStateChanged(playbackState: Int) {
                if (enabled && playbackState == Player.STATE_ENDED) {
                    currentOnEnded()
                }
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
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
private fun PlayerArea(
    state: SceneState,
    player: ExoPlayer,
    fullscreen: FullscreenState,
    pictureInPicture: PictureInPictureState,
    options: PlayerOptions,
    extras: PlayerExtras,
    immersive: Boolean,
    onFindAlternates: () -> Unit,
    onOpenScene: (SceneSummary) -> Unit,
) {
    var boxModifier = Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO)
    if (immersive) {
        boxModifier = Modifier.fillMaxSize()
    }
    Box(boxModifier.background(VeilColors.canvas), contentAlignment = Alignment.Center) {
        val active = state.active
        if (active != null) {
            ScenePlayerView(
                player,
                stream = active.playable,
                title = (state.detail as? LoadState.Loaded)?.value?.title.orEmpty(),
                fullscreen = fullscreen,
                pictureInPicture = pictureInPicture,
                options = options,
                extras = extras,
                related = state.related,
                onOpenScene = onOpenScene,
                modifier = Modifier.fillMaxSize(),
            )
        }
        if (active == null) {
            PlayerPlaceholder(state, onFindAlternates)
        }
        val problem = state.playbackProblem
        if (problem != null) {
            Text(
                problem,
                modifier = Modifier.align(Alignment.BottomCenter).background(VeilColors.imageLabel).padding(VeilSpacing.small),
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
    Box(Modifier.fillMaxSize().background(VeilColors.scrim), contentAlignment = Alignment.Center) {
        val searching = state.fetchingDetail || state.resolvingStreamId != null
        if (searching) {
            Spinner()
            return@Box
        }
        if (state.streams.isEmpty()) {
            Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
                Text("No playable source found", color = VeilColors.content, style = MaterialTheme.typography.titleSmall)
                PrimaryButton("Find other sources", onClick = onFindAlternates, icon = VeilIcons.Alternates)
            }
        }
    }
}
