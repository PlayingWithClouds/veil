package com.playingwithclouds.veil.ui.people

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.SnackbarResult
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PeopleRepository
import com.playingwithclouds.veil.data.PersonDetail
import com.playingwithclouds.veil.data.SearchFilter
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SiteStyles
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.BarSearchField
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.SearchButton
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilMenu
import com.playingwithclouds.veil.ui.design.VeilMenuItem
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.search.FilterChipRow
import com.playingwithclouds.veil.ui.search.FilterChoices
import com.playingwithclouds.veil.ui.search.FilterSheet
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatVideoCount
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Videos per page of the list. */
private const val VIDEO_PAGE = 24

/** How long typing has to pause before their videos are narrowed. */
private const val VIDEO_FILTER_DELAY_MILLISECONDS = 300L

/** Galleries in the row. */
private const val GALLERY_LIMIT = 30

/** Tags the filter sheet offers to hide. */
private const val TAG_COUNT_LIMIT = 30

/** The list asks for more this close to its end. */
private const val LOAD_MORE_DISTANCE = 5

/** Lines of the bio before it's tapped open. */
private const val BIO_LINES = 3

private val PictureSize = 96.dp
private val GalleryWidth = 140.dp

/**
 * A performer's or studio's page: who they are, then their galleries and videos. Opening a
 * performer reads their pages on the sites not read yet, as the web does; pulling down reads every
 * page again, which can bring in new videos.
 */
class PersonViewModel(private val kind: MemberKind, private val personId: String) : ViewModel() {

    private val detailState = MutableStateFlow<LoadState<PersonDetail>>(LoadState.Loading)
    private val galleriesState = MutableStateFlow(emptyList<GallerySummary>())
    private val refreshingState = MutableStateFlow(false)
    private val followIdState = MutableStateFlow<String?>(null)
    private val blockedState = MutableStateFlow(false)
    private val filterState = MutableStateFlow(SearchFilter())
    private val choicesState = MutableStateFlow(FilterChoices())
    private val messageState = MutableStateFlow<String?>(null)

    val detail: StateFlow<LoadState<PersonDetail>> = detailState.asStateFlow()
    val galleries: StateFlow<List<GallerySummary>> = galleriesState.asStateFlow()

    /** Whether a pull-down refresh is reading their site pages. */
    val refreshing: StateFlow<Boolean> = refreshingState.asStateFlow()

    /** The follow on them; null while they aren't followed. */
    val followId: StateFlow<String?> = followIdState.asStateFlow()

    /** Whether they were just blocked (the page offers Undo, then closes). */
    val blocked: StateFlow<Boolean> = blockedState.asStateFlow()

    /** The filter on their videos: length, quality, rating, sites, sort, hidden tags. */
    val filter: StateFlow<SearchFilter> = filterState.asStateFlow()

    /** The sites and tags the filter sheet offers. */
    val filterChoices: StateFlow<FilterChoices> = choicesState.asStateFlow()

    /** A note for the snackbar, e.g. why a refresh failed. */
    val message: StateFlow<String?> = messageState.asStateFlow()

    private val videoTextState = MutableStateFlow("")
    private var videoTextJob: Job? = null

    /** What's typed in the bar's search box. */
    val videoText: StateFlow<String> = videoTextState.asStateFlow()

    /** Their videos, newest first unless the filter sorts them, once the page is loaded; narrowed by [videoText] and [filter]. */
    val videos = PagedList<VideoSummary>(viewModelScope, VIDEO_PAGE, { video -> video.id }) { offset ->
        val person = loadedPerson() ?: return@PagedList emptyList()
        PeopleRepository.videos(person, videoTextState.value, filterState.value, VIDEO_PAGE, offset)
    }

    init {
        load()
    }

    /** Takes the box's text and narrows their videos once typing pauses. */
    fun setVideoText(text: String) {
        videoTextState.value = text
        videoTextJob?.cancel()
        videoTextJob = viewModelScope.launch {
            delay(VIDEO_FILTER_DELAY_MILLISECONDS)
            videos.refresh()
        }
    }

    /** Applies a filter to their videos and recounts the tags it leaves. */
    fun setFilter(filter: SearchFilter) {
        filterState.value = filter
        videos.refresh()
        loadTagCounts()
    }

