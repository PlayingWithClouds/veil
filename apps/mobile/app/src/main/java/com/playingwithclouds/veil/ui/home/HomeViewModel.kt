package com.playingwithclouds.veil.ui.home

import android.content.Context
import android.util.Log
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.withTimeoutOrNull
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

/** Picks in the top row; More for you starts with the rest of the first load and goes on endlessly. */
private const val TOP_PICK_COUNT = 5

/** Recommendations in the first load. */
private const val FIRST_LOAD = 40

/** Videos per page More for you loads as it's scrolled to its end. */
private const val FOR_YOU_PAGE = 20

/** Where More for you goes on: more recommendations, then the library newest first, then it ends. */
private enum class ForYouStage { RANKED, LIBRARY, DONE }

/** The best recommendations the hero and the top picks are drawn from, freshly mixed on each open. */
private const val OPENING_POOL = 12

/** How long Home waits for the covers' face check before it shows the opening anyway. */
private const val FACE_SCAN_MILLISECONDS = 1500L

/** Where Home remembers the last hero, so the next open leads with another. */
private const val PREFERENCES = "home"
private const val LAST_HERO_KEY = "lastHeroId"

/**
 * How this open of Home starts: the hero, then the top picks in a fresh order. Repeating the same
 * pictures dulls them (Koukounas & Over), so both change on every load.
 */
private data class Opening(val heroId: String? = null, val topOrder: List<String> = emptyList())

/** Cards in each of the other rows. */
private const val ROW_LIMIT = 20

/** Log tag of rows that failed to load. */
private const val TAG = "VeilHome"

/** The recommendations Home shows: the hero, the top picks row, then the More for you list. */
data class HomePicks(val hero: RecommendedVideo?, val top: List<RecommendedVideo>, val forYou: List<RecommendedVideo>)

/**
 * Home's rows, as the web home has them: top picks and More for you from the recommendations,
 * Continue watching, Live now, Continue reading, Stories, Audio and Performers. Each row loads on
 * its own and stays empty (hidden) when it fails; only the recommendations report failure. Videos
 * being watched are never recommended here: they have their own row.
 */
class HomeViewModel(private val context: Context) : ViewModel() {

    private val recommended = MutableStateFlow<LoadState<List<RecommendedVideo>>>(LoadState.Loading)
    private var opening = Opening()
    private val continueWatchingRow = MutableStateFlow(emptyList<VideoSummary>())
    private val liveRow = MutableStateFlow(emptyList<CamSummary>())
    private val continueReadingRow = MutableStateFlow(emptyList<StorySummary>())
    private val storiesRow = MutableStateFlow(emptyList<StorySummary>())
    private val audiosRow = MutableStateFlow(emptyList<AudioSummary>())
    private val performersRow = MutableStateFlow(emptyList<PerformerSummary>())
    private val refreshingState = MutableStateFlow(false)
    private val moreForYou = MutableStateFlow(emptyList<RecommendedVideo>())
    private val loadingMoreState = MutableStateFlow(false)
    private var forYouStage = ForYouStage.RANKED
    private var rankedOffset = FIRST_LOAD
    private var libraryOffset = 0

    /**
     * The recommendations without the videos being watched, split into top picks and More for you,
     * which goes on with what [loadMoreForYou] adds.
     */
    val picks: StateFlow<LoadState<HomePicks>> = combine(recommended, continueWatchingRow, moreForYou) { state, watching, more ->
        picksOf(state, more, watching.map { video -> video.id }.toSet(), opening)
    }.stateIn(viewModelScope, SharingStarted.Eagerly, LoadState.Loading)

    /** Whether More for you is loading its next page. */
    val loadingMore: StateFlow<Boolean> = loadingMoreState.asStateFlow()

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

    /**
     * Adds the next page to More for you: further recommendations, and once they run out the
     * library's videos newest first. Does nothing while a page loads or after the end.
     */
    fun loadMoreForYou() {
        if (loadingMoreState.value || forYouStage == ForYouStage.DONE || recommended.value !is LoadState.Loaded) {
            return
        }
        loadingMoreState.value = true
        viewModelScope.launch {
            try {
                val page = nextForYouPage()
                moreForYou.value = moreForYou.value + page
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                Log.w(TAG, "more for you failed to load", error)
            } finally {
                loadingMoreState.value = false
            }
        }
    }

