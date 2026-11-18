package com.playingwithclouds.veil.ui.compilation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.SceneMatchVideo
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.scenes.CompilationClip
import com.playingwithclouds.veil.scenes.CompilationOptions
import com.playingwithclouds.veil.scenes.CompilationPlanner
import com.playingwithclouds.veil.scenes.CompilationSource
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlin.random.Random

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

/** A planned clip with the video it's cut from, waiting for its stream. */
private class PlannedClip(val clip: CompilationClip, val video: VideoSummary)

/**
 * The compilation mode: only the stretches of liked scenes (or [CompilationOptions.tags] when
 * given), from [firstVideoId] (when the page was opened from a video), then the queue's videos in
 * order, then, unless [queueOnly], the videos with the most matching seconds. Each batch of videos
 * is cut into clips as the options say (shortest clip, queue order or shuffled; with [queueOnly]
 * the first video and the queue are one batch, so shuffle mixes them all). Streams resolve a few
 * clips ahead of playback, once per video.
 */
class CompilationViewModel(
    private val firstVideoId: String,
    private val options: CompilationOptions = CompilationOptions(),
    private val queueOnly: Boolean = false,
) : ViewModel() {

    private val mutableState = MutableStateFlow(CompilationState())
    val state: StateFlow<CompilationState> = mutableState.asStateFlow()

    private val usedVideoIds = mutableSetOf<String>()
    private val pending = ArrayDeque<PlannedClip>()
    private val streams = mutableMapOf<String, PlayableStream?>()
    private var stage = Stage.FIRST
    private var matchOffset = 0

    /** Where the list of videos comes from next. */
    private enum class Stage { FIRST, QUEUE, MATCHES, DONE }

    init {
        viewModelScope.launch { start() }
    }

    /** Reads the liked scenes (unless tags were picked) and prepares the first clips. */
    private suspend fun start() {
        try {
            var tags = options.tags
            if (tags.isEmpty()) {
                tags = SceneRepository.refreshLikedTags()
            }
            mutableState.update { state -> state.copy(likedTags = tags) }
            if (tags.isEmpty()) {
                mutableState.update { state -> state.copy(loading = false, finished = true, problem = "Say yes to the scenes you like in Your taste first.") }
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

    /** Resolves clips until [PREPARE_AHEAD] more are ready or none are left. */
    private suspend fun prepareMore() {
        mutableState.update { state -> state.copy(loading = true) }
        var added = 0
        while (added < PREPARE_AHEAD) {
            val planned = nextClip() ?: break
            val stream = streamOf(planned.video.id) ?: continue
            mutableState.update { state -> state.copy(clips = state.clips + ReadyClip(planned.clip, planned.video, stream)) }
            added++
        }
        mutableState.update { state ->
            var problem: String? = null
            if (state.clips.isEmpty() && stage == Stage.DONE) {
                problem = "No video here has those scenes yet. Scanned videos show up here."
            }
            state.copy(loading = false, finished = stage == Stage.DONE && pending.isEmpty(), problem = problem)
        }
    }

    /** The stream a video's clips play from, resolved once; null when nothing plays. */
    private suspend fun streamOf(videoId: String): PlayableStream? {
        if (streams.containsKey(videoId)) {
            return streams[videoId]
        }
        val playable = runCatching { SceneRepository.playable(videoId, readPages = false) }.getOrNull()
            ?: runCatching { SceneRepository.playable(videoId, readPages = true) }.getOrNull()
        streams[videoId] = playable?.stream
        return playable?.stream
    }

    /** The next planned clip, cutting the next batch of videos when the planned ones run out; null at the end. */
    private suspend fun nextClip(): PlannedClip? {
        while (pending.isEmpty()) {
            if (stage == Stage.DONE) {
                return null
            }
            val fresh = readStage().filter { match -> usedVideoIds.add(match.video.id) }
            pending.addAll(plan(fresh))
        }
        return pending.removeFirst()
    }

    /** One batch of videos cut into clips as the options say. */
    private fun plan(matches: List<SceneMatchVideo>): List<PlannedClip> {
        val videos = matches.associate { match -> match.video.id to match.video }
        val sources = matches.map { match -> CompilationSource(match.video.id, match.segments, match.video.durationSeconds) }
        val clips = CompilationPlanner.plan(sources, state.value.likedTags, options.order, options.minClipSeconds, Random.Default)
        return clips.mapNotNull { clip -> videos[clip.videoId]?.let { video -> PlannedClip(clip, video) } }
    }

    /** Reads the videos of the current stage and moves on to the next. */
    private suspend fun readStage(): List<SceneMatchVideo> {
        val tags = state.value.likedTags
        when (stage) {
            Stage.FIRST -> {
                stage = Stage.QUEUE
                if (queueOnly) {
                    stage = Stage.DONE
                    val ids = (listOf(firstVideoId).filter { id -> id.isNotEmpty() } + SceneRepository.queueIds()).distinct()
                    return matchesOf(tags, ids)
                }
                if (firstVideoId.isEmpty()) {
                    return emptyList()
                }
                return SceneRepository.sceneMatches(tags, listOf(firstVideoId), 1, 0)
            }
            Stage.QUEUE -> {
                stage = Stage.MATCHES
                return matchesOf(tags, SceneRepository.queueIds())
            }
            Stage.MATCHES -> {
                val page = SceneRepository.sceneMatches(tags, null, MATCH_PAGE, matchOffset)
                matchOffset += page.size
                if (page.size < MATCH_PAGE) {
                    stage = Stage.DONE
                }
                return page
            }
            Stage.DONE -> return emptyList()
        }
    }

    /** The matches among [ids], in their order. */
    private suspend fun matchesOf(tags: Set<String>, ids: List<String>): List<SceneMatchVideo> {
        if (ids.isEmpty()) {
            return emptyList()
        }
        return SceneRepository.sceneMatches(tags, ids, ids.size, 0)
    }
}
