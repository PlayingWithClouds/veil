package com.playingwithclouds.veil.ui.studios

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
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.StudioCard
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.paging.SearchablePagedList

/** The studio index with a name search. */
class StudiosViewModel : ViewModel() {

    /** Studios matching the search box, in name order. */
    val studios = SearchablePagedList(viewModelScope, PAGE_SIZE, { studio -> studio.id }) { query, offset ->
        EntityRepository.studios(query.ifEmpty { null }, null, PAGE_SIZE, offset)
    }

    companion object {
        private const val PAGE_SIZE = 60
    }
}

/** All studios and channels as a searchable grid. */
@Composable
fun StudiosScreen(navigator: AppNavigator) {
    val viewModel = viewModel { StudiosViewModel() }
    val query by viewModel.studios.query.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Studios", onBack = navigator::back) }) { padding ->
        Column(Modifier.padding(padding)) {
            SearchField(query, viewModel.studios::setQuery, "Search studios")
            PagedGrid(
                paged = viewModel.studios.paged,
                keyOf = { studio -> studio.id },
                emptyText = "No studios found.",
                cellWidth = AvatarCellWidth,
            ) { _, studio ->
                StudioCard(studio, onClick = { navigator.openStudio(studio.id) })
            }
        }
    }
}
