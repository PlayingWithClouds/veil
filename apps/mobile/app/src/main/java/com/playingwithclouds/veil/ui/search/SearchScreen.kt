package com.playingwithclouds.veil.ui.search

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyGridScope
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Business
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.NotificationsActive
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Sell
import androidx.compose.material3.AssistChip
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.graphics.vector.ImageVector
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
import com.playingwithclouds.veil.ui.entity.GalleryRow
import com.playingwithclouds.veil.ui.entity.PerformerRow
import com.playingwithclouds.veil.ui.entity.StudioRow

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
            Row(verticalAlignment = Alignment.CenterVertically) {
                IconButton(onClick = navigator::back) { Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back") }
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
                LinearProgressIndicator(Modifier.fillMaxWidth())
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
    LazyColumn {
        items(suggestions, key = { suggestion -> "${suggestion.kind}:${suggestion.text}:${suggestion.entityId}" }) { suggestion ->
            ListItem(
                headlineContent = { Text(suggestion.text) },
                supportingContent = { suggestion.detail?.let { detail -> Text(detail) } },
                leadingContent = { SuggestionIcon(suggestion) },
                trailingContent = {
                    if (suggestion.kind == "RECENT") {
                        IconButton(onClick = { viewModel.forget(suggestion) }) {
                            Icon(Icons.Filled.Close, contentDescription = "Forget")
                        }
                    }
                },
                modifier = Modifier.clickable { chooseSuggestion(suggestion, viewModel, navigator) },
            )
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
        "RECENT" -> Icons.Filled.History
        "TAG" -> Icons.Filled.Sell
        "PERFORMER" -> Icons.Filled.Person
        "STUDIO" -> Icons.Filled.Business
        else -> Icons.Filled.Search
    }
    Icon(icon, contentDescription = null)
}

/** Scope chips, the follow button and the site filter. */
@Composable
private fun ResultFilters(state: SearchState, viewModel: SearchViewModel) {
    LazyRow(contentPadding = PaddingValues(horizontal = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        item {
            FollowSearchChip(state.isFollowed, viewModel::followSearch)
        }
        items(SearchScope.entries) { scope ->
            FilterChip(selected = scope == state.scope, onClick = { viewModel.setScope(scope) }, label = { Text(scope.label) })
        }
    }
    if (state.sites.isNotEmpty()) {
        LazyRow(contentPadding = PaddingValues(horizontal = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            items(state.sites, key = { site -> site.name }) { site ->
                FilterChip(
                    selected = site.name in state.selectedSites,
                    onClick = { viewModel.toggleSite(site.name) },
                    label = { Text(site.label) },
                )
            }
        }
    }
}

/** The chip that follows the search, or confirms it is followed. */
@Composable
private fun FollowSearchChip(isFollowed: Boolean, onFollow: () -> Unit) {
    if (isFollowed) {
        AssistChip(onClick = {}, label = { Text("Following") }, leadingIcon = { Icon(Icons.Filled.Check, contentDescription = null) })
        return
    }
    AssistChip(
        onClick = onFollow,
        label = { Text("Follow search") },
        leadingIcon = { Icon(Icons.Filled.NotificationsActive, contentDescription = null) },
    )
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
        contentPadding = PaddingValues(12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
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
        contentPadding = PaddingValues(12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
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
        contentPadding = PaddingValues(12.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
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
