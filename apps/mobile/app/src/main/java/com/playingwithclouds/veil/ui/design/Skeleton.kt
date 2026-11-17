package com.playingwithclouds.veil.ui.design

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.graphicsLayer
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes

/** How long a skeleton takes to fade from dim to bright. */
private const val PULSE_MILLISECONDS = 800

/** The dimmest a skeleton gets while pulsing. */
private const val PULSE_LOW_ALPHA = 0.45f

/**
 * A grey block standing in for text, a pill or a picture until it loads, pulsing gently. The caller
 * sizes it to what it stands in for, so the page doesn't jump when the real thing arrives.
 */
@Composable
fun SkeletonBlock(modifier: Modifier = Modifier, shape: Shape = VeilShapes.badge) {
    val pulse = rememberInfiniteTransition(label = "skeleton")
    val alpha by pulse.animateFloat(
        initialValue = PULSE_LOW_ALPHA,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(tween(PULSE_MILLISECONDS), RepeatMode.Reverse),
        label = "skeleton alpha",
    )
    Box(modifier.graphicsLayer { this.alpha = alpha }.clip(shape).background(VeilColors.surfaceHigh))
}
