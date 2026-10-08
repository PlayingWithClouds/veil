package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.CompositingStrategy
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.layout
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.offset
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * Widens the element by [horizontal] on both sides, past its parent's padding. Lets a
 * horizontally scrolling row inside a padded grid run to the screen edges; give the row the same
 * amount as content padding so its first item still lines up with the gutter.
 */
fun Modifier.bleed(horizontal: Dp = VeilSpacing.gutter): Modifier {
    return layout { measurable, constraints ->
        val extra = horizontal.roundToPx()
        val placeable = measurable.measure(constraints.offset(horizontal = extra * 2))
        val width = (placeable.width - extra * 2).coerceIn(constraints.minWidth, constraints.maxWidth)
        layout(width, placeable.height) {
            placeable.place(-extra, 0)
        }
    }
}

/** Insets the element by the screen gutter on both sides, for content in edge-to-edge lists. */
fun Modifier.gutterPadding(): Modifier {
    return this.padding(horizontal = VeilSpacing.gutter)
}

/** Content padding of a screen-wide scrolling row: items start at the gutter and scroll to the edge. */
val GutterRowPadding = PaddingValues(horizontal = VeilSpacing.gutter)

/** Fades the left and right [width] of the element to transparent, so clipped rows trail off softly. */
fun Modifier.fadingEdges(width: Dp = VeilSpacing.gutter): Modifier {
    return this
        .graphicsLayer { compositingStrategy = CompositingStrategy.Offscreen }
        .drawWithContent {
            drawContent()
            val fraction = (width.toPx() / size.width).coerceIn(0f, 0.5f)
            val mask = Brush.horizontalGradient(
                0f to Color.Transparent,
                fraction to Color.Black,
                1f - fraction to Color.Black,
                1f to Color.Transparent,
            )
            drawRect(mask, blendMode = BlendMode.DstIn)
        }
}
