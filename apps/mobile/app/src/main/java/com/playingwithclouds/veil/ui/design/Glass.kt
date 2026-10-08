package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import dev.chrisbanes.haze.HazeInput
import dev.chrisbanes.haze.HazeState
import dev.chrisbanes.haze.blur.HazeBlurStyle
import dev.chrisbanes.haze.blur.HazeColorEffect
import dev.chrisbanes.haze.blur.hazeBlur

/** The backdrop glass surfaces blur: the shell's page content. Null where nothing is captured. */
val LocalGlassBackdrop = staticCompositionLocalOf<HazeState?> { null }

/** Frosted dark glass: heavy blur, a dark tint so text stays legible, a little grain. */
private val glassStyle = HazeBlurStyle {
    blurRadius(28.dp)
    noiseFactor(0.08f)
    backgroundColor(VeilColors.canvas)
    colorEffects(listOf(HazeColorEffect.tint(Color(0xFF101013).copy(alpha = 0.62f))))
}

/** Edge highlight of glass surfaces. */
private val glassEdge = Color.White.copy(alpha = 0.10f)

/**
 * Turns the element into glass of [shape]: the page behind it, blurred and tinted, with a hairline
 * edge. Without a backdrop it falls back to a translucent fill.
 */
@Composable
fun Modifier.glass(shape: Shape): Modifier {
    val backdrop = LocalGlassBackdrop.current
    val clipped = this.clip(shape)
    val surface = if (backdrop != null) {
        clipped.hazeBlur(HazeInput.Sources(backdrop), glassStyle)
    } else {
        clipped.background(VeilColors.surfaceHigh.copy(alpha = 0.92f))
    }
    return surface.border(1.dp, glassEdge, shape)
}

/** A round glass button holding one icon. */
@Composable
fun GlassIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 40.dp,
) {
    Box(
        modifier.size(size).pressClickable(onClick).glass(CircleShape),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = contentDescription, tint = VeilColors.content, modifier = Modifier.size(size * 0.5f))
    }
}

/** A round button on a plain dark fill, for places with nothing behind to blur. */
@Composable
fun RoundIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 40.dp,
    tint: Color = VeilColors.content,
) {
    Box(
        modifier.size(size).pressClickable(onClick).clip(CircleShape).background(VeilColors.surfaceHigh),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = contentDescription, tint = tint, modifier = Modifier.size(size * 0.5f))
    }
}
