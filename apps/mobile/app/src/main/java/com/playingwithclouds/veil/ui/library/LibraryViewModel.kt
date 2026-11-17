package com.playingwithclouds.veil.ui.library

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.CollectionEntry
import com.playingwithclouds.veil.data.CollectionPage
import com.playingwithclouds.veil.data.CollectionTile
import com.playingwithclouds.veil.data.DownloadItem
import com.playingwithclouds.veil.data.LibraryRepository
import com.playingwithclouds.veil.graphql.type.CollectionKind
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.loadInto
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** The Library's tabs, in order, as the web's library page has them, plus Favourites. */
enum class LibraryTab(val label: String) {
    WATCHLIST("Watchlist"),
    FAVORITES("Favourites"),
    QUEUE("Queue"),
    COLLECTIONS("Collections"),
    DOWNLOADS("Downloads"),
}

/**
 * The Library tab: the watchlist, the queue, the user's collections and the downloads. A tab loads
 * the first time it's shown and keeps what it loaded, so switching back is instant.
 */
class LibraryViewModel : ViewModel() {

    private val selectedTab = MutableStateFlow(LibraryTab.WATCHLIST)
    private val watchlistState = MutableStateFlow<LoadState<CollectionPage>>(LoadState.Loading)
    private val queueState = MutableStateFlow<LoadState<CollectionPage>>(LoadState.Loading)
    private val favoritesState = MutableStateFlow<LoadState<CollectionPage>>(LoadState.Loading)
    private val collectionsState = MutableStateFlow<LoadState<List<CollectionTile>>>(LoadState.Loading)
    private val downloadsState = MutableStateFlow<LoadState<List<DownloadItem>>>(LoadState.Loading)
    private val messageState = MutableStateFlow<String?>(null)
    private val loadedTabs = mutableSetOf<LibraryTab>()

    val tab: StateFlow<LibraryTab> = selectedTab.asStateFlow()
    val watchlist: StateFlow<LoadState<CollectionPage>> = watchlistState.asStateFlow()
    val queue: StateFlow<LoadState<CollectionPage>> = queueState.asStateFlow()
    val favorites: StateFlow<LoadState<CollectionPage>> = favoritesState.asStateFlow()
    val collections: StateFlow<LoadState<List<CollectionTile>>> = collectionsState.asStateFlow()
    val downloads: StateFlow<LoadState<List<DownloadItem>>> = downloadsState.asStateFlow()

    /** A note for the snackbar. */
    val message: StateFlow<String?> = messageState.asStateFlow()

    init {
        select(LibraryTab.WATCHLIST)
    }

    /** Shows a tab, loading it the first time. */
    fun select(tab: LibraryTab) {
        selectedTab.value = tab
        if (loadedTabs.add(tab)) {
            load(tab)
        }
    }

    /** Loads the tab shown again, e.g. after coming back to the screen. */
    fun reload() {
        load(selectedTab.value)
    }

    /** The downloads' progress, read again without showing the spinner. */
    fun refreshDownloads() {
        viewModelScope.launch {
            val found = runCatching { LibraryRepository.downloads() }.getOrNull() ?: return@launch
            downloadsState.value = LoadState.Loaded(found)
        }
    }

    /** Takes an entry out of the watchlist, the queue or the favourites; it vanishes at once. */
    fun remove(page: CollectionPage, entry: CollectionEntry) {
        if (page.id == LibraryRepository.FAVORITES_ID) {
            favoritesState.update { current -> withoutEntry(current, entry) }
            write("Couldn't remove the favourite") { LibraryRepository.unfavorite(entry) }
            return
        }
        val state = stateOf(page)
        state.update { current -> withoutEntry(current, entry) }
        write("Couldn't remove it") { LibraryRepository.remove(page.id, entry) }
    }

    /** Creates a collection by name and lists it. */
    fun createCollection(name: String) {
        write("Couldn't create the collection") {
            LibraryRepository.createCollection(name.trim())
            load(LibraryTab.COLLECTIONS)
        }
    }

    /** Stops a download and deletes it with its file; it vanishes at once. */
    fun deleteDownload(item: DownloadItem) {
        downloadsState.update { current ->
            if (current !is LoadState.Loaded) {
                return@update current
            }
            LoadState.Loaded(current.value.filter { download -> download.jobId != item.jobId })
        }
        write("Couldn't delete the download") { LibraryRepository.deleteDownload(item.jobId) }
    }

    /** The snackbar showed the message. */
    fun messageShown() {
        messageState.value = null
    }

    /** Loads one tab's content. */
    private fun load(tab: LibraryTab) {
        when (tab) {
            LibraryTab.WATCHLIST -> viewModelScope.loadInto(watchlistState) { LibraryRepository.builtInCollection(CollectionKind.watchlist) }
            LibraryTab.QUEUE -> viewModelScope.loadInto(queueState) { LibraryRepository.builtInCollection(CollectionKind.queue) }
            LibraryTab.FAVORITES -> viewModelScope.loadInto(favoritesState) { LibraryRepository.favorites() }
            LibraryTab.COLLECTIONS -> viewModelScope.loadInto(collectionsState) { LibraryRepository.collections() }
            LibraryTab.DOWNLOADS -> viewModelScope.loadInto(downloadsState) { LibraryRepository.downloads() }
        }
    }

    /** The state holding a built-in collection: the watchlist's, else the queue's. */
    private fun stateOf(page: CollectionPage): MutableStateFlow<LoadState<CollectionPage>> {
        val watchlist = watchlistState.value
        if (watchlist is LoadState.Loaded && watchlist.value.id == page.id) {
            return watchlistState
        }
        return queueState
    }

    /** Runs a write; on failure the snackbar says [failure]. */
    private fun write(failure: String, action: suspend () -> Unit) {
        viewModelScope.launch {
            try {
                action()
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                messageState.value = "$failure: ${error.displayMessage()}"
            }
        }
    }
}

/** A loaded collection without one entry. */
internal fun withoutEntry(state: LoadState<CollectionPage>, entry: CollectionEntry): LoadState<CollectionPage> {
    if (state !is LoadState.Loaded) {
        return state
    }
    val page = state.value
    return LoadState.Loaded(page.copy(entries = page.entries - entry, itemCount = page.itemCount - 1))
}
