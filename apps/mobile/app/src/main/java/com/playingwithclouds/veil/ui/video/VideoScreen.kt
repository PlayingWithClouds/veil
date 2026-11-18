package com.playingwithclouds.veil.ui.video

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.ScenePreview
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
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
import androidx.compose.ui.platform.LocalContext
import com.playingwithclouds.veil.data.VideoDetail
import com.playingwithclouds.veil.data.VideoSegment
import com.playingwithclouds.veil.scenes.OpenScanProgress
import com.playingwithclouds.veil.scenes.SceneSections
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.media3.common.Player
import com.playingwithclouds.veil.data.PlayQueue
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.compilation.CompilationSheet
import kotlinx.coroutines.delay
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.design.OverlayIconButton
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** The small spinner in the corner of the preview while the stream resolves. */
private val PreviewSpinnerSize = 20.dp

/** The player box's shape on the page; fullscreen and picture-in-picture fill the screen instead. */
private const val PLAYER_ASPECT_RATIO = 16f / 9f

/** How long playback runs before the scene scan starts beside it. */
private const val SCENE_SCAN_DELAY_MILLISECONDS = 5_000L

private const val MILLISECONDS_PER_SECOND = 1000.0

/**
 * A video's page: the player under the status bar (no bar of its own; Back is in its controls), then what the video is, what to do with it and Up
 * next. Fullscreen and picture-in-picture show only the player. Watching is logged for as long as
 * the page is open. [startSeconds] (a short's moment) starts playback there instead of the resume point.
 */
@Composable
fun VideoScreen(videoId: String, navigator: AppNavigator, startSeconds: Double? = null) {
    val viewModel = viewModel(key = "$videoId@$startSeconds") { VideoViewModel(videoId, startSeconds) }
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
    val context = LocalContext.current
    LaunchedEffect(state.active != null) {
        if (state.active == null) {
            return@LaunchedEffect
        }
        // Playback gets its first seconds of buffer before the scan starts reading the same stream.
        delay(SCENE_SCAN_DELAY_MILLISECONDS)
        viewModel.scanScenes(context.applicationContext, player.currentPosition / MILLISECONDS_PER_SECOND)
    }
    PlaybackErrors(player, onError = viewModel::onPlayerError)
    WatchTracking(player, onEvent = viewModel::onWatchEvent)
    VideoSnackbars(state, viewModel, snackbarHostState)
    LaunchedEffect(videoId) { runCatching { PlayQueue.refresh() } }
    QueueAdvance(player, videoId, onAdvance = navigator::playQueued)
    var showingQueue by remember { mutableStateOf(false) }
    var choosingCompilation by remember { mutableStateOf(false) }
    if (showingQueue) {
        QueueSheet(
            playing = (state.detail as? LoadState.Loaded)?.value?.summary ?: state.card,
            playingId = videoId,
            onPlay = { id ->
                showingQueue = false
                if (id != videoId) {
                    navigator.playQueued(id)
                }
            },
            onCompilation = compilationAction {
                showingQueue = false
                choosingCompilation = true
            },
            onDismiss = { showingQueue = false },
        )
    }
    if (choosingCompilation) {
        CompilationSheet(
            onDismiss = { choosingCompilation = false },
            onStart = { options ->
                choosingCompilation = false
                player.pause()
                navigator.openQueueCompilation(videoId, options)
            },
        )
    }

    val detail = state.detail
    if (detail is LoadState.Failed) {
        DetailScreen("Video", onBack = navigator::back) {
            FailedMessage(detail.message, onRetry = viewModel::load, modifier = Modifier.padding(top = LocalTopBarInset.current))
        }
        return
    }
    val immersive = fullscreen.isFullscreen || pictureInPicture.isActive
    val title = (detail as? LoadState.Loaded)?.value?.title.orEmpty()
    ShowAmbientGlow((detail as? LoadState.Loaded)?.value?.summary?.coverUrl)
    val queueContents by PlayQueue.contents.collectAsState()
    val queueRepeat by PlayQueue.repeat.collectAsState()
    val queueShuffle by PlayQueue.shuffle.collectAsState()
    val skipTarget = remember(queueContents, queueRepeat, queueShuffle, videoId) { PlayQueue.skip(videoId) }
    var onNext: (() -> Unit)? = null
    if (skipTarget != null) {
        onNext = { navigator.playQueued(skipTarget) }
    }
    Scaffold(
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
        bottomBar = {
            if (!immersive) {
                QueueBar(videoId, onOpen = { showingQueue = true }, Modifier.navigationBarsPadding())
            }
        },
    ) { padding ->
        // No bar above the player: it sits right under the status bar, its controls carry Back.
        var columnModifier = Modifier.padding(bottom = padding.calculateBottomPadding())
        if (!immersive) {
            columnModifier = columnModifier.statusBarsPadding()
        }
        Column(columnModifier) {
            val extras = PlayerExtras(
                markers = markersInPlayOrder(SceneSections.chapters(pageSegments(detail, state.sceneScan))),
                onEnterPictureInPicture = pictureInPicture::enterPictureInPicture,
                onBack = navigator::back,
                onNext = onNext,
                onOpenQueue = { showingQueue = true },
            )
            var playerModifier = Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO)
            if (immersive) {
                playerModifier = Modifier.fillMaxSize()
            }
            Box(playerModifier.background(VeilColors.canvas), contentAlignment = Alignment.Center) {
                PlayerArea(state, player, title, fullscreen, pictureInPicture, options, extras, navigator)
            }
            if (!immersive) {
                VideoContent(state, viewModel, player, navigator, onCompilation = compilationAction { choosingCompilation = true })
            }
        }
    }
}

