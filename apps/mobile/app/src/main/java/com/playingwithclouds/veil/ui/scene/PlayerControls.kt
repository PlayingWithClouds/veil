package com.playingwithclouds.veil.ui.scene

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.gestures.detectVerticalDragGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.displayCutoutPadding
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.media3.common.Player
import androidx.media3.ui.compose.state.rememberPlayPauseButtonState
import androidx.media3.ui.compose.state.rememberProgressStateWithTickInterval
import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.ui.design.GlassIconButton
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.glass
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.delay

/** How long the controls stay up after the last touch while the video plays. */
private const val AUTO_HIDE_MILLISECONDS = 3_000L

/** How long the double-tap seek label lingers after the last tap. */
private const val SEEK_FEEDBACK_MILLISECONDS = 700L

/** Media-time step of the elapsed-time text and seek bar updates. */
private const val PROGRESS_TICK_MILLISECONDS = 250L

private const val MILLISECONDS_PER_SECOND = 1000L

private val PlayButtonSize = 64.dp
private val ControlButtonSize = 36.dp

/** How far a swipe up has to travel to open the related panel. */
private val PanelSwipeDistance = 48.dp

private val TopScrim = Brush.verticalGradient(listOf(VeilColors.scrim, Color.Transparent))
private val BottomScrim = Brush.verticalGradient(listOf(Color.Transparent, VeilColors.scrim))

/** Whether the player is playing, and whether it is waiting for data while it should be. */
@Stable
private class PlaybackStatus {
    var isPlaying by mutableStateOf(false)
    var isBuffering by mutableStateOf(false)

    /** Takes the player's current state. */
    fun update(player: Player) {
        isPlaying = player.isPlaying
        isBuffering = player.playbackState == Player.STATE_BUFFERING && player.playWhenReady
    }
}

/** The player's [PlaybackStatus], kept current through a listener. */
@Composable
private fun rememberPlaybackStatus(player: Player): PlaybackStatus {
    val status = remember(player) { PlaybackStatus().apply { update(player) } }
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onEvents(player: Player, events: Player.Events) {
                status.update(player)
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
    }
    return status
}

/**
 * What the controls show beyond playback itself: the scene's markers (chips and ticks on the seek
 * bar), the most-replayed graph, a next-in-queue action and the picture-in-picture action.
 */
data class PlayerExtras(
    val markers: List<SceneMarker> = emptyList(),
    val heatmap: List<Float> = emptyList(),
    val onNext: (() -> Unit)? = null,
    val onEnterPictureInPicture: (() -> Unit)? = null,
)

/**
 * The controls drawn over the video. A tap shows or hides them (they hide by themselves after a few
 * seconds of playback), a double tap on the left or right half jumps 10 s, and in fullscreen a swipe
 * up from the lower half asks for the related panel through [onOpenRelated].
 */
