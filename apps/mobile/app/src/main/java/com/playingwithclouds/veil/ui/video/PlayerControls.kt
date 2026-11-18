package com.playingwithclouds.veil.ui.video

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
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Icon
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.Shadow
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.media3.common.Player
import androidx.media3.ui.compose.state.rememberPlayPauseButtonState
import androidx.media3.ui.compose.state.rememberProgressStateWithTickInterval
import com.playingwithclouds.veil.data.VideoSegment
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.OverlayIconButton
import com.playingwithclouds.veil.ui.design.OverlayRoundButton
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

private val PlayButtonSize = 60.dp
private val ControlButtonSize = 36.dp

/** How far a swipe up has to travel to open the related panel. */
private val PanelSwipeDistance = 48.dp


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
 * bar), the most-replayed graph, a next-in-queue action and the picture-in-picture action. [live]
 * is a live room: no frame steps, speed or loop, and instead of the clock a LIVE badge, or how far
 * the playhead is rewound into the room's buffer with "Go live". [onBack] leaves the page from the
 * top row, for pages that have no bar above the player. [onOpenQueue] shows the queue (also in
 * fullscreen).
 */
data class PlayerExtras(
    val markers: List<VideoSegment> = emptyList(),
    val heatmap: List<Float> = emptyList(),
    val onNext: (() -> Unit)? = null,
    val onEnterPictureInPicture: (() -> Unit)? = null,
    val live: Boolean = false,
    val onBack: (() -> Unit)? = null,
    val onOpenQueue: (() -> Unit)? = null,
)

/**
 * The controls drawn over the video, laid out like YouTube's. A tap shows or hides them (they hide
 * by themselves after a few seconds of playback), a double tap on the left or right half jumps 10 s,
 * holding plays at 2× until let go, a swipe up goes fullscreen and a swipe down leaves it; in
 * fullscreen a swipe up from the lower half asks for the related panel through [onOpenRelated].
 * While hidden, a thin line along the bottom shows how far the video is.
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
    val boost = remember(player) { SpeedBoost(player) }

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
    val swipe = VerticalSwipeActions(isFullscreen, onToggleFullscreen, openRelated)

    Box(
        modifier
            .fillMaxSize()
            .playerGestures(
                onTap = { visible = !visible },
                onDoubleTap = seekByDoubleTap,
                onHoldChange = { holding -> boost.hold(holding && !extras.live) },
                onVerticalSwipe = swipe::onSwipe,
            ),
    ) {
        SeekFeedbackOverlay(seekFeedback)
        if (boost.active) {
            SpeedBoostLabel(Modifier.align(Alignment.TopCenter).padding(top = VeilSpacing.medium))
        }
        if (!visible && !isFullscreen && !extras.live) {
            HiddenProgressLine(player, Modifier.align(Alignment.BottomCenter))
        }
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

/**
 * Taps, double taps (with the side they landed on), holds (true on a long press, false when let
 * go) and vertical swipes (where they started, the player's height, how far they went).
 */
@Composable
private fun Modifier.playerGestures(
    onTap: () -> Unit,
    onDoubleTap: (direction: Int) -> Unit,
    onHoldChange: (Boolean) -> Unit,
    onVerticalSwipe: (startY: Float, height: Float, dragY: Float, threshold: Float) -> Unit,
): Modifier {
    val currentOnTap by rememberUpdatedState(onTap)
    val currentOnDoubleTap by rememberUpdatedState(onDoubleTap)
    val currentOnHoldChange by rememberUpdatedState(onHoldChange)
    val currentOnVerticalSwipe by rememberUpdatedState(onVerticalSwipe)
    return this
        .pointerInput(Unit) {
            detectTapGestures(
                onTap = { currentOnTap() },
                onDoubleTap = { offset -> currentOnDoubleTap(seekDirectionAt(offset.x, size.width.toFloat())) },
                onLongPress = { currentOnHoldChange(true) },
                onPress = {
                    tryAwaitRelease()
                    currentOnHoldChange(false)
                },
            )
        }
        .pointerInput(Unit) {
            var startY = 0f
            var dragY = 0f
            detectVerticalDragGestures(
                onDragStart = { offset ->
                    startY = offset.y
                    dragY = 0f
                },
                onDragEnd = { currentOnVerticalSwipe(startY, size.height.toFloat(), dragY, PanelSwipeDistance.toPx()) },
                onVerticalDrag = { change, amount ->
                    change.consume()
                    dragY += amount
                },
            )
        }
}

