package com.playingwithclouds.veil.ui.collections

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CollectionRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.CollectionCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.PillRow
import com.playingwithclouds.veil.ui.design.PinnedTitleBar
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** Which collections the index lists. */
enum class CollectionOrigin(val label: String, val backendValue: String?) {
    ALL("All", null),
    MINE("Mine", "user"),
    SCRAPED("From sites", "scraped"),
}

/** The collection index with an origin filter. */
class CollectionsViewModel : ViewModel() {

    private val mutableOrigin = MutableStateFlow(CollectionOrigin.ALL)

    /** The selected origin filter. */
    val origin: StateFlow<CollectionOrigin> = mutableOrigin

    /** Collections of the selected origin. */
    val collections = PagedList(viewModelScope, PAGE_SIZE, { collection -> collection.id }) { offset ->
        CollectionRepository.collections(mutableOrigin.value.backendValue, PAGE_SIZE, offset)
    }

    init {
        collections.loadMore()
    }

    /** Switches the origin filter and reloads. */
    fun selectOrigin(origin: CollectionOrigin) {
        mutableOrigin.value = origin
        collections.refresh()
    }

    /** Creates a collection and hands its id on. */
    fun create(name: String, onCreated: (String) -> Unit) {
        viewModelScope.launch {
            try {
                val created = CollectionRepository.create(name)
                collections.refresh()
                onCreated(created.id)
            } catch (error: Exception) {
                // The index simply stays as it was; the dialog already closed.
            }
        }
    }

    companion object {
        private const val PAGE_SIZE = 40
    }
}

/**
 * The Collections tab: playlists of the user and series from the sites. The large header scrolls
 * away and a slim title bar takes its place.
 */
@Composable
fun CollectionsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { CollectionsViewModel() }
    val origin by viewModel.origin.collectAsStateWithLifecycle()
    var creating by remember { mutableStateOf(false) }
    val gridState = rememberLazyGridState()
    val headerScrolledAway by remember { derivedStateOf { gridState.firstVisibleItemIndex > 0 } }

    if (creating) {
        TextInputDialog(
            title = "New collection",
            label = "Name",
            confirmLabel = "Create",
            onConfirm = { name ->
                creating = false
                viewModel.create(name, navigator::openCollection)
            },
            onDismiss = { creating = false },
        )
    }

    Scaffold { padding ->
        Box(Modifier.padding(padding)) {
            PagedGrid(
                paged = viewModel.collections,
                keyOf = { collection -> collection.id },
                emptyText = "No collections yet.",
                cellWidth = 220.dp,
                gridState = gridState,
                header = {
                    fullWidthItem("header") {
                        LargeHeader("Collections") {
                            RoundIconButton(VeilIcons.Plus, contentDescription = "New collection", onClick = { creating = true })
                        }
                    }
                    fullWidthItem("origins") {
                        PillRow(
                            CollectionOrigin.entries,
                            labelOf = { option -> option.label },
                            onClick = { option -> viewModel.selectOrigin(option) },
                            modifier = Modifier.bleed().padding(vertical = VeilSpacing.extraSmall),
                            isSelected = { option -> option == origin },
                        )
                    }
                },
            ) { _, collection ->
                CollectionCard(collection, onClick = { navigator.openCollection(collection.id) })
            }
            PinnedTitleBar("Collections", visible = headerScrolledAway, modifier = Modifier.align(Alignment.TopCenter))
        }
    }
}
