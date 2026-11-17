package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.theme.VeilType
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LocalOpenDrawer
import com.playingwithclouds.veil.ui.design.DetailBar
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.NoticeDot
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.update.UpdateChecker

/**
 * A detail screen: its back bar (round back button, [title], round actions) floats over the
 * content, see-through until the content scrolls under it, then glass. The content draws from the
 * top edge; lists start below the bar with [LocalTopBarInset] (`PagedGrid` does on its own).
 */
@Composable
fun DetailScreen(
    title: String,
    onBack: (() -> Unit)?,
    modifier: Modifier = Modifier,
    snackbarHostState: SnackbarHostState? = null,
    actions: @Composable RowScope.() -> Unit = {},
    content: @Composable () -> Unit,
) {
    Scaffold(
        modifier,
        snackbarHost = {
            if (snackbarHostState != null) {
                VeilSnackbarHost(snackbarHostState)
            }
        },
    ) { padding ->
        GlassTopBarLayout(
            bar = { DetailBar(title, onBack, Modifier.statusBarsPadding(), actions) },
            modifier = Modifier.padding(bottom = padding.calculateBottomPadding()),
            content = content,
        )
    }
}

/** Height of a tab screen's top bar under the status bar. */
private val TabTopBarHeight = 52.dp

/**
 * The slim bar of the tab screens: ☰, then the wordmark on Home or the tab's [title], then search.
 * The magnifier opens the site search unless the tab searches itself through [onSearch].
 * Draws under the status bar on its own.
 */
@Composable
fun TabTopBar(navigator: AppNavigator, title: String? = null, onSearch: (() -> Unit)? = null) {
    Row(
        Modifier.fillMaxWidth().statusBarsPadding().height(TabTopBarHeight).padding(horizontal = VeilSpacing.extraSmall),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
    ) {
        MenuButton()
        if (title == null) {
            Text("Veil", Modifier.weight(1f), style = VeilType.wordmark, color = VeilColors.content, maxLines = 1)
        } else {
            Text(title, Modifier.weight(1f), style = MaterialTheme.typography.titleLarge, color = VeilColors.content, maxLines = 1)
        }
        if (onSearch == null) {
            SearchButton(onClick = navigator::openSearch)
        } else {
            SearchButton(onClick = onSearch)
        }
    }
}

/** The magnifier in the tab screens' top bars. */
@Composable
fun SearchButton(onClick: () -> Unit) {
    IconTap(VeilIcons.Search, contentDescription = "Search", onClick = onClick, tint = VeilColors.content)
}

/** The ☰ at the start of the tab screens' top bars; opens the side menu. A dot marks an update the user hasn't skipped. */
@Composable
fun MenuButton() {
    val openDrawer = LocalOpenDrawer.current
    val updateMarked by UpdateChecker.hasUnseenUpdate.collectAsStateWithLifecycle()
    Box {
        IconTap(VeilIcons.Menu, contentDescription = "Menu", onClick = openDrawer, tint = VeilColors.content)
        if (updateMarked) {
            NoticeDot(Modifier.align(Alignment.TopEnd).padding(VeilSpacing.extraSmall))
        }
    }
}
