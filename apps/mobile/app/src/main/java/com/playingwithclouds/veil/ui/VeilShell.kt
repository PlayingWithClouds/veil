package com.playingwithclouds.veil.ui

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.PagerState
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.Scaffold
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveableStateHolder
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.navigation.NavBackStackEntry
import androidx.navigation.NavGraphBuilder
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.playingwithclouds.veil.ui.cams.CamRoomScreen
import com.playingwithclouds.veil.ui.cams.CamsScreen
import com.playingwithclouds.veil.ui.design.LocalGlassBackdrop
import com.playingwithclouds.veil.ui.designkit.DesignKitScreen
import com.playingwithclouds.veil.ui.galleries.GalleriesScreen
import com.playingwithclouds.veil.ui.galleries.GalleryScreen
import com.playingwithclouds.veil.ui.home.HomeScreen
import com.playingwithclouds.veil.ui.media.AudioScreen
import com.playingwithclouds.veil.ui.media.AudiosScreen
import com.playingwithclouds.veil.ui.media.StoriesScreen
import com.playingwithclouds.veil.ui.media.StoryReaderScreen
import com.playingwithclouds.veil.ui.media.StoryScreen
import com.playingwithclouds.veil.ui.library.CollectionScreen
import com.playingwithclouds.veil.ui.library.LibraryScreen
import com.playingwithclouds.veil.data.VideoScope
import com.playingwithclouds.veil.ui.categories.CategoriesScreen
import com.playingwithclouds.veil.ui.categories.ScopedVideosScreen
import com.playingwithclouds.veil.ui.plugins.PluginsScreen
import com.playingwithclouds.veil.ui.taste.TasteScreen
import com.playingwithclouds.veil.ui.people.MemberKind
import com.playingwithclouds.veil.ui.people.MembersScreen
import com.playingwithclouds.veil.ui.people.PersonScreen
import com.playingwithclouds.veil.ui.search.SearchScreen
import androidx.navigation.NavType
import androidx.navigation.navArgument
import com.playingwithclouds.veil.ui.settings.SettingsScreen
import com.playingwithclouds.veil.ui.video.VideoScreen
import dev.chrisbanes.haze.hazeSource
import dev.chrisbanes.haze.rememberHazeState
import kotlinx.coroutines.launch

/** Opens the side menu; the tab screens' top bars call it from their menu button. */
val LocalOpenDrawer = staticCompositionLocalOf<() -> Unit> { {} }

/**
 * The app's frame: the side menu with every top-level page, over the tab screens (Home, Cams,
 * Library) with their floating bar and the screens opened from them. The menu opens from the menu
 * button or a swipe from the left edge while a tab shows; the tabs change only by tapping, so the
 * swipe stays the menu's.
 */
