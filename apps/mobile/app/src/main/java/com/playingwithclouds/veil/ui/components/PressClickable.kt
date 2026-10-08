package com.playingwithclouds.veil.ui.components

import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.graphicsLayer

/** How far a pressed card shrinks. */
private const val PRESSED_SCALE = 0.96f

/** Clickable that springs the element down while pressed and back on release; the app has no ripples. */
@Composable
fun Modifier.pressClickable(onClick: () -> Unit): Modifier {
    return pressClickable(enabled = true, onLongClick = null, onClick = onClick)
}

/** [pressClickable] that can be disabled and can take a long press as a second action. */
@Composable
fun Modifier.pressClickable(enabled: Boolean, onLongClick: (() -> Unit)? = null, onClick: () -> Unit): Modifier {
    val interactionSource = remember { MutableInteractionSource() }
    val pressed by interactionSource.collectIsPressedAsState()
    val scale by animateFloatAsState(
        targetValue = if (pressed) PRESSED_SCALE else 1f,
        animationSpec = spring(dampingRatio = Spring.DampingRatioMediumBouncy, stiffness = Spring.StiffnessMediumLow),
        label = "press",
    )
    return this
        .graphicsLayer {
            scaleX = scale
            scaleY = scale
        }
        .combinedClickable(
            interactionSource = interactionSource,
            indication = null,
            enabled = enabled,
            onLongClick = onLongClick,
            onClick = onClick,
        )
}
