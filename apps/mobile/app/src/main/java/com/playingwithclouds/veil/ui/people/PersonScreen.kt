package com.playingwithclouds.veil.ui.people

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
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
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatVideoCount
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Videos per page of the list. */
private const val VIDEO_PAGE = 24

/** Galleries in the row. */
private const val GALLERY_LIMIT = 30

/** The list asks for more this close to its end. */
private const val LOAD_MORE_DISTANCE = 5

/** Lines of the bio before it's tapped open. */
private const val BIO_LINES = 3

private val PictureSize = 96.dp
private val GalleryWidth = 140.dp

/**
 * A performer's or studio's page: who they are, then their galleries and videos. Opening a
 * performer reads their pages on the sites not read yet, as the web does.
 */
class PersonViewModel(private val kind: MemberKind, private val personId: String) : ViewModel() {

    private val detailState = MutableStateFlow<LoadState<PersonDetail>>(LoadState.Loading)
    private val galleriesState = MutableStateFlow(emptyList<GallerySummary>())

    val detail: StateFlow<LoadState<PersonDetail>> = detailState.asStateFlow()
    val galleries: StateFlow<List<GallerySummary>> = galleriesState.asStateFlow()

    /** Their videos, newest first, once the page is loaded. */
    val videos = PagedList<VideoSummary>(viewModelScope, VIDEO_PAGE, { video -> video.id }) { offset ->
        val person = (detailState.value as? LoadState.Loaded)?.value ?: return@PagedList emptyList()
        PeopleRepository.videos(person, VIDEO_PAGE, offset)
    }

    init {
        load()
    }

    /** Loads who they are, then their videos and galleries. */
    fun load() {
        detailState.value = LoadState.Loading
        viewModelScope.launch {
            val found = runCatching { requireNotNull(fetchPerson()) { "They're gone from the library." } }
            val person = found.getOrElse { error ->
                detailState.value = LoadState.Failed(error.displayMessage())
                return@launch
            }
            detailState.value = LoadState.Loaded(person)
            videos.refresh()
            galleriesState.value = runCatching { PeopleRepository.galleries(person, GALLERY_LIMIT) }.getOrDefault(emptyList())
        }
    }

    /** The studio as stored, or the performer after reading their site pages. */
    private suspend fun fetchPerson(): PersonDetail? {
        if (kind == MemberKind.STUDIOS) {
            return PeopleRepository.studio(personId)
        }
        return PeopleRepository.readPerformer(personId)
    }

    /** Flips the favourite mark. */
    fun toggleFavorite() {
        val person = (detailState.value as? LoadState.Loaded)?.value ?: return
        val favorite = !person.favorite
        detailState.update { LoadState.Loaded(person.copy(favorite = favorite)) }
        viewModelScope.launch {
            val saved = runCatching { PeopleRepository.setFavorite(person, favorite) }
            if (saved.isFailure) {
                detailState.update { LoadState.Loaded(person) }
            }
        }
    }
}

/** A performer's or studio's page. */
@Composable
fun PersonScreen(kind: MemberKind, personId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "${kind.name}-$personId") { PersonViewModel(kind, personId) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val title = (detail as? LoadState.Loaded)?.value?.name.orEmpty()
    ShowAmbientGlow((detail as? LoadState.Loaded)?.value?.imageUrl)
    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }) { padding ->
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { person ->
            PersonContent(person, viewModel, navigator)
        }
    }
}

/** The header, the galleries row and the videos, growing as it's scrolled to its end. */
@Composable
private fun PersonContent(person: PersonDetail, viewModel: PersonViewModel, navigator: AppNavigator) {
    val galleries by viewModel.galleries.collectAsStateWithLifecycle()
    val videos by viewModel.videos.state.collectAsStateWithLifecycle()
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
    LazyColumn(state = listState, contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
        item("header") { PersonHeader(person, onFavorite = viewModel::toggleFavorite, onTag = navigator::openTag) }
        if (galleries.isNotEmpty()) {
            item("galleries") {
                Shelf("Galleries", galleries, key = { gallery -> gallery.id }, itemWidth = GalleryWidth) { gallery, _ ->
                    GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
                }
            }
        }
        if (videos.items.isNotEmpty()) {
            item("videos-heading") { SectionHeading("Videos", modifier = Modifier.gutterPadding()) }
        }
        items(videos.items, key = { video -> video.id }) { video ->
            VideoCard(video, onClick = { navigator.openVideo(video.id) }, edgeToEdge = true)
        }
        if (videos.isLoading) {
            item("loading") { Row(Modifier.fillMaxWidth().padding(VeilSpacing.large), horizontalArrangement = Arrangement.Center) { Spinner() } }
        }
    }
}

/** Picture, name, counts and facts, Favourite, the bio and the tags. */
@Composable
private fun PersonHeader(person: PersonDetail, onFavorite: () -> Unit, onTag: (TagSummary) -> Unit) {
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
        Row {
            var icon = VeilIcons.Heart
            if (person.favorite) {
                icon = VeilIcons.HeartFilled
            }
            Pill("Favourite", onClick = onFavorite, selected = person.favorite, icon = icon)
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
