package com.playingwithclouds.veil.ui.scene

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.AlikeCandidate
import com.playingwithclouds.veil.data.CollectionRepository
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.ResumePoint
import com.playingwithclouds.veil.data.SceneDetail
import com.playingwithclouds.veil.data.SceneMarker
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.SceneSite
import com.playingwithclouds.veil.data.SceneSites
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.StreamOption
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.data.WatchProgressStore
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** The stream being played and what the player needs to load it. */
data class ActiveStream(val option: StreamOption, val playable: PlayableStream)

/** Everything the scene page shows besides the player itself. */
data class SceneState(
    val detail: LoadState<SceneDetail> = LoadState.Loading,
    val streams: List<StreamOption> = emptyList(),
    /** The detail fetch from the origin site is running (stub scenes). */
    val fetchingDetail: Boolean = false,
    val active: ActiveStream? = null,
    val resolvingStreamId: String? = null,
    val playbackProblem: String? = null,
    val resume: ResumePoint? = null,
    val markers: List<SceneMarker> = emptyList(),
    /** How much each slice of the scene gets rewatched, 0 to 1; empty until known. */
    val heatmap: List<Float> = emptyList(),
    val verdict: Verdict? = null,
    val oCount: Int = 0,
    val onWatchlist: Boolean = false,
    val related: List<SceneSummary> = emptyList(),
    val site: SceneSite? = null,
    val downloadingStreamId: String? = null,
    /** A one-off note for a snackbar. */
    val message: String? = null,
)

/** State of the find-alternate-sources sheet. */
data class AlikeState(
    val candidates: List<AlikeCandidate> = emptyList(),
    val isLoading: Boolean = false,
    val isLoaded: Boolean = false,
    val attachingUrl: String? = null,
)

