package com.playingwithclouds.veil.ui.cams

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.platform.LocalFocusManager
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CamFilter
import com.playingwithclouds.veil.data.CamModel
import com.playingwithclouds.veil.data.CamOrder
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.data.CamSource
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.GlassTopBarLayout
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.TabTopBar
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Taste-ranked rooms in the top row. */
private const val TOP_PICK_COUNT = 12

/** Tag chips offered over the live grid. */
private const val POPULAR_TAG_COUNT = 12

/** Width of a room in the top picks row. */
private val TopPickWidth = 240.dp

/** Everything around the live grid: the filter, the sites, the user's rooms, picks and running recordings. */
data class CamsState(
    val filter: CamFilter = CamFilter(),
    val sources: List<CamSource> = emptyList(),
    val topPicks: List<CamModel> = emptyList(),
    val yourRooms: List<CamModel> = emptyList(),
    val recordings: List<CamRecording> = emptyList(),
    /** The tags most of the first live rooms carry, for the tag chips. */
    val popularTags: List<String> = emptyList(),
    /** What's typed in the room search box; the filter takes it once typing pauses. */
    val queryText: String = "",
) {

    /** Ids of the rooms being recorded now. */
    val recordingModelIds: Set<String>
        get() = recordings.map { recording -> recording.camModelId }.toSet()
}

/** The live rooms across the cam sites, with the rows the web's cams page has above them. */
class CamsViewModel(initialTag: String?) : ViewModel() {

    private val mutableState = MutableStateFlow(CamsState(filter = CamFilter(tag = initialTag)))

    /** The page state. */
    val state: StateFlow<CamsState> = mutableState

    /** Rooms live now under the filter, in its order. */
    val rooms = PagedList<CamModel>(viewModelScope, PAGE_SIZE, { room -> room.id }) { offset ->
        val page = CamRepository.rooms(mutableState.value.filter, PAGE_SIZE, offset)
        if (offset == 0 && mutableState.value.filter.tag == null) {
            mutableState.update { current -> current.copy(popularTags = popularTags(page)) }
        }
        page
    }

    init {
        rooms.loadMore()
        viewModelScope.launch {
            val sources = runCatching { CamRepository.sources() }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(sources = sources) }
        }
        if (initialTag == null) {
            viewModelScope.launch {
                val picks = runCatching { CamRepository.recommended(TOP_PICK_COUNT) }.getOrDefault(emptyList())
                mutableState.update { current -> current.copy(topPicks = picks) }
            }
        }
    }

    /** Reloads the user's rooms and the running recordings, which change on their own. */
    fun refreshLive() {
        viewModelScope.launch {
            val yourRooms = runCatching { CamRepository.yourRooms() }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(yourRooms = yourRooms) }
        }
        viewModelScope.launch {
            val recordings = runCatching { CamRepository.activeRecordings() }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(recordings = recordings) }
        }
    }

    /** Narrows to a gender, or all with null. */
    fun setGender(gender: String?) {
        applyFilter(mutableState.value.filter.copy(gender = gender))
    }

    /** Orders the rooms. */
    fun setOrder(order: CamOrder) {
        applyFilter(mutableState.value.filter.copy(order = order))
    }

    /** Includes or excludes a site. */
    fun toggleSource(name: String) {
        val sources = mutableState.value.filter.sources.toMutableList()
        if (!sources.remove(name)) {
            sources.add(name)
        }
        applyFilter(mutableState.value.filter.copy(sources = sources))
    }

    /** Narrows to a tag, or drops the tag filter when it's the one set. */
    fun toggleTag(tag: String) {
        var next: String? = tag
        if (mutableState.value.filter.tag == tag) {
            next = null
        }
        applyFilter(mutableState.value.filter.copy(tag = next))
    }

    /** Narrows the live rooms to the cam sites' keyword search for [text] once typing pauses; blank lists all. */
    fun setQuery(text: String) {
        mutableState.update { current -> current.copy(queryText = text) }
        queryJob?.cancel()
        queryJob = viewModelScope.launch {
            delay(QUERY_DEBOUNCE_MILLISECONDS)
            applyFilter(mutableState.value.filter.copy(query = text.trim().ifEmpty { null }))
        }
    }

    /** Stops a running recording; it becomes a video among the room's recordings. */
    fun stopRecording(recording: CamRecording) {
        mutableState.update { current -> current.copy(recordings = current.recordings - recording) }
        viewModelScope.launch {
            runCatching { CamRepository.stopRecording(recording.camModelId) }
            refreshLive()
        }
    }

    /** Lists the rooms again under [filter]. */
    private fun applyFilter(filter: CamFilter) {
        mutableState.update { current -> current.copy(filter = filter) }
        rooms.refresh()
    }

    private var queryJob: Job? = null

    companion object {
        private const val PAGE_SIZE = 48
        private const val QUERY_DEBOUNCE_MILLISECONDS = 350L
    }
}

/** The tags most of the rooms carry, most common first. */
private fun popularTags(rooms: List<CamModel>): List<String> {
    return rooms.flatMap { room -> room.tags }
        .groupingBy { tag -> tag }
        .eachCount()
        .entries
        .sortedByDescending { entry -> entry.value }
        .take(POPULAR_TAG_COUNT)
        .map { entry -> entry.key }
}

/**
 * Cams, after the web's cams page: top picks for you, your rooms, the recordings running now, then
 * the live rooms with gender, order, site and tag filters. As the Cams tab it has the tab bar on
 * top; opened for a [tag] (from a room's tags) it lists only rooms with it, under a back bar.
 */
