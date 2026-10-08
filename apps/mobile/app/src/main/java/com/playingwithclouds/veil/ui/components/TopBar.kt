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
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilMenu
import com.playingwithclouds.veil.ui.design.VeilMenuItem

/** The bar of every detail screen: a round back button, the title and round actions. */
@Composable
fun VeilTopBar(
    title: String,
    onBack: (() -> Unit)? = null,
    actions: @Composable RowScope.() -> Unit = {},
) {
    DetailBar(title, onBack, Modifier.statusBarsPadding(), actions)
}

/** The gear of the tab screens: a menu with Random, Plugins and Settings. */
@Composable
fun SettingsMenuButton(navigator: AppNavigator) {
    var open by remember { mutableStateOf(false) }
    Box {
        RoundIconButton(VeilIcons.Settings, contentDescription = "More", onClick = { open = true })
        VeilMenu(expanded = open, onDismissRequest = { open = false }) {
            VeilMenuItem("Random", icon = VeilIcons.Random, onClick = {
                open = false
                navigator.openRandom()
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
