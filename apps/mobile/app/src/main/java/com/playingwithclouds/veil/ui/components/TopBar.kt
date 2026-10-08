package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.design.DetailBar
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
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

/** The gear of the tab screens: a menu with Random, Plugins and Settings. */
@Composable
fun SettingsMenuButton(navigator: AppNavigator) {
    var open by remember { mutableStateOf(false) }
    Box {
        RoundIconButton(VeilIcons.Settings, contentDescription = "More", onClick = { open = true })
        DropdownMenu(
            expanded = open,
            onDismissRequest = { open = false },
            shape = RoundedCornerShape(18.dp),
            containerColor = VeilColors.surfaceHigh,
        ) {
            MenuItem("Random", VeilIcons.Random) {
                open = false
                navigator.openRandom()
            }
            MenuItem("Plugins", VeilIcons.Plugins) {
                open = false
                navigator.openPlugins()
            }
            MenuItem("Settings", VeilIcons.Settings) {
                open = false
                navigator.openSettings()
            }
        }
    }
}

/** One entry of [SettingsMenuButton]. */
@Composable
private fun MenuItem(label: String, icon: ImageVector, onClick: () -> Unit) {
    DropdownMenuItem(
        text = { Text(label) },
        leadingIcon = { Icon(icon, contentDescription = null) },
        onClick = onClick,
    )
}
