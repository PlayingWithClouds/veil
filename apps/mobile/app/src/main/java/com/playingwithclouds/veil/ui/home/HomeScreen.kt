package com.playingwithclouds.veil.ui.home

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.LazyGridLayoutInfo
import androidx.compose.foundation.lazy.grid.LazyGridScope
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.FloatingBarInset
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.AvatarShelf
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SceneQuickActionsSheet
import com.playingwithclouds.veil.ui.components.SceneShelf
import com.playingwithclouds.veil.ui.components.SearchButton
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.rememberHideOnScrollState
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

/** How long the feed has to rest before the card in focus starts its preview. */
private const val PREVIEW_SETTLE_MILLISECONDS = 250L

/** Feed cards between two shelves. */
private const val CARDS_PER_SHELF = 5

/** Height of the top bar under the status bar. */
private val TopBarHeight = 48.dp

/**
 * The Home tab: a slim top bar (wordmark, search, menu) that always stays, a chip row under it that
 * hides while scrolling down, then the personalized feed with shelves between runs of cards,
 * impression logging and a preview playing on the card in focus. Pull down to re-rank; swipe a
 * card right for "not interested"; a long press or the ⋮ opens a scene's quick actions.
 */
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val gridState = rememberLazyGridState()
    val chipRowVisibility = rememberHideOnScrollState()
    val snackbarHostState = remember { SnackbarHostState() }
    val scope = rememberCoroutineScope()
    var quickActionsFor by remember { mutableStateOf<SceneSummary?>(null) }

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.refreshProgress() }

    Scaffold(snackbarHost = { VeilSnackbarHost(snackbarHostState, Modifier.padding(bottom = FloatingBarInset)) }) { padding ->
        Column(Modifier.padding(padding).nestedScroll(chipRowVisibility.connection)) {
            HomeTopBar(navigator)
            AnimatedVisibility(chipRowVisibility.visible, enter = expandVertically(), exit = shrinkVertically()) {
                HomeFilterRow(viewModel)
            }
            VideoFeed(
                viewModel,
                gridState,
                navigator,
                onMenu = { scene -> quickActionsFor = scene },
                onNotInterested = { scene ->
                    viewModel.notInterested(scene)
                    scope.launch { snackbarHostState.showSnackbar("Not interested. You will see less like this.") }
                },
            )
        }
    }

    val selected = quickActionsFor
    if (selected != null) {
        SceneQuickActionsSheet(
            scene = selected,
            onDismiss = { quickActionsFor = null },
            onNotInterested = { viewModel.dropFromFeed(selected.id) },
        )
    }
}

