package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.design.DetailBar
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilMenu
import com.playingwithclouds.veil.ui.design.VeilMenuItem
import com.playingwithclouds.veil.ui.theme.VeilColors

/** The bar of every detail screen: a round back button, the title and round actions. */
@Composable
fun VeilTopBar(
    title: String,
    onBack: (() -> Unit)? = null,
    actions: @Composable RowScope.() -> Unit = {},
) {
    DetailBar(title, onBack, Modifier.statusBarsPadding(), actions)
}

/** The magnifier in the headers of the tab screens; opens the search screen. */
@Composable
fun SearchButton(navigator: AppNavigator, bare: Boolean = false) {
    if (bare) {
        IconTap(VeilIcons.Search, contentDescription = "Search", onClick = { navigator.openSearch() }, tint = VeilColors.content)
        return
    }
    RoundIconButton(VeilIcons.Search, contentDescription = "Search", onClick = { navigator.openSearch() })
}

/**
 * The gear of the tab screens: a menu with Random, Plugins and Settings. [bare] draws just the
 * icon, for the compact Home bar.
 */
@Composable
fun SettingsMenuButton(navigator: AppNavigator, bare: Boolean = false) {
    var open by remember { mutableStateOf(false) }
    Box {
        if (bare) {
            IconTap(VeilIcons.Settings, contentDescription = "More", onClick = { open = true }, tint = VeilColors.content)
        } else {
            RoundIconButton(VeilIcons.Settings, contentDescription = "More", onClick = { open = true })
        }
        VeilMenu(expanded = open, onDismissRequest = { open = false }) {
            VeilMenuItem("Random", icon = VeilIcons.Random, onClick = {
                open = false
                navigator.openRandom()
            })
            VeilMenuItem("Previews", icon = VeilIcons.Play, onClick = {
                open = false
                navigator.openPreviewFeed()
            })
            VeilMenuItem("Taste", icon = VeilIcons.Like, onClick = {
                open = false
                navigator.openTaste()
            })
            VeilMenuItem("Plugins", icon = VeilIcons.Plugins, onClick = {
                open = false
                navigator.openPlugins()
            })
            VeilMenuItem("Settings", icon = VeilIcons.Settings, onClick = {
                open = false
                navigator.openSettings()
            })
        }
    }
}