/**
 * At the end of the video, the queue goes on: the same video again on repeat one, else what
 * [PlayQueue.next] names, opened through [onAdvance]; nothing past the end without repeat.
 */
@Composable
private fun QueueAdvance(player: Player, videoId: String, onAdvance: (String) -> Unit) {
    val currentOnAdvance by rememberUpdatedState(onAdvance)
    DisposableEffect(player, videoId) {
        val listener = object : Player.Listener {
            override fun onPlaybackStateChanged(playbackState: Int) {
                if (playbackState != Player.STATE_ENDED) {
                    return
                }
                val next = PlayQueue.next(videoId) ?: return
                if (next == videoId) {
                    player.seekTo(0)
                    player.play()
                    return
                }
                currentOnAdvance(next)
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
    }
}

/** The Compilation action, offered only with a scene model installed. */
@Composable
private fun compilationAction(open: () -> Unit): (() -> Unit)? {
    val model by SceneModelStore.state.collectAsState()
    if (model.ready == null) {
        return null
    }
    return open
}

/** The video's segments, the running scan's in place of the stored ai ones. */
private fun pageSegments(detail: LoadState<VideoDetail>, scan: OpenScanProgress?): List<VideoSegment> {
    val stored = (detail as? LoadState.Loaded)?.value?.segments.orEmpty()
    if (scan == null) {
        return stored
    }
    return SceneSections.withScanned(stored, scan.segments, scan.done)
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
    Box(Modifier.fillMaxSize()) {
        PlayerPlaceholder(state)
        val onBack = extras.onBack
        if (onBack != null) {
            OverlayIconButton(VeilIcons.Back, contentDescription = "Back", onClick = onBack, modifier = Modifier.padding(VeilSpacing.medium))
        }
    }
}

/**
 * Before a source plays: the cover with the preview looping over it from the moment the card is
 * read, so the page moves at once while the stream resolves; a small spinner while it does, or why
 * nothing plays. A spinner alone until even the card is there.
 */
@Composable
private fun BoxScope.PlayerPlaceholder(state: VideoState) {
    val summary = (state.detail as? LoadState.Loaded)?.value?.summary ?: state.card
    if (summary == null) {
        LoadingIndicator()
        return
    }
    RemoteImage(summary.coverUrl, Modifier.fillMaxSize(), ContentScale.Crop)
    val previewUrl = summary.previewUrl
    val problem = state.problem
    if (previewUrl != null && problem == null) {
        ScenePreview(previewUrl, Modifier.fillMaxSize())
        Spinner(Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.medium), size = PreviewSpinnerSize)
        return
    }
    Box(Modifier.fillMaxSize().background(VeilColors.scrim), contentAlignment = Alignment.Center) {
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
