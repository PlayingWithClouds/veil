package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CollectionEntry
import com.playingwithclouds.veil.data.CollectionPage
import com.playingwithclouds.veil.data.LibraryRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.loadInto
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** One of the user's collections with its items. */
class CollectionViewModel(private val collectionId: String) : ViewModel() {

    private val pageState = MutableStateFlow<LoadState<CollectionPage>>(LoadState.Loading)
    val page: StateFlow<LoadState<CollectionPage>> = pageState.asStateFlow()

    init {
        load()
    }

    /** Loads the collection; a collection that's gone fails. */
    fun load() {
        viewModelScope.loadInto(pageState) { requireNotNull(LibraryRepository.collection(collectionId)) { "This collection is gone." } }
    }

    /** Takes an entry out; it vanishes at once, and comes back with the page when that fails. */
    fun remove(entry: CollectionEntry) {
        pageState.update { current -> withoutEntry(current, entry) }
        viewModelScope.launch {
            val removed = runCatching { LibraryRepository.remove(collectionId, entry) }
            if (removed.isFailure) {
                load()
            }
        }
    }
}

/** A collection's page: its name on top, then its videos and galleries. */
@Composable
fun CollectionScreen(collectionId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "collection-$collectionId") { CollectionViewModel(collectionId) }
    val state by viewModel.page.collectAsStateWithLifecycle()
    val title = (state as? LoadState.Loaded)?.value?.name ?: "Collection"
    DetailScreen(title, onBack = navigator::back) {
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(top = LocalTopBarInset.current)) { page ->
            CollectionEntries(page, "Nothing in this collection yet.", navigator, onRemove = viewModel::remove)
        }
    }
}
