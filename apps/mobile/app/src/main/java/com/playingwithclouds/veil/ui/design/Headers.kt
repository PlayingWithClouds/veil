package com.playingwithclouds.veil.ui.design

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * The bold title that opens a tab screen and scrolls away with it (first item of its grid, which
 * supplies the gutter), with round actions on the right. Pair with [PinnedTitleBar] so the title
 * stays reachable once scrolled off.
 */
@Composable
fun LargeHeader(title: String, modifier: Modifier = Modifier, actions: @Composable RowScope.() -> Unit = {}) {
    Row(
        modifier.fillMaxWidth().padding(vertical = VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            title,
            modifier = Modifier.weight(1f),
            style = MaterialTheme.typography.headlineLarge,
            color = VeilColors.content,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically, content = actions)
    }
}

/**
 * A slim solid bar with [title] (opaque, so text scrolling under it does not show through) that fades in over the top of a tab screen while [visible],
 * i.e. once its [LargeHeader] has scrolled away. Draw it above the scrolling content.
 */
@Composable
fun PinnedTitleBar(title: String, visible: Boolean, modifier: Modifier = Modifier) {
    AnimatedVisibility(visible, modifier, enter = fadeIn(), exit = fadeOut()) {
        Box(
            Modifier
                .fillMaxWidth()
                .height(48.dp)
                .background(VeilColors.elevated)
                .drawBehind {
                    drawLine(VeilColors.glassEdge, Offset(0f, size.height), Offset(size.width, size.height), strokeWidth = 1.dp.toPx())
                }
                .padding(horizontal = VeilSpacing.gutter),
            contentAlignment = Alignment.CenterStart,
        ) {
            Text(title, style = MaterialTheme.typography.titleMedium, color = VeilColors.content, maxLines = 1)
        }
    }
}

/** The bar of a detail screen: round back button, title, then round actions. */
@Composable
fun DetailBar(
    title: String,
    onBack: (() -> Unit)?,
    modifier: Modifier = Modifier,
    actions: @Composable RowScope.() -> Unit = {},
) {
    Row(
        modifier.fillMaxWidth().height(60.dp).padding(horizontal = VeilSpacing.medium),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (onBack != null) {
            RoundIconButton(VeilIcons.Back, contentDescription = "Back", onClick = onBack)
        }
        var titleModifier = Modifier.weight(1f)
        if (onBack == null) {
            titleModifier = titleModifier.padding(start = VeilSpacing.small)
        }
        Text(
            title,
            modifier = titleModifier,
            style = MaterialTheme.typography.titleLarge,
            color = VeilColors.content,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically, content = actions)
    }
}

/**
 * A section heading, with a chevron when it leads somewhere. Draws no side padding: inside a
 * grid the grid's gutter applies, in an edge-to-edge list add [gutterPadding].
 */
@Composable
fun SectionHeading(title: String, modifier: Modifier = Modifier, onClick: (() -> Unit)? = null) {
    var rowModifier = modifier
    if (onClick != null) {
        rowModifier = rowModifier.pressClickable(onClick)
    }
    Row(
        rowModifier.padding(top = VeilSpacing.large, bottom = VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
        if (onClick != null) {
            Icon(
                VeilIcons.Chevron,
                contentDescription = null,
                tint = VeilColors.contentMuted,
                modifier = Modifier.padding(start = VeilSpacing.extraSmall).size(18.dp),
            )
        }
    }
}

/** A small rounded label over an image: runtime, counts, NEW. */
@Composable
fun Badge(
    text: String,
    modifier: Modifier = Modifier,
    background: Color = VeilColors.imageLabel,
    foreground: Color = VeilColors.content,
) {
    Text(
        text,
        modifier = modifier
            .clip(VeilShapes.badge)
            .background(background)
            .padding(horizontal = 6.dp, vertical = VeilSpacing.hairline),
        color = foreground,
        style = MaterialTheme.typography.labelSmall,
        fontWeight = FontWeight.SemiBold,
    )
}