@Composable
fun PlayerControls(
    player: Player,
    options: PlayerOptions,
    extras: PlayerExtras,
    title: String,
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    onOpenRelated: (() -> Unit)?,
    modifier: Modifier = Modifier,
) {
    val status = rememberPlaybackStatus(player)
    var visible by remember { mutableStateOf(true) }
    var scrubbing by remember { mutableStateOf(false) }
    var menuOpen by remember { mutableStateOf(false) }
    var touches by remember { mutableIntStateOf(0) }
    var seekFeedback by remember { mutableStateOf<SeekFeedback?>(null) }

    LaunchedEffect(visible, status.isPlaying, scrubbing, menuOpen, touches) {
        if (!shouldAutoHide(visible, status.isPlaying, scrubbing, menuOpen)) {
            return@LaunchedEffect
        }
        delay(AUTO_HIDE_MILLISECONDS)
        visible = false
    }
    val feedbackId = seekFeedback?.id
    LaunchedEffect(feedbackId) {
        if (feedbackId == null) {
            return@LaunchedEffect
        }
        delay(SEEK_FEEDBACK_MILLISECONDS)
        seekFeedback = null
    }

    val seekByDoubleTap = { direction: Int ->
        val delta = direction * DOUBLE_TAP_SEEK_SECONDS * MILLISECONDS_PER_SECOND
        player.seekTo(clampedSeekTarget(player.currentPosition, delta, player.duration))
        seekFeedback = accumulateSeek(seekFeedback, direction)
    }
    var openRelated: (() -> Unit)? = null
    if (onOpenRelated != null) {
        openRelated = {
            visible = false
            onOpenRelated()
        }
    }

    Box(
        modifier
            .fillMaxSize()
            .playerGestures(
                onTap = { visible = !visible },
                onDoubleTap = seekByDoubleTap,
                onSwipeUp = openRelated,
            ),
    ) {
        SeekFeedbackOverlay(seekFeedback)
        AnimatedVisibility(visible, enter = fadeIn(), exit = fadeOut(), modifier = Modifier.fillMaxSize()) {
            ControlsLayer(
                player = player,
                options = options,
                extras = extras,
                title = title,
                isFullscreen = isFullscreen,
                isBuffering = status.isBuffering,
                onToggleFullscreen = onToggleFullscreen,
                onOpenRelated = openRelated,
                onTouch = { touches += 1 },
                onScrubbingChange = { value -> scrubbing = value },
                onMenuChange = { value -> menuOpen = value },
            )
        }
        if (status.isBuffering) {
            Spinner(Modifier.align(Alignment.Center), size = PlayButtonSize / 2)
        }
    }
}

/** Taps, double taps (with the side they landed on) and, when [onSwipeUp] is set, a swipe up from the lower half. */
@Composable
private fun Modifier.playerGestures(onTap: () -> Unit, onDoubleTap: (direction: Int) -> Unit, onSwipeUp: (() -> Unit)?): Modifier {
    val currentOnTap by rememberUpdatedState(onTap)
    val currentOnDoubleTap by rememberUpdatedState(onDoubleTap)
    val currentOnSwipeUp by rememberUpdatedState(onSwipeUp)
    val swipeEnabled = onSwipeUp != null
    return this
        .pointerInput(Unit) {
            detectTapGestures(
                onTap = { currentOnTap() },
                onDoubleTap = { offset -> currentOnDoubleTap(seekDirectionAt(offset.x, size.width.toFloat())) },
            )
        }
        .pointerInput(swipeEnabled) {
            if (!swipeEnabled) {
                return@pointerInput
            }
            var startY = 0f
            var dragY = 0f
            detectVerticalDragGestures(
                onDragStart = { offset ->
                    startY = offset.y
                    dragY = 0f
                },
                onDragEnd = {
                    if (isRelatedPanelSwipe(startY, size.height.toFloat(), dragY, PanelSwipeDistance.toPx())) {
                        currentOnSwipeUp?.invoke()
                    }
                },
                onVerticalDrag = { change, amount ->
                    change.consume()
                    dragY += amount
                },
            )
        }
}

/** Scrims, the top row, the centred play button and the bottom bar with the seek bar. */
@Composable
private fun ControlsLayer(
    player: Player,
    options: PlayerOptions,
    extras: PlayerExtras,
    title: String,
    isFullscreen: Boolean,
    isBuffering: Boolean,
    onToggleFullscreen: () -> Unit,
    onOpenRelated: (() -> Unit)?,
    onTouch: () -> Unit,
    onScrubbingChange: (Boolean) -> Unit,
    onMenuChange: (Boolean) -> Unit,
) {
    Box(Modifier.fillMaxSize()) {
        Box(Modifier.fillMaxWidth().fillMaxHeight(SCRIM_HEIGHT_FRACTION).align(Alignment.TopCenter).background(TopScrim))
        Box(Modifier.fillMaxWidth().fillMaxHeight(SCRIM_HEIGHT_FRACTION).align(Alignment.BottomCenter).background(BottomScrim))
        TopRow(player, options, title, isFullscreen, onToggleFullscreen, onMenuChange, Modifier.align(Alignment.TopCenter))
        if (!isBuffering) {
            CentreButtons(player, onTouch, Modifier.align(Alignment.Center))
        }
        BottomBar(
            player = player,
            options = options,
            extras = extras,
            isFullscreen = isFullscreen,
            onToggleFullscreen = onToggleFullscreen,
            onOpenRelated = onOpenRelated,
            onTouch = onTouch,
            onScrubbingChange = onScrubbingChange,
            modifier = Modifier.align(Alignment.BottomCenter),
        )
    }
}

