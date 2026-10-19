package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Width of a reaction button, wider than its circle so captions fit. */
private val ReactionButtonWidth = 68.dp

/** Diameter of a reaction button's circle. */
private val ReactionCircleSize = 48.dp

/** Side inset of a row of reaction buttons that puts the first circle on the gutter. */
val ReactionRowInset = VeilSpacing.gutter - (ReactionButtonWidth - ReactionCircleSize) / 2

/**
 * A round icon button over a caption, for the reaction rows of scene and gallery pages: white with
 * a dark icon while [active]; a long press can trigger a second action.
 */
@Composable
fun ReactionButton(
    icon: ImageVector,
    label: String,
    active: Boolean = false,
    onLongClick: (() -> Unit)? = null,
    onClick: () -> Unit,
) {
    var background = VeilColors.surfaceHigh
    var tint = VeilColors.content
    if (active) {
        background = VeilColors.content
        tint = VeilColors.canvas
    }
    Column(
        Modifier.width(ReactionButtonWidth).pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick).padding(vertical = VeilSpacing.small),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        Box(Modifier.size(ReactionCircleSize).clip(VeilShapes.capsule).background(background), contentAlignment = Alignment.Center) {
            Icon(icon, contentDescription = label, tint = tint, modifier = Modifier.size(22.dp))
        }
        Text(label, style = MaterialTheme.typography.labelSmall, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
    }
}

/** The like icon, filled when chosen. */
fun likeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return VeilIcons.LikeFilled
    }
    return VeilIcons.Like
}

/** The dislike icon, filled when chosen. */
fun dislikeIcon(chosen: Boolean): ImageVector {
    if (chosen) {
        return VeilIcons.DislikeFilled
    }
    return VeilIcons.Dislike
}