    /** Reads what the filter sheet offers, the first time it opens. */
    fun loadFilterChoices() {
        if (choicesState.value.siteChoices.isNotEmpty()) {
            return
        }
        viewModelScope.launch {
            val sites = runCatching { SearchRepository.siteChoices() }.getOrDefault(emptyList())
            choicesState.update { current -> current.copy(siteChoices = sites, siteStyles = SiteStyles.of(sites)) }
        }
        viewModelScope.launch {
            val povTag = runCatching { SearchRepository.povTag() }.getOrNull()
            choicesState.update { current -> current.copy(povTag = povTag) }
        }
        loadTagCounts()
    }

    /** Counts the tags of their videos matching the filter, for hiding. */
    private fun loadTagCounts() {
        val person = loadedPerson() ?: return
        viewModelScope.launch {
            val counts = runCatching { PeopleRepository.tagCounts(person, filterState.value, TAG_COUNT_LIMIT) }.getOrDefault(emptyList())
            choicesState.update { current -> current.copy(tagCounts = counts.filter { (tag, _) -> tag.id !in filterState.value.hiddenTags }) }
        }
    }

    /** Loads who they are, then their videos, galleries and whether they're followed. */
    fun load() {
        detailState.value = LoadState.Loading
        viewModelScope.launch {
            val found = runCatching { requireNotNull(fetchPerson()) { "They're gone from the library." } }
            val person = found.getOrElse { error ->
                detailState.value = LoadState.Failed(error.displayMessage())
                return@launch
            }
            showPerson(person)
            followIdState.value = runCatching { PeopleRepository.followId(person) }.getOrNull()
        }
    }

    /** Reads every one of their site pages again (a studio is only read anew), then reloads their videos and galleries. */
    fun refresh() {
        if (refreshingState.value) {
            return
        }
        refreshingState.value = true
        viewModelScope.launch {
            val found = runCatching { requireNotNull(refetchPerson()) { "They're gone from the library." } }
            refreshingState.value = false
            val person = found.getOrElse { error ->
                messageState.value = error.displayMessage()
                return@launch
            }
            showPerson(person)
        }
    }

    /** Reloads them as stored after an edit, without the page going blank. */
    fun reloadAfterEdit() {
        viewModelScope.launch {
            val person = runCatching { fetchPerson() }.getOrNull() ?: return@launch
            detailState.value = LoadState.Loaded(person)
        }
    }

    /** Shows them and reloads what's credited to them. */
    private suspend fun showPerson(person: PersonDetail) {
        detailState.value = LoadState.Loaded(person)
        videos.refresh()
        galleriesState.value = runCatching { PeopleRepository.galleries(person, GALLERY_LIMIT) }.getOrDefault(emptyList())
    }

    /** The studio as stored, or the performer after reading their site pages not read yet. */
    private suspend fun fetchPerson(): PersonDetail? {
        if (kind == MemberKind.STUDIOS) {
            return PeopleRepository.studio(personId)
        }
        return PeopleRepository.readPerformer(personId)
    }

    /** The studio as stored, or the performer after reading every one of their site pages again. */
    private suspend fun refetchPerson(): PersonDetail? {
        if (kind == MemberKind.STUDIOS) {
            return PeopleRepository.studio(personId)
        }
        return PeopleRepository.refreshPerformer(personId)
    }

    /** Flips the favourite mark. */
    fun toggleFavorite() {
        val person = loadedPerson() ?: return
        val favorite = !person.favorite
        detailState.update { LoadState.Loaded(person.copy(favorite = favorite)) }
        viewModelScope.launch {
            val saved = runCatching { PeopleRepository.setFavorite(person, favorite) }
            if (saved.isFailure) {
                detailState.update { LoadState.Loaded(person) }
            }
        }
    }

    /** Follows them, or stops following: followed, their new videos come in on their own. */
    fun toggleFollow() {
        val person = loadedPerson() ?: return
        val current = followIdState.value
        viewModelScope.launch {
            if (current != null) {
                followIdState.value = null
                runCatching { PeopleRepository.unfollow(current) }.onFailure { error ->
                    followIdState.value = current
                    messageState.value = error.displayMessage()
                }
                return@launch
            }
            runCatching { PeopleRepository.follow(person) }
                .onSuccess { id -> followIdState.value = id }
                .onFailure { error -> messageState.value = error.displayMessage() }
        }
    }

    /** Blocks them, or takes the block back (Undo). */
    fun setBlocked(blocked: Boolean) {
        val person = loadedPerson() ?: return
        viewModelScope.launch {
            runCatching { PeopleRepository.setBlocked(person.kind, person.id, blocked) }
                .onSuccess { blockedState.value = blocked }
                .onFailure { error -> messageState.value = error.displayMessage() }
        }
    }

