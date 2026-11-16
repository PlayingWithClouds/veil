package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
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
 * A screen whose content scrolls under its top bar, the bar frosted glass over it. The content
 * draws from the top edge and reads the bar's height from [LocalTopBarInset].
 */
@Composable
fun GlassTopBarLayout(bar: @Composable () -> Unit, modifier: Modifier = Modifier, content: @Composable () -> Unit) {
    val backdrop = rememberHazeState()
    val density = LocalDensity.current
    var barHeight by remember { mutableStateOf<Dp>(0.dp) }
    Box(modifier.fillMaxSize()) {
        Box(Modifier.fillMaxSize().hazeSource(backdrop)) {
            CompositionLocalProvider(LocalTopBarInset provides barHeight) { content() }
        }
        CompositionLocalProvider(LocalGlassBackdrop provides backdrop) {
            Column(
                Modifier
                    .fillMaxWidth()
                    .align(Alignment.TopCenter)
                    .onSizeChanged { size -> barHeight = with(density) { size.height.toDp() } }
                    .glassBar(),
            ) {
                bar()
            }
        }
    }
}
