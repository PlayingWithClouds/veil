package com.playingwithclouds.veil.ui.tags

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.paging.SearchablePagedList
import com.playingwithclouds.veil.util.formatCount
import com.playingwithclouds.veil.util.tagLabel

/** The tag index with a name search. */
class TagsViewModel : ViewModel() {

    /** Tags matching the search box. */
    val tags = SearchablePagedList(viewModelScope, PAGE_SIZE, { tag -> tag.id }) { query, offset ->
        EntityRepository.tags(query.ifEmpty { null }, PAGE_SIZE, offset)
    }

    companion object {
        private const val PAGE_SIZE = 100
    }
}

/** All tags as a searchable grid of tiles. */
@Composable
fun TagsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { TagsViewModel() }
    val query by viewModel.tags.query.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Tags", onBack = navigator::back) }) { padding ->
        Column(Modifier.padding(padding)) {
            SearchField(query, viewModel.tags::setQuery, "Search tags")
            PagedGrid(
                paged = viewModel.tags.paged,
                keyOf = { tag -> tag.id },
                emptyText = "No tags found.",
                cellWidth = 160.dp,
            ) { _, tag ->
                TagTile(tag, onClick = { navigator.openTag(tag.id) })
            }
        }
    }
}

/** A tag name with how many videos carry it. */
@Composable
private fun TagTile(tag: TagSummary, onClick: () -> Unit) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MaterialTheme.colorScheme.surfaceVariant,
        modifier = Modifier.fillMaxWidth().clickable(onClick = onClick),
    ) {
        Row(Modifier.padding(12.dp), horizontalArrangement = Arrangement.SpaceBetween) {
            Text(tagLabel(tag.name), modifier = Modifier.weight(1f), maxLines = 1)
            Text(formatCount(tag.sceneCount), color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}
