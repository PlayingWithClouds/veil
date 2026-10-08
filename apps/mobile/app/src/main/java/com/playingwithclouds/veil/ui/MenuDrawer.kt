package com.playingwithclouds.veil.ui

import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material.icons.filled.Groups
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material.icons.filled.Business
import androidx.compose.material.icons.filled.Sell
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Shuffle
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalDrawerSheet
import androidx.compose.material3.NavigationDrawerItem
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp

/** One entry of the drawer. */
private data class MenuEntry(val label: String, val icon: ImageVector, val open: (AppNavigator) -> Unit)

private val browseEntries = listOf(
    MenuEntry("Random", Icons.Filled.Shuffle) { navigator -> navigator.openRandom() },
    MenuEntry("Performers", Icons.Filled.Groups) { navigator -> navigator.openPerformers() },
    MenuEntry("Studios", Icons.Filled.Business) { navigator -> navigator.openStudios() },
    MenuEntry("Tags", Icons.Filled.Sell) { navigator -> navigator.openTags() },
    MenuEntry("Galleries", Icons.Filled.PhotoLibrary) { navigator -> navigator.openGalleries() },
)

private val manageEntries = listOf(
    MenuEntry("Plugins", Icons.Filled.Extension) { navigator -> navigator.openPlugins() },
    MenuEntry("Settings", Icons.Filled.Settings) { navigator -> navigator.openSettings() },
)

/** The side menu: browse pages that are not in the bottom bar, then plugins and settings. */
@Composable
fun MenuDrawer(navigator: AppNavigator, closeDrawer: () -> Unit) {
    ModalDrawerSheet {
        Text("Veil", Modifier.padding(horizontal = 28.dp, vertical = 24.dp), style = MaterialTheme.typography.titleLarge)
        MenuItems(browseEntries, navigator, closeDrawer)
        Spacer(Modifier.height(8.dp))
        HorizontalDivider(Modifier.padding(horizontal = 28.dp))
        Spacer(Modifier.height(8.dp))
        MenuItems(manageEntries, navigator, closeDrawer)
    }
}

/** A group of drawer items; choosing one closes the drawer and opens its page. */
@Composable
private fun MenuItems(entries: List<MenuEntry>, navigator: AppNavigator, closeDrawer: () -> Unit) {
    for (entry in entries) {
        NavigationDrawerItem(
            label = { Text(entry.label) },
            icon = { Icon(entry.icon, contentDescription = null) },
            selected = false,
            onClick = {
                closeDrawer()
                entry.open(navigator)
            },
            modifier = Modifier.padding(horizontal = 12.dp),
        )
    }
}
