package com.playingwithclouds.veil.ui.galleries

import androidx.activity.compose.BackHandler
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.foundation.layout.Box
import com.playingwithclouds.veil.data.SearchFilter
import com.playingwithclouds.veil.ui.search.FilterChipRow
import com.playingwithclouds.veil.ui.search.FilterChoices
import com.playingwithclouds.veil.ui.search.FilterSheet
import kotlinx.coroutines.flow.update
import com.playingwithclouds.veil.ui.components.SearchButton
import com.playingwithclouds.veil.ui.components.BarSearchField
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** Galleries per page of the feed. */
private const val FEED_PAGE = 24

/** How long typing has to pause before the sites are searched. */
private const val QUERY_DEBOUNCE_MILLISECONDS = 500L

/** Tags offered for hiding in the search's filter sheet. */
private const val TAG_COUNT_LIMIT = 24

/** Galleries in each row above the feed. */
private const val ROW_LIMIT = 20

/** Width of a gallery in the rows above the feed. */
private val RowCardWidth = 150.dp

/** Where the feed is: recommended galleries first, then the rest of the library, then the gallery sites. */
private enum class FeedStage { RANKED, REST, SITES }

/** The galleries overview: rows of recently viewed and favourites, and the endless feed. */
class GalleriesViewModel : ViewModel() {

    private val recentState = MutableStateFlow(emptyList<GallerySummary>())
    private val favoritesState = MutableStateFlow(emptyList<GallerySummary>())
    private var stage = FeedStage.RANKED
    private var stageOffset = 0

    val recent: StateFlow<List<GallerySummary>> = recentState.asStateFlow()
    val favorites: StateFlow<List<GallerySummary>> = favoritesState.asStateFlow()

    /** Gallery sites still listing, taking turns as the web's `siteSource` does; null until first asked. */
    private var sites: MutableList<String>? = null
    private val siteOffsets = mutableMapOf<String, Int>()
    private var siteTurn = 0

    /** Recommended galleries, then the rest of the library newest first, then new ones from the gallery sites. */
    val feed = PagedList<GallerySummary>(
        viewModelScope,
        FEED_PAGE,
        { gallery -> gallery.id },
        isLastPage = { page -> page.isEmpty() },
    ) { offset ->
        if (offset == 0) {
            stage = FeedStage.RANKED
            stageOffset = 0
            sites = null
            siteOffsets.clear()
        }
        nextFeedPage()
    }

    private val queryState = MutableStateFlow("")
    private var queryJob: Job? = null

    /** What's typed in the search box. */
    val query: StateFlow<String> = queryState.asStateFlow()

    /** What the gallery sites find for the query, page by page; nothing while it's blank. */
    val results = PagedList<GallerySummary>(
        viewModelScope,
        FEED_PAGE,
        { gallery -> gallery.id },
        isLastPage = { page -> page.isEmpty() },
    ) { offset ->
        val text = queryState.value.trim()
        if (text.isEmpty()) {
            emptyList()
        } else {
            GalleryRepository.searchPage(text, offset / FEED_PAGE, filterState.value)
        }
    }

    private val filterState = MutableStateFlow(SearchFilter())
    private val choicesState = MutableStateFlow(FilterChoices())

    /** The search's filter: sites, rating, not viewed, favourites, hidden tags. */
    val filter: StateFlow<SearchFilter> = filterState.asStateFlow()

    /** The gallery sites and the tags the filter sheet offers. */
    val filterChoices: StateFlow<FilterChoices> = choicesState.asStateFlow()

    /** Takes the box's text and searches the sites once typing pauses. */
    fun setQuery(text: String) {
        queryState.value = text
        queryJob?.cancel()
        queryJob = viewModelScope.launch {
            delay(QUERY_DEBOUNCE_MILLISECONDS)
            results.refresh()
        }
    }

    /** Searches again under a changed filter, and reads the tags offered for hiding under it. */
    fun setFilter(filter: SearchFilter) {
        filterState.value = filter
        results.refresh()
        loadTagCounts()
    }

    /** Reads what the filter sheet offers, the first time search opens. */
    fun loadFilterChoices() {
        if (choicesState.value.siteChoices.isNotEmpty()) {
            return
        }
        viewModelScope.launch {
            val sites = runCatching { GalleryRepository.siteChoices() }.getOrDefault(emptyList())
            choicesState.update { current -> current.copy(siteChoices = sites) }
        }
        loadTagCounts()
    }

    /** Reads the tags most common among the galleries under the filter. */
    private fun loadTagCounts() {
        viewModelScope.launch {
            val counts = runCatching { GalleryRepository.tagCounts(filterState.value, TAG_COUNT_LIMIT) }.getOrDefault(emptyList())
            choicesState.update { current -> current.copy(tagCounts = counts) }
        }
    }

    init {
        feed.loadMore()
        viewModelScope.launch { recentState.value = runCatching { GalleryRepository.recentlyViewed(ROW_LIMIT) }.getOrDefault(emptyList()) }
        viewModelScope.launch { favoritesState.value = runCatching { GalleryRepository.favorites(ROW_LIMIT) }.getOrDefault(emptyList()) }
    }

