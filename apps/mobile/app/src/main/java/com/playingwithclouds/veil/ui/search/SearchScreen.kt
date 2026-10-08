package com.playingwithclouds.veil.ui.search

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyGridScope
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.SearchSuggestion
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.AvatarCellWidth
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PerformerCard
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SceneCellWidth
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.StudioCard
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.PillRow
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.entity.GalleryRow
import com.playingwithclouds.veil.ui.entity.PerformerRow
import com.playingwithclouds.veil.ui.entity.StudioRow

/** Padding of the result grids: the screen gutter at the sides. */
private val ResultGridPadding = PaddingValues(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.medium)

/** Search: suggestions while typing, then library and live site results. */
@Composable
fun SearchScreen(initialQuery: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "search-$initialQuery") { SearchViewModel(initialQuery) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val focusRequester = remember { FocusRequester() }
    LaunchedEffect(Unit) {
        if (initialQuery.isEmpty()) {
            focusRequester.requestFocus()
        }
    }

    Scaffold(
        topBar = {
            Row(
                Modifier.statusBarsPadding().padding(start = VeilSpacing.gutter, top = VeilSpacing.small, bottom = VeilSpacing.extraSmall),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                RoundIconButton(VeilIcons.Back, contentDescription = "Back", onClick = navigator::back)
                SearchField(
                    value = state.query,
                    onValueChange = viewModel::onQueryChange,
                    placeholder = "Search videos, performers, studios",
                    onSearch = { viewModel.submit(state.query) },
                    focusRequester = focusRequester,
                )
            }
        },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            if (state.isLoading || state.isSearchingSites) {
                LoadingBar(Modifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall))
            }
            if (state.submitted == null) {
                SuggestionList(state.suggestions, viewModel, navigator)
                return@Column
            }
            ResultFilters(state, viewModel)
            Results(state, navigator)
        }
    }
}

/** Recent searches, taste-based picks and completions for the typed text. */
@Composable
private fun SuggestionList(suggestions: List<SearchSuggestion>, viewModel: SearchViewModel, navigator: AppNavigator) {
    LazyColumn(contentPadding = PaddingValues(vertical = VeilSpacing.small)) {
        items(suggestions, key = { suggestion -> "${suggestion.kind}:${suggestion.text}:${suggestion.entityId}" }) { suggestion ->
            SuggestionRow(
                suggestion,
                onChoose = { chooseSuggestion(suggestion, viewModel, navigator) },
                onForget = { viewModel.forget(suggestion) },
            )
        }
    }
}

