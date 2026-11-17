package com.playingwithclouds.veil.ui.search

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.PerformerBubble
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.TrackShown
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType

/** The next page is asked for this close to the end of the results. */
private const val LOAD_MORE_DISTANCE = 4

/** Width of a performer in the results' performer row. */
private val PerformerWidth = 88.dp

/** The list key of a result card. */
private fun resultKey(videoId: String): String = "result:$videoId"

/**
 * Search: the box on top (focused when opened empty) with the filter chips under it, the recent
 * searches while nothing is asked, else the performers named so and the videos the sites found,
 * filling in as each site answers. Changing the filter searches again with it.
 */
@Composable
fun SearchScreen(initialQuery: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "search-$initialQuery") { SearchViewModel(initialQuery) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val focusRequester = remember { FocusRequester() }
    val focusManager = LocalFocusManager.current
    var filtersOpen by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        if (initialQuery.isEmpty()) {
            focusRequester.requestFocus()
        }
    }
    Column(Modifier.fillMaxSize().statusBarsPadding()) {
        Row(Modifier.padding(start = VeilSpacing.extraSmall), verticalAlignment = Alignment.CenterVertically) {
            IconTap(VeilIcons.Back, contentDescription = "Back", onClick = navigator::back, tint = VeilColors.content)
            SearchField(
                value = state.text,
                onValueChange = viewModel::onTextChange,
                placeholder = "Search the sites",
                onSearch = {
                    focusManager.clearFocus()
                    viewModel.submit()
                },
                focusRequester = focusRequester,
                modifier = Modifier.weight(1f),
            )
        }
        var chipRecent = state.recent
        if (state.query == null) {
            chipRecent = emptyList()
        }
        FilterChipRow(state.filter, onOpenFilters = { filtersOpen = true }, onFilter = viewModel::setFilter, recent = chipRecent, onRecent = { text ->
            focusManager.clearFocus()
            viewModel.searchFor(text)
        })
        if (filtersOpen) {
            val choices = FilterChoices(state.siteChoices, state.siteStyles, state.tagCounts)
            FilterSheet(state.filter, choices, onFilter = viewModel::setFilter, onDismiss = { filtersOpen = false })
        }
        if (state.query == null) {
            RecentSearches(state.recent, onSearch = { text ->
                focusManager.clearFocus()
                viewModel.searchFor(text)
            })
        } else {
            Results(state, viewModel, navigator)
        }
    }
}

/** What the user searched lately; tapping one searches it again. */
@Composable
private fun RecentSearches(recent: List<String>, onSearch: (String) -> Unit) {
    LazyColumn(contentPadding = PaddingValues(vertical = VeilSpacing.small)) {
        items(recent, key = { text -> text }) { text ->
            Row(
                Modifier.fillMaxWidth().pressClickable { onSearch(text) }.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.medium),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(VeilIcons.History, contentDescription = null, tint = VeilColors.contentMuted, modifier = Modifier.size(RecentIconSize))
                Text(text, style = MaterialTheme.typography.bodyLarge, color = VeilColors.content, maxLines = 1)
            }
        }
    }
}

private val RecentIconSize = 20.dp

/** The results: progress while sites answer, the performers named so, then the videos. */
@Composable
private fun Results(state: SearchState, viewModel: SearchViewModel, navigator: AppNavigator) {
    val listState = rememberLazyListState()
    LaunchedEffect(listState) {
        snapshotFlow {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisible >= listState.layoutInfo.totalItemsCount - LOAD_MORE_DISTANCE
        }.collect { nearEnd ->
            if (nearEnd) {
                viewModel.loadMore()
            }
        }
    }
    val videos = state.videos
    TrackShown(listState, videos.map { video -> resultKey(video.id) }) { position ->
        AppServices.impressions.recordShown(videos[position].id, RecordKind.video, ImpressionSurface.SEARCH, position)
    }
    if (!state.searching && videos.isEmpty() && state.performers.isEmpty()) {
        EmptyMessage(state.error ?: "No site found anything for “${state.query}”. Try fewer words, or another spelling.")
        return
    }
    LazyColumn(state = listState, contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
        if (state.searching) {
            item("progress") { SearchProgress(state) }
        }
        if (state.performers.isNotEmpty()) {
            item("performers") {
                Shelf("Performers", state.performers, key = { performer -> performer.id }, itemWidth = PerformerWidth) { performer, _ ->
                    PerformerBubble(performer, onClick = { navigator.openPerformer(performer.id) })
                }
            }
        }
        if (videos.isNotEmpty()) {
            item("videos-heading") { SectionHeading("Videos", modifier = Modifier.gutterPadding()) }
        }
        itemsIndexed(videos, key = { _, video -> resultKey(video.id) }) { position, video ->
            VideoCard(
                video,
                onClick = {
                    AppServices.impressions.recordClicked(video.id, RecordKind.video, ImpressionSurface.SEARCH, position)
                    navigator.openVideo(video.id)
                },
                edgeToEdge = true,
            )
        }
    }
}

/** "Asking 12 sites · 5 answered" over a sweeping bar, while a page is being asked. */
@Composable
private fun SearchProgress(state: SearchState) {
    Column(Modifier.gutterPadding().padding(vertical = VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
        var text = "Asking the sites"
        if (state.sitesAsked > 0) {
            text = "Asking ${state.sitesAsked} sites · ${state.sitesAnswered} answered"
        }
        Text(text, style = VeilType.cardMeta, color = VeilColors.contentMuted)
        LoadingBar()
    }
}