@Composable
fun CamsScreen(tag: String, navigator: AppNavigator) {
    val initialTag = tag.ifEmpty { null }
    val viewModel = viewModel(key = "cams-$tag") { CamsViewModel(initialTag) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    PollWhileStarted(CAM_POLL_MILLISECONDS) { viewModel.refreshLive() }

    if (initialTag == null) {
        var searching by rememberSaveable { mutableStateOf(false) }
        BackHandler(enabled = searching) {
            searching = false
            viewModel.setQuery("")
        }
        ShowAmbientGlow(state.topPicks.firstOrNull()?.thumbnailUrl)
        GlassTopBarLayout(bar = {
            if (searching) {
                RoomSearchBar(state.queryText, onQueryChange = viewModel::setQuery, onClose = {
                    searching = false
                    viewModel.setQuery("")
                })
            } else {
                TabTopBar(navigator, "Cams", onSearch = { searching = true })
            }
        }) {
            CamsGrid(state, viewModel, navigator, showOverview = !searching, Modifier.fillMaxSize())
        }
        return
    }
    DetailScreen("#$initialTag", onBack = navigator::back) {
        CamsGrid(state, viewModel, navigator, showOverview = false, Modifier.fillMaxSize())
    }
}

/** The Cams tab's bar while searching rooms: close, then the box, focused when it opens. */
@Composable
private fun RoomSearchBar(text: String, onQueryChange: (String) -> Unit, onClose: () -> Unit) {
    val focusRequester = remember { FocusRequester() }
    val focusManager = LocalFocusManager.current
    LaunchedEffect(Unit) { focusRequester.requestFocus() }
    Row(
        Modifier.fillMaxWidth().statusBarsPadding().padding(start = VeilSpacing.extraSmall),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconTap(VeilIcons.Back, contentDescription = "Close search", onClick = onClose, tint = VeilColors.content)
        SearchField(
            value = text,
            onValueChange = onQueryChange,
            placeholder = "Search live rooms",
            onSearch = { focusManager.clearFocus() },
            focusRequester = focusRequester,
            modifier = Modifier.weight(1f),
        )
    }
}

/** The live grid, with the overview rows above it when [showOverview]. */
@Composable
private fun CamsGrid(state: CamsState, viewModel: CamsViewModel, navigator: AppNavigator, showOverview: Boolean, modifier: Modifier) {
    PagedGrid(
        paged = viewModel.rooms,
        keyOf = { room -> room.id },
        emptyText = "No rooms live match. Is a cam plugin turned on?",
        modifier = modifier,
        cellWidth = CamCellWidth,
        header = {
            if (showOverview) {
                fullWidthItem("top-picks") {
                    Shelf("Top picks for you", state.topPicks, key = { room -> room.id }, itemWidth = TopPickWidth, modifier = Modifier.bleed()) { room, _ ->
                        RoomCard(room, recording = room.id in state.recordingModelIds, onClick = { navigator.openCam(room.id) })
                    }
                }
                fullWidthItem("your-rooms") { YourRoomsRow(state.yourRooms, onOpen = { room -> navigator.openCam(room.id) }) }
                fullWidthItem("recording") {
                    RecordingStrip(
                        state.recordings,
                        onOpen = { recording -> navigator.openCam(recording.camModelId) },
                        onStop = viewModel::stopRecording,
                    )
                }
                fullWidthItem("live-heading") { SectionHeading("Live now") }
            }
            fullWidthItem("filters") { CamFilters(state, viewModel) }
        },
    ) { _, room ->
        RoomCard(room, recording = room.id in state.recordingModelIds, onClick = { navigator.openCam(room.id) })
    }
}

/** The filter rows: genders and orders, then the sites (when more than one) and the popular tags. */
@Composable
private fun CamFilters(state: CamsState, viewModel: CamsViewModel) {
    Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        ChipRow {
            items(CAM_GENDERS, key = { gender -> "gender-${gender.first}" }) { (value, label) ->
                FeedChip(label, onClick = { viewModel.setGender(value) }, selected = state.filter.gender == value)
            }
            items(CamOrder.entries, key = { order -> "order-${order.name}" }) { order ->
                FeedChip(order.label, onClick = { viewModel.setOrder(order) }, selected = state.filter.order == order)
            }
        }
        if (state.sources.size > 1) {
            ChipRow {
                items(state.sources, key = { source -> "source-${source.name}" }) { source ->
                    FeedChip(
                        source.displayName,
                        onClick = { viewModel.toggleSource(source.name) },
                        selected = source.name in state.filter.sources,
                        leading = { SourceIcon(source.iconUrl) },
                    )
                }
            }
        }
        val tags = tagChips(state)
        if (tags.isNotEmpty()) {
            ChipRow {
                items(tags, key = { tag -> "tag-$tag" }) { tag ->
                    FeedChip("#$tag", onClick = { viewModel.toggleTag(tag) }, selected = state.filter.tag == tag)
                }
            }
        }
    }
}

/** The popular tags, with the picked one first even when it isn't among them. */
private fun tagChips(state: CamsState): List<String> {
    val picked = state.filter.tag ?: return state.popularTags
    return listOf(picked) + state.popularTags.filter { tag -> tag != picked }
}

/** A row of chips running edge to edge and fading where clipped. */
@Composable
private fun ChipRow(content: LazyListScope.() -> Unit) {
    LazyRow(
        Modifier.bleed().fillMaxWidth().fadingEdges(),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        content = content,
    )
}

/** A site's small icon in its chip. */
@Composable
private fun SourceIcon(iconUrl: String?) {
    if (iconUrl == null) {
        return
    }
    RemoteImage(iconUrl, Modifier.size(SourceIconSize).clip(VeilShapes.badge))
}

private val SourceIconSize = 16.dp
