package com.playingwithclouds.veil.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.AnimationVector1D
import androidx.compose.animation.core.FastOutLinearInEasing
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.glass
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.launch
import kotlin.math.roundToInt

/** Height of the floating bar plus its gap to the screen edge. */
private val FloatingBarHeight = 64.dp
private val FloatingBarGap = VeilSpacing.medium

/** Scroll distance in pixels that hides or reveals the bar, so small jitters don't toggle it. */
private const val HIDE_THRESHOLD_PIXELS = 24f

/**
 * Bottom space a tab screen's scrolling content needs so its last item can scroll clear of the
 * floating bar. Zero on screens without the bar.
 */
val LocalFloatingBarInset = compositionLocalOf { 0.dp }

/** The inset a tab screen provides through [LocalFloatingBarInset]. */
val FloatingBarInset: Dp = FloatingBarHeight + FloatingBarGap * 2

/** A page of the tab pager, in swipe order. */
enum class Tab(val label: String, val selectedIcon: ImageVector, val icon: ImageVector) {
    HOME("Home", VeilIcons.HomeFilled, VeilIcons.Home),
    FOLLOWING("Following", VeilIcons.FollowingFilled, VeilIcons.Following),
    COLLECTIONS("Collections", VeilIcons.CollectionsFilled, VeilIcons.Collections),
    LIBRARY("Library", VeilIcons.LibraryFilled, VeilIcons.Library),
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

/** Inner padding of the glass pill around the tab slots. */
private val PillPadding = 6.dp

/**
 * The detached glass pill with the tabs, and the round glass search button beside it.
 * [pagerPosition] is the tab pager's position in pages (fractional mid-swipe), so the selection
 * pill follows the finger.
 */
@Composable
fun FloatingNavBar(
    pagerPosition: () -> Float,
    selected: Tab,
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
            Modifier.fillMaxWidth().navigationBarsPadding().padding(horizontal = FloatingBarGap, vertical = FloatingBarGap),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(Modifier.weight(1f)) {
                TabPill(pagerPosition, selected, onTab)
            }
            Box(
                Modifier.size(FloatingBarHeight).pressClickable(onSearch).glass(VeilShapes.capsule),
                contentAlignment = Alignment.Center,
            ) {
                Icon(VeilIcons.Search, contentDescription = "Search", tint = VeilColors.contentMuted, modifier = Modifier.size(22.dp))
            }
        }
    }
}

/** The tab slots over a selection pill that sits at the pager's position. */
@Composable
private fun TabPill(pagerPosition: () -> Float, selected: Tab, onTab: (Tab) -> Unit) {
    BoxWithConstraints(Modifier.fillMaxWidth().height(FloatingBarHeight).glass(VeilShapes.capsule).padding(PillPadding)) {
        val tabWidth = maxWidth / Tab.entries.size
        Box(
            Modifier
                .offset { IntOffset((tabWidth.toPx() * pagerPosition()).roundToInt(), 0) }
                .width(tabWidth)
                .fillMaxHeight()
                .clip(VeilShapes.capsule)
                .background(VeilColors.glassSelection),
        )
        Row(Modifier.fillMaxHeight(), verticalAlignment = Alignment.CenterVertically) {
            for (tab in Tab.entries) {
                Box(Modifier.width(tabWidth)) {
                    TabItem(tab, selected = tab == selected, onClick = { onTab(tab) })
                }
            }
        }
    }
}

/** How far a tab icon shrinks when tapped before it springs back. */
private const val TAP_SCALE = 0.72f

/** One tab slot: icon over label, in the accent and filled when selected; a tap plops the icon. */
@Composable
private fun TabItem(tab: Tab, selected: Boolean, onClick: () -> Unit) {
    val tint by animateColorAsState(if (selected) VeilColors.accent else VeilColors.contentMuted, label = "tint")
    val iconScale = remember { Animatable(1f) }
    val scope = rememberCoroutineScope()
    Column(
        Modifier
            .fillMaxWidth()
            .fillMaxHeight()
            .clickable(interactionSource = null, indication = null) {
                scope.launch { plop(iconScale) }
                onClick()
            },
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(
            if (selected) tab.selectedIcon else tab.icon,
            contentDescription = null,
            tint = tint,
            modifier = Modifier.size(22.dp).graphicsLayer {
                scaleX = iconScale.value
                scaleY = iconScale.value
            },
        )
        // Long labels may run a hair past a narrow slot rather than being cut or wrapped.
        Text(tab.label, color = tint, style = MaterialTheme.typography.labelSmall, maxLines = 1, softWrap = false, overflow = TextOverflow.Visible)
    }
}

/** Squeezes [scale] down quickly, then lets it bounce back past full size and settle. */
private suspend fun plop(scale: Animatable<Float, AnimationVector1D>) {
    scale.animateTo(TAP_SCALE, tween(durationMillis = 90, easing = FastOutLinearInEasing))
    scale.animateTo(1f, spring(dampingRatio = Spring.DampingRatioHighBouncy, stiffness = Spring.StiffnessMedium))
}
