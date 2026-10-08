package com.playingwithclouds.veil.ui.scene

import android.content.pm.ActivityInfo
import android.content.res.Configuration
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.hls.HlsMediaSource
import androidx.media3.exoplayer.source.MediaSource
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.compose.ContentFrame
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.util.findActivity
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** Seconds between progress saves while playing. */
private const val PROGRESS_SAVE_INTERVAL_MILLISECONDS = 15_000L

/** How long to keep portrait locked after leaving fullscreen, so the rotation settles. */
private const val ORIENTATION_RELEASE_DELAY_MILLISECONDS = 600L

/** An ExoPlayer that lives as long as the page it is remembered on. */
@Composable
fun rememberScenePlayer(): ExoPlayer {
    val context = LocalContext.current
    val player = remember { ExoPlayer.Builder(context).build() }
    DisposableEffect(player) { onDispose { player.release() } }
    return player
}

/** Loads a resolved stream (HLS or progressive, with its request headers) and starts playing at the position. */
fun ExoPlayer.playStream(stream: PlayableStream, startPositionMilliseconds: Long) {
    setMediaSource(streamMediaSource(stream), startPositionMilliseconds)
    prepare()
    playWhenReady = true
}

/** The HLS or progressive media source for a stream, sending its request headers. */
fun streamMediaSource(stream: PlayableStream): MediaSource {
    val httpFactory = DefaultHttpDataSource.Factory()
        .setAllowCrossProtocolRedirects(true)
        .setDefaultRequestProperties(stream.headers)
    return createMediaSource(stream, httpFactory, MediaItem.fromUri(stream.url))
}

/** The HLS or progressive media source for a stream. */
private fun createMediaSource(stream: PlayableStream, httpFactory: DefaultHttpDataSource.Factory, mediaItem: MediaItem): MediaSource {
    if (stream.isHls) {
        return HlsMediaSource.Factory(httpFactory).createMediaSource(mediaItem)
    }
    return ProgressiveMediaSource.Factory(httpFactory).createMediaSource(mediaItem)
}

/**
 * The video with the app's own controls over it, keeping the screen on while shown. Scrubbing
 * shows frames of [stream] above the seek bar. In fullscreen, when there are [related] scenes, a
 * swipe up (or the related button) raises them over the video.
 */
@Composable
fun ScenePlayerView(
    player: ExoPlayer,
    stream: PlayableStream,
    title: String,
    fullscreen: FullscreenState,
    related: List<SceneSummary>,
    onOpenScene: (SceneSummary) -> Unit,
    modifier: Modifier = Modifier,
) {
    var showingRelated by remember { mutableStateOf(false) }
    val relatedAvailable = fullscreen.isFullscreen && related.isNotEmpty()
    LaunchedEffect(relatedAvailable) {
        if (!relatedAvailable) {
            showingRelated = false
        }
    }
    KeepScreenOn()
    Box(modifier) {
        ContentFrame(player, Modifier.fillMaxSize())
        var openRelated: (() -> Unit)? = null
        if (relatedAvailable) {
            openRelated = { showingRelated = true }
        }
        CompositionLocalProvider(LocalSeekPreview provides rememberSeekPreview(stream)) {
            PlayerControls(
                player = player,
                title = title,
                isFullscreen = fullscreen.isFullscreen,
                onToggleFullscreen = fullscreen::toggle,
                onOpenRelated = openRelated,
            )
        }
        RelatedPanel(
            visible = showingRelated,
            related = related,
            onDismiss = { showingRelated = false },
            onOpenScene = { scene ->
                player.pause()
                onOpenScene(scene)
            },
        )
    }
}

/** Keeps the display awake while composed, as the player is on screen. */
@Composable
private fun KeepScreenOn() {
    val view = LocalView.current
    DisposableEffect(view) {
        view.keepScreenOn = true
        onDispose { view.keepScreenOn = false }
    }
}

/**
 * Reports player errors and saves the position while playing: every few seconds, on pause and
 * when the page closes.
 */
@Composable
fun PlaybackReporting(
    player: ExoPlayer,
    onError: (String) -> Unit,
    onSaveProgress: (positionSeconds: Double, durationSeconds: Double?) -> Unit,
) {
    val saveProgress = {
        val durationMilliseconds = player.duration
        var durationSeconds: Double? = null
        if (durationMilliseconds > 0) {
            durationSeconds = durationMilliseconds / MILLISECONDS_PER_SECOND
        }
        onSaveProgress(player.currentPosition / MILLISECONDS_PER_SECOND, durationSeconds)
    }
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onPlayerError(error: PlaybackException) {
                onError(error.message ?: "Playback failed")
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                if (!isPlaying && player.mediaItemCount > 0) {
                    saveProgress()
                }
            }
        }
        player.addListener(listener)
        onDispose {
            if (player.mediaItemCount > 0) {
                saveProgress()
            }
            player.removeListener(listener)
        }
    }
    LaunchedEffect(player) {
        while (true) {
            delay(PROGRESS_SAVE_INTERVAL_MILLISECONDS)
            if (player.isPlaying) {
                saveProgress()
            }
        }
    }
}

private const val MILLISECONDS_PER_SECOND = 1000.0

/** Whether the page is in fullscreen, and how to enter and leave it. */
@Stable
class FullscreenState {
    var isFullscreen by mutableStateOf(false)
        internal set
    internal var enter: () -> Unit = {}
    internal var exit: () -> Unit = {}

    /** Switches between the inline player and fullscreen. */
    fun toggle() {
        if (isFullscreen) {
            exit()
            return
        }
        enter()
    }
}

/**
 * Fullscreen playback: landscape, system bars hidden, back leaves it. Turning the phone to
 * landscape enters it too and turning back leaves it. The window is restored when the page closes.
 */
@Composable
fun rememberFullscreenState(): FullscreenState {
    val context = LocalContext.current
    val view = LocalView.current
    val scope = rememberCoroutineScope()
    val state = remember { FullscreenState() }
    val isLandscape = LocalConfiguration.current.orientation == Configuration.ORIENTATION_LANDSCAPE
    val activity = context.findActivity()

    state.enter = {
        state.isFullscreen = true
        activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE
    }
    state.exit = {
        state.isFullscreen = false
        activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        scope.launch {
            delay(ORIENTATION_RELEASE_DELAY_MILLISECONDS)
            activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }
    }

    LaunchedEffect(isLandscape) { state.isFullscreen = isLandscape }
    BackHandler(enabled = state.isFullscreen) { state.exit() }
    DisposableEffect(state.isFullscreen) {
        val window = activity?.window
        if (window != null && state.isFullscreen) {
            val controller = WindowCompat.getInsetsController(window, view)
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            controller.hide(WindowInsetsCompat.Type.systemBars())
        }
        onDispose {
            if (window != null) {
                WindowCompat.getInsetsController(window, view).show(WindowInsetsCompat.Type.systemBars())
            }
        }
    }
    DisposableEffect(Unit) {
        onDispose { activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED }
    }
    return state
}
