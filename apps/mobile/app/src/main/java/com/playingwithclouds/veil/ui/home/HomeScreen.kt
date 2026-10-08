package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.LazyGridLayoutInfo
import androidx.compose.foundation.lazy.grid.LazyGridScope
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SceneQuickActionsSheet
import com.playingwithclouds.veil.ui.components.SceneShelf
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SegmentedControl
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.collectLatest

/** Header rows before the feed in the grid; impressions and previews count feed positions only. */
private const val HEADER_ROWS = 5

/** How long the feed has to rest before the card in focus starts its preview. */
private const val PREVIEW_SETTLE_MILLISECONDS = 250L

/**
 * The Home tab: a videos/galleries switch and site badges, then either the continue-watching shelf
 * and the personalized feed (with impression logging and a preview playing on the card in focus),
 * or the newest galleries. Pull down to re-rank. The large header scrolls away and a slim title bar
 * takes its place; a long press on a scene opens its quick actions.
 */
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val mode by viewModel.mode.collectAsStateWithLifecycle()
    val videoGrid = rememberLazyGridState()
    val galleryGrid = rememberLazyGridState()
    var gridState = videoGrid
    if (mode == HomeMode.GALLERIES) {
        gridState = galleryGrid
    }
    val headerScrolledAway by remember(gridState) { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }
    var quickActionsFor by remember { mutableStateOf<SceneSummary?>(null) }

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.refreshProgress() }

    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            if (mode == HomeMode.GALLERIES) {
                GalleryFeed(viewModel, galleryGrid, navigator)
            } else {
                VideoFeed(viewModel, videoGrid, navigator, onLongPress = { scene -> quickActionsFor = scene })
            }
            PinnedTitleBar("Home", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
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

/** The continue-watching shelf and the recommendation feed, logging impressions and previewing the card in focus. */
@Composable
private fun VideoFeed(
    viewModel: HomeViewModel,
    gridState: LazyGridState,
    navigator: AppNavigator,
    onLongPress: (SceneSummary) -> Unit,
) {
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val continueWatching by viewModel.continueWatching.collectAsStateWithLifecycle()
    var previewSceneId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(gridState) {
        snapshotFlow { gridState.layoutInfo }.collect { layoutInfo ->
            for (index in feedIndicesShown(layoutInfo)) {
                val item = state.items.getOrNull(index) ?: continue
                viewModel.recordShown(item, index)
            }
        }
    }
    LaunchedEffect(gridState) {
        snapshotFlow { feedIndexInFocus(gridState.layoutInfo) }.collectLatest { index ->
            delay(PREVIEW_SETTLE_MILLISECONDS)
            previewSceneId = index?.let { position -> state.items.getOrNull(position)?.scene?.id }
        }
    }

    PagedGrid(
        paged = viewModel.feed,
        keyOf = { item -> item.scene.id },
        emptyText = "Nothing to recommend yet. Search for something to get started.",
        gridState = gridState,
        header = {
            homeHeader(viewModel, navigator)
            fullWidthItem("continue") {
                SceneShelf(
                    title = "Continue watching",
                    scenes = continueWatching,
                    onOpen = { scene -> navigator.openScene(scene.id) },
                    onSeeAll = navigator::openHistory,
                    onLongPress = onLongPress,
                )
            }
            fullWidthItem("for-you") {
                if (continueWatching.isNotEmpty()) {
                    SectionHeading("For you")
                }
            }
        },
    ) { position, item ->
        SceneCard(
            scene = item.scene,
            onClick = {
                viewModel.recordClicked(item, position)
                navigator.openScene(item.scene.id)
            },
            onLongClick = { onLongPress(item.scene) },
            previewing = item.scene.id == previewSceneId,
        )
    }
}

/** The newest galleries of the chosen sites. */
@Composable
private fun GalleryFeed(viewModel: HomeViewModel, gridState: LazyGridState, navigator: AppNavigator) {
    PagedGrid(
        paged = viewModel.galleries,
        keyOf = { gallery -> gallery.id },
        emptyText = "No galleries from these sites yet.",
        cellWidth = PosterCellWidth,
        gridState = gridState,
        header = { homeHeader(viewModel, navigator) },
    ) { _, gallery ->
        GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
    }
}

/** The rows above either list: title, the videos/galleries switch and the site badges. */
private fun LazyGridScope.homeHeader(viewModel: HomeViewModel, navigator: AppNavigator) {
    fullWidthItem("header") {
        LargeHeader("Home") { SettingsMenuButton(navigator) }
    }
    fullWidthItem("mode") {
        val mode by viewModel.mode.collectAsStateWithLifecycle()
        SegmentedControl(
            labels = HomeMode.entries.map { option -> option.label },
            selectedIndex = mode.ordinal,
            onSelect = { index -> viewModel.selectMode(HomeMode.entries[index]) },
            modifier = Modifier.padding(vertical = VeilSpacing.small),
        )
    }
    fullWidthItem("sources") {
        val mode by viewModel.mode.collectAsStateWithLifecycle()
        val sites by viewModel.sites.collectAsStateWithLifecycle()
        val selected by viewModel.selectedSources.collectAsStateWithLifecycle()
        SiteBadges(
            sites = remember(sites, mode) { viewModel.sitesFor(mode) },
            selected = selected,
            onToggle = viewModel::toggleSource,
            onAll = viewModel::clearSources,
        )
    }
}

/** "All" and one badge per site with its icon, edge to edge; several sites may be picked at once. */
@Composable
private fun SiteBadges(sites: List<SearchSite>, selected: Set<String>, onToggle: (String) -> Unit, onAll: () -> Unit) {
    if (sites.size < 2) {
        return
    }
    LazyRow(
        Modifier.bleed().fillMaxWidth().fadingEdges().padding(vertical = VeilSpacing.extraSmall),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item(key = "all") {
            Pill("All", onClick = onAll, selected = selected.isEmpty())
        }
        items(sites, key = { site -> site.name }) { site ->
            Pill(
                site.label,
                onClick = { onToggle(site.name) },
                selected = selected.contains(site.name),
                leading = { SiteIcon(site.iconUrl) },
            )
        }
    }
}

/** A site's small icon inside its badge; nothing when the site has none. */
@Composable
private fun SiteIcon(iconUrl: String?) {
    if (iconUrl == null) {
        return
    }
    RemoteImage(iconUrl, Modifier.size(16.dp).clip(VeilShapes.badge), contentScale = ContentScale.Fit)
}

/** Feed positions (not grid indices) of the cards on screen enough to count as shown. */
private fun feedIndicesShown(layoutInfo: LazyGridLayoutInfo): List<Int> {
    return VisibleImpressions.shownIndices(feedExtents(layoutInfo), layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}

/** Feed position of the card whose preview should play, or null. */
private fun feedIndexInFocus(layoutInfo: LazyGridLayoutInfo): Int? {
    val extents = feedExtents(layoutInfo).filter { extent -> extent.index >= 0 }
    return VisibleImpressions.previewIndex(extents, layoutInfo.viewportStartOffset, layoutInfo.viewportEndOffset)
}

/** The laid-out grid items as feed positions; header rows get negative positions. */
private fun feedExtents(layoutInfo: LazyGridLayoutInfo): List<ItemExtent> {
    return layoutInfo.visibleItemsInfo.map { item -> ItemExtent(item.index - HEADER_ROWS, item.offset.y, item.size.height) }
}