    /** The note was shown. */
    fun messageShown() {
        messageState.value = null
    }

    /** Who the page shows, once loaded. */
    private fun loadedPerson(): PersonDetail? {
        return (detailState.value as? LoadState.Loaded)?.value
    }
}

/**
 * A performer's or studio's page. The magnifier searches their videos in the bar; ⋯ edits (a
 * performer) or blocks them.
 */
@Composable
fun PersonScreen(kind: MemberKind, personId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "${kind.name}-$personId") { PersonViewModel(kind, personId) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val videoText by viewModel.videoText.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    var searching by rememberSaveable { mutableStateOf(false) }
    var editing by rememberSaveable { mutableStateOf(false) }
    val closeSearch = {
        searching = false
        viewModel.setVideoText("")
    }
    BackHandler(enabled = searching, onBack = closeSearch)
    PersonSnackbars(viewModel, snackbarHostState, onBlockKept = navigator::back)

    val person = (detail as? LoadState.Loaded)?.value
    ShowAmbientGlow(person?.imageUrl)
    var searchBar: (@Composable () -> Unit)? = null
    if (searching) {
        searchBar = { BarSearchField(videoText, onTextChange = viewModel::setVideoText, placeholder = "Search their videos", onClose = closeSearch) }
    }
    DetailScreen(
        person?.name.orEmpty(),
        onBack = navigator::back,
        snackbarHostState = snackbarHostState,
        actions = {
            SearchButton(onClick = { searching = true })
            if (person != null) {
                PersonMenu(person, onEdit = { editing = true }, onBlock = { viewModel.setBlocked(true) })
            }
        },
        searchBar = searchBar,
    ) {
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(top = LocalTopBarInset.current)) { loaded ->
            val refreshing by viewModel.refreshing.collectAsStateWithLifecycle()
            PullToRefreshBox(isRefreshing = refreshing, onRefresh = viewModel::refresh, modifier = Modifier.fillMaxSize()) {
                PersonContent(loaded, viewModel, navigator)
            }
        }
    }
    if (editing) {
        PerformerEditSheet(personId, onSaved = viewModel::reloadAfterEdit, onDismiss = { editing = false })
    }
}

/** ⋯ in the bar: Edit (performers only, as on the desktop) and Block. */
@Composable
private fun PersonMenu(person: PersonDetail, onEdit: () -> Unit, onBlock: () -> Unit) {
    var open by remember { mutableStateOf(false) }
    Box {
        IconTap(VeilIcons.More, contentDescription = "More", onClick = { open = true }, tint = VeilColors.content)
        VeilMenu(expanded = open, onDismissRequest = { open = false }) {
            if (person.kind == RecordKind.performer) {
                VeilMenuItem("Edit", icon = VeilIcons.Edit, onClick = {
                    open = false
                    onEdit()
                })
            }
            VeilMenuItem("Block", icon = VeilIcons.Block, onClick = {
                open = false
                onBlock()
            })
        }
    }
}

