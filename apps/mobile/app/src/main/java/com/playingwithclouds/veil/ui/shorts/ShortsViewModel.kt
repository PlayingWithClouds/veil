package com.playingwithclouds.veil.ui.shorts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.ShortItem
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Shorts read at once; the backend ranks them all on every read. */
private const val SHORTS_LIMIT = 80

/** A short swiped away before this share of it played counts as skipped. */
private const val SKIP_SHARE = 0.5

/** The surface skips are logged under, as the engine reads them. */
const val SHORTS_SURFACE = "shorts"

/** The feed and the streams resolved so far, by short key (null: tried, nothing plays). */
data class ShortsState(
    val shorts: LoadState<List<ShortItem>> = LoadState.Loading,
    val streams: Map<String, PlayableStream?> = emptyMap(),
)

/**
 * The shorties feed: the backend's `shorts` (saved clips, segments, the scene model's liked
 * scenes, most replayed moments, short vertical videos), each resolved to its video's stream when
 * it comes near the screen. Swiping one away early logs a skip, which the engine ranks down.
 */
class ShortsViewModel : ViewModel() {

    private val mutableState = MutableStateFlow(ShortsState())
    val state: StateFlow<ShortsState> = mutableState.asStateFlow()

    private val resolving = mutableSetOf<String>()

    init {
        load()
    }

    /** Reads the feed. */
    fun load() {
        mutableState.update { state -> state.copy(shorts = LoadState.Loading) }
        viewModelScope.launch {
            try {
                val shorts = SceneRepository.shorts(SHORTS_LIMIT)
                mutableState.update { state -> state.copy(shorts = LoadState.Loaded(shorts)) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { state -> state.copy(shorts = LoadState.Failed(error.displayMessage())) }
            }
        }
    }

    /** Resolves the stream of [short] once, so it can play when it's on screen. */
    fun prepare(short: ShortItem) {
        if (!resolving.add(short.key)) {
            return
        }
        viewModelScope.launch {
            val playable = runCatching { SceneRepository.playable(short.video.id, readPages = false) }.getOrNull()
                ?: runCatching { SceneRepository.playable(short.video.id, readPages = true) }.getOrNull()
            mutableState.update { state -> state.copy(streams = state.streams + (short.key to playable?.stream)) }
        }
    }

    /** [short] left the screen after [playedSeconds]: a skip when less than half of it played. */
    fun left(short: ShortItem, playedSeconds: Double) {
        val length = short.endSeconds - short.startSeconds
        if (length <= 0 || playedSeconds >= length * SKIP_SHARE) {
            return
        }
        val data = mapOf("surface" to SHORTS_SURFACE, "position" to playedSeconds)
        AppServices.events.log(LoggedEvent(EventKind.skip, short.itemKind, short.itemId, data))
    }
}
