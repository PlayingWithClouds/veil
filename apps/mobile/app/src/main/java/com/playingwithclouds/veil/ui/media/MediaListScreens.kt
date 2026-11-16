package com.playingwithclouds.veil.ui.media

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.AudioSummary
import com.playingwithclouds.veil.data.HomeRepository
import com.playingwithclouds.veil.data.MediaRepository
import com.playingwithclouds.veil.data.StorySummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.AudioCard
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.PosterCellWidth
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.StoryCard
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** Cards per page of the lists. */
private const val PAGE_SIZE = 30

/** Cards in the "continue" row. */
private const val CONTINUE_LIMIT = 20

/** Width of a card in the "continue" row. */
private val RowCardWidth = 130.dp

/** Stories newest first, and those being read. */
class StoriesViewModel : ViewModel() {

    private val continueState = MutableStateFlow(emptyList<StorySummary>())
    val continueReading: StateFlow<List<StorySummary>> = continueState.asStateFlow()

    val stories = PagedList<StorySummary>(viewModelScope, PAGE_SIZE, { story -> story.id }) { offset ->
        MediaRepository.stories(PAGE_SIZE, offset)
    }

    init {
        stories.loadMore()
        viewModelScope.launch { continueState.value = runCatching { HomeRepository.continueReading(CONTINUE_LIMIT) }.getOrDefault(emptyList()) }
    }
}

/** Audio newest first, and those being listened to. */
class AudiosViewModel : ViewModel() {

    private val continueState = MutableStateFlow(emptyList<AudioSummary>())
    val continueListening: StateFlow<List<AudioSummary>> = continueState.asStateFlow()

    val audios = PagedList<AudioSummary>(viewModelScope, PAGE_SIZE, { audio -> audio.id }) { offset ->
        MediaRepository.audios(PAGE_SIZE, offset)
    }

    init {
        audios.loadMore()
        viewModelScope.launch { continueState.value = runCatching { MediaRepository.continueListening(CONTINUE_LIMIT) }.getOrDefault(emptyList()) }
    }
}

/** Stories: Continue reading, then every story, newest first. */
@Composable
fun StoriesScreen(navigator: AppNavigator) {
    val viewModel = viewModel { StoriesViewModel() }
    val continueReading by viewModel.continueReading.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Stories", onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.stories,
            keyOf = { story -> story.id },
            emptyText = "No stories yet. They come in from the story sites.",
            modifier = Modifier.padding(padding),
            cellWidth = PosterCellWidth,
            header = {
                fullWidthItem("continue") {
                    Shelf("Continue reading", continueReading, key = { story -> story.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { story, _ ->
                        StoryCard(story, onClick = { navigator.openStory(story.id) })
                    }
                }
                fullWidthItem("all-heading") { SectionHeading("Newest") }
            },
        ) { _, story ->
            StoryCard(story, onClick = { navigator.openStory(story.id) })
        }
    }
}

/** Audio: Continue listening, then every audio, newest first. */
@Composable
fun AudiosScreen(navigator: AppNavigator) {
    val viewModel = viewModel { AudiosViewModel() }
    val continueListening by viewModel.continueListening.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Audio", onBack = navigator::back) }) { padding ->
        PagedGrid(
            paged = viewModel.audios,
            keyOf = { audio -> audio.id },
            emptyText = "No audio yet. It comes in from the audio sites.",
            modifier = Modifier.padding(padding),
            cellWidth = PosterCellWidth,
            header = {
                fullWidthItem("continue") {
                    Shelf("Continue listening", continueListening, key = { audio -> audio.id }, itemWidth = RowCardWidth, modifier = Modifier.bleed()) { audio, _ ->
                        AudioCard(audio, onClick = { navigator.openAudio(audio.id) })
                    }
                }
                fullWidthItem("all-heading") { SectionHeading("Newest") }
            },
        ) { _, audio ->
            AudioCard(audio, onClick = { navigator.openAudio(audio.id) })
        }
    }
}