@Composable
fun VeilShell() {
    val navController = rememberNavController()
    val pagerState = rememberPagerState { Tab.entries.size }
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val navigator = remember(navController) {
        AppNavigator(navController) { tab -> scope.launch { pagerState.animateScrollToPage(tab.ordinal) } }
    }
    val hideOnScroll = rememberHideOnScrollState()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route
    val onTabs = currentRoute == Routes.TABS
    val glassBackdrop = rememberHazeState()

    var currentDestination: Destination? = null
    if (onTabs) {
        currentDestination = Destination.entries.first { destination -> destination.tab == Tab.entries[pagerState.currentPage] }
    }

    LaunchedEffect(currentRoute, pagerState.currentPage) { hideOnScroll.show() }
    BackHandler(enabled = drawerState.isOpen) { scope.launch { drawerState.close() } }
    BackHandler(enabled = onTabs && drawerState.isClosed && pagerState.currentPage != Tab.HOME.ordinal) {
        navigator.openTab(Tab.HOME)
    }

    ModalNavigationDrawer(
        drawerState = drawerState,
        gesturesEnabled = onTabs || drawerState.isOpen,
        drawerContent = {
            MainDrawer(
                current = currentDestination,
                onDestination = { destination ->
                    scope.launch { drawerState.close() }
                    navigator.openDestination(destination)
                },
                onSettings = {
                    scope.launch { drawerState.close() }
                    navigator.openSettings()
                },
            )
        },
    ) {
        CompositionLocalProvider(LocalOpenDrawer provides { scope.launch { drawerState.open() } }) {
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
    }
}

/**
 * The tab screens side by side, changed by the bar only. Each page keeps its saved state (scroll
 * position) while another shows, and has room at the bottom for the floating bar.
 */
@Composable
private fun TabPager(pagerState: PagerState, navigator: AppNavigator) {
    val savedStates = rememberSaveableStateHolder()
    HorizontalPager(
        pagerState,
        Modifier.fillMaxSize(),
        userScrollEnabled = false,
        key = { page -> Tab.entries[page].name },
    ) { page ->
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
        Tab.CAMS -> CamsScreen("", navigator)
        Tab.LIBRARY -> LibraryScreen(navigator)
    }
}

/** Every destination opened from a tab, the menu or a card. */
private fun NavGraphBuilder.detailScreens(navigator: AppNavigator) {
    composable(Routes.SETTINGS) { SettingsScreen(navigator) }
    composable(Routes.DESIGN_KIT) { DesignKitScreen(navigator) }
    composable(Routes.CATEGORIES) { CategoriesScreen(navigator) }
    composable(Routes.CATEGORY, arguments = listOf(optionalText("name"))) { entry ->
        ScopedVideosScreen(VideoScope.Category(entry.argument("id")), entry.argument("name"), navigator)
    }
    composable(Routes.TAG, arguments = listOf(optionalText("name"))) { entry ->
        ScopedVideosScreen(VideoScope.Tag(entry.argument("id")), entry.argument("name"), navigator)
    }
    composable(Routes.TASTE) { TasteScreen(navigator) }
    composable(Routes.PLUGINS) { PluginsScreen(navigator) }
    composable(Routes.VIDEO) { entry -> VideoScreen(entry.argument("id"), navigator) }
    composable(Routes.CAM) { entry -> CamRoomScreen(entry.argument("id"), navigator) }
    composable(Routes.COLLECTION) { entry -> CollectionScreen(entry.argument("id"), navigator) }
    composable(Routes.STORIES) { StoriesScreen(navigator) }
    composable(Routes.STORY) { entry -> StoryScreen(entry.argument("id"), navigator) }
    composable(Routes.STORY_READER) { entry ->
        StoryReaderScreen(entry.argument("id"), entry.argument("position").toIntOrNull() ?: 0, navigator)
    }
    composable(Routes.AUDIOS) { AudiosScreen(navigator) }
    composable(Routes.AUDIO) { entry -> AudioScreen(entry.argument("id"), navigator) }
    composable(Routes.GALLERIES) { GalleriesScreen(navigator) }
    composable(Routes.GALLERY) { entry -> GalleryScreen(entry.argument("id"), navigator) }
    composable(Routes.PERFORMERS) { MembersScreen(MemberKind.PERFORMERS, navigator) }
    composable(Routes.STUDIOS) { MembersScreen(MemberKind.STUDIOS, navigator) }
    composable(Routes.PERFORMER) { entry -> PersonScreen(MemberKind.PERFORMERS, entry.argument("id"), navigator) }
    composable(Routes.STUDIO) { entry -> PersonScreen(MemberKind.STUDIOS, entry.argument("id"), navigator) }
    composable(Routes.SEARCH, arguments = listOf(optionalText("query"))) { entry -> SearchScreen(entry.argument("query"), navigator) }
    composable(Routes.CAMS_TAGGED) { entry -> CamsScreen(entry.argument("tag"), navigator) }
}

/** A text query argument that may be left out; it's empty then. */
private fun optionalText(name: String) = navArgument(name) {
    type = NavType.StringType
    defaultValue = ""
}

/** A string argument of the destination; empty when missing. */
private fun NavBackStackEntry.argument(name: String): String {
    return arguments?.getString(name).orEmpty()
}
