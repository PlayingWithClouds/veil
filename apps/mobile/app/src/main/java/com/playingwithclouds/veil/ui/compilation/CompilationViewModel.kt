package com.playingwithclouds.veil.ui.compilation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.SceneMatchVideo
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.scenes.CompilationClip
import com.playingwithclouds.veil.scenes.CompilationPlanner
import com.playingwithclouds.veil.scenes.CompilationSource
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Matching videos read per page once the first video and the queue are used up. */
private const val MATCH_PAGE = 12

/** More clips are prepared when playback gets this close to the last prepared one. */
const val PREPARE_AHEAD = 3

/** One clip ready to play: the stretch, its video and the stream it plays from. */
data class ReadyClip(val clip: CompilationClip, val video: VideoSummary, val stream: PlayableStream)

/**
 * What the compilation page shows. [clips] grows as videos resolve; [loading] while more are
 * prepared; [finished] once no video is left; [problem] when nothing could be prepared.
 */
data class CompilationState(
    val likedTags: Set<String> = emptySet(),
    val clips: List<ReadyClip> = emptyList(),
    val loading: Boolean = true,
    val finished: Boolean = false,
    val problem: String? = null,
)

/**
 * The compilation mode: only the stretches of liked scenes, from [firstVideoId] (when the page
 * was opened from a video), then the queue's videos in order, then the videos with the most
 * matching seconds. Videos resolve to streams one at a time, a few clips ahead of playback.
 */
class CompilationViewModel(private val firstVideoId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(CompilationState())
    val state: StateFlow<CompilationState> = mutableState.asStateFlow()

    private val usedVideoIds = mutableSetOf<String>()
    private var pending = ArrayDeque<SceneMatchVideo>()
    private var stage = Stage.FIRST
    private var matchOffset = 0

    /** Where the list of videos comes from next. */
    private enum class Stage { FIRST, QUEUE, MATCHES, DONE }

    init {
        viewModelScope.launch { start() }
    }

    /** Reads the liked scenes and prepares the first clips. */
    private suspend fun start() {
        try {
            val liked = SceneRepository.refreshLikedTags()
            mutableState.update { state -> state.copy(likedTags = liked) }
            if (liked.isEmpty()) {
                mutableState.update { state -> state.copy(loading = false, finished = true, problem = "Pick the scenes you like in Settings first.") }
                return
            }
            prepareMore()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            mutableState.update { state -> state.copy(loading = false, problem = error.displayMessage()) }
        }
    }

    /** Playback reached clip [index]: prepares more when few are left ahead. */
    fun onClipPlaying(index: Int) {
        val current = state.value
        if (current.loading || current.finished || index < current.clips.size - PREPARE_AHEAD) {
            return
        }
        viewModelScope.launch { runCatching { prepareMore() } }
    }

    /** Resolves videos until at least [PREPARE_AHEAD] more clips are ready or none are left. */
    private suspend fun prepareMore() {
        mutableState.update { state -> state.copy(loading = true) }
        var added = 0
        while (added < PREPARE_AHEAD) {
            val video = nextVideo() ?: break
            added += prepare(video)
        }
        mutableState.update { state ->
            var problem: String? = null
            if (state.clips.isEmpty() && stage == Stage.DONE) {
                problem = "No video has the scenes you like yet. Scanned videos show up here."
            }
            state.copy(loading = false, finished = stage == Stage.DONE && pending.isEmpty(), problem = problem)
        }
    }

    /** Adds one video's clips once its stream resolves; returns how many. */
    private suspend fun prepare(match: SceneMatchVideo): Int {
        val source = CompilationSource(match.video.id, match.segments, match.video.durationSeconds)
        val clips = CompilationPlanner.clips(source, state.value.likedTags)
        if (clips.isEmpty()) {
            return 0
        }
        val playable = runCatching { SceneRepository.playable(match.video.id, readPages = false) }.getOrNull()
            ?: runCatching { SceneRepository.playable(match.video.id, readPages = true) }.getOrNull()
            ?: return 0
        val ready = clips.map { clip -> ReadyClip(clip, match.video, playable.stream) }
        mutableState.update { state -> state.copy(clips = state.clips + ready) }
        return ready.size
    }

    /** The next video not used yet, reading the next source when the pending ones run out; null at the end. */
    private suspend fun nextVideo(): SceneMatchVideo? {
        while (true) {
            val next = pending.removeFirstOrNull()
            if (next != null) {
                if (usedVideoIds.add(next.video.id)) {
                    return next
                }
                continue
            }
            if (stage == Stage.DONE) {
                return null
            }
            pending = ArrayDeque(readStage())
        }
    }

    /** Reads the videos of the current stage and moves on to the next. */
    private suspend fun readStage(): List<SceneMatchVideo> {
        val liked = state.value.likedTags
        when (stage) {
            Stage.FIRST -> {
                stage = Stage.QUEUE
                if (firstVideoId.isEmpty()) {
                    return emptyList()
                }
                return SceneRepository.sceneMatches(liked, listOf(firstVideoId), 1, 0)
            }
            Stage.QUEUE -> {
                stage = Stage.MATCHES
                val queue = SceneRepository.queueIds()
                if (queue.isEmpty()) {
                    return emptyList()
                }
                return SceneRepository.sceneMatches(liked, queue, queue.size, 0)
            }
            Stage.MATCHES -> {
                val page = SceneRepository.sceneMatches(liked, null, MATCH_PAGE, matchOffset)
                matchOffset += page.size
                if (page.size < MATCH_PAGE) {
                    stage = Stage.DONE
                }
                return page
            }
            Stage.DONE -> return emptyList()
        }
    }
}
