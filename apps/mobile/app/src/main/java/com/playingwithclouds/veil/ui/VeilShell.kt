package com.playingwithclouds.veil.ui

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.PagerState
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveableStateHolder
import kotlinx.coroutines.launch
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
    val pagerState = rememberPagerState { Tab.entries.size }
    val scope = rememberCoroutineScope()
    val navigator = remember(navController) {
        AppNavigator(navController) { tab -> scope.launch { pagerState.animateScrollToPage(tab.ordinal) } }
    }
    val hideOnScroll = rememberHideOnScrollState()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route
    val onTabs = currentRoute == Routes.TABS

    val glassBackdrop = rememberHazeState()

    LaunchedEffect(currentRoute, pagerState.currentPage) { hideOnScroll.show() }
    BackHandler(enabled = onTabs && pagerState.currentPage != Tab.HOME.ordinal) {
        navigator.openTab(Tab.HOME)
    }

    Scaffold { innerPadding ->
        Box(Modifier.fillMaxSize()) {
            NavHost(
                navController = navController,
                startDestination = Routes.TABS,
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
                composable(Routes.TABS) { TabPager(pagerState, navigator) }
                detailScreens(navigator)
            }
            CompositionLocalProvider(LocalGlassBackdrop provides glassBackdrop) {
                FloatingNavBar(
                    pagerPosition = { pagerState.currentPage + pagerState.currentPageOffsetFraction },
                    selected = Tab.entries[pagerState.currentPage],
                    visible = onTabs && hideOnScroll.visible,
                    onTab = navigator::openTab,
                    modifier = Modifier.align(Alignment.BottomCenter),
                )
            }
        }
    }
}

/**
 * The tab screens side by side: swipe sideways to move between them. Each page keeps its saved
 * state (scroll position) while swiped away, and has room at the bottom for the floating bar.
 */
@Composable
private fun TabPager(pagerState: PagerState, navigator: AppNavigator) {
    val savedStates = rememberSaveableStateHolder()
    HorizontalPager(pagerState, Modifier.fillMaxSize(), key = { page -> Tab.entries[page].name }) { page ->
        val tab = Tab.entries[page]
        savedStates.SaveableStateProvider(tab.name) {
            CompositionLocalProvider(LocalFloatingBarInset provides FloatingBarInset) {
                TabScreen(tab, navigator)
            }
        }
    }
}

/** The screen of one tab. */
@Composable
private fun TabScreen(tab: Tab, navigator: AppNavigator) {
    when (tab) {
        Tab.HOME -> HomeScreen(navigator)
        Tab.FOLLOWING -> SubscriptionsScreen(navigator)
        Tab.GALLERIES -> GalleriesScreen(navigator)
        Tab.LIBRARY -> LibraryHubScreen(navigator)
    }
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