/** The page's notes, and Undo after a block; a block kept closes the page through [onBlockKept]. */
@Composable
private fun PersonSnackbars(viewModel: PersonViewModel, snackbarHostState: SnackbarHostState, onBlockKept: () -> Unit) {
    val message by viewModel.message.collectAsStateWithLifecycle()
    val blocked by viewModel.blocked.collectAsStateWithLifecycle()
    LaunchedEffect(message) {
        val note = message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    LaunchedEffect(blocked) {
        if (!blocked) {
            return@LaunchedEffect
        }
        val result = snackbarHostState.showSnackbar("Blocked: they and their videos are gone from every list.", actionLabel = "Undo", duration = SnackbarDuration.Short)
        if (result == SnackbarResult.ActionPerformed) {
            viewModel.setBlocked(false)
            return@LaunchedEffect
        }
        onBlockKept()
    }
}

/** The header, the galleries row, the filter chips and the videos, growing as it's scrolled to its end. */
@Composable
private fun PersonContent(person: PersonDetail, viewModel: PersonViewModel, navigator: AppNavigator) {
    val galleries by viewModel.galleries.collectAsStateWithLifecycle()
    val videos by viewModel.videos.state.collectAsStateWithLifecycle()
    val videoText by viewModel.videoText.collectAsStateWithLifecycle()
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val followId by viewModel.followId.collectAsStateWithLifecycle()
    var filtersOpen by remember { mutableStateOf(false) }
    if (filtersOpen) {
        val choices by viewModel.filterChoices.collectAsStateWithLifecycle()
        LaunchedEffect(Unit) { viewModel.loadFilterChoices() }
        FilterSheet(filter, choices, onFilter = viewModel::setFilter, onDismiss = { filtersOpen = false })
    }
    val listState = rememberLazyListState()
    LaunchedEffect(listState) {
        snapshotFlow {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisible >= listState.layoutInfo.totalItemsCount - LOAD_MORE_DISTANCE
        }.collect { nearEnd ->
            if (nearEnd) {
                viewModel.videos.loadMore()
            }
        }
    }
    val narrowed = videoText.isNotBlank() || !filter.isEmpty
    LazyColumn(state = listState, contentPadding = PaddingValues(top = LocalTopBarInset.current, bottom = VeilSpacing.huge)) {
        item("header") {
            PersonHeader(
                person,
                following = followId != null,
                onFavorite = viewModel::toggleFavorite,
                onFollow = viewModel::toggleFollow,
                onTag = navigator::openTag,
            )
        }
        if (galleries.isNotEmpty()) {
            item("galleries") {
                Shelf("Galleries", galleries, key = { gallery -> gallery.id }, itemWidth = GalleryWidth) { gallery, _ ->
                    GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                }
            }
        }
        if (videos.items.isNotEmpty() || narrowed) {
            item("videos-heading") { SectionHeading("Videos", modifier = Modifier.gutterPadding()) }
            item("videos-filters") { FilterChipRow(filter, onOpenFilters = { filtersOpen = true }, onFilter = viewModel::setFilter) }
        }
        if (videos.hasLoaded && !videos.isLoading && videos.items.isEmpty() && narrowed) {
            item("videos-none") { EmptyMessage("None of their videos match.") }
        }
        items(videos.items, key = { video -> video.id }) { video ->
            VideoCard(video, onClick = { navigator.openVideo(video.id) }, edgeToEdge = true)
        }
        if (videos.isLoading) {
            item("loading") { Row(Modifier.fillMaxWidth().padding(VeilSpacing.large), horizontalArrangement = Arrangement.Center) { Spinner() } }
        }
    }
}

/** Picture, name, counts and facts, Favourite and Subscribe, the bio and the tags. */
@Composable
private fun PersonHeader(person: PersonDetail, following: Boolean, onFavorite: () -> Unit, onFollow: () -> Unit, onTag: (TagSummary) -> Unit) {
    var bioOpen by rememberSaveable { mutableStateOf(false) }
    Column(Modifier.gutterPadding().padding(top = VeilSpacing.large), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large), verticalAlignment = Alignment.CenterVertically) {
            Avatar(person.imageUrl, Modifier.size(PictureSize))
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                Text(person.name, style = MaterialTheme.typography.headlineSmall, color = VeilColors.content)
                Text(countsLine(person), style = VeilType.cardMeta, color = VeilColors.contentMuted)
                if (person.aliases.isNotEmpty()) {
                    Text("Also " + person.aliases.joinToString(", "), style = VeilType.cardMeta, color = VeilColors.contentFaint, maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
            }
        }
        if (person.facts.isNotEmpty()) {
            Text(person.facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            var heart = VeilIcons.Heart
            if (person.favorite) {
                heart = VeilIcons.HeartFilled
            }
            Pill("Favourite", onClick = onFavorite, selected = person.favorite, icon = heart)
            var followLabel = "Subscribe"
            var followIcon = VeilIcons.Follow
            if (following) {
                followLabel = "Subscribed"
                followIcon = VeilIcons.FollowingFilled
            }
            Pill(followLabel, onClick = onFollow, selected = following, icon = followIcon)
        }
        val bio = person.details
        if (!bio.isNullOrBlank()) {
            var maxLines = BIO_LINES
            if (bioOpen) {
                maxLines = Int.MAX_VALUE
            }
            Text(
                bio,
                Modifier.pressClickable { bioOpen = !bioOpen },
                style = MaterialTheme.typography.bodySmall,
                color = VeilColors.contentMuted,
                maxLines = maxLines,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (person.tags.isNotEmpty()) {
            FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                for (tag in person.tags) {
                    TagPill(tag.name, onClick = { onTag(tag) })
                }
            }
        }
    }
}

/** "120 videos · 4 galleries", galleries left out when there are none. */
private fun countsLine(person: PersonDetail): String {
    val parts = mutableListOf(formatVideoCount(person.videoCount))
    if (person.galleryCount == 1) {
        parts.add("1 gallery")
    }
    if (person.galleryCount > 1) {
        parts.add("${person.galleryCount} galleries")
    }
    return parts.joinToString(" · ")
}