/**
 * What a vertical swipe on the video does: inline a swipe up goes fullscreen; in fullscreen a swipe
 * down leaves it and a swipe up from the lower half opens the related panel when there is one.
 */
private class VerticalSwipeActions(
    private val isFullscreen: Boolean,
    private val onToggleFullscreen: () -> Unit,
    private val onOpenRelated: (() -> Unit)?,
) {

    /** Acts on a finished swipe that started at [startY] and moved [dragY] (negative is up). */
    fun onSwipe(startY: Float, height: Float, dragY: Float, threshold: Float) {
        if (!isFullscreen) {
            if (dragY <= -threshold) {
                onToggleFullscreen()
            }
            return
        }
        if (dragY >= threshold) {
            onToggleFullscreen()
            return
        }
        if (onOpenRelated != null && isRelatedPanelSwipe(startY, height, dragY, threshold)) {
            onOpenRelated()
        }
    }
}

/** Speed while the video is held down. */
private const val BOOST_SPEED = 2f

/** Plays at [BOOST_SPEED] while held, then back at the speed it had. */
@Stable
private class SpeedBoost(private val player: Player) {
    var active by mutableStateOf(false)
        private set
    private var speedBefore = 1f

    /** Starts the boost when [holding], ends it when let go. */
    fun hold(holding: Boolean) {
        if (holding && !active) {
            speedBefore = player.playbackParameters.speed
            player.setPlaybackSpeed(BOOST_SPEED)
            active = true
            return
        }
        if (!holding && active) {
            player.setPlaybackSpeed(speedBefore)
            active = false
        }
    }
}

/** "2×" at the top while the video is held. */
@Composable
private fun SpeedBoostLabel(modifier: Modifier = Modifier) {
    Row(
        modifier.clip(VeilShapes.capsule).background(VeilColors.playerDim).padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text("2×", style = MaterialTheme.typography.labelLarge, color = VeilColors.content)
        Icon(VeilIcons.FastForward, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(SmallIconSize))
    }
}

/** A thin played line along the bottom edge of the inline player while its controls are hidden. */
@Composable
private fun HiddenProgressLine(player: Player, modifier: Modifier = Modifier) {
    val progress = rememberProgressStateWithTickInterval(player, PROGRESS_TICK_MILLISECONDS)
    val played = fractionOf(progress.currentPositionMs, progress.durationMs)
    Box(modifier.fillMaxWidth().height(HiddenLineHeight).background(VeilColors.playerDim)) {
        Box(Modifier.fillMaxWidth(played).fillMaxHeight().background(VeilColors.playerProgress))
    }
}

private val HiddenLineHeight = 2.dp
private val SmallIconSize = 16.dp

/** An even dim over the video, the top row, the centred play button and the bottom bar with the seek bar. */
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
        Box(Modifier.fillMaxSize().background(VeilColors.playerDim))
        TopRow(player, options, title, isFullscreen, extras, onToggleFullscreen, onOpenRelated, onMenuChange, Modifier.align(Alignment.TopCenter))
        if (!isBuffering) {
            CentreButtons(player, extras.live, onTouch, Modifier.align(Alignment.Center))
        }
        BottomBar(
            player = player,
            options = options,
            extras = extras,
            isFullscreen = isFullscreen,
            onToggleFullscreen = onToggleFullscreen,
            onTouch = onTouch,
            onScrubbingChange = onScrubbingChange,
            modifier = Modifier.align(Alignment.BottomCenter),
        )
    }
}

/**
 * Keeps controls clear of the camera cutout in fullscreen. Inline the player sits below the status
 * bar, where the cutout's inset would only push the controls down.
 */
private fun Modifier.fullscreenCutoutPadding(isFullscreen: Boolean): Modifier {
    if (!isFullscreen) {
        return this
    }
    return this.displayCutoutPadding()
}


/**
 * In fullscreen a way out and the title, inline the page's back button if it has one; on the right,
 * as YouTube's, the secondary actions (picture-in-picture, related) and the settings gear.
 */
