package com.playingwithclouds.veil.ui.galleries

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.AlikeCandidate
import com.playingwithclouds.veil.data.CollectionRepository
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.GalleryDetail
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Everything the gallery page shows: the set and the user's reactions to it. */
data class GalleryState(
    val gallery: LoadState<GalleryDetail> = LoadState.Loading,
    /** The page of a stub gallery is being fetched from its site. */
    val fetchingImages: Boolean = false,
    val verdict: Verdict? = null,
    val oCount: Int = 0,
    /** The user's collections holding the gallery, Favorites included. */
    val collectionIds: Set<String> = emptySet(),
    /** A one-off note for a snackbar. */
    val message: String? = null,
) {

    /** Whether the gallery is in the Favorites collection. */
    val isFavorite: Boolean
        get() = CollectionRepository.FAVORITES_ID in collectionIds
}

/** The "find the video" search: videos the set may come from. */
data class VideoSearchState(
    val candidates: List<AlikeCandidate> = emptyList(),
    val isLoading: Boolean = false,
    val isLoaded: Boolean = false,
    val linkingUrl: String? = null,
)

/** One gallery; a stub gallery has its page fetched from the origin site on first open. */
class GalleryViewModel(private val galleryId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(GalleryState())
    private val mutableVideoSearch = MutableStateFlow(VideoSearchState())
    private val mutableUserCollections = MutableStateFlow<List<CollectionSummary>>(emptyList())

    /** The page state. */
    val state: StateFlow<GalleryState> = mutableState

    /** The search for the set's video. */
    val videoSearch: StateFlow<VideoSearchState> = mutableVideoSearch

    /** The user's own collections, for the add-to-collection dialog. */
    val userCollections: StateFlow<List<CollectionSummary>> = mutableUserCollections

    init {
        load()
        loadReactions()
    }

    /** Loads the stored gallery, then fetches its images while it has none. */
    fun load() {
        mutableState.update { current -> current.copy(gallery = LoadState.Loading) }
        viewModelScope.launch {
            try {
                var detail = requireNotNull(GalleryRepository.gallery(galleryId)) { "Gallery not found" }
                mutableState.update { current -> current.copy(gallery = LoadState.Loaded(detail)) }
                if (detail.images.isEmpty()) {
                    mutableState.update { current -> current.copy(fetchingImages = true) }
                    detail = GalleryRepository.ensureImages(galleryId) ?: detail
                    mutableState.update { current -> current.copy(gallery = LoadState.Loaded(detail), fetchingImages = false) }
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(gallery = LoadState.Failed(error.displayMessage()), fetchingImages = false) }
            }
        }
    }

    /** Loads the like state, the O-count and the collections holding the gallery. */
    private fun loadReactions() {
        viewModelScope.launch {
            val verdict = runCatching { SceneRepository.verdict(galleryId) }.getOrNull()
            mutableState.update { current -> current.copy(verdict = verdict) }
        }
        viewModelScope.launch {
            val count = runCatching { SceneRepository.oCount(galleryId) }.getOrDefault(0)
            mutableState.update { current -> current.copy(oCount = count) }
        }
        viewModelScope.launch {
            val ids = runCatching { SceneRepository.collectionIds(galleryId) }.getOrDefault(emptySet())
            mutableState.update { current -> current.copy(collectionIds = ids) }
        }
    }

    /** Likes or dislikes the gallery; choosing the active verdict again clears it. */
    fun react(verdict: Verdict) {
        val previous = mutableState.value.verdict
        if (previous == verdict) {
            mutableState.update { current -> current.copy(verdict = null) }
            viewModelScope.launch { runCatching { SceneRepository.clearVerdict(galleryId) } }
            return
        }
        mutableState.update { current -> current.copy(verdict = verdict) }
        viewModelScope.launch {
            val result = runCatching { SceneRepository.setVerdict(galleryId, verdict) }
            if (result.isFailure) {
                mutableState.update { current -> current.copy(verdict = previous) }
            }
        }
    }

    /** Adds an O-counter event. */
    fun incrementOCount() {
        viewModelScope.launch {
            val total = runCatching { SceneRepository.incrementOCount(galleryId) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(oCount = total) }
        }
    }

    /** Removes the latest O-counter event. */
    fun decrementOCount() {
        viewModelScope.launch {
            val total = runCatching { SceneRepository.decrementOCount(galleryId) }.getOrNull() ?: return@launch
            mutableState.update { current -> current.copy(oCount = total) }
        }
    }

    /** Puts the gallery into the Favorites collection or takes it out. */
    fun toggleFavorite() {
        setInCollection(CollectionRepository.FAVORITES_ID, !mutableState.value.isFavorite)
    }

    /** Loads the user's collections for the add-to-collection dialog. */
    fun loadUserCollections() {
        viewModelScope.launch {
            mutableUserCollections.value = runCatching { CollectionRepository.collections("user", COLLECTION_LIMIT, 0) }
                .getOrDefault(emptyList())
        }
    }

    /** Adds the gallery to a collection or removes it from there. */
    fun setInCollection(collection: CollectionSummary, included: Boolean) {
        setInCollection(collection.id, included)
    }

    /** Adds the gallery to the collection [collectionId] or removes it, undoing the tick on failure. */
    private fun setInCollection(collectionId: String, included: Boolean) {
        val before = mutableState.value.collectionIds
        var after = before - collectionId
        if (included) {
            after = before + collectionId
        }
        mutableState.update { current -> current.copy(collectionIds = after) }
        viewModelScope.launch {
            val result = runCatching { SceneRepository.setInCollection(collectionId, galleryId, included) }
            if (result.isFailure) {
                mutableState.update { current -> current.copy(collectionIds = before, message = "Could not update the collection") }
            }
        }
    }

    /** Creates a collection that already contains the gallery. */
    fun createCollectionWithGallery(name: String) {
        viewModelScope.launch {
            try {
                val created = CollectionRepository.create(name)
                CollectionRepository.add(created.id, galleryId)
                mutableUserCollections.value = mutableUserCollections.value + created
                mutableState.update { current -> current.copy(collectionIds = current.collectionIds + created.id) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = error.displayMessage()) }
            }
        }
    }

    /** Searches every scene site for the video the set comes from, once. */
    fun findVideo() {
        if (mutableVideoSearch.value.isLoading || mutableVideoSearch.value.isLoaded) {
            return
        }
        mutableVideoSearch.value = VideoSearchState(isLoading = true)
        viewModelScope.launch {
            val candidates = runCatching { GalleryRepository.findVideos(galleryId) }.getOrDefault(emptyList())
            mutableVideoSearch.value = VideoSearchState(candidates = candidates, isLoaded = true)
        }
    }

    /** Stores the picked video as the gallery's and hands the scene to [onLinked]. */
    fun linkVideo(candidate: AlikeCandidate, onLinked: (String) -> Unit) {
        mutableVideoSearch.update { current -> current.copy(linkingUrl = candidate.sourceUrl) }
        viewModelScope.launch {
            try {
                val sceneId = GalleryRepository.linkVideo(galleryId, candidate)
                mutableState.update { current -> current.withVideo(EntityRef(sceneId, candidate.title, null)) }
                onLinked(sceneId)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(message = "Could not open that video") }
            }
            mutableVideoSearch.update { current -> current.copy(linkingUrl = null) }
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

/** The state with [video] as the loaded gallery's video. */
private fun GalleryState.withVideo(video: EntityRef): GalleryState {
    val loaded = gallery as? LoadState.Loaded ?: return this
    return copy(gallery = LoadState.Loaded(loaded.value.copy(video = video)))
}
