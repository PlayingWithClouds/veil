package com.playingwithclouds.veil.ui.people

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.activity.compose.BackHandler
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.MemberSummary
import com.playingwithclouds.veil.data.PeopleRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.AvatarCellWidth
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.BarSearchField
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.SearchButton
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatVideoCount
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** Members per page of the grid. */
private const val PAGE_SIZE = 60

/** How long typing has to pause before the list narrows. */
private const val FILTER_DELAY_MILLISECONDS = 300L

/** Which list a [MembersScreen] shows. */
enum class MemberKind(val title: String) {
    PERFORMERS("Performers"),
    STUDIOS("Studios"),
}

/** Performers or studios by name, narrowed by the text typed. */
class MembersViewModel(private val kind: MemberKind) : ViewModel() {

    private val textState = MutableStateFlow("")
    private var filterJob: Job? = null

    /** The filter box's text. */
    val text: StateFlow<String> = textState.asStateFlow()

    val members = PagedList<MemberSummary>(viewModelScope, PAGE_SIZE, { member -> member.id }) { offset ->
        val text = textState.value.trim()
        if (kind == MemberKind.STUDIOS) {
            PeopleRepository.studios(text, PAGE_SIZE, offset)
        } else {
            PeopleRepository.performers(text, PAGE_SIZE, offset)
        }
    }

    init {
        members.loadMore()
    }

    /** Takes the box's text and narrows the list once typing pauses. */
    fun onTextChange(text: String) {
        textState.value = text
        filterJob?.cancel()
        filterJob = viewModelScope.launch {
            delay(FILTER_DELAY_MILLISECONDS)
            members.refresh()
        }
    }
}

/** The library's performers or studios as round pictures with names; the magnifier filters them by name. */
@Composable
fun MembersScreen(kind: MemberKind, navigator: AppNavigator) {
    val viewModel = viewModel(key = "members-${kind.name}") { MembersViewModel(kind) }
    val text by viewModel.text.collectAsStateWithLifecycle()
    var searching by rememberSaveable { mutableStateOf(false) }
    val closeSearch = {
        searching = false
        viewModel.onTextChange("")
    }
    BackHandler(enabled = searching, onBack = closeSearch)
    var searchBar: (@Composable () -> Unit)? = null
    if (searching) {
        searchBar = { BarSearchField(text, onTextChange = viewModel::onTextChange, placeholder = "Filter by name", onClose = closeSearch) }
    }
    DetailScreen(kind.title, onBack = navigator::back, actions = { SearchButton(onClick = { searching = true }) }, searchBar = searchBar) {
        PagedGrid(
            paged = viewModel.members,
            keyOf = { member -> member.id },
            emptyText = "Nobody here yet. They come in with the videos you open.",
            cellWidth = AvatarCellWidth,
        ) { _, member ->
            MemberTile(member, onClick = {
                if (kind == MemberKind.STUDIOS) {
                    navigator.openStudio(member.id)
                } else {
                    navigator.openPerformer(member.id)
                }
            })
        }
    }
}

/** A performer or studio in the grid: round picture, name, how many videos. */
@Composable
private fun MemberTile(member: MemberSummary, onClick: () -> Unit) {
    Column(
        Modifier.fillMaxWidth().pressClickable(onClick).padding(VeilSpacing.small),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
    ) {
        Avatar(member.imageUrl, Modifier.fillMaxWidth().aspectRatio(1f))
        Text(member.name, style = VeilType.cardTitle, color = VeilColors.content, textAlign = TextAlign.Center, maxLines = 2, overflow = TextOverflow.Ellipsis)
        Text(formatVideoCount(member.videoCount), style = VeilType.cardMeta, color = VeilColors.contentMuted)
    }
}
