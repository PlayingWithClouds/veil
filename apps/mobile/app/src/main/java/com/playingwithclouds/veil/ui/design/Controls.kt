package com.playingwithclouds.veil.ui.design

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.rotate
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import kotlin.math.floor

/** A capsule toggle: accent track with a dark knob when on, a dark track with a grey knob when off. */
@Composable
fun VeilSwitch(checked: Boolean, onCheckedChange: (Boolean) -> Unit, modifier: Modifier = Modifier) {
    val knobOffset by animateDpAsState(
        if (checked) 20.dp else 0.dp,
        animationSpec = spring(dampingRatio = Spring.DampingRatioMediumBouncy, stiffness = Spring.StiffnessMedium),
        label = "knob",
    )
    val track by animateColorAsState(if (checked) VeilColors.accent else VeilColors.surfaceHigh, label = "track")
    val knob by animateColorAsState(if (checked) VeilColors.onAccent else VeilColors.contentMuted, label = "knobColor")
    Box(
        modifier
            .size(width = 50.dp, height = 30.dp)
            .pressClickable { onCheckedChange(!checked) }
            .clip(CircleShape)
            .background(track)
            .padding(3.dp),
    ) {
        Box(Modifier.offset(x = knobOffset).size(24.dp).clip(CircleShape).background(knob))
    }
}

/** A round check mark: accent with a dark tick when checked, an empty ring otherwise. */
@Composable
fun VeilCheck(checked: Boolean, modifier: Modifier = Modifier) {
    val fill by animateColorAsState(if (checked) VeilColors.accent else Color.Transparent, label = "check")
    var ring = Modifier.border(1.5.dp, VeilColors.contentFaint, CircleShape)
    if (checked) {
        ring = Modifier
    }
    Box(modifier.size(24.dp).clip(CircleShape).background(fill).then(ring), contentAlignment = Alignment.Center) {
        if (checked) {
            Icon(VeilIcons.Check, contentDescription = null, tint = VeilColors.onAccent, modifier = Modifier.size(15.dp))
        }
    }
}

/** A thin rounded bar filled to [progress] (0 to 1). */
@Composable
fun ProgressBar(
    progress: Float,
    modifier: Modifier = Modifier,
    color: Color = VeilColors.accent,
    trackColor: Color = VeilColors.surfaceHigh,
    height: Dp = 4.dp,
) {
    Box(modifier.fillMaxWidth().height(height).clip(CircleShape).background(trackColor)) {
        Box(Modifier.fillMaxWidth(progress.coerceIn(0f, 1f)).fillMaxHeight().clip(CircleShape).background(color))
    }
}

/** A thin bar with a segment sweeping across it: work of unknown length. */
@Composable
fun LoadingBar(modifier: Modifier = Modifier, color: Color = VeilColors.accent) {
    val transition = rememberInfiniteTransition(label = "loadingBar")
    val position by transition.animateFloat(
        initialValue = -0.35f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(tween(1100, easing = LinearEasing)),
        label = "sweep",
    )
    BoxWithConstraints(modifier.fillMaxWidth().height(3.dp).clip(CircleShape).background(VeilColors.surface)) {
        Box(
            Modifier
                .offset(x = maxWidth * position)
                .width(maxWidth * 0.35f)
                .fillMaxHeight()
                .clip(CircleShape)
                .background(color),
        )
    }
}

/** Spokes of the activity spinner. */
private const val SPOKE_COUNT = 8

/** An activity spinner of fading spokes that turns step by step. */
@Composable
fun Spinner(modifier: Modifier = Modifier, size: Dp = 28.dp, color: Color = VeilColors.content) {
    val transition = rememberInfiniteTransition(label = "spinner")
    val turn by transition.animateFloat(
        initialValue = 0f,
        targetValue = SPOKE_COUNT.toFloat(),
        animationSpec = infiniteRepeatable(tween(900, easing = LinearEasing)),
        label = "turn",
    )
    Canvas(modifier.size(size)) {
        val step = floor(turn)
        val strokeWidth = this.size.minDimension * 0.09f
        val outer = this.size.minDimension / 2f - strokeWidth
        val inner = outer * 0.5f
        for (spoke in 0 until SPOKE_COUNT) {
            val age = (step - spoke + SPOKE_COUNT) % SPOKE_COUNT
            val alpha = 1f - age / SPOKE_COUNT * 0.85f
            rotate(spoke * 360f / SPOKE_COUNT) {
                drawLine(
                    color = color.copy(alpha = alpha),
                    start = Offset(center.x, center.y - inner),
                    end = Offset(center.x, center.y - outer),
                    strokeWidth = strokeWidth,
                    cap = StrokeCap.Round,
                )
            }
        }
    }
}