@Composable
private fun TopRow(
    player: Player,
    options: PlayerOptions,
    title: String,
    isFullscreen: Boolean,
    extras: PlayerExtras,
    onToggleFullscreen: () -> Unit,
    onOpenRelated: (() -> Unit)?,
    onMenuChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier.fillMaxWidth().fullscreenCutoutPadding(isFullscreen).padding(horizontal = VeilSpacing.small, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (isFullscreen) {
            OverlayIconButton(VeilIcons.Back, contentDescription = "Leave fullscreen", onClick = onToggleFullscreen, size = ControlButtonSize)
            Text(
                title,
                modifier = Modifier.weight(1f),
                style = MaterialTheme.typography.titleSmall,
                color = VeilColors.content,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        val onBack = extras.onBack
        if (!isFullscreen && onBack != null) {
            OverlayIconButton(VeilIcons.Back, contentDescription = "Back", onClick = onBack, size = ControlButtonSize)
        }
        if (!isFullscreen) {
            Box(Modifier.weight(1f))
        }
        if (extras.onEnterPictureInPicture != null) {
            OverlayIconButton(VeilIcons.PictureInPicture, contentDescription = "Picture-in-picture", onClick = extras.onEnterPictureInPicture, size = ControlButtonSize)
        }
        if (onOpenRelated != null) {
            OverlayIconButton(VeilIcons.Collections, contentDescription = "Related videos", onClick = onOpenRelated, size = ControlButtonSize)
        }
        if (extras.onOpenQueue != null) {
            OverlayIconButton(VeilIcons.Queue, contentDescription = "Queue", onClick = extras.onOpenQueue, size = ControlButtonSize)
        }
        PlayerSettingsButton(player, options, ControlButtonSize, live = extras.live, onOpenChange = onMenuChange)
    }
}

/**
 * Three dark round buttons in the middle, as YouTube's: back, play/pause, forward. Back and forward
 * jump 10 s while playing and step one frame while paused; a live room has only play/pause.
 */
@Composable
private fun CentreButtons(player: Player, live: Boolean, onTouch: () -> Unit, modifier: Modifier = Modifier) {
    val paused = rememberPlayPauseButtonState(player).showPlay
    Row(modifier, horizontalArrangement = Arrangement.spacedBy(CentreGap), verticalAlignment = Alignment.CenterVertically) {
        if (!live) {
            SideButton(player, direction = -1, paused, onTouch)
        }
        PlayPauseButton(player, onTouch)
        if (!live) {
            SideButton(player, direction = 1, paused, onTouch)
        }
    }
}

/** A round button beside play: a 10 s jump, or a frame step while [paused]; [direction] -1 is back. */
@Composable
private fun SideButton(player: Player, direction: Int, paused: Boolean, onTouch: () -> Unit) {
    var icon = VeilIcons.SeekForward
    var label = "Forward $DOUBLE_TAP_SEEK_SECONDS seconds"
    if (direction < 0) {
        icon = VeilIcons.SeekBack
        label = "Back $DOUBLE_TAP_SEEK_SECONDS seconds"
    }
    if (paused && direction < 0) {
        icon = VeilIcons.StepBack
        label = "Previous frame"
    }
    if (paused && direction > 0) {
        icon = VeilIcons.StepForward
        label = "Next frame"
    }
    OverlayRoundButton(
        icon,
        contentDescription = label,
        onClick = {
            if (paused) {
                player.stepFrame(direction)
            } else {
                val delta = direction * DOUBLE_TAP_SEEK_SECONDS * MILLISECONDS_PER_SECOND
                player.seekTo(clampedSeekTarget(player.currentPosition, delta, player.duration))
            }
            onTouch()
        },
        size = SideButtonSize,
    )
}

private val SideButtonSize = 44.dp
private val CentreGap = 40.dp

/** The large round play/pause button in the middle. */
@Composable
private fun PlayPauseButton(player: Player, onTouch: () -> Unit, modifier: Modifier = Modifier) {
    val state = rememberPlayPauseButtonState(player)
    var icon = VeilIcons.Pause
    var label = "Pause"
    if (state.showPlay) {
        icon = VeilIcons.Play
        label = "Play"
    }
    OverlayRoundButton(
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

/**
 * Elapsed / total time and the chapter playing on the left, the loop, next and fullscreen on the
 * right, then the seek bar: along the bottom edge inline, as YouTube's, lifted and inset in
 * fullscreen.
 */
@Composable
private fun BottomBar(
    player: Player,
    options: PlayerOptions,
    extras: PlayerExtras,
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    onTouch: () -> Unit,
    onScrubbingChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    val progress = rememberProgressStateWithTickInterval(player, PROGRESS_TICK_MILLISECONDS)
    var barModifier: Modifier = Modifier
    var columnModifier = modifier.fillMaxWidth()
    if (isFullscreen) {
        barModifier = barModifier.padding(horizontal = VeilSpacing.large)
        columnModifier = columnModifier.fullscreenCutoutPadding(true).padding(bottom = FullscreenBarLift)
    }
    Column(columnModifier) {
        if (isFullscreen) {
            MarkerChips(extras.markers, player, Modifier.padding(horizontal = VeilSpacing.medium))
        }
        Row(
            Modifier.fillMaxWidth().padding(start = VeilSpacing.medium, end = VeilSpacing.small),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (extras.live) {
                LiveStatus(
                    liveLagMilliseconds(progress.currentPositionMs, progress.durationMs),
                    onGoLive = {
                        player.seekToDefaultPosition()
                        player.play()
                        onTouch()
                    },
                    modifier = Modifier.weight(1f),
                )
            } else {
                TimeAndChapter(progress.currentPositionMs, progress.durationMs, extras.markers, Modifier.weight(1f))
            }
            if (options.loop.startMilliseconds != null) {
                OverlayIconButton(VeilIcons.Repeat, contentDescription = "Loop: ${loopLabel(options.loop)}", onClick = options::clearLoop, size = ControlButtonSize)
            }
            if (extras.onNext != null) {
                OverlayIconButton(VeilIcons.SkipNext, contentDescription = "Next in queue", onClick = extras.onNext, size = ControlButtonSize)
            }
            OverlayIconButton(
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
            modifier = barModifier,
            loop = options.loop,
            markerFractions = markerFractions(extras.markers, progress.durationMs / MILLISECONDS_PER_SECOND_DOUBLE),
            heatmap = extras.heatmap,
        )
    }
}

private const val MILLISECONDS_PER_SECOND_DOUBLE = 1000.0

/** How far the seek bar sits above the screen's bottom edge in fullscreen. */
private val FullscreenBarLift = 16.dp

/** Small white text over video, shadowed so it reads on bright frames. */
private val OverlayLabel: TextStyle
    @Composable get() = MaterialTheme.typography.labelMedium.copy(shadow = Shadow(VeilColors.playerShadow, blurRadius = TEXT_SHADOW_BLUR))

private const val TEXT_SHADOW_BLUR = 6f

/** "1:23 / 6:40", then " · " and the chapter playing (the marker reached last), as YouTube shows it. */
@Composable
private fun TimeAndChapter(positionMilliseconds: Long, durationMilliseconds: Long, markers: List<VideoSegment>, modifier: Modifier = Modifier) {
    val activeIndex = activeMarkerIndex(markers, positionMilliseconds / MILLISECONDS_PER_SECOND_DOUBLE)
    val chapter = markers.getOrNull(activeIndex)?.title?.takeIf { title -> title.isNotBlank() }
    Row(modifier, verticalAlignment = Alignment.CenterVertically) {
        Text(formatPlaybackClock(positionMilliseconds), style = OverlayLabel, color = VeilColors.content, maxLines = 1)
        if (durationMilliseconds > 0) {
            Text(" / ${formatPlaybackClock(durationMilliseconds)}", style = OverlayLabel, color = VeilColors.contentMuted, maxLines = 1)
        }
        if (chapter != null) {
            Text(
                " · $chapter",
                style = OverlayLabel,
                color = VeilColors.contentMuted,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

/**
 * Where a live room's playhead is: a red LIVE badge at the live edge, else how far it is rewound
 * and a "Go live" button that jumps back to the edge.
 */
@Composable
private fun LiveStatus(lagMilliseconds: Long, onGoLive: () -> Unit, modifier: Modifier = Modifier) {
    Row(modifier, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
        if (!isBehindLive(lagMilliseconds)) {
            Badge("LIVE", background = VeilColors.error, foreground = VeilColors.content)
            return@Row
        }
        Text(formatBehindLive(lagMilliseconds), style = MaterialTheme.typography.labelMedium, color = VeilColors.content)
        Text(
            "Go live",
            modifier = Modifier.clip(VeilShapes.capsule).pressClickable(onGoLive).glass(VeilShapes.capsule).padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.extraSmall),
            style = MaterialTheme.typography.labelMedium,
            color = VeilColors.content,
        )
    }
}

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