/**
 * Keeps controls clear of the camera cutout in fullscreen. Inline the player sits below the top
 * bar, where the cutout's inset would only push the controls down.
 */
private fun Modifier.fullscreenCutoutPadding(isFullscreen: Boolean): Modifier {
    if (!isFullscreen) {
        return this
    }
    return this.displayCutoutPadding()
}

/** Share of the player's height each gradient scrim covers. */
private const val SCRIM_HEIGHT_FRACTION = 0.4f

/** In fullscreen a way out and the title; always the settings gear on the right. */
@Composable
private fun TopRow(
    player: Player,
    options: PlayerOptions,
    title: String,
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    onMenuChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier.fillMaxWidth().fullscreenCutoutPadding(isFullscreen).padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (isFullscreen) {
            RoundIconButton(VeilIcons.Back, contentDescription = "Leave fullscreen", onClick = onToggleFullscreen, size = ControlButtonSize)
            Text(
                title,
                modifier = Modifier.weight(1f),
                style = MaterialTheme.typography.titleSmall,
                color = VeilColors.content,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (!isFullscreen) {
            Box(Modifier.weight(1f))
        }
        PlayerSettingsButton(player, options, ControlButtonSize, onOpenChange = onMenuChange)
    }
}

/** The play/pause button in the middle, with frame-step buttons on either side while paused. */
@Composable
private fun CentreButtons(player: Player, onTouch: () -> Unit, modifier: Modifier = Modifier) {
    val paused = rememberPlayPauseButtonState(player).showPlay
    Row(modifier, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraLarge), verticalAlignment = Alignment.CenterVertically) {
        FrameStepSlot(paused) {
            RoundIconButton(
                VeilIcons.StepBack,
                contentDescription = "Previous frame",
                onClick = {
                    player.stepFrame(-1)
                    onTouch()
                },
                size = ControlButtonSize,
            )
        }
        PlayPauseButton(player, onTouch)
        FrameStepSlot(paused) {
            RoundIconButton(
                VeilIcons.StepForward,
                contentDescription = "Next frame",
                onClick = {
                    player.stepFrame(1)
                    onTouch()
                },
                size = ControlButtonSize,
            )
        }
    }
}

/** Room for a frame-step button that only shows while paused, so the play button does not shift. */
@Composable
private fun FrameStepSlot(paused: Boolean, content: @Composable () -> Unit) {
    Box(Modifier.size(ControlButtonSize)) {
        if (paused) {
            content()
        }
    }
}

/** The large glass play/pause button in the middle. */
@Composable
private fun PlayPauseButton(player: Player, onTouch: () -> Unit, modifier: Modifier = Modifier) {
    val state = rememberPlayPauseButtonState(player)
    var icon = VeilIcons.Pause
    var label = "Pause"
    if (state.showPlay) {
        icon = VeilIcons.Play
        label = "Play"
    }
    GlassIconButton(
        icon,
        contentDescription = label,
        onClick = {
            state.onClick()
            onTouch()
        },
        modifier = modifier,
        size = PlayButtonSize,
    )
}

