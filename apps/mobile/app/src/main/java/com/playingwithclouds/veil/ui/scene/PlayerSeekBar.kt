package com.playingwithclouds.veil.ui.scene

import androidx.compose.animation.core.animateDpAsState
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.design.glass
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlin.math.roundToInt

/** Height of the bar's touch area; the drawn track is much thinner. */
private val SeekBarTouchHeight = 28.dp

private val TrackThickness = 3.dp
private val TrackThicknessScrubbing = 5.dp
private val ThumbDiameter = 12.dp
private val ThumbDiameterScrubbing = 20.dp

/** Unplayed part of the track. */
private val TrackColor = VeilColors.content.copy(alpha = 0.22f)

/** Buffered but not yet played. */
private val BufferedColor = VeilColors.content.copy(alpha = 0.45f)

/**
 * A thin accent seek bar: played part in accent, buffered part lighter, a small thumb that grows
 * while dragged with the drag time floating above it. Dragging only previews; the player seeks on
 * release (or on a tap), so HLS streams are not asked for every position on the way.
 */
@Composable
fun PlayerSeekBar(
    positionMilliseconds: Long,
    bufferedMilliseconds: Long,
    durationMilliseconds: Long,
    onSeek: (Long) -> Unit,
    onScrubbingChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    var scrubFraction by remember { mutableStateOf<Float?>(null) }
    val scrubbing = scrubFraction != null
    val thickness by animateDpAsState(scrubThickness(scrubbing), label = "trackThickness")
    val thumb by animateDpAsState(thumbDiameter(scrubbing), label = "thumbDiameter")
    val currentDuration by rememberUpdatedState(durationMilliseconds)
    val currentOnSeek by rememberUpdatedState(onSeek)
    val currentOnScrubbingChange by rememberUpdatedState(onScrubbingChange)
    val seekable = durationMilliseconds > 0

    var shownFraction = fractionOf(positionMilliseconds, durationMilliseconds)
    val dragged = scrubFraction
    if (dragged != null) {
        shownFraction = dragged
    }
    val bufferedFraction = fractionOf(bufferedMilliseconds, durationMilliseconds)

    BoxWithConstraints(
        modifier
            .fillMaxWidth()
            .height(SeekBarTouchHeight)
            .pointerInput(seekable) {
                if (!seekable) {
                    return@pointerInput
                }
                detectTapGestures { offset -> currentOnSeek(positionAt(fractionAt(offset.x, size.width.toFloat()), currentDuration)) }
            }
            .pointerInput(seekable) {
                if (!seekable) {
                    return@pointerInput
                }
                detectHorizontalDragGestures(
                    onDragStart = { offset ->
                        scrubFraction = fractionAt(offset.x, size.width.toFloat())
                        currentOnScrubbingChange(true)
                    },
                    onDragEnd = {
                        val released = scrubFraction
                        if (released != null) {
                            currentOnSeek(positionAt(released, currentDuration))
                        }
                        scrubFraction = null
                        currentOnScrubbingChange(false)
                    },
                    onDragCancel = {
                        scrubFraction = null
                        currentOnScrubbingChange(false)
                    },
                    onHorizontalDrag = { change, _ ->
                        change.consume()
                        scrubFraction = fractionAt(change.position.x, size.width.toFloat())
                    },
                )
            },
        contentAlignment = Alignment.CenterStart,
    ) {
        Canvas(Modifier.fillMaxWidth().height(SeekBarTouchHeight)) {
            drawTrack(shownFraction, bufferedFraction, thickness, thumb, seekable)
        }
        if (dragged != null) {
            ScrubTimeLabel(positionAt(dragged, durationMilliseconds), dragged, maxWidth, Modifier.align(Alignment.TopStart))
        }
    }
}

/** Track thickness, thicker while scrubbing. */
private fun scrubThickness(scrubbing: Boolean): Dp {
    if (scrubbing) {
        return TrackThicknessScrubbing
    }
    return TrackThickness
}

/** Thumb size, bigger while scrubbing. */
private fun thumbDiameter(scrubbing: Boolean): Dp {
    if (scrubbing) {
        return ThumbDiameterScrubbing
    }
    return ThumbDiameter
}

/** Draws the track, buffered and played parts and, when seekable, the thumb at [played]. */
private fun DrawScope.drawTrack(played: Float, buffered: Float, thickness: Dp, thumb: Dp, seekable: Boolean) {
    val stroke = thickness.toPx()
    val centerY = size.height / 2f
    val start = Offset(0f, centerY)
    drawLine(TrackColor, start, Offset(size.width, centerY), stroke, StrokeCap.Round)
    if (buffered > 0f) {
        drawLine(BufferedColor, start, Offset(size.width * buffered, centerY), stroke, StrokeCap.Round)
    }
    if (played > 0f) {
        drawLine(VeilColors.accent, start, Offset(size.width * played, centerY), stroke, StrokeCap.Round)
    }
    if (seekable) {
        drawCircle(VeilColors.accent, radius = thumb.toPx() / 2f, center = Offset(size.width * played, centerY))
    }
}

/** The time under the finger, in a glass capsule above the thumb, kept inside the bar's width. */
@Composable
private fun ScrubTimeLabel(positionMilliseconds: Long, fraction: Float, barWidth: Dp, modifier: Modifier = Modifier) {
    var labelWidth by remember { mutableIntStateOf(0) }
    var labelHeight by remember { mutableIntStateOf(0) }
    val barWidthPixels = with(LocalDensity.current) { barWidth.toPx() }
    val gapPixels = with(LocalDensity.current) { VeilSpacing.extraSmall.toPx() }
    Box(
        modifier
            .onSizeChanged { measured ->
                labelWidth = measured.width
                labelHeight = measured.height
            }
            .offset {
                val centered = barWidthPixels * fraction - labelWidth / 2f
                val x = centered.coerceIn(0f, maxOf(0f, barWidthPixels - labelWidth))
                IntOffset(x.roundToInt(), -(labelHeight + gapPixels).roundToInt())
            }
            .glass(VeilShapes.capsule)
            .padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.extraSmall),
    ) {
        Text(formatPlaybackClock(positionMilliseconds), style = MaterialTheme.typography.labelLarge, color = VeilColors.content)
    }
}