    /** The next page of More for you, moving on to the next stage when one runs out. */
    private suspend fun nextForYouPage(): List<RecommendedVideo> {
        while (forYouStage != ForYouStage.DONE) {
            val page = when (forYouStage) {
                ForYouStage.RANKED -> HomeRepository.recommendedVideos(FOR_YOU_PAGE, rankedOffset).also { rankedOffset += FOR_YOU_PAGE }
                ForYouStage.LIBRARY -> HomeRepository.newestVideos(FOR_YOU_PAGE, libraryOffset)
                    .also { libraryOffset += FOR_YOU_PAGE }
                    .map { video -> RecommendedVideo(video, null) }
                ForYouStage.DONE -> emptyList()
            }
            if (page.isNotEmpty()) {
                return page
            }
            forYouStage = ForYouStage.entries[forYouStage.ordinal + 1]
        }
        return emptyList()
    }

    /** Starts More for you over, after the first load came back again. */
    private fun resetMoreForYou() {
        moreForYou.value = emptyList()
        forYouStage = ForYouStage.RANKED
        rankedOffset = FIRST_LOAD
        libraryOffset = 0
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
            val loaded = HomeRepository.recommendedVideos(FIRST_LOAD)
            resetMoreForYou()
            opening = arrangeOpening(loaded)
            recommended.value = LoadState.Loaded(loaded)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            recommended.value = LoadState.Failed(error.displayMessage())
        }
    }

    /**
     * Picks the hero among the best recommendations: one with a preview clip, not the last open's
     * hero, a cover with one or two faces when the check finds any in time; then mixes the rest of
     * the pool into the top picks, faces first.
     */
    private suspend fun arrangeOpening(loaded: List<RecommendedVideo>): Opening {
        val pool = loaded.take(OPENING_POOL)
        if (pool.isEmpty()) {
            return Opening()
        }
        val preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        val lastHeroId = preferences.getString(LAST_HERO_KEY, null)
        val withFaces = facedIds(pool)
        val hero = pickHero(pool, lastHeroId, withFaces)
        preferences.edit().putString(LAST_HERO_KEY, hero.video.id).apply()
        val rest = (pool - hero).shuffled()
        val topOrder = rest.filter { pick -> pick.video.id in withFaces } + rest.filter { pick -> pick.video.id !in withFaces }
        return Opening(hero.video.id, topOrder.map { pick -> pick.video.id })
    }

    /** Ids of the picks whose cover shows one or two faces, as far as the check gets in [FACE_SCAN_MILLISECONDS]. */
    private suspend fun facedIds(pool: List<RecommendedVideo>): Set<String> = coroutineScope {
        val found = java.util.concurrent.ConcurrentHashMap.newKeySet<String>()
        withTimeoutOrNull(FACE_SCAN_MILLISECONDS) {
            pool.map { pick ->
                async {
                    if (coverShowsFaces(context, pick.video.coverUrl)) {
                        found.add(pick.video.id)
                    }
                }
            }.awaitAll()
        }
        found.toSet()
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

/**
 * Splits the loaded recommendations, then the pages loaded since ([more]), minus [watchingIds] and
 * repeats, into the hero, the top picks (in this open's order) and More for you.
 */
private fun picksOf(
    state: LoadState<List<RecommendedVideo>>,
    more: List<RecommendedVideo>,
    watchingIds: Set<String>,
    opening: Opening,
): LoadState<HomePicks> {
    return when (state) {
        is LoadState.Loading -> LoadState.Loading
        is LoadState.Failed -> state
        is LoadState.Loaded -> {
            val fresh = (state.value + more).filter { pick -> pick.video.id !in watchingIds }.distinctBy { pick -> pick.video.id }
            val hero = fresh.firstOrNull { pick -> pick.video.id == opening.heroId }
            val rest = fresh.filter { pick -> pick != hero }
            val ordered = opening.topOrder.mapNotNull { id -> rest.firstOrNull { pick -> pick.video.id == id } }
            val top = (ordered + rest).distinctBy { pick -> pick.video.id }.take(TOP_PICK_COUNT)
            LoadState.Loaded(HomePicks(hero, top, rest.filter { pick -> pick !in top }))
        }
    }
}

/** The hero: among the pool's picks with a preview clip (else all) not shown last time, one with faces when any. */
private fun pickHero(pool: List<RecommendedVideo>, lastHeroId: String?, withFaces: Set<String>): RecommendedVideo {
    val playable = pool.filter { pick -> pick.video.previewUrl != null }.ifEmpty { pool }
    val fresh = playable.filter { pick -> pick.video.id != lastHeroId }.ifEmpty { playable }
    val faced = fresh.filter { pick -> pick.video.id in withFaces }
    return faced.ifEmpty { fresh }.random()
}
