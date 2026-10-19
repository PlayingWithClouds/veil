package com.playingwithclouds.veil.ui.cams

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.CamModel
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

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
    /** Finished recordings of this room that became scenes, newest first. */
    val pastRecordings: List<CamRecording> = emptyList(),
    val recordBusy: Boolean = false,
    val autoRecordBusy: Boolean = false,
    /** A one-off note for a snackbar. */
    val message: String? = null,
)

/**
 * One live room: its stream (asked for again on errors, since the site's token URL expires within
 * about a minute), following, recording and auto-record.
 */
class CamRoomViewModel(private val camModelId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(CamRoomState())
    private var streamAttempts = 0

    /** The page state. */
    val state: StateFlow<CamRoomState> = mutableState

    init {
        load()
    }

    /** Loads the room with its online status, then its stream when live. */
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

    /** Reloads this room's recordings: the running one and the finished scenes. */
    fun refreshRecordings() {
        viewModelScope.launch {
            val recordings = runCatching { CamRepository.recordings(camModelId) }.getOrNull() ?: return@launch
            mutableState.update { current ->
                current.copy(
                    recording = recordings.firstOrNull { recording -> recording.isRecording },
                    pastRecordings = recordings.filter { recording -> !recording.isRecording && recording.sceneId != null },
                )
            }
        }
    }

    /** Follows or unfollows the room. */
    fun toggleFollow() {
        val model = loadedModel() ?: return
        viewModelScope.launch {
            try {
                val updated = CamRepository.setFollowed(camModelId, !model.followed)
                mutableState.update { current -> current.copy(model = LoadState.Loaded(updated.withOnline(model.online))) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
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
                val recording = CamRepository.startRecording(camModelId, quality, fromSecondsAgo)
                mutableState.update { current -> current.copy(recording = recording, recordBusy = false) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(recordBusy = false, message = error.displayMessage()) }
            }
        }
    }

    /** Stops the running recording; it becomes a scene among the finished ones. */
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

    /** Turns auto-record on or off; [quality] stays as it was when null. On also follows. */
    fun setAutoRecord(enabled: Boolean, quality: String? = null) {
        val model = loadedModel() ?: return
        mutableState.update { current -> current.copy(autoRecordBusy = true) }
        viewModelScope.launch {
            try {
                val updated = CamRepository.setAutoRecord(camModelId, enabled, quality)
                mutableState.update { current ->
                    current.copy(model = LoadState.Loaded(updated.withOnline(model.online)), autoRecordBusy = false)
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
 * The room as a mutation returned it, keeping the online status the page checked: the mutations
 * answer from the stored row, which may lag behind the probe.
 */
private fun CamModel.withOnline(online: Boolean): CamModel {
    return copy(online = online)
}
