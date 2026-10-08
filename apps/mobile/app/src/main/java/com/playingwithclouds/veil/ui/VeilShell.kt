package com.playingwithclouds.veil.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.navigation.NavBackStackEntry
import com.playingwithclouds.veil.ui.design.LocalGlassBackdrop
import dev.chrisbanes.haze.hazeSource
import dev.chrisbanes.haze.rememberHazeState
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.playingwithclouds.veil.ui.collections.CollectionScreen
import com.playingwithclouds.veil.ui.collections.CollectionsScreen
import com.playingwithclouds.veil.ui.designkit.DesignKitScreen
import com.playingwithclouds.veil.ui.galleries.GalleriesScreen
import com.playingwithclouds.veil.ui.galleries.GalleryCategoryScreen
import com.playingwithclouds.veil.ui.galleries.GalleryScreen
import com.playingwithclouds.veil.ui.history.HistoryScreen
import com.playingwithclouds.veil.ui.home.HomeScreen
import com.playingwithclouds.veil.ui.library.LibraryHubScreen
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

/** The screens with the floating tab bar and system back. */
@Composable
fun VeilShell() {
    val navController = rememberNavController()
    val navigator = remember(navController) { AppNavigator(navController) }
    val hideOnScroll = rememberHideOnScrollState()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route
    val onTab = Tab.entries.any { tab -> tab.route == currentRoute }

    val glassBackdrop = rememberHazeState()

    LaunchedEffect(currentRoute) { hideOnScroll.show() }

    Scaffold { innerPadding ->
        Box(Modifier.fillMaxSize()) {
            NavHost(
                navController = navController,
                startDestination = Routes.HOME,
                modifier = Modifier
                    .padding(innerPadding)
                    .consumeWindowInsets(innerPadding)
                    .nestedScroll(hideOnScroll.connection)
                    .hazeSource(glassBackdrop),
                enterTransition = { PageTransitions.enter(initialState, targetState) },
                exitTransition = { PageTransitions.exit(initialState, targetState) },
                popEnterTransition = { PageTransitions.popEnter(initialState, targetState) },
                popExitTransition = { PageTransitions.popExit(initialState, targetState) },
            ) {
                tabScreens(navigator)
                detailScreens(navigator)
            }
            CompositionLocalProvider(LocalGlassBackdrop provides glassBackdrop) {
                FloatingNavBar(
                    currentRoute = currentRoute,
                    visible = onTab && hideOnScroll.visible,
                    onTab = { tab -> navigator.openTab(tab.route) },
                    onSearch = { navigator.openSearch() },
                    modifier = Modifier.align(Alignment.BottomCenter),
                )
            }
        }
    }
}

/** A tab screen, given room at the bottom for the floating bar. */
private fun NavGraphBuilder.tab(route: String, content: @Composable () -> Unit) {
    composable(route) {
        CompositionLocalProvider(LocalFloatingBarInset provides FloatingBarInset) { content() }
    }
}

/** The three tab destinations. */
private fun NavGraphBuilder.tabScreens(navigator: AppNavigator) {
    tab(Routes.HOME) { HomeScreen(navigator) }
    tab(Routes.SUBSCRIPTIONS) { SubscriptionsScreen(navigator) }
    tab(Routes.LIBRARY) { LibraryHubScreen(navigator) }
}

/** Every destination opened from a tab, the Library hub or a card. */
private fun NavGraphBuilder.detailScreens(navigator: AppNavigator) {
    composable(
        Routes.SEARCH,
        arguments = listOf(navArgument("query") { type = NavType.StringType; defaultValue = "" }),
    ) { entry -> SearchScreen(entry.argument("query"), navigator) }
    composable(Routes.COLLECTIONS) { CollectionsScreen(navigator) }
    composable(Routes.HISTORY) { HistoryScreen(navigator) }
    composable(Routes.LIBRARY_SECTION) { entry -> LibraryScreen(entry.argument("section"), navigator) }
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
    composable(Routes.DESIGN_KIT) { DesignKitScreen(navigator) }
    composable(Routes.SUBSCRIPTION) { entry -> SubscriptionScreen(entry.argument("id"), navigator) }
}

/** A string argument of the destination; empty when missing. */
private fun NavBackStackEntry.argument(name: String): String {
    return arguments?.getString(name).orEmpty()
}