/** The compact bar: the wordmark on the left, search and the menu on the right. */
@Composable
private fun HomeTopBar(navigator: AppNavigator) {
    Row(
        Modifier.fillMaxWidth().height(TopBarHeight).padding(start = VeilSpacing.gutter, end = VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text("Veil", Modifier.weight(1f), style = VeilType.wordmark, color = VeilColors.content, maxLines = 1)
        SearchButton(navigator, bare = true)
        SettingsMenuButton(navigator, bare = true)
    }
}

/** The recommendation feed as full-width cards with shelves between them, logging impressions and previewing the card in focus. */
@Composable
private fun VideoFeed(
    viewModel: HomeViewModel,
    gridState: LazyGridState,
    navigator: AppNavigator,
    onMenu: (SceneSummary) -> Unit,
    onNotInterested: (SceneSummary) -> Unit,
) {
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val shelves by viewModel.shelves.collectAsStateWithLifecycle()
    var previewSceneId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(gridState) {
        snapshotFlow { feedExtents(gridState.layoutInfo, state.items) }.collect { extents ->
            for (index in VisibleImpressions.shownIndices(extents, gridState.layoutInfo.viewportStartOffset, gridState.layoutInfo.viewportEndOffset)) {
                val item = state.items.getOrNull(index) ?: continue
                viewModel.recordShown(item, index)
            }
        }
    }
    LaunchedEffect(gridState) {
        snapshotFlow { feedIndexInFocus(gridState.layoutInfo, state.items) }.collectLatest { index ->
            delay(PREVIEW_SETTLE_MILLISECONDS)
            previewSceneId = index?.let { position -> state.items.getOrNull(position)?.scene?.id }
        }
    }

    val filter by viewModel.filter.collectAsStateWithLifecycle()
    var shownShelves = shelves
    if (filter.isActive) {
        shownShelves = emptyList()
    }

    PagedGrid(
        paged = viewModel.feed,
        keyOf = { item -> item.scene.id },
        emptyText = "Nothing to show. Search for something to get started.",
        gridState = gridState,
        contentPadding = PaddingValues(bottom = VeilSpacing.large + LocalFloatingBarInset.current),
        listContent = { items ->
            feedWithShelves(
                items,
                shownShelves,
                feedCard = { position, item ->
                    SwipeToDismissCard(onDismiss = { onNotInterested(item.scene) }) {
                        SceneCard(
                            scene = item.scene,
                            onClick = {
                                viewModel.recordClicked(item, position)
                                navigator.openScene(item.scene.id)
                            },
                            onLongClick = { onMenu(item.scene) },
                            onMenu = { onMenu(item.scene) },
                            previewing = item.scene.id == previewSceneId,
                            edgeToEdge = true,
                        )
                    }
                },
                shelfRow = { shelf -> HomeShelfRow(shelf, navigator, onMenu) },
            )
        },
        itemContent = { _, _ -> },
    )
}

/**
 * Lays the feed out as runs of [CARDS_PER_SHELF] cards, each followed by the next shelf while there
 * are any left. Cards are keyed by scene id so impressions and previews can find them again.
 */
private fun LazyGridScope.feedWithShelves(
    items: List<RecommendedScene>,
    shelves: List<HomeShelf>,
    feedCard: @Composable (position: Int, item: RecommendedScene) -> Unit,
    shelfRow: @Composable (HomeShelf) -> Unit,
) {
    items.chunked(CARDS_PER_SHELF).forEachIndexed { runIndex, run ->
        itemsIndexed(run, key = { _, item -> item.scene.id }) { indexInRun, item ->
            Box(Modifier.animateItem()) { feedCard(runIndex * CARDS_PER_SHELF + indexInRun, item) }
        }
        val shelf = shelves.getOrNull(runIndex)
        if (shelf != null && run.size == CARDS_PER_SHELF) {
            fullWidthItem("shelf-${shelf.key}") { shelfRow(shelf) }
        }
    }
}

/** One shelf: scene rows open the scene, performer and studio rows their page. */
@Composable
private fun HomeShelfRow(shelf: HomeShelf, navigator: AppNavigator, onMenu: (SceneSummary) -> Unit) {
    when (shelf) {
        is HomeShelf.Scenes -> SceneShelf(
            title = shelf.title,
            scenes = shelf.scenes,
            onOpen = { scene -> navigator.openScene(scene.id) },
            onLongPress = onMenu,
            edgeToEdge = true,
        )
        is HomeShelf.Performers -> AvatarShelf("Performers for you", shelf.performers, onOpen = { performer -> navigator.openPerformer(performer.id) })
        is HomeShelf.Studios -> AvatarShelf("Studios for you", shelf.studios, onOpen = { studio -> navigator.openStudio(studio.id) })
    }
}

/** Feed positions of the cards laid out on screen; shelves and the footer have none and are skipped. */
private fun feedExtents(layoutInfo: LazyGridLayoutInfo, items: List<RecommendedScene>): List<ItemExtent> {
    val positionBySceneId = items.withIndex().associate { (position, item) -> item.scene.id to position }
    return layoutInfo.visibleItemsInfo.mapNotNull { cell ->
        val position = positionBySceneId[cell.key]
        if (position == null) {
            null
        } else {
            ItemExtent(position, cell.offset.y, cell.size.height)
        }
    }
}

/** Feed position of the card whose preview should play, or null. */
private fun feedIndexInFocus(layoutInfo: LazyGridLayoutInfo, items: List<RecommendedScene>): Int? {
    return VisibleImpressions.previewIndex(feedExtents(layoutInfo, items), layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}