    /**
     * The next page of the feed; when the recommendations run short, the library's rest fills it up,
     * and when that runs out the gallery sites do. Empty only once every site is exhausted.
     */
    private suspend fun nextFeedPage(): List<GallerySummary> {
        if (stage == FeedStage.RANKED) {
            val ranked = GalleryRepository.recommended(FEED_PAGE, stageOffset)
            stageOffset += FEED_PAGE
            if (ranked.size == FEED_PAGE) {
                return ranked.distinctBy { gallery -> gallery.id }
            }
            stage = FeedStage.REST
            stageOffset = 0
            return (ranked + nextRestPage()).distinctBy { gallery -> gallery.id }
        }
        if (stage == FeedStage.REST) {
            return nextRestPage()
        }
        return nextSitePage()
    }

    /** The next page of the library's galleries, newest first; the sites' turn once one comes back short. */
    private suspend fun nextRestPage(): List<GallerySummary> {
        val page = GalleryRepository.listing(FEED_PAGE, stageOffset)
        stageOffset += FEED_PAGE
        if (page.size < FEED_PAGE) {
            stage = FeedStage.SITES
        }
        if (page.isEmpty()) {
            return nextSitePage()
        }
        return page
    }

    /** The next page from the gallery sites in turn; a site leaves the turns when it lists nothing more or fails. */
    private suspend fun nextSitePage(): List<GallerySummary> {
        val remaining = sites ?: GalleryRepository.siteNames().toMutableList().also { names -> sites = names }
        while (remaining.isNotEmpty()) {
            siteTurn %= remaining.size
            val site = remaining[siteTurn]
            val offset = siteOffsets[site] ?: 0
            val page = runCatching { GalleryRepository.browse(site, FEED_PAGE, offset) }.getOrDefault(emptyList())
            if (page.isEmpty()) {
                remaining.removeAt(siteTurn)
                continue
            }
            siteOffsets[site] = offset + page.size
            siteTurn++
            return page
        }
        return emptyList()
    }
}

/**
 * Galleries: recently viewed and favourites as rows, then more for you, endlessly. The magnifier
 * searches the gallery sites instead, as the web's top bar search on its galleries page does.
 */
@Composable
fun GalleriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { GalleriesViewModel() }
    var searching by rememberSaveable { mutableStateOf(false) }
    BackHandler(enabled = searching) {
        searching = false
        viewModel.setQuery("")
    }
    val query by viewModel.query.collectAsStateWithLifecycle()
    var searchBar: (@Composable () -> Unit)? = null
    if (searching) {
        searchBar = {
            BarSearchField(query, onTextChange = viewModel::setQuery, placeholder = "Search the gallery sites", onClose = {
                searching = false
                viewModel.setQuery("")
            })
        }
    }
    DetailScreen("Galleries", onBack = navigator::back, actions = { SearchButton(onClick = { searching = true }) }, searchBar = searchBar) {
        if (searching) {
            GallerySearch(viewModel, navigator)
        } else {
            GalleriesOverview(viewModel, navigator)
        }
    }
}

/**
 * The filter chips (the video search's, without length, quality and sort) over what the gallery
 * sites find for the text in the bar's search box.
 */
@Composable
private fun GallerySearch(viewModel: GalleriesViewModel, navigator: AppNavigator) {
    val query by viewModel.query.collectAsStateWithLifecycle()
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val choices by viewModel.filterChoices.collectAsStateWithLifecycle()
    var filtersOpen by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) { viewModel.loadFilterChoices() }
    if (filtersOpen) {
        FilterSheet(filter, choices, onFilter = viewModel::setFilter, onDismiss = { filtersOpen = false }, forVideos = false)
    }
    var emptyText = "Type to search the gallery sites."
    if (query.isNotBlank()) {
        emptyText = "No gallery site found anything."
    }
    PagedGrid(
        paged = viewModel.results,
        keyOf = { gallery -> gallery.id },
        emptyText = emptyText,
        cellWidth = PosterCellWidth,
        header = {
            fullWidthItem("filters") {
                Box(Modifier.bleed()) {
                    FilterChipRow(filter, onOpenFilters = { filtersOpen = true }, onFilter = viewModel::setFilter)
                }
            }
        },
    ) { _, gallery ->
        GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
    }
}

/** Recently viewed and favourites as rows, then the endless feed. */
@Composable
private fun GalleriesOverview(viewModel: GalleriesViewModel, navigator: AppNavigator) {
    val recent by viewModel.recent.collectAsStateWithLifecycle()
    val favorites by viewModel.favorites.collectAsStateWithLifecycle()
    PagedGrid(
        paged = viewModel.feed,
        keyOf = { gallery -> gallery.id },
        emptyText = "No galleries yet. They come in from the gallery sites.",
        cellWidth = PosterCellWidth,
        header = {
            fullWidthItem("recent") {
                Shelf("Recently viewed", recent, key = { gallery -> gallery.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { gallery, _ ->
                    GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                }
            }
            fullWidthItem("favorites") {
                Shelf("Favourites", favorites, key = { gallery -> gallery.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { gallery, _ ->
                    GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                }
            }
            fullWidthItem("feed-heading") { SectionHeading("More for you") }
        },
    ) { _, gallery ->
        GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
    }
}
