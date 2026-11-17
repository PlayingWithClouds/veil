package com.playingwithclouds.veil.ui.video

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.HomeRepository
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.PlayerSource
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.VideoDetail
import com.playingwithclouds.veil.data.VideoRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID

/** Library videos like this one that Up next starts with. */
private const val LIBRARY_RELATED = 30

/** Videos per page of what the sites list as related. */
private const val SITE_RELATED_PAGE = 20

/** Recommendations Up next ends with. */
private const val PICKS = 60

/** A video watched this far counts as finished, so it starts over instead of resuming. */
private const val FINISHED_FRACTION = 0.95

/** The source playing and what it resolved to. */
data class ActiveSource(val sourceId: String, val playable: PlayableStream)

/** A thumbs up or down given on this page. */
enum class Verdict { LIKED, DISLIKED }

/** Search pages read for one performer or studio name before the next name's turn. */
private const val PAGES_PER_NAME = 3

/**
 * Where Up next is in its chain: the library's related, then the sites', then the sites' search for
 * its performers and studio by name, then recommendations.
 */
private enum class RelatedStage { LIBRARY, SITES, NAMES, PICKS, DONE }

/**
 * What the video page shows. [problem] says why nothing plays; [message] is a note for the
 * snackbar; [pendingDislike] asks the page to offer Undo for a thumbs down.
 */
data class VideoState(
    val detail: LoadState<VideoDetail> = LoadState.Loading,
    /** The video's card, read first so the cover and preview show while the pages are read. */
    val card: VideoSummary? = null,
    val active: ActiveSource? = null,
    val resolving: Boolean = false,
    val problem: String? = null,
    val favorite: Boolean = false,
    val rating: Double? = null,
    val verdict: Verdict? = null,
    val pendingDislike: Boolean = false,
    val onWatchlist: Boolean = false,
    val downloading: Boolean = false,
    val oCount: Int = 0,
    val related: List<VideoSummary> = emptyList(),
    val relatedLoading: Boolean = false,
    val relatedDone: Boolean = false,
    val message: String? = null,
)

/**
 * The video page, as the web's video page works: opens the video (reading its source pages), plays
 * a local copy first, else the sites' streams in order (the next one when one fails), and resumes
 * where it was left. Up next chains the library's related videos, the sites' related lists (once
 * the pages were read), a site search for its performers and studio, and recommendations. Watching goes to the event log with one session id.
 */
