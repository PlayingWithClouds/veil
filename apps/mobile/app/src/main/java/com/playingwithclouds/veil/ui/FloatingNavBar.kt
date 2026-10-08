package com.playingwithclouds.veil.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.GlassIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.glass
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Height of the floating bar plus its gap to the screen edge. */
private val FloatingBarHeight = 64.dp
private val FloatingBarGap = 12.dp

/** Scroll distance in pixels that hides or reveals the bar, so small jitters don't toggle it. */
private const val HIDE_THRESHOLD_PIXELS = 24f

/**
 * Bottom space a tab screen's scrolling content needs so its last item can scroll clear of the
 * floating bar. Zero on screens without the bar.
 */
val LocalFloatingBarInset = compositionLocalOf { 0.dp }

/** The inset a tab screen provides through [LocalFloatingBarInset]. */
val FloatingBarInset: Dp = FloatingBarHeight + FloatingBarGap * 2

/** A destination of the floating bar. */
enum class Tab(val route: String, val label: String, val selectedIcon: ImageVector, val icon: ImageVector) {
    HOME(Routes.HOME, "Home", VeilIcons.HomeFilled, VeilIcons.Home),
    FOLLOWING(Routes.SUBSCRIPTIONS, "Following", VeilIcons.FollowingFilled, VeilIcons.Following),
    LIBRARY(Routes.LIBRARY, "Library", VeilIcons.LibraryFilled, VeilIcons.Library),
}

/** Whether the floating bar is shown: it slides away while scrolling down and back on scrolling up. */
class HideOnScrollState {
    var visible by mutableStateOf(true)
        private set

    private var travelled = 0f

    /** Feeds into the shell's nested scroll so every scrolling list drives the bar. */
    val connection = object : NestedScrollConnection {
        override fun onPreScroll(available: Offset, source: NestedScrollSource): Offset {
            track(available.y)
            return Offset.Zero
        }
    }

    /** Shows the bar again, e.g. after switching screens. */
    fun show() {
        visible = true
        travelled = 0f
    }

    /** Accumulates scroll in one direction and flips visibility once it passes the threshold. */
    private fun track(deltaY: Float) {
        if (deltaY == 0f) {
            return
        }
        if ((deltaY < 0f) != (travelled < 0f)) {
            travelled = 0f
        }
        travelled += deltaY
        if (travelled < -HIDE_THRESHOLD_PIXELS) {
            visible = false
        }
        if (travelled > HIDE_THRESHOLD_PIXELS) {
            visible = true
        }
    }
}

@Composable
fun rememberHideOnScrollState(): HideOnScrollState {
    return remember { HideOnScrollState() }
}

/** Width of one tab slot; the selection pill slides between slots. */
private val TabWidth = 84.dp

/** Inner padding of the glass pill around the tab slots. */
private val PillPadding = 6.dp

/** The detached glass pill with the three tabs, and the round glass search button beside it. */
@Composable
fun FloatingNavBar(
    currentRoute: String?,
    visible: Boolean,
    onTab: (Tab) -> Unit,
    onSearch: () -> Unit,
    modifier: Modifier = Modifier,
) {
    AnimatedVisibility(
        visible = visible,
        modifier = modifier,
        enter = slideInVertically(spring(dampingRatio = Spring.DampingRatioLowBouncy, stiffness = Spring.StiffnessMediumLow)) { height -> height } + fadeIn(),
        exit = slideOutVertically { height -> height } + fadeOut(),
    ) {
        Row(
            Modifier.navigationBarsPadding().padding(horizontal = 20.dp, vertical = FloatingBarGap),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TabPill(currentRoute, onTab)
            GlassIconButton(VeilIcons.Search, contentDescription = "Search", onClick = onSearch, size = FloatingBarHeight)
        }
    }
}

/** The tab slots over a selection pill that springs to the current tab. */
@Composable
private fun TabPill(currentRoute: String?, onTab: (Tab) -> Unit) {
    val selectedIndex = Tab.entries.indexOfFirst { tab -> tab.route == currentRoute }.coerceAtLeast(0)
    val indicatorOffset by animateDpAsState(
        TabWidth * selectedIndex,
        animationSpec = spring(dampingRatio = Spring.DampingRatioLowBouncy, stiffness = Spring.StiffnessMediumLow),
        label = "indicator",
    )
    Box(Modifier.height(FloatingBarHeight).glass(CircleShape).padding(PillPadding)) {
        Box(
            Modifier
                .offset(x = indicatorOffset)
                .width(TabWidth)
                .fillMaxHeight()
                .clip(CircleShape)
                .background(Color.White.copy(alpha = 0.14f)),
        )
        Row(Modifier.fillMaxHeight(), verticalAlignment = Alignment.CenterVertically) {
            for (tab in Tab.entries) {
                TabItem(tab, selected = tab.route == currentRoute, onClick = { onTab(tab) })
            }
        }
    }
}

/** One tab slot: icon over label, bright and filled when selected. */
@Composable
private fun TabItem(tab: Tab, selected: Boolean, onClick: () -> Unit) {
    val tint by animateColorAsState(if (selected) VeilColors.content else VeilColors.contentMuted, label = "tint")
    Column(
        Modifier.width(TabWidth).fillMaxHeight().clip(CircleShape).pressClickable(onClick),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(if (selected) tab.selectedIcon else tab.icon, contentDescription = null, tint = tint, modifier = Modifier.size(22.dp))
        Text(tab.label, color = tint, style = MaterialTheme.typography.labelSmall)
    }
}
