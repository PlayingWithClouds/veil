package com.playingwithclouds.veil.ui

import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Collections
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.VideoLibrary
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.Icon
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.NavBackStackEntry
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.playingwithclouds.veil.ui.collections.CollectionScreen
import com.playingwithclouds.veil.ui.collections.CollectionsScreen
import com.playingwithclouds.veil.ui.galleries.GalleriesScreen
import com.playingwithclouds.veil.ui.galleries.GalleryCategoryScreen
import com.playingwithclouds.veil.ui.galleries.GalleryScreen
import com.playingwithclouds.veil.ui.history.HistoryScreen
import com.playingwithclouds.veil.ui.home.HomeScreen
import com.playingwithclouds.veil.ui.library.LibraryScreen
import com.playingwithclouds.veil.ui.performers.PerformerScreen
import com.playingwithclouds.veil.ui.performers.PerformersScreen
import com.playingwithclouds.veil.ui.plugins.PluginsScreen
import com.playingwithclouds.veil.ui.random.RandomScreen
import com.playingwithclouds.veil.ui.scene.SceneScreen
import com.playingwithclouds.veil.ui.search.SearchScreen
import com.playingwithclouds.veil.ui.settings.SettingsScreen
import com.playingwithclouds.veil.ui.studios.StudioScreen
import com.playingwithclouds.veil.ui.studios.StudiosScreen
import com.playingwithclouds.veil.ui.subscriptions.SubscriptionScreen
import com.playingwithclouds.veil.ui.subscriptions.SubscriptionsScreen
import com.playingwithclouds.veil.ui.tags.TagScreen
import com.playingwithclouds.veil.ui.tags.TagsScreen
import kotlinx.coroutines.launch

/** A destination of the bottom bar. */
private enum class Tab(val route: String, val label: String, val icon: ImageVector) {
    HOME(Routes.HOME, "Home", Icons.Filled.Home),
    SUBSCRIPTIONS(Routes.SUBSCRIPTIONS, "Following", Icons.Filled.Notifications),
    LIBRARY(Routes.LIBRARY, "Library", Icons.Filled.VideoLibrary),
    COLLECTIONS(Routes.COLLECTIONS, "Collections", Icons.Filled.Collections),
    HISTORY(Routes.HISTORY, "History", Icons.Filled.History),
}

/** The screens with the bottom bar, the menu drawer and system back. */
@Composable
fun VeilShell() {
    val navController = rememberNavController()
    val navigator = remember(navController) { AppNavigator(navController) }
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route
    val openMenu: () -> Unit = { scope.launch { drawerState.open() } }

    ModalNavigationDrawer(
        drawerState = drawerState,
        // The edge swipe would fight the system back gesture; the menu button opens it.
        gesturesEnabled = drawerState.isOpen,
        drawerContent = {
            MenuDrawer(
                navigator = navigator,
                closeDrawer = { scope.launch { drawerState.close() } },
            )
        },
    ) {
        Scaffold(
            bottomBar = {
                if (Tab.entries.any { tab -> tab.route == currentRoute }) {
                    VeilBottomBar(currentRoute, navigator)
                }
            },
        ) { innerPadding ->
            NavHost(
                navController = navController,
                startDestination = Routes.HOME,
                modifier = Modifier.padding(innerPadding).consumeWindowInsets(innerPadding),
                enterTransition = { fadeIn() },
                exitTransition = { fadeOut() },
            ) {
                tabScreens(navigator, openMenu)
                detailScreens(navigator)
            }
        }
    }
}

/** The bottom bar with one item per tab. */
@Composable
private fun VeilBottomBar(currentRoute: String?, navigator: AppNavigator) {
    NavigationBar {
        for (tab in Tab.entries) {
            NavigationBarItem(
                selected = tab.route == currentRoute,
                onClick = { navigator.openTab(tab.route) },
                icon = { Icon(tab.icon, contentDescription = tab.label) },
                label = { Text(tab.label) },
            )
        }
    }
}

/** The five tab destinations. */
private fun NavGraphBuilder.tabScreens(navigator: AppNavigator, openMenu: () -> Unit) {
    composable(Routes.HOME) { HomeScreen(navigator, openMenu) }
    composable(Routes.SUBSCRIPTIONS) { SubscriptionsScreen(navigator, openMenu) }
    composable(Routes.LIBRARY) { LibraryScreen(navigator, openMenu) }
    composable(Routes.COLLECTIONS) { CollectionsScreen(navigator, openMenu) }
    composable(Routes.HISTORY) { HistoryScreen(navigator, openMenu) }
}

/** Every destination opened from a tab, the menu or a card. */
private fun NavGraphBuilder.detailScreens(navigator: AppNavigator) {
    composable(
        Routes.SEARCH,
        arguments = listOf(navArgument("query") { type = NavType.StringType; defaultValue = "" }),
    ) { entry -> SearchScreen(entry.argument("query"), navigator) }
    composable(Routes.SCENE) { entry -> SceneScreen(entry.argument("id"), navigator) }
    composable(Routes.PERFORMERS) { PerformersScreen(navigator) }
    composable(Routes.PERFORMER) { entry -> PerformerScreen(entry.argument("id"), navigator) }
    composable(Routes.STUDIOS) { StudiosScreen(navigator) }
    composable(Routes.STUDIO) { entry -> StudioScreen(entry.argument("id"), navigator) }
    composable(Routes.TAGS) { TagsScreen(navigator) }
    composable(Routes.TAG) { entry -> TagScreen(entry.argument("id"), navigator) }
    composable(Routes.COLLECTION) { entry -> CollectionScreen(entry.argument("id"), navigator) }
    composable(Routes.GALLERIES) { GalleriesScreen(navigator) }
    composable(Routes.GALLERY_CATEGORY) { entry -> GalleryCategoryScreen(entry.argument("name"), navigator) }
    composable(Routes.GALLERY) { entry -> GalleryScreen(entry.argument("id"), navigator) }
    composable(Routes.PLUGINS) { PluginsScreen(navigator) }
    composable(Routes.SETTINGS) { SettingsScreen(navigator) }
    composable(Routes.RANDOM) { RandomScreen(navigator) }
    composable(Routes.SUBSCRIPTION) { entry -> SubscriptionScreen(entry.argument("id"), navigator) }
}

/** A string argument of the destination; empty when missing. */
private fun NavBackStackEntry.argument(name: String): String {
    return arguments?.getString(name).orEmpty()
}
