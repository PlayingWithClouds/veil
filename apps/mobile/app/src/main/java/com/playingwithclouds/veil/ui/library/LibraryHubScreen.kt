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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
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
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.ShelfHeading
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** How many recently watched scenes the hub's shelf shows. */
private const val RECENT_COUNT = 12

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
    HubEntry("Galleries", VeilIcons.Galleries) { navigator -> navigator.openGalleries() },
)

/** The recently watched scenes shown at the top of the hub. */
class LibraryHubViewModel : ViewModel() {
    private val mutableRecent = MutableStateFlow<List<SceneSummary>>(emptyList())
    val recent: StateFlow<List<SceneSummary>> = mutableRecent

    /** Reloads the shelf. */
    fun reload() {
        viewModelScope.launch {
            try {
                mutableRecent.value = FeedRepository.watchHistory(RECENT_COUNT, 0).mapNotNull { entry -> entry.scene }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                // Keep the previous shelf; the grid below works without it.
            }
        }
    }
}

/** The Library tab: recently watched, then a tile for every saved or browsable kind of thing. */
@Composable
fun LibraryHubScreen(navigator: AppNavigator) {
    val viewModel = viewModel { LibraryHubViewModel() }
    val recent by viewModel.recent.collectAsStateWithLifecycle()
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { viewModel.reload() }

    Scaffold { padding ->
        LazyVerticalGrid(
            columns = GridCells.Adaptive(160.dp),
            modifier = Modifier.padding(padding),
            contentPadding = PaddingValues(start = 12.dp, top = 4.dp, end = 12.dp, bottom = 12.dp + LocalFloatingBarInset.current),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            item(span = { GridItemSpan(maxLineSpan) }) {
                LargeHeader("Library") { SettingsMenuButton(navigator) }
            }
            item(span = { GridItemSpan(maxLineSpan) }) {
                SceneShelf(
                    title = "Recently watched",
                    scenes = recent,
                    onOpen = { scene -> navigator.openScene(scene.id) },
                    onSeeAll = navigator::openHistory,
                    modifier = Modifier.padding(bottom = 8.dp),
                )
            }
            item(span = { GridItemSpan(maxLineSpan) }) { ShelfHeading("Browse") }
            items(hubEntries, key = { entry -> entry.label }) { entry ->
                HubTile(entry.label, entry.icon, onClick = { entry.open(navigator) })
            }
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
            .clip(RoundedCornerShape(20.dp))
            .background(VeilColors.surface)
            .pressClickable(onClick)
            .padding(horizontal = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            Modifier.size(38.dp).clip(CircleShape).background(VeilColors.accent.copy(alpha = 0.16f)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = VeilColors.accent, modifier = Modifier.size(20.dp))
        }
        Text(label, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Medium)
    }
}
