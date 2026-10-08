package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.SceneShelf
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.components.SearchButton
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.Tab
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.glassControl
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** How many history entries the hub looks through for its shelves. */
private const val HISTORY_SCAN = 40

/** How many scenes a hub shelf shows at most. */
private const val SHELF_SIZE = 12

/** One entry of the hub grid. */
private data class HubEntry(val label: String, val icon: ImageVector, val open: (AppNavigator) -> Unit)

private val hubEntries = listOf(
    HubEntry("Downloads", VeilIcons.Download) { navigator -> navigator.openLibrarySection(LibrarySection.DOWNLOADED.name) },
    HubEntry("Watchlist", VeilIcons.Bookmark) { navigator -> navigator.openLibrarySection(LibrarySection.WATCHLIST.name) },
    HubEntry("Collections", VeilIcons.Collections) { navigator -> navigator.openCollections() },
    HubEntry("History", VeilIcons.History) { navigator -> navigator.openHistory() },
    HubEntry("Performers", VeilIcons.Performers) { navigator -> navigator.openPerformers() },
    HubEntry("Studios", VeilIcons.Studios) { navigator -> navigator.openStudios() },
    HubEntry("Tags", VeilIcons.Tags) { navigator -> navigator.openTags() },
    HubEntry("Galleries", VeilIcons.Galleries) { navigator -> navigator.openTab(Tab.GALLERIES) },
)

/** The started-but-unfinished and the recently watched scenes shown at the top of the hub. */
class LibraryHubViewModel : ViewModel() {
    private val mutableContinueWatching = MutableStateFlow<List<SceneSummary>>(emptyList())

    /** Started but unfinished scenes, most recent first. */
    val continueWatching: StateFlow<List<SceneSummary>> = mutableContinueWatching

    private val mutableRecent = MutableStateFlow<List<SceneSummary>>(emptyList())

    /** The other recently watched scenes, most recent first. */
    val recent: StateFlow<List<SceneSummary>> = mutableRecent

    /** Reloads both shelves from the watch history. */
    fun reload() {
        viewModelScope.launch {
            try {
                val history = FeedRepository.watchHistory(HISTORY_SCAN, 0)
                val unfinished = history.filter { entry -> !entry.completed && entry.progressSeconds > 0 }
                val unfinishedIds = unfinished.map { entry -> entry.mediaId }.toSet()
                mutableContinueWatching.value = unfinished.mapNotNull { entry -> entry.scene }.take(SHELF_SIZE)
                mutableRecent.value = history
                    .filter { entry -> entry.mediaId !in unfinishedIds }
                    .mapNotNull { entry -> entry.scene }
                    .take(SHELF_SIZE)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                // Keep the previous shelf; the grid below works without it.
            }
        }
    }
}

/**
 * The Library tab: continue watching and recently watched, then a tile for every saved or browsable kind of thing. The
 * large header scrolls away and a slim title bar takes its place.
 */
@Composable
fun LibraryHubScreen(navigator: AppNavigator) {
    val viewModel = viewModel { LibraryHubViewModel() }
    val continueWatching by viewModel.continueWatching.collectAsStateWithLifecycle()
    val recent by viewModel.recent.collectAsStateWithLifecycle()
    val gridState = rememberLazyGridState()
    val headerScrolledAway by remember { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.reload() }

    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            LazyVerticalGrid(
                columns = GridCells.Adaptive(160.dp),
                state = gridState,
                contentPadding = PaddingValues(
                    start = VeilSpacing.gutter,
                    top = VeilSpacing.extraSmall,
                    end = VeilSpacing.gutter,
                    bottom = VeilSpacing.medium + LocalFloatingBarInset.current,
                ),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
            ) {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    LargeHeader("Library") {
                        SearchButton(navigator)
                        SettingsMenuButton(navigator)
                    }
                }
                item(span = { GridItemSpan(maxLineSpan) }) {
                    SceneShelf(
                        title = "Continue watching",
                        scenes = continueWatching,
                        onOpen = { scene -> navigator.openScene(scene.id) },
                        onSeeAll = navigator::openHistory,
                    )
                }
                item(span = { GridItemSpan(maxLineSpan) }) {
                    SceneShelf(
                        title = "Recently watched",
                        scenes = recent,
                        onOpen = { scene -> navigator.openScene(scene.id) },
                        onSeeAll = navigator::openHistory,
                        modifier = Modifier.padding(bottom = VeilSpacing.small),
                    )
                }
                item(span = { GridItemSpan(maxLineSpan) }) { SectionHeading("Browse") }
                items(hubEntries, key = { entry -> entry.label }) { entry ->
                    HubTile(entry.label, entry.icon, onClick = { entry.open(navigator) })
                }
            }
            PinnedTitleBar("Library", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
        }
    }
}

/** A rounded tile with a tinted icon badge and a label. */
@Composable
private fun HubTile(label: String, icon: ImageVector, onClick: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(64.dp)
            .pressClickable(onClick)
            .glassControl(VeilShapes.panel)
            .padding(horizontal = VeilSpacing.medium),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier.size(38.dp).clip(VeilShapes.capsule).background(VeilColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = VeilColors.accent, modifier = Modifier.size(20.dp))
        }
        Text(label, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Medium)
    }
}