class VideoViewModel(private val videoId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(VideoState())
    val state: StateFlow<VideoState> = mutableState.asStateFlow()

    private val sessionId = UUID.randomUUID().toString()
    private val triedSourceIds = mutableSetOf<String>()
    private var relatedStage = RelatedStage.LIBRARY
    private var siteRelatedOffset = 0

    /** Names Up next still searches the sites for, read from the opened video when that stage starts. */
    private var creditNames: MutableList<String>? = null
    private var namePage = 0

    /**
     * Settles once the video's pages were read (null when it couldn't open). Reading them stores what
     * the sites list as related and names its credits, so the later Up next stages wait for it.
     */
    private val opened = CompletableDeferred<VideoDetail?>()

    init {
        load()
    }

    /** Opens the video and starts its first source; Up next starts loading alongside. */
    fun load() {
        mutableState.update { state -> state.copy(detail = LoadState.Loading, problem = null) }
        viewModelScope.launch {
            val card = runCatching { VideoRepository.card(videoId) }.getOrNull() ?: return@launch
            mutableState.update { state -> state.copy(card = card) }
        }
        viewModelScope.launch { loadMoreRelated() }
        viewModelScope.launch {
            try {
                val detail = VideoRepository.open(videoId)
                opened.complete(detail)
                if (detail == null) {
                    mutableState.update { state -> state.copy(detail = LoadState.Failed("This video is gone.")) }
                    return@launch
                }
                mutableState.update { state ->
                    state.copy(detail = LoadState.Loaded(detail), favorite = detail.favorite, rating = detail.rating, oCount = detail.oCount)
                }
                log(EventKind.view)
                startNextSource()
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                opened.complete(null)
                mutableState.update { state -> state.copy(detail = LoadState.Failed(error.displayMessage())) }
            }
        }
    }

    /** Where playback starts: the resume point, unless the video was watched (nearly) to the end. */
    fun startPositionMilliseconds(): Long {
        val detail = (state.value.detail as? LoadState.Loaded)?.value ?: return 0
        val progress = detail.summary.progressSeconds ?: return 0
        val duration = detail.summary.durationSeconds
        if (duration != null && progress >= duration * FINISHED_FRACTION) {
            return 0
        }
        return (progress * MILLISECONDS_PER_SECOND).toLong()
    }

    /** The player failed on the playing source: the next untried one starts, else the page says why. */
    fun onPlayerError(message: String) {
        viewModelScope.launch {
            val started = startNextSource()
            if (!started) {
                mutableState.update { state -> state.copy(active = null, problem = message) }
            }
        }
    }

    /** Plays the source picked in the source menu. */
    fun selectSource(source: PlayerSource) {
        viewModelScope.launch { start(source) }
    }

    /** Appends the next videos of Up next; does nothing while a page loads or when it has ended. */
    suspend fun loadMoreRelated() {
        val current = state.value
        if (current.relatedLoading || current.relatedDone) {
            return
        }
        mutableState.update { state -> state.copy(relatedLoading = true) }
        var found = emptyList<VideoSummary>()
        try {
            found = nextRelated()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            relatedStage = RelatedStage.DONE
        }
        mutableState.update { state ->
            val knownIds = state.related.map { video -> video.id }.toSet() + videoId
            val fresh = found.filter { video -> video.id !in knownIds }.distinctBy { video -> video.id }
            state.copy(related = state.related + fresh, relatedLoading = false, relatedDone = relatedStage == RelatedStage.DONE)
        }
    }

    /** Logs what the watch session reports: progress, played stretches, the finish. */
    fun onWatchEvent(event: WatchEvent) {
        when (event) {
            is WatchEvent.Progress -> log(EventKind.progress, mapOf("position" to event.positionSeconds))
            is WatchEvent.Span -> log(EventKind.heatSpan, mapOf("start" to event.startSeconds, "end" to event.endSeconds))
            is WatchEvent.Finish -> log(EventKind.finish, mapOf("position" to event.positionSeconds), flush = true)
        }
    }

    /** Sends what the session still holds when the page closes; a thumbs down still offering Undo stands. */
    override fun onCleared() {
        if (state.value.pendingDislike) {
            log(EventKind.like, mapOf("value" to -1))
        }
        AppServices.events.flushNow()
    }

    /** Thumbs up: the engine recommends more like it. */
    fun like() {
        if (state.value.verdict == Verdict.LIKED) {
            return
        }
        mutableState.update { state -> state.copy(verdict = Verdict.LIKED) }
        log(EventKind.like, mapOf("value" to 1))
    }

    /** Thumbs down: the page offers Undo before it's logged for good (the log can't take one back). */
    fun dislike() {
        if (state.value.verdict == Verdict.DISLIKED) {
            return
        }
        mutableState.update { state -> state.copy(verdict = Verdict.DISLIKED, pendingDislike = true) }
    }

    /** The thumbs down stood past its Undo: never recommend the video again. */
    fun settleDislike() {
        mutableState.update { state -> state.copy(pendingDislike = false) }
        log(EventKind.like, mapOf("value" to -1), flush = true)
    }

    /** The thumbs down was taken back. */
    fun undoDislike() {
        mutableState.update { state -> state.copy(verdict = null, pendingDislike = false) }
    }

    /** Counts an O at the playing moment. */
    fun logO(positionSeconds: Double) {
        mutableState.update { state -> state.copy(oCount = state.oCount + 1) }
        log(EventKind.o, mapOf("position" to positionSeconds), flush = true)
    }

    /** Flips the favourite mark. */
    fun toggleFavorite() {
        val favorite = !state.value.favorite
        mutableState.update { state -> state.copy(favorite = favorite) }
        runAction("Couldn't save the favourite") { VideoRepository.setFavorite(videoId, favorite) }
    }

    /** Rates the video 0..10; null clears the rating. */
    fun rate(rating: Double?) {
        mutableState.update { state -> state.copy(rating = rating) }
        runAction("Couldn't save the rating") { VideoRepository.rate(videoId, rating) }
    }

    /** Puts the video on the watchlist. */
    fun saveForLater() {
        mutableState.update { state -> state.copy(onWatchlist = true) }
        runAction("Couldn't add to the watchlist") { VideoRepository.addToWatchlist(videoId) }
    }

    /** Downloads the video's best stream in the background. */
    fun download() {
        mutableState.update { state -> state.copy(downloading = true) }
        viewModelScope.launch {
            try {
                VideoRepository.download(videoId)
                showMessage("Downloading. It plays from your library once it's done.")
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { state -> state.copy(downloading = false) }
                showMessage("Couldn't download: ${error.displayMessage()}")
            }
        }
    }

    /** The snackbar showed the message. */
    fun messageShown() {
        mutableState.update { state -> state.copy(message = null) }
    }

    /** Starts the first source not tried yet; false when every one was tried. */
    private suspend fun startNextSource(): Boolean {
        val detail = (state.value.detail as? LoadState.Loaded)?.value ?: return false
        for (source in detail.sources) {
            if (source.id in triedSourceIds) {
                continue
            }
            if (start(source)) {
                return true
            }
        }
        if (detail.sources.isEmpty()) {
            mutableState.update { state -> state.copy(problem = "No site has a stream of this video.") }
        }
        return false
    }

    /** Resolves a source and hands it to the player; false when it gives nothing playable. */
    private suspend fun start(source: PlayerSource): Boolean {
        triedSourceIds.add(source.id)
        val local = source.local
        if (local != null) {
            mutableState.update { state -> state.copy(active = ActiveSource(source.id, local), problem = null) }
            return true
        }
        val streamUrl = source.streamUrl ?: return false
        mutableState.update { state -> state.copy(resolving = true) }
        try {
            val playable = VideoRepository.play(streamUrl)
            if (playable == null) {
                mutableState.update { state -> state.copy(resolving = false, problem = "This stream could not be resolved.") }
                return false
            }
            mutableState.update { state -> state.copy(resolving = false, active = ActiveSource(source.id, playable), problem = null) }
            return true
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            mutableState.update { state -> state.copy(resolving = false, problem = error.displayMessage()) }
            return false
        }
    }

    /** The next videos of Up next, moving to the next stage of the chain when one runs out. */
    private suspend fun nextRelated(): List<VideoSummary> {
        while (relatedStage != RelatedStage.DONE) {
            val found = when (relatedStage) {
                RelatedStage.LIBRARY -> VideoRepository.related(videoId, LIBRARY_RELATED)
                RelatedStage.SITES -> siteRelatedPage()
                RelatedStage.NAMES -> nextNameSearch()
                RelatedStage.PICKS -> HomeRepository.recommendedVideos(PICKS).map { pick -> pick.video }
                RelatedStage.DONE -> emptyList()
            }
            advanceRelated(found.size)
            if (found.isNotEmpty()) {
                return found
            }
        }
        return emptyList()
    }

    /** The next page of what the sites list as related, once the video's pages were read. */
    private suspend fun siteRelatedPage(): List<VideoSummary> {
        opened.await()
        return VideoRepository.siteRelated(videoId, SITE_RELATED_PAGE, siteRelatedOffset)
    }

    /**
     * The next page of a site search for the video's performers, then its studio, quoted so only
     * results naming them come back. A name's turn ends at an empty page or after [PAGES_PER_NAME].
     */
    private suspend fun nextNameSearch(): List<VideoSummary> {
        val names = creditNames ?: creditNamesOf(opened.await()).also { names -> creditNames = names }
        while (names.isNotEmpty()) {
            if (namePage < PAGES_PER_NAME) {
                val found = SearchRepository.collectPage("\"${names.first()}\"", namePage)
                namePage++
                if (found.isNotEmpty()) {
                    return found
                }
            }
            names.removeAt(0)
            namePage = 0
        }
        return emptyList()
    }

    /** Moves past the stage that just answered with [count] videos; the sites' lists and name searches go on until empty. */
    private fun advanceRelated(count: Int) {
        if (relatedStage == RelatedStage.SITES && count > 0) {
            siteRelatedOffset += count
            return
        }
        if (relatedStage == RelatedStage.NAMES && count > 0) {
            return
        }
        relatedStage = RelatedStage.entries[relatedStage.ordinal + 1]
    }

    /** Runs a user-state write; on failure the snackbar says [failure]. */
    private fun runAction(failure: String, write: suspend () -> Unit) {
        viewModelScope.launch {
            try {
                write()
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                showMessage("$failure: ${error.displayMessage()}")
            }
        }
    }

    /** Puts a note in the snackbar. */
    private fun showMessage(text: String) {
        mutableState.update { state -> state.copy(message = text) }
    }

    /** Logs an event about this video in this page's watch session. */
    private fun log(kind: EventKind, data: Map<String, Any> = emptyMap(), flush: Boolean = false) {
        AppServices.events.log(LoggedEvent(kind, RecordKind.video, videoId, data, sessionId), flush)
    }

    private companion object {
        const val MILLISECONDS_PER_SECOND = 1000.0
    }
}

/** The video's performers' names, then its studio's; none when it couldn't open. */
private fun creditNamesOf(detail: VideoDetail?): MutableList<String> {
    if (detail == null) {
        return mutableListOf()
    }
    val names = detail.performers.map { performer -> performer.name }.toMutableList()
    val studio = detail.studio
    if (studio != null) {
        names.add(studio.name)
    }
    return names
}
