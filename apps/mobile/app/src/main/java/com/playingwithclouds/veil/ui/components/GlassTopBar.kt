package com.playingwithclouds.veil.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.design.LocalGlassBackdrop
import com.playingwithclouds.veil.ui.design.glassBar
import dev.chrisbanes.haze.hazeSource
import dev.chrisbanes.haze.rememberHazeState

/**
 * Top space a screen's scrolling content needs so its first item starts below a [GlassTopBarLayout]'s
 * bar; zero on screens without one. Lists add it to their top content padding.
 */
val LocalTopBarInset = compositionLocalOf { 0.dp }

/**
 * A screen whose content scrolls under its top bar. The bar is see-through while the content sits
 * at the top and turns to frosted glass once anything scrolled under it. The content draws from the
 * top edge and reads the bar's height from [LocalTopBarInset].
 */
@Composable
fun GlassTopBarLayout(bar: @Composable () -> Unit, modifier: Modifier = Modifier, content: @Composable () -> Unit) {
    val backdrop = rememberHazeState()
    val density = LocalDensity.current
    var barHeight by remember { mutableStateOf<Dp>(0.dp) }
    val scrolled = remember { ScrollDistance() }
    val glassAlpha by animateFloatAsState(glassAlphaFor(scrolled.pixels), label = "top bar glass")
    Box(modifier.fillMaxSize()) {
        Box(Modifier.fillMaxSize().nestedScroll(scrolled).hazeSource(backdrop)) {
            CompositionLocalProvider(LocalTopBarInset provides barHeight) { content() }
        }
        CompositionLocalProvider(LocalGlassBackdrop provides backdrop) {
            Box(
                Modifier
                    .fillMaxWidth()
                    .align(Alignment.TopCenter)
                    .onSizeChanged { size -> barHeight = with(density) { size.height.toDp() } },
            ) {
                Box(Modifier.matchParentSize().alpha(glassAlpha).glassBar())
                Column { bar() }
            }
        }
    }
}

/** The bar's glass: none while the content is at the top, full once it scrolled. */
private fun glassAlphaFor(scrolledPixels: Float): Float {
    if (scrolledPixels > 0f) {
        return 1f
    }
    return 0f
}

/**
 * How far the content has scrolled down from its top, summed from what its lists consumed. Nested
 * scrolling sees every scrollable inside, so it works for any list without access to its state.
 */
private class ScrollDistance : NestedScrollConnection {

    var pixels by mutableFloatStateOf(0f)
        private set

    /** Adds what a vertical scroll inside moved, never below the top. */
    override fun onPostScroll(consumed: Offset, available: Offset, source: NestedScrollSource): Offset {
        pixels = (pixels - consumed.y).coerceAtLeast(0f)
        return Offset.Zero
    }
}
