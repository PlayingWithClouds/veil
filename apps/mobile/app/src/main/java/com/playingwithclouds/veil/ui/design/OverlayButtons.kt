package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes

/** How far the shadow under an overlay icon drops, and how soft it is. */
private val ShadowDrop = 1.dp
private val ShadowBlur = 3.dp

/**
 * A flat white icon over video, with no backdrop and a soft shadow so it reads on bright frames:
 * the player's small controls, as YouTube draws them.
 */
@Composable
fun OverlayIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 40.dp,
    iconSize: Dp = 24.dp,
) {
    Box(modifier.size(size).pressClickable(onClick), contentAlignment = Alignment.Center) {
        ShadowedIcon(icon, contentDescription, iconSize)
    }
}

/** A large white icon in a dark see-through circle over video: the player's play button. */
@Composable
fun OverlayRoundButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 64.dp,
) {
    Box(
        modifier.size(size).clip(VeilShapes.capsule).background(VeilColors.playerDim).pressClickable(onClick),
        contentAlignment = Alignment.Center,
    ) {
        ShadowedIcon(icon, contentDescription, size / 2)
    }
}

/** A white icon over a blurred dark copy of itself. */
@Composable
private fun ShadowedIcon(icon: ImageVector, contentDescription: String, iconSize: Dp) {
    Box(contentAlignment = Alignment.Center) {
        Icon(icon, contentDescription = null, tint = VeilColors.playerShadow, modifier = Modifier.size(iconSize).offset(y = ShadowDrop).blur(ShadowBlur))
        Icon(icon, contentDescription = contentDescription, tint = VeilColors.content, modifier = Modifier.size(iconSize))
    }
}
