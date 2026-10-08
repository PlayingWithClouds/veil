package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Opacity of a control that cannot be used right now. */
const val DISABLED_ALPHA = 0.4f

/** The opacity a control draws at, dimmed while disabled. */
fun enabledAlpha(enabled: Boolean): Float {
    if (enabled) {
        return 1f
    }
    return DISABLED_ALPHA
}

/** The main action of a screen or dialog: a white capsule with dark text. */
@Composable
fun PrimaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    icon: ImageVector? = null,
) {
    CapsuleButton(text, onClick, VeilColors.content, VeilColors.canvas, modifier, enabled, icon)
}

/** A secondary action: a dark capsule with light text. */
@Composable
fun SecondaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    icon: ImageVector? = null,
) {
    CapsuleButton(text, onClick, VeilColors.surfaceHigh, VeilColors.content, modifier, enabled, icon)
}

/** A capsule button in the given colors, with an optional leading icon. */
@Composable
private fun CapsuleButton(
    text: String,
    onClick: () -> Unit,
    background: Color,
    foreground: Color,
    modifier: Modifier,
    enabled: Boolean,
    icon: ImageVector?,
) {
    Row(
        modifier
            .height(44.dp)
            .alpha(enabledAlpha(enabled))
            .pressClickable(enabled = enabled, onClick = onClick)
            .clip(CircleShape)
            .background(background)
            .padding(horizontal = 20.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (icon != null) {
            Icon(icon, contentDescription = null, tint = foreground, modifier = Modifier.size(18.dp))
        }
        Text(text, color = foreground, style = MaterialTheme.typography.labelLarge, maxLines = 1)
    }
}

/** A text-only action, for inline links and low-weight choices. */
@Composable
fun TextAction(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    color: Color = VeilColors.accent,
    icon: ImageVector? = null,
) {
    Row(
        modifier
            .alpha(enabledAlpha(enabled))
            .pressClickable(enabled = enabled, onClick = onClick)
            .padding(horizontal = 8.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (icon != null) {
            Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(18.dp))
        }
        Text(text, color = color, style = MaterialTheme.typography.labelLarge, maxLines = 1)
    }
}

/** A bare icon tap target, for rows and cards where a filled circle would be too heavy. */
@Composable
fun IconTap(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    tint: Color = VeilColors.contentMuted,
    size: Dp = 40.dp,
) {
    Box(
        modifier
            .size(size)
            .alpha(enabledAlpha(enabled))
            .pressClickable(enabled = enabled, onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = contentDescription, tint = tint, modifier = Modifier.size(22.dp))
    }
}
