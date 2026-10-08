package com.playingwithclouds.veil.ui.performers

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.AvatarCellWidth
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PerformerCard
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.paging.SearchablePagedList

/** The performer index with a name search. */
class PerformersViewModel : ViewModel() {

    /** Performers matching the search box, in name order. */
    val performers = SearchablePagedList(viewModelScope, PAGE_SIZE, { performer -> performer.id }) { query, offset ->
        EntityRepository.performers(query.ifEmpty { null }, null, PAGE_SIZE, offset)
    }

    companion object {
        private const val PAGE_SIZE = 60
    }
}

/** All performers as a searchable grid. */
@Composable
fun PerformersScreen(navigator: AppNavigator) {
    val viewModel = viewModel { PerformersViewModel() }
    val query by viewModel.performers.query.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Performers", onBack = navigator::back) }) { padding ->
        Column(Modifier.padding(padding)) {
            SearchField(query, viewModel.performers::setQuery, "Search performers")
            PagedGrid(
                paged = viewModel.performers.paged,
                keyOf = { performer -> performer.id },
                emptyText = "No performers found.",
                cellWidth = AvatarCellWidth,
            ) { _, performer ->
                PerformerCard(performer, onClick = { navigator.openPerformer(performer.id) })
            }
        }
    }
}