/** Elapsed / total time and the buttons above the seek bar. */
@Composable
private fun BottomBar(
    player: Player,
    options: PlayerOptions,
    extras: PlayerExtras,
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    onOpenRelated: (() -> Unit)?,
    onTouch: () -> Unit,
    onScrubbingChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    val progress = rememberProgressStateWithTickInterval(player, PROGRESS_TICK_MILLISECONDS)
    Column(modifier.fillMaxWidth().fullscreenCutoutPadding(isFullscreen).padding(horizontal = VeilSpacing.medium)) {
        if (isFullscreen) {
            MarkerChips(extras.markers, player)
        }
        Row(
            Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                formatPlaybackTime(progress.currentPositionMs, progress.durationMs),
                modifier = Modifier.weight(1f),
                style = MaterialTheme.typography.labelMedium,
                color = VeilColors.content,
            )
            if (options.loop.startMilliseconds != null) {
                RoundIconButton(VeilIcons.Repeat, contentDescription = "Loop: ${loopLabel(options.loop)}", onClick = options::clearLoop, size = ControlButtonSize)
            }
            if (extras.onNext != null) {
                RoundIconButton(VeilIcons.SkipNext, contentDescription = "Next in queue", onClick = extras.onNext, size = ControlButtonSize)
            }
            if (extras.onEnterPictureInPicture != null) {
                RoundIconButton(VeilIcons.PictureInPicture, contentDescription = "Picture-in-picture", onClick = extras.onEnterPictureInPicture, size = ControlButtonSize)
            }
            if (onOpenRelated != null) {
                RoundIconButton(VeilIcons.Collections, contentDescription = "Related videos", onClick = onOpenRelated, size = ControlButtonSize)
            }
            RoundIconButton(
                fullscreenIcon(isFullscreen),
                contentDescription = fullscreenLabel(isFullscreen),
                onClick = onToggleFullscreen,
                size = ControlButtonSize,
            )
        }
        PlayerSeekBar(
            positionMilliseconds = progress.currentPositionMs,
            bufferedMilliseconds = progress.bufferedPositionMs,
            durationMilliseconds = progress.durationMs,
            onSeek = { position ->
                player.seekTo(position)
                onTouch()
            },
            onScrubbingChange = onScrubbingChange,
            loop = options.loop,
            markerFractions = markerFractions(extras.markers, progress.durationMs / MILLISECONDS_PER_SECOND_DOUBLE),
            heatmap = extras.heatmap,
        )
    }
}

private const val MILLISECONDS_PER_SECOND_DOUBLE = 1000.0

/** Enter or leave fullscreen, by the current mode. */
private fun fullscreenIcon(isFullscreen: Boolean): ImageVector {
    if (isFullscreen) {
        return VeilIcons.ExitFullscreen
    }
    return VeilIcons.Fullscreen
}

/** Spoken label of the fullscreen button. */
private fun fullscreenLabel(isFullscreen: Boolean): String {
    if (isFullscreen) {
        return "Leave fullscreen"
    }
    return "Fullscreen"
}

/** A soft glow on the tapped half with the running seek total, while double taps seek. */
@Composable
private fun BoxScope.SeekFeedbackOverlay(feedback: SeekFeedback?) {
    var shown by remember { mutableStateOf(feedback) }
    if (feedback != null) {
        shown = feedback
    }
    val last = shown ?: return
    var alignment = Alignment.CenterEnd
    if (last.direction < 0) {
        alignment = Alignment.CenterStart
    }
    AnimatedVisibility(
        feedback != null,
        enter = fadeIn(),
        exit = fadeOut(),
        modifier = Modifier.fillMaxHeight().fillMaxWidth(HALF).align(alignment),
    ) {
        Box(Modifier.fillMaxSize().background(SeekGlow), contentAlignment = Alignment.Center) {
            Text(
                seekFeedbackLabel(last),
                modifier = Modifier.glass(VeilShapes.capsule).padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.small),
                style = MaterialTheme.typography.labelLarge,
                color = VeilColors.content,
            )
        }
    }
}

private const val HALF = 0.5f

/** The glow behind the seek label. */
private val SeekGlow = Brush.radialGradient(listOf(VeilColors.glassControlTop, Color.Transparent))
