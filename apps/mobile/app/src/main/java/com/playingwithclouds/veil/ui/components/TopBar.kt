package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.RowScope
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.TopAppBarScrollBehavior
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Shuffle
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.style.TextOverflow
import com.playingwithclouds.veil.ui.AppNavigator

/** The app bar of every screen: a back arrow on detail screens, none on tab screens. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VeilTopBar(
    title: String,
    onBack: (() -> Unit)? = null,
    actions: @Composable RowScope.() -> Unit = {},
) {
    VeilTopBarLayout(title, onBack, actions, scrollBehavior = null)
}

/** [VeilTopBar] that slides away while [scrollBehavior]'s content scrolls down. */
@ExperimentalMaterial3Api
@Composable
fun ScrollingTopBar(
    title: String,
    scrollBehavior: TopAppBarScrollBehavior,
    actions: @Composable RowScope.() -> Unit = {},
) {
    VeilTopBarLayout(title, onBack = null, actions, scrollBehavior)
}

/** The bar shared by [VeilTopBar] and [ScrollingTopBar]. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun VeilTopBarLayout(
    title: String,
    onBack: (() -> Unit)?,
    actions: @Composable RowScope.() -> Unit,
    scrollBehavior: TopAppBarScrollBehavior?,
) {
    TopAppBar(
        title = { Text(title, maxLines = 1, overflow = TextOverflow.Ellipsis) },
        navigationIcon = {
            if (onBack != null) {
                IconButton(onClick = onBack) {
                    Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                }
            }
        },
        actions = actions,
        scrollBehavior = scrollBehavior,
        colors = TopAppBarDefaults.topAppBarColors(
            containerColor = MaterialTheme.colorScheme.background,
            scrolledContainerColor = MaterialTheme.colorScheme.background,
        ),
    )
}

/** The gear of the tab screens: a menu with Random, Plugins and Settings. */
@Composable
fun SettingsMenuButton(navigator: AppNavigator) {
    var open by remember { mutableStateOf(false) }
    IconButton(onClick = { open = true }) {
        Icon(Icons.Filled.Settings, contentDescription = "More")
    }
    DropdownMenu(expanded = open, onDismissRequest = { open = false }) {
        MenuItem("Random", Icons.Filled.Shuffle) {
            open = false
            navigator.openRandom()
        }
        MenuItem("Plugins", Icons.Filled.Extension) {
            open = false
            navigator.openPlugins()
        }
        MenuItem("Settings", Icons.Filled.Settings) {
            open = false
            navigator.openSettings()
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