/** One scene's page: detail, sources, playback choice and the user's reactions. */
class SceneViewModel(private val sceneId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(SceneState())
    private val mutableAlike = MutableStateFlow(AlikeState())
    private val mutableUserCollections = MutableStateFlow<List<CollectionSummary>>(emptyList())
    private val mutableCollectionIds = MutableStateFlow<Set<String>>(emptySet())
    private var autoplayed = false

    /** The page state. */
    val state: StateFlow<SceneState> = mutableState

    /** The alternate-sources search. */
    val alike: StateFlow<AlikeState> = mutableAlike

    /** The user's own collections, for the add-to-collection dialog. */
    val userCollections: StateFlow<List<CollectionSummary>> = mutableUserCollections

    /** Ids of the user's collections that contain this scene. */
    val collectionIds: StateFlow<Set<String>> = mutableCollectionIds

    init {
        load()
    }

    /** Loads the scene, then everything around it, and follows source and related updates. */
    fun load() {
        autoplayed = false
        mutableState.update { SceneState() }
        viewModelScope.launch {
            try {
                val (detail, streams) = SceneRepository.load(sceneId)
                if (detail == null) {
                    mutableState.update { current -> current.copy(detail = LoadState.Failed("Scene not found")) }
                    return@launch
                }
                mutableState.update { current -> current.copy(detail = LoadState.Loaded(detail), streams = streams) }
                onDetailLoaded(detail, streams)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(detail = LoadState.Failed(error.displayMessage())) }
            }
        }
    }

    /** Starts the side loads and subscriptions once the scene exists. */
    private fun onDetailLoaded(detail: SceneDetail, streams: List<StreamOption>) {
        loadSite(detail.sourceUrl)
        loadReactions()
        followStreams()
        followRelated()
        if (streams.isEmpty()) {
            fetchDetailFromOrigin()
        }
        autoplayIfReady()
    }

    /** Scrapes a stub scene's page for its sources and details, then reloads the detail. */
    private fun fetchDetailFromOrigin() {
        viewModelScope.launch {
            mutableState.update { current -> current.copy(fetchingDetail = true) }
            try {
                val streams = SceneRepository.ensureStreams(sceneId)
                if (streams.isNotEmpty()) {
                    mutableState.update { current -> current.copy(streams = streams) }
                }
                val (detail, _) = SceneRepository.load(sceneId)
                if (detail != null) {
                    mutableState.update { current -> current.copy(detail = LoadState.Loaded(detail)) }
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
            }
            mutableState.update { current -> current.copy(fetchingDetail = false) }
            autoplayIfReady()
        }
    }

    /** Resolves which site the scene came from, for the channel row. */
    private fun loadSite(sourceUrl: String) {
        viewModelScope.launch {
            val sites = runCatching { SearchRepository.installedSites() }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(site = SceneSites.resolve(sourceUrl, sites)) }
        }
    }

    /** Loads markers, the like state, O-count, watchlist state, collections and the resume point. */
    private fun loadReactions() {
        viewModelScope.launch {
            val markers = runCatching { SceneRepository.markers(sceneId) }.getOrDefault(emptyList())
            mutableState.update { current -> current.copy(markers = markers) }
        }
        viewModelScope.launch {
            val verdict = runCatching { SceneRepository.verdict(sceneId) }.getOrNull()
            mutableState.update { current -> current.copy(verdict = verdict) }
        }
        viewModelScope.launch {
            val count = runCatching { SceneRepository.oCount(sceneId) }.getOrDefault(0)
            mutableState.update { current -> current.copy(oCount = count) }
        }
        viewModelScope.launch {
            val onWatchlist = runCatching { SceneRepository.isOnWatchlist(sceneId) }.getOrDefault(false)
            mutableState.update { current -> current.copy(onWatchlist = onWatchlist) }
        }
        viewModelScope.launch {
            mutableCollectionIds.value = runCatching { SceneRepository.collectionIds(sceneId) }.getOrDefault(emptySet())
        }
        viewModelScope.launch {
            val resume = runCatching { SceneRepository.resumePoint(sceneId) }.getOrNull()
            mutableState.update { current -> current.copy(resume = resume) }
            autoplayIfReady()
        }
    }

    /** Takes live source updates as plugins resolve them. */
    private fun followStreams() {
        viewModelScope.launch {
            SceneRepository.streamsChanged(sceneId).catch { }.collect { streams ->
                mutableState.update { current -> current.copy(streams = streams) }
                autoplayIfReady()
            }
        }
    }

    /** Takes the related list now and whenever more scenes get linked. */
    private fun followRelated() {
        viewModelScope.launch {
            SceneRepository.relatedChanged(sceneId).catch { }.collect { related ->
                mutableState.update { current -> current.copy(related = related.filter { scene -> scene.id != sceneId }) }
            }
        }
    }

    /** Starts the best source once, as soon as there is one. */
    private fun autoplayIfReady() {
        val current = mutableState.value
        val best = current.streams.firstOrNull() ?: return
        if (autoplayed || current.active != null || current.detail !is LoadState.Loaded) {
            return
        }
        autoplayed = true
        play(best)
    }

    /** Resolves a source and hands it to the player. */
    fun play(stream: StreamOption) {
        viewModelScope.launch {
            mutableState.update { current -> current.copy(resolvingStreamId = stream.id, playbackProblem = null) }
            try {
                val playable = SceneRepository.resolve(stream)
                mutableState.update { current -> current.copy(active = ActiveStream(stream, playable), resolvingStreamId = null) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(resolvingStreamId = null, playbackProblem = "Could not resolve this source") }
            }
        }
    }

    /** Notes that the player failed on the active source. */
    fun onPlayerError(message: String) {
        mutableState.update { current -> current.copy(playbackProblem = message) }
    }

    /** Stores the playback position for resume. */
    fun saveProgress(positionSeconds: Double, durationSeconds: Double?) {
        if (positionSeconds <= 0) {
            return
        }
        AppScope.launch {
            runCatching { SceneRepository.saveProgress(sceneId, positionSeconds, durationSeconds) }
        }
    }

    /** Refreshes the resume bars after leaving the player. */
    fun refreshProgressStore() {
        AppScope.launch { WatchProgressStore.refresh() }
    }

    /** Likes or dislikes the scene; choosing the active verdict again clears it. */
    fun react(verdict: Verdict) {
        val previous = mutableState.value.verdict
        if (previous == verdict) {
            mutableState.update { current -> current.copy(verdict = null) }
            viewModelScope.launch { runCatching { SceneRepository.clearVerdict(sceneId) } }
            return
        }
        mutableState.update { current -> current.copy(verdict = verdict) }
        viewModelScope.launch {
            val result = runCatching { SceneRepository.setVerdict(sceneId, verdict) }
            if (result.isFailure) {
                mutableState.update { current -> current.copy(verdict = previous) }
            }
        }
    }

    /** Adds an O-counter event. */
    fun incrementOCount() {
        viewModelScope.launch {
            val total = runCatching { SceneRepository.incrementOCount(sceneId) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(oCount = total) }
        }
    }

    /** Removes the latest O-counter event. */
    fun decrementOCount() {
        viewModelScope.launch {
            val total = runCatching { SceneRepository.decrementOCount(sceneId) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(oCount = total) }
        }
    }

    /** Puts the scene on or takes it off the watchlist. */
    fun toggleWatchlist() {
        val target = !mutableState.value.onWatchlist
        mutableState.update { current -> current.copy(onWatchlist = target) }
        viewModelScope.launch {
            val result = runCatching { SceneRepository.setOnWatchlist(sceneId, target) }
            if (result.isFailure) {
                mutableState.update { current -> current.copy(onWatchlist = !target, message = "Could not update the watchlist") }
            }
        }
    }

    /** Queues the active source (or the best one) for download. */
    fun download(stream: StreamOption) {
        val title = (mutableState.value.detail as? LoadState.Loaded)?.value?.title ?: return
        viewModelScope.launch {
            mutableState.update { current -> current.copy(downloadingStreamId = stream.id) }
            try {
                SceneRepository.queueDownload(stream, title)
                mutableState.update { current -> current.copy(message = "Download started") }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
            }
            mutableState.update { current -> current.copy(downloadingStreamId = null) }
        }
    }

    /** Adds a personal marker at the given time. */
    fun addMarker(seconds: Double, tagName: String?, label: String?) {
        viewModelScope.launch {
            runCatching { SceneRepository.createMarker(sceneId, seconds, tagName, label) }
            val markers = runCatching { SceneRepository.markers(sceneId) }.getOrDefault(mutableState.value.markers)
            mutableState.update { current -> current.copy(markers = markers) }
        }
    }

    /** Deletes a personal marker. */
    fun deleteMarker(marker: SceneMarker) {
        mutableState.update { current -> current.copy(markers = current.markers - marker) }
        viewModelScope.launch { runCatching { SceneRepository.deleteMarker(marker.id) } }
    }

    /** Blocks a tag, performer or studio from recommendations. */
    fun block(kind: String, targetId: String, label: String) {
        viewModelScope.launch {
            val result = runCatching { SceneRepository.block(kind, targetId, label) }
            var note = "Could not block $label"
            if (result.isSuccess) {
                note = "Blocked $label"
            }
            mutableState.update { current -> current.copy(message = note) }
        }
    }

    /** Loads the user's collections for the add-to-collection dialog. */
    fun loadUserCollections() {
        viewModelScope.launch {
            mutableUserCollections.value = runCatching { CollectionRepository.collections("user", COLLECTION_LIMIT, 0) }
                .getOrDefault(emptyList())
        }
    }

    /** Adds the scene to a collection or removes it from there. */
    fun setInCollection(collection: CollectionSummary, included: Boolean) {
        val before = mutableCollectionIds.value
        if (included) {
            mutableCollectionIds.value = before + collection.id
        } else {
            mutableCollectionIds.value = before - collection.id
        }
        viewModelScope.launch {
            val result = runCatching { SceneRepository.setInCollection(collection.id, sceneId, included) }
            if (result.isFailure) {
                mutableCollectionIds.value = before
            }
        }
    }

    /** Creates a collection that already contains the scene. */
    fun createCollectionWithScene(name: String) {
        viewModelScope.launch {
            try {
                val created = CollectionRepository.create(name)
                CollectionRepository.add(created.id, sceneId)
                mutableUserCollections.value = mutableUserCollections.value + created
                mutableCollectionIds.value = mutableCollectionIds.value + created.id
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
            }
        }
    }

    /** Searches other sites for copies of the scene, once. */
    fun findAlike() {
        if (mutableAlike.value.isLoading || mutableAlike.value.isLoaded) {
            return
        }
        viewModelScope.launch {
            mutableAlike.update { current -> current.copy(isLoading = true) }
            val found = runCatching { SceneRepository.findAlikeSources(sceneId) }.getOrDefault(emptyList())
            mutableAlike.update { current -> current.copy(candidates = found, isLoading = false, isLoaded = true) }
        }
    }

    /** Attaches an alternate source to the scene and plays its best quality. */
    fun attachAlike(candidate: AlikeCandidate) {
        viewModelScope.launch {
            mutableAlike.update { current -> current.copy(attachingUrl = candidate.sourceUrl) }
            try {
                val streams = SceneRepository.attachAlikeSource(sceneId, candidate)
                mutableState.update { current -> current.copy(streams = streams) }
                val fromPlugin = streams.filter { stream -> stream.pluginName == candidate.plugin }
                val target = fromPlugin.firstOrNull() ?: streams.firstOrNull()
                if (target != null) {
                    play(target)
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
            }
            mutableAlike.update { current -> current.copy(attachingUrl = null) }
        }
    }

    /** Marks the snackbar note as shown. */
    fun messageShown() {
        mutableState.update { current -> current.copy(message = null) }
    }

    companion object {
        private const val COLLECTION_LIMIT = 100
    }
}
