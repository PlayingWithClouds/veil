package com.playingwithclouds.veil.ui.design

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.dp
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors

/** A filled capsule label, optionally with a leading icon: white with dark text when selected, a dark fill otherwise. */
@Composable
fun Pill(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    selected: Boolean = false,
    icon: ImageVector? = null,
) {
    val background by animateColorAsState(if (selected) VeilColors.content else VeilColors.surfaceHigh, label = "pill")
    val foreground by animateColorAsState(if (selected) VeilColors.canvas else VeilColors.content, label = "pillText")
    Row(
        modifier
            .height(36.dp)
            .pressClickable(onClick)
            .clip(CircleShape)
            .background(background)
            .padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (icon != null) {
            Icon(icon, contentDescription = null, tint = foreground, modifier = Modifier.size(16.dp))
        }
        Text(text, color = foreground, style = MaterialTheme.typography.labelLarge, maxLines = 1)
    }
}

/** A horizontally scrolling row of pills, one of which may be selected. */
@Composable
fun <T> PillRow(
    options: List<T>,
    labelOf: (T) -> String,
    onClick: (T) -> Unit,
    modifier: Modifier = Modifier,
    isSelected: (T) -> Boolean = { false },
    contentPadding: PaddingValues = PaddingValues(horizontal = 16.dp),
) {
    LazyRow(modifier, contentPadding = contentPadding, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        itemsIndexed(options) { _, option ->
            Pill(labelOf(option), onClick = { onClick(option) }, selected = isSelected(option))
        }
    }
}

/**
 * Equal-width segments in a dark capsule with a white thumb that springs to the selected one:
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
            .clip(CircleShape)
            .background(VeilColors.surfaceHigh)
            .padding(3.dp)
            .onSizeChanged { size -> width = size.width },
    ) {
        Box(Modifier.offset(x = thumbOffset).width(segmentWidth).fillMaxHeight().clip(CircleShape).background(VeilColors.content))
        Row(Modifier.fillMaxWidth().fillMaxHeight()) {
            labels.forEachIndexed { index, label ->
                val color by animateColorAsState(
                    if (index == selectedIndex) VeilColors.canvas else VeilColors.contentMuted,
                    label = "segment",
                )
                Box(
                    Modifier.weight(1f).fillMaxHeight().clip(CircleShape).pressClickable { onSelect(index) },
                    contentAlignment = Alignment.Center,
                ) {
                    Text(label, color = color, style = MaterialTheme.typography.labelLarge, maxLines = 1)
                }
            }
        }
    }
}
