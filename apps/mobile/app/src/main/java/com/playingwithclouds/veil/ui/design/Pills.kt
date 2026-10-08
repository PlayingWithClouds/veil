package com.playingwithclouds.veil.ui.design

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * A choice capsule, optionally with a leading icon (or any [leading] content, e.g. a site's
 * favicon): accent with dark text when selected, a dark fill otherwise. For filters and options;
 * tags use [TagPill].
 */
@Composable
fun Pill(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    selected: Boolean = false,
    icon: ImageVector? = null,
    leading: (@Composable () -> Unit)? = null,
) {
    val foreground by animateColorAsState(if (selected) VeilColors.onAccent else VeilColors.content, label = "pillText")
    var surface = Modifier.glassControl(VeilShapes.capsule)
    if (selected) {
        surface = Modifier.clip(VeilShapes.capsule).background(VeilColors.accent)
    }
    Row(
        modifier
            .height(36.dp)
            .pressClickable(onClick)
            .then(surface)
            .padding(horizontal = VeilSpacing.large),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (icon != null) {
            Icon(icon, contentDescription = null, tint = foreground, modifier = Modifier.size(16.dp))
        }
        if (leading != null) {
            leading()
        }
        Text(text, color = foreground, style = MaterialTheme.typography.labelLarge, maxLines = 1)
    }
}

/** A tag as a link: a lighter outlined capsule, so tags read differently from choices. */
@Composable
fun TagPill(text: String, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Text(
        text,
        modifier = modifier
            .height(32.dp)
            .pressClickable(onClick)
            .glassControl(VeilShapes.capsule)
            .padding(horizontal = VeilSpacing.medium, vertical = 7.dp),
        color = VeilColors.contentMuted,
        style = MaterialTheme.typography.labelMedium,
        maxLines = 1,
    )
}

/**
 * A horizontally scrolling row of pills, one of which may be selected. Meant to span the screen
 * (add [bleed] inside a padded grid): pills start at the gutter and fade out where clipped.
 */
@Composable
fun <T> PillRow(
    options: List<T>,
    labelOf: (T) -> String,
    onClick: (T) -> Unit,
    modifier: Modifier = Modifier,
    isSelected: (T) -> Boolean = { false },
) {
    LazyRow(
        modifier.fillMaxWidth().fadingEdges(),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        itemsIndexed(options) { _, option ->
            Pill(labelOf(option), onClick = { onClick(option) }, selected = isSelected(option))
        }
    }
}

/**
 * Equal-width segments in a glass capsule with a lighter thumb that springs to the selected one,
 * like the tab bar:
 * the app's replacement for tab rows.
 */
@Composable
fun SegmentedControl(
    labels: List<String>,
    selectedIndex: Int,
    onSelect: (Int) -> Unit,
    modifier: Modifier = Modifier,
) {
    var width by remember { mutableIntStateOf(0) }
    val segmentWidth = with(LocalDensity.current) { (width / labels.size.coerceAtLeast(1)).toDp() }
    val thumbOffset by animateDpAsState(
        segmentWidth * selectedIndex,
        animationSpec = spring(dampingRatio = Spring.DampingRatioLowBouncy, stiffness = Spring.StiffnessMediumLow),
        label = "thumb",
    )
    Box(
        modifier
            .fillMaxWidth()
            .height(40.dp)
            .glassControl(VeilShapes.capsule)
            .padding(3.dp)
            .onSizeChanged { size -> width = size.width },
    ) {
        Box(Modifier.offset(x = thumbOffset).width(segmentWidth).fillMaxHeight().clip(VeilShapes.capsule).background(VeilColors.glassSelection))
        Row(Modifier.fillMaxWidth().fillMaxHeight()) {
            labels.forEachIndexed { index, label ->
                val color by animateColorAsState(
                    if (index == selectedIndex) VeilColors.content else VeilColors.contentMuted,
                    label = "segment",
                )
                Box(
                    Modifier.weight(1f).fillMaxHeight().clip(VeilShapes.capsule).pressClickable { onSelect(index) },
                    contentAlignment = Alignment.Center,
                ) {
                    Text(label, color = color, style = MaterialTheme.typography.labelLarge, maxLines = 1)
                }
            }
        }
    }
}
