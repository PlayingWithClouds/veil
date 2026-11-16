package com.playingwithclouds.veil.ui.home

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.AudioSummary
import com.playingwithclouds.veil.data.CamSummary
import com.playingwithclouds.veil.data.HomeRepository
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.RecommendedVideo
import com.playingwithclouds.veil.data.StorySummary
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

/** Picks in the top row; More for you lists the next [FOR_YOU_COUNT]. */
private const val TOP_PICK_COUNT = 5
private const val FOR_YOU_COUNT = 15

/** Asked for beyond what's shown, to make up for videos being watched (they have their own row). */
private const val RECOMMENDATION_SLACK = 20

/** Cards in each of the other rows. */
private const val ROW_LIMIT = 20

/** Log tag of rows that failed to load. */
private const val TAG = "VeilHome"

/** The recommendations Home shows: the top picks row, then the More for you list. */
data class HomePicks(val top: List<RecommendedVideo>, val forYou: List<RecommendedVideo>)

/**
 * Home's rows, as the web home has them: top picks and More for you from the recommendations,
 * Continue watching, Live now, Continue reading, Stories, Audio and Performers. Each row loads on
 * its own and stays empty (hidden) when it fails; only the recommendations report failure. Videos
 * being watched are never recommended here: they have their own row.
 */
class HomeViewModel : ViewModel() {

    private val recommended = MutableStateFlow<LoadState<List<RecommendedVideo>>>(LoadState.Loading)
    private val continueWatchingRow = MutableStateFlow(emptyList<VideoSummary>())
    private val liveRow = MutableStateFlow(emptyList<CamSummary>())
    private val continueReadingRow = MutableStateFlow(emptyList<StorySummary>())
    private val storiesRow = MutableStateFlow(emptyList<StorySummary>())
    private val audiosRow = MutableStateFlow(emptyList<AudioSummary>())
    private val performersRow = MutableStateFlow(emptyList<PerformerSummary>())
    private val refreshingState = MutableStateFlow(false)

    /** The recommendations without the videos being watched, split into top picks and More for you. */
    val picks: StateFlow<LoadState<HomePicks>> = combine(recommended, continueWatchingRow) { state, watching ->
        picksOf(state, watching.map { video -> video.id }.toSet())
    }.stateIn(viewModelScope, SharingStarted.Eagerly, LoadState.Loading)

    val continueWatching: StateFlow<List<VideoSummary>> = continueWatchingRow.asStateFlow()
    val live: StateFlow<List<CamSummary>> = liveRow.asStateFlow()
    val continueReading: StateFlow<List<StorySummary>> = continueReadingRow.asStateFlow()
    val stories: StateFlow<List<StorySummary>> = storiesRow.asStateFlow()
    val audios: StateFlow<List<AudioSummary>> = audiosRow.asStateFlow()
    val performers: StateFlow<List<PerformerSummary>> = performersRow.asStateFlow()

    /** Whether a pull-to-refresh is running. */
    val refreshing: StateFlow<Boolean> = refreshingState.asStateFlow()

    init {
        viewModelScope.launch { loadAll() }
    }

    /** Loads every row again, e.g. on pull-to-refresh; shown cards may be logged again. */
    fun refresh() {
        if (refreshingState.value) {
            return
        }
        AppServices.impressions.resetShown(ImpressionSurface.HOME)
        AppServices.impressions.resetShown(ImpressionSurface.CONTINUE)
        viewModelScope.launch {
            refreshingState.value = true
            loadAll()
            refreshingState.value = false
        }
    }

    /** Loads the recommendations again after they failed. */
    fun retryPicks() {
        viewModelScope.launch { loadRecommendations() }
    }

    /** Logs a video card as shown on a surface at a position. */
    fun recordShown(videoId: String, surface: String, position: Int) {
        AppServices.impressions.recordShown(videoId, RecordKind.video, surface, position)
    }

    /** Logs a video card as opened from a surface at a position. */
    fun recordClicked(videoId: String, surface: String, position: Int) {
        AppServices.impressions.recordClicked(videoId, RecordKind.video, surface, position)
    }

    /** Loads every row side by side and returns once all have answered. */
    private suspend fun loadAll() = coroutineScope {
        launch { loadRecommendations() }
        launch { loadRow("continue watching", continueWatchingRow) { HomeRepository.continueWatching(ROW_LIMIT) } }
        launch { loadRow("live", liveRow) { HomeRepository.live(ROW_LIMIT) } }
        launch { loadRow("continue reading", continueReadingRow) { HomeRepository.continueReading(ROW_LIMIT) } }
        launch { loadRow("stories", storiesRow) { HomeRepository.stories(ROW_LIMIT) } }
        launch { loadRow("audio", audiosRow) { HomeRepository.audios(ROW_LIMIT) } }
        launch { loadRow("performers", performersRow) { HomeRepository.performers(ROW_LIMIT) } }
    }

    /** Loads the recommendations; while a refresh runs the shown ones stay until the new arrive. */
    private suspend fun loadRecommendations() {
        if (recommended.value !is LoadState.Loaded) {
            recommended.value = LoadState.Loading
        }
        try {
            val wanted = TOP_PICK_COUNT + FOR_YOU_COUNT + RECOMMENDATION_SLACK
            recommended.value = LoadState.Loaded(HomeRepository.recommendedVideos(wanted))
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            recommended.value = LoadState.Failed(error.displayMessage())
        }
    }

    /** Fills a row when its data arrives; a row that fails keeps what it showed and logs why. */
    private suspend fun <T> loadRow(name: String, row: MutableStateFlow<List<T>>, load: suspend () -> List<T>) {
        try {
            row.value = load()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            Log.w(TAG, "home row $name failed to load", error)
        }
    }
}

/** Splits the loaded recommendations, minus [watchingIds], into top picks and More for you. */
private fun picksOf(state: LoadState<List<RecommendedVideo>>, watchingIds: Set<String>): LoadState<HomePicks> {
    return when (state) {
        is LoadState.Loading -> LoadState.Loading
        is LoadState.Failed -> state
        is LoadState.Loaded -> {
            val fresh = state.value.filter { pick -> pick.video.id !in watchingIds }
            LoadState.Loaded(HomePicks(fresh.take(TOP_PICK_COUNT), fresh.drop(TOP_PICK_COUNT).take(FOR_YOU_COUNT)))
        }
    }
}
