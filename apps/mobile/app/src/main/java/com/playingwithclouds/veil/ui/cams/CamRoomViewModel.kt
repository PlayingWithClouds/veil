package com.playingwithclouds.veil.ui.cams

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.CamModel
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID

/** The surface rooms are logged as watched from, as the web logs it. */
private const val CAMS_SURFACE = "cams"

/** How long the room played is logged every this many seconds of playing. */
private const val WATCH_REPORT_SECONDS = 30

/** Everything a room's page shows. */
data class CamRoomState(
    val model: LoadState<CamModel> = LoadState.Loading,
    /** The live stream to play; replaced by a fresh one (new [streamGeneration]) after an error. */
    val stream: PlayableStream? = null,
    val streamGeneration: Int = 0,
    /** A stream URL is being asked for. */
    val connecting: Boolean = false,
    /** The room is not live, or its stream failed every retry. */
    val offline: Boolean = false,
    /** The recording running for this room, if any. */
    val recording: CamRecording? = null,
    /** The room's finished recordings, as the videos they became, newest first. */
    val pastRecordings: List<VideoSummary> = emptyList(),
    val recordBusy: Boolean = false,
    val autoRecordBusy: Boolean = false,
    /** A one-off note for a snackbar. */
    val message: String? = null,
)

/**
 * One cam room: its stream (asked for again on errors, since the site's token URL expires within
 * about a minute), following, recording and auto-record. Watching is logged as the web's
 * `trackRoomWatch` does: a view on opening, then how long it played.
 */
class CamRoomViewModel(private val camModelId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(CamRoomState())
    private var streamAttempts = 0
    private val sessionId = UUID.randomUUID().toString()
    private var watchedSeconds = 0
    private var reportedSeconds = 0

    /** The page state. */
    val state: StateFlow<CamRoomState> = mutableState

    init {
        log(EventKind.view, mapOf("surface" to CAMS_SURFACE))
        load()
    }

    /** Loads the room with its live state, then its stream when live. */
    fun load() {
        streamAttempts = 0
        mutableState.update { current -> current.copy(model = LoadState.Loading, offline = false) }
        viewModelScope.launch {
            try {
                val model = requireNotNull(CamRepository.model(camModelId)) { "Room not found" }
                mutableState.update { current -> current.copy(model = LoadState.Loaded(model), offline = !model.online) }
                if (model.online) {
                    requestStream()
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(model = LoadState.Failed(error.displayMessage())) }
            }
        }
        refreshRecordings()
    }

    /** Asks for a fresh stream URL; a room whose stream cannot be had counts as offline. */
    private fun requestStream() {
        mutableState.update { current -> current.copy(connecting = true) }
        viewModelScope.launch {
            try {
                val stream = CamRepository.stream(camModelId)
                mutableState.update { current ->
                    current.copy(stream = stream, streamGeneration = current.streamGeneration + 1, connecting = false, offline = false)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(connecting = false, offline = true, stream = null) }
            }
        }
    }

    /** The player failed: try a fresh URL a few times before showing the room as offline. */
    fun onPlayerError() {
        streamAttempts++
        if (streamAttempts > MAX_STREAM_RETRIES) {
            mutableState.update { current -> current.copy(offline = true, stream = null) }
            return
        }
        requestStream()
    }

    /** The stream plays, so later errors get their retries again. */
    fun onPlaying() {
        streamAttempts = 0
    }

    /** The broadcast ended. */
    fun onStreamEnded() {
        mutableState.update { current -> current.copy(offline = true, stream = null) }
    }

    /** One more second of the room playing; how long it played is logged every so often. */
    fun onWatchedSecond() {
        watchedSeconds++
        if (watchedSeconds % WATCH_REPORT_SECONDS == 0) {
            reportWatched()
        }
    }

    /** Reloads this room's recordings: the running one and the finished videos. */
    fun refreshRecordings() {
        viewModelScope.launch {
            val running = runCatching { CamRepository.activeRecordings() }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(recording = running.firstOrNull { recording -> recording.camModelId == camModelId }) }
        }
        viewModelScope.launch {
            val finished = runCatching { CamRepository.recordedVideos(camModelId) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(pastRecordings = finished) }
        }
    }

    /** Follows or unfollows the room. */
    fun toggleFollow() {
        val model = loadedModel() ?: return
        val followed = !model.followed
        mutableState.update { current -> current.copy(model = LoadState.Loaded(model.copy(followed = followed))) }
        viewModelScope.launch {
            try {
                CamRepository.setFollowed(camModelId, followed)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(model = LoadState.Loaded(model), message = error.displayMessage()) }
            }
        }
    }

    /**
     * Starts recording at [quality] ("source" or a height); with [fromSecondsAgo] from that far
     * back in the room's live buffer.
     */
    fun startRecording(quality: String, fromSecondsAgo: Int? = null) {
        mutableState.update { current -> current.copy(recordBusy = true) }
        viewModelScope.launch {
            try {
                CamRepository.startRecording(camModelId, quality, fromSecondsAgo)
                mutableState.update { current -> current.copy(recordBusy = false) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(recordBusy = false, message = error.displayMessage()) }
            }
            refreshRecordings()
        }
    }

    /** Stops the running recording; it becomes a video among the finished ones. */
    fun stopRecording() {
        mutableState.update { current -> current.copy(recordBusy = true) }
        viewModelScope.launch {
            try {
                CamRepository.stopRecording(camModelId)
                mutableState.update { current -> current.copy(recording = null, recordBusy = false) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(recordBusy = false, message = error.displayMessage()) }
            }
            refreshRecordings()
        }
    }

    /** Turns auto-record on or off; [quality] stays as it was when null. */
    fun setAutoRecord(enabled: Boolean, quality: String? = null) {
        val model = loadedModel() ?: return
        mutableState.update { current -> current.copy(autoRecordBusy = true) }
        viewModelScope.launch {
            try {
                val updated = CamRepository.setAutoRecord(camModelId, enabled, quality)
                mutableState.update { current ->
                    current.copy(model = LoadState.Loaded(updated.withOnline(model)), autoRecordBusy = false)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(autoRecordBusy = false, message = error.displayMessage()) }
            }
        }
    }

    /** Marks the snackbar note as shown. */
    fun messageShown() {
        mutableState.update { current -> current.copy(message = null) }
    }

    /** Logs how long the room played once more as the page closes. */
    override fun onCleared() {
        reportWatched()
        AppServices.events.flushNow()
    }

    /** Logs how long the room has played, unless that was logged already. */
    private fun reportWatched() {
        if (watchedSeconds == reportedSeconds) {
            return
        }
        reportedSeconds = watchedSeconds
        log(EventKind.progress, mapOf("watched" to watchedSeconds))
    }

    /** Logs an event about this room in this page's session. */
    private fun log(kind: EventKind, data: Map<String, Any>) {
        AppServices.events.log(LoggedEvent(kind, RecordKind.camModel, camModelId, data, sessionId))
    }

    /** The room, once loaded. */
    private fun loadedModel(): CamModel? {
        return (mutableState.value.model as? LoadState.Loaded)?.value
    }

    companion object {
        /** Fresh stream URLs asked for after an error before the room shows as offline. */
        private const val MAX_STREAM_RETRIES = 3
    }
}

/**
 * The room as a mutation returned it, keeping the live state the page loaded: the mutation answers
 * from the stored row, which may lag behind the site.
 */
private fun CamModel.withOnline(loaded: CamModel): CamModel {
    return copy(online = loaded.online, viewers = loaded.viewers, subject = loaded.subject, thumbnailUrl = loaded.thumbnailUrl)
}