/** One suggestion: icon or photo, text with its detail, and a forget button on recent searches. */
@Composable
private fun SuggestionRow(suggestion: SearchSuggestion, onChoose: () -> Unit, onForget: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(onChoose).padding(start = VeilSpacing.gutter, end = VeilSpacing.small, top = VeilSpacing.small, bottom = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        SuggestionIcon(suggestion)
        Column(Modifier.weight(1f)) {
            Text(suggestion.text, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            val detail = suggestion.detail
            if (detail != null) {
                Text(detail, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted, maxLines = 1)
            }
        }
        if (suggestion.kind == "RECENT") {
            IconTap(VeilIcons.Close, contentDescription = "Forget", onClick = onForget)
        }
    }
}

/** Opens the entity a suggestion names, or runs it as a search. */
private fun chooseSuggestion(suggestion: SearchSuggestion, viewModel: SearchViewModel, navigator: AppNavigator) {
    val entityId = suggestion.entityId
    if (entityId != null && suggestion.kind == "TAG") {
        navigator.openTag(entityId)
        return
    }
    if (entityId != null && suggestion.kind == "PERFORMER") {
        navigator.openPerformer(entityId)
        return
    }
    if (entityId != null && suggestion.kind == "STUDIO") {
        navigator.openStudio(entityId)
        return
    }
    viewModel.submit(suggestion.text)
}

/** The icon or photo in front of a suggestion. */
@Composable
private fun SuggestionIcon(suggestion: SearchSuggestion) {
    if (suggestion.imageUrl != null) {
        Avatar(suggestion.imageUrl, Modifier.size(40.dp))
        return
    }
    val icon: ImageVector = when (suggestion.kind) {
        "RECENT" -> VeilIcons.History
        "TAG" -> VeilIcons.Tags
        "PERFORMER" -> VeilIcons.Performer
        "STUDIO" -> VeilIcons.Studios
        else -> VeilIcons.Search
    }
    Box(Modifier.size(40.dp).clip(VeilShapes.capsule).background(VeilColors.surfaceHigh), contentAlignment = Alignment.Center) {
        Icon(icon, contentDescription = null, tint = VeilColors.contentMuted, modifier = Modifier.size(20.dp))
    }
}

/** Scope pills, the follow pill and the site filter. */
@Composable
private fun ResultFilters(state: SearchState, viewModel: SearchViewModel) {
    LazyRow(
        Modifier.fadingEdges(),
        contentPadding = PaddingValues(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item {
            FollowSearchPill(state.isFollowed, viewModel::followSearch)
        }
        items(SearchScope.entries) { scope ->
            Pill(scope.label, selected = scope == state.scope, onClick = { viewModel.setScope(scope) })
        }
    }
    if (state.sites.isNotEmpty()) {
        PillRow(
            state.sites,
            labelOf = { site -> site.label },
            onClick = { site -> viewModel.toggleSite(site.name) },
            isSelected = { site -> site.name in state.selectedSites },
            modifier = Modifier.padding(vertical = VeilSpacing.extraSmall),
        )
    }
}

/** The pill that follows the search, or confirms it is followed. */
@Composable
private fun FollowSearchPill(isFollowed: Boolean, onFollow: () -> Unit) {
    if (isFollowed) {
        Pill("Following", onClick = {}, icon = VeilIcons.Check)
        return
    }
    Pill("Follow search", onClick = onFollow, icon = VeilIcons.Follow)
}

/** The results of the chosen scope. */
@Composable
private fun Results(state: SearchState, navigator: AppNavigator) {
    val nothingFound = state.scenes.isEmpty() && state.galleries.isEmpty() && state.performers.isEmpty() && state.studios.isEmpty()
    if (nothingFound && !state.isLoading && !state.isSearchingSites) {
        EmptyMessage(state.error ?: "Nothing found.")
        return
    }
    when (state.scope) {
        SearchScope.ALL -> AllResults(state, navigator)
        SearchScope.SCENES -> SceneGrid(state, navigator) {}
        SearchScope.GALLERIES -> GalleryGrid(state, navigator)
        SearchScope.PERFORMERS -> AvatarGrid(state, navigator, performers = true)
        SearchScope.STUDIOS -> AvatarGrid(state, navigator, performers = false)
    }
}

/** Performers and studios as rows on top, then the videos, then the galleries. */
@Composable
private fun AllResults(state: SearchState, navigator: AppNavigator) {
    SceneGrid(state, navigator) {
        fullWidthItem("performers") { PerformerRow("Performers", state.performers, navigator::openPerformer) }
        fullWidthItem("studios") { StudioRow("Studios", state.studios, navigator::openStudio) }
    }
}

/** The scene grid; [header] adds rows above the videos, and the galleries follow below. */
@Composable
private fun SceneGrid(state: SearchState, navigator: AppNavigator, header: LazyGridScope.() -> Unit) {
    LazyVerticalGrid(
        columns = GridCells.Adaptive(SceneCellWidth),
        contentPadding = ResultGridPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        header()
        items(state.scenes, key = { scene -> scene.id }) { scene ->
            SceneCard(scene, onClick = { navigator.openScene(scene.id) })
        }
        if (state.scope == SearchScope.ALL) {
            fullWidthItem("galleries") { GalleryRow("Galleries", state.galleries, navigator::openGallery) }
        }
    }
}

/** Galleries as a poster grid. */
@Composable
private fun GalleryGrid(state: SearchState, navigator: AppNavigator) {
    LazyVerticalGrid(
        columns = GridCells.Adaptive(PosterCellWidth),
        contentPadding = ResultGridPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        items(state.galleries, key = { gallery -> gallery.id }) { gallery ->
            GalleryCard(gallery, onClick = { navigator.openGallery(gallery.id) })
        }
    }
}

/** Performers or studios as a round-photo grid. */
@Composable
private fun AvatarGrid(state: SearchState, navigator: AppNavigator, performers: Boolean) {
    LazyVerticalGrid(
        columns = GridCells.Adaptive(AvatarCellWidth),
        contentPadding = ResultGridPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        if (performers) {
            items(state.performers, key = { performer -> performer.id }) { performer ->
                PerformerCard(performer, onClick = { navigator.openPerformer(performer.id) })
            }
            return@LazyVerticalGrid
        }
        items(state.studios, key = { studio -> studio.id }) { studio ->
            StudioCard(studio, onClick = { navigator.openStudio(studio.id) })
        }
    }
}
