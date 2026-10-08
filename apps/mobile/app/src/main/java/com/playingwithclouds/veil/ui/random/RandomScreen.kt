package com.playingwithclouds.veil.ui.random

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.paging.PagedList

/** An endless stream of random stored scenes. */
class RandomViewModel : ViewModel() {

    /** Every page is a fresh random pick, so the offset is ignored. */
    val scenes = PagedList<SceneSummary>(viewModelScope, PAGE_SIZE, { scene -> scene.id }) { _ ->
        FeedRepository.randomScenes(PAGE_SIZE)
    }

    init {
        scenes.loadMore()
    }

    companion object {
        private const val PAGE_SIZE = 24
    }
}

/** Random scenes from the library. */
@Composable
fun RandomScreen(navigator: AppNavigator) {
    val viewModel = viewModel { RandomViewModel() }
    Scaffold(
        topBar = {
            VeilTopBar(
                title = "Random",
                onBack = navigator::back,
                actions = {
                    RoundIconButton(VeilIcons.Random, contentDescription = "Shuffle", onClick = { viewModel.scenes.refresh() })
                },
            )
        },
    ) { padding ->
        PagedGrid(
            paged = viewModel.scenes,
            keyOf = { scene -> scene.id },
            emptyText = "The library has no videos yet.",
            modifier = Modifier.padding(padding),
        ) { _, scene ->
            SceneCard(scene, onClick = { navigator.openScene(scene.id) })
        }
    }
}
