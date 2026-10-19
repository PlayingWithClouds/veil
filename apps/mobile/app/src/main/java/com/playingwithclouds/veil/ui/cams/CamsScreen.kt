package com.playingwithclouds.veil.ui.cams

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Scaffold
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
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.data.CamSource
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Everything around the room grid: the filter, the sites, followed rooms and running recordings. */
data class CamsState(
    val filter: CamFilter = CamFilter(),
    /** The text in the search box, applied on submit. */
    val queryText: String = "",
    val sources: List<CamSource> = emptyList(),
    val followed: List<CamModel> = emptyList(),
    val recordings: List<CamRecording> = emptyList(),
) {

    /** Ids of the rooms being recorded now. */
    val recordingModelIds: Set<String>
        get() = recordings.map { recording -> recording.camModelId }.toSet()
}

/** The live room listing across the cam sites. */
class CamsViewModel(initialTag: String?) : ViewModel() {

    private val mutableState = MutableStateFlow(CamsState(filter = CamFilter(tag = initialTag)))

    /** The page state. */
    val state: StateFlow<CamsState> = mutableState

    /** Rooms online now, in the sites' popularity order. */
    val rooms = PagedList<CamModel>(viewModelScope, PAGE_SIZE, { room -> room.id }) { offset ->
        CamRepository.rooms(mutableState.value.filter, PAGE_SIZE, offset).rooms
    }

    init {
        rooms.loadMore()
        viewModelScope.launch {
            val sources = runCatching { CamRepository.sources() }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(sources = sources) }
        }
    }

    /** Reloads the followed rooms and the running recordings, which change on their own. */
    fun refreshLive() {
        viewModelScope.launch {
            val followed = runCatching { CamRepository.followed() }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(followed = followed) }
        }
        viewModelScope.launch {
            val recordings = runCatching { CamRepository.recordings(activeOnly = true) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(recordings = recordings) }
        }
    }

    /** Takes the text of the search box. */
    fun onQueryChange(text: String) {
        mutableState.update { current -> current.copy(queryText = text) }
    }

    /** Searches rooms for the box's text; an empty box lists every room. */
    fun submitQuery() {
        val query = mutableState.value.queryText.trim().ifEmpty { null }
        applyFilter(mutableState.value.filter.copy(query = query))
    }

    /** Narrows to a gender, or all with null. */
    fun setGender(gender: String?) {
        applyFilter(mutableState.value.filter.copy(gender = gender))
    }

    /** Includes or excludes a site. */
    fun toggleSource(name: String) {
        val sources = mutableState.value.filter.sources.toMutableList()
        if (!sources.remove(name)) {
            sources.add(name)
        }
        applyFilter(mutableState.value.filter.copy(sources = sources))
    }

    /** Drops the tag filter. */
    fun clearTag() {
        applyFilter(mutableState.value.filter.copy(tag = null))
    }

    /** Stops a running recording; its scene appears in the room's recordings. */
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

    companion object {
        private const val PAGE_SIZE = 48
    }
}

/** Live cams: search, gender and site filters, running recordings, followed rooms, then the rooms. */
@Composable
fun CamsScreen(initialTag: String, navigator: AppNavigator) {
    val tag = initialTag.ifEmpty { null }
    val viewModel = viewModel(key = "cams-$initialTag") { CamsViewModel(tag) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    PollWhileStarted(CAM_POLL_MILLISECONDS) { viewModel.refreshLive() }

    Scaffold(topBar = { VeilTopBar("Cams", onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.rooms,
            keyOf = { room -> room.id },
            emptyText = "No rooms online match. Is a cam plugin turned on?",
            modifier = Modifier.padding(padding),
            cellWidth = CamCellWidth,
            header = {
                fullWidthItem("search") {
                    SearchField(
                        value = state.queryText,
                        onValueChange = viewModel::onQueryChange,
                        placeholder = "Search rooms",
                        onSearch = viewModel::submitQuery,
                        modifier = Modifier.bleed(),
                    )
                }
                fullWidthItem("recording") {
                    RecordingStrip(
                        state.recordings,
                        onOpen = { recording -> navigator.openCam(recording.camModelId) },
                        onStop = viewModel::stopRecording,
                    )
                }
                fullWidthItem("following") { FollowedCamsRow(state.followed, onOpen = { model -> navigator.openCam(model.id) }) }
                fullWidthItem("filters") { CamFilters(state, viewModel) }
            },
        ) { _, room ->
            CamCard(room, recording = room.id in state.recordingModelIds, onClick = { navigator.openCam(room.id) })
        }
    }
}

/** The tag filter (when set), the genders and, with more than one, the sites. */
@Composable
private fun CamFilters(state: CamsState, viewModel: CamsViewModel) {
    LazyRow(
        Modifier.bleed().fillMaxWidth().fadingEdges(),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        val tag = state.filter.tag
        if (tag != null) {
            item { FeedChip("#$tag", onClick = viewModel::clearTag, selected = true, icon = VeilIcons.Close) }
        }
        items(CAM_GENDERS, key = { gender -> "gender-${gender.first}" }) { (value, label) ->
            FeedChip(label, onClick = { viewModel.setGender(value) }, selected = state.filter.gender == value)
        }
        if (state.sources.size > 1) {
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
}

/** A site's small icon in its chip. */
@Composable
private fun SourceIcon(iconUrl: String?) {
    if (iconUrl == null) {
        return
    }
    RemoteImage(iconUrl, Modifier.size(16.dp).clip(VeilShapes.badge))
}
