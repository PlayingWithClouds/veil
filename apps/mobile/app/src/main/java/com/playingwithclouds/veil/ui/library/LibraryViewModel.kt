package com.playingwithclouds.veil.ui.library

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.FeedRepository
import com.playingwithclouds.veil.data.JobQueue
import com.playingwithclouds.veil.data.JobRepository
import com.playingwithclouds.veil.data.MediaCard
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.loadInto
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** Downloads, the download queue and the watchlist. */
class LibraryViewModel : ViewModel() {

    private val mutableDownloaded = MutableStateFlow<LoadState<List<SceneSummary>>>(LoadState.Loading)
    private val mutableQueue = MutableStateFlow<LoadState<JobQueue>>(LoadState.Loading)
    private val mutableWatchlist = MutableStateFlow<LoadState<List<MediaCard>>>(LoadState.Loading)

    /** Scenes with a completed download. */
    val downloaded: StateFlow<LoadState<List<SceneSummary>>> = mutableDownloaded

    /** Download and background jobs. */
    val queue: StateFlow<LoadState<JobQueue>> = mutableQueue

    /** Scenes saved for later. */
    val watchlist: StateFlow<LoadState<List<MediaCard>>> = mutableWatchlist

    /** Loads the data of one section. */
    fun loadSection(section: LibrarySection) {
        when (section) {
            LibrarySection.DOWNLOADED -> viewModelScope.loadInto(mutableDownloaded) { FeedRepository.downloadedScenes() }
            LibrarySection.QUEUE -> viewModelScope.loadInto(mutableQueue) { JobRepository.jobs() }
            LibrarySection.WATCHLIST -> viewModelScope.loadInto(mutableWatchlist) { FeedRepository.watchlist() }
        }
    }

    /** Refreshes the queue in place, without showing the spinner again. */
    fun reloadQueue() {
        viewModelScope.launch {
            try {
                mutableQueue.value = LoadState.Loaded(JobRepository.jobs())
            } catch (error: Exception) {
                // Keep showing the last known queue.
            }
        }
    }

    /** Queues a failed download again. */
    fun retry(jobId: String) {
        viewModelScope.launch {
            runCatching { JobRepository.retry(jobId) }
            reloadQueue()
        }
    }

    /** Removes a job from the queue. */
    fun delete(jobId: String) {
        viewModelScope.launch {
            runCatching { JobRepository.delete(jobId) }
            reloadQueue()
        }
    }

    /** Removes every background job. */
    fun clearBackgroundJobs() {
        val jobs = (mutableQueue.value as? LoadState.Loaded)?.value ?: return
        viewModelScope.launch {
            for (kind in jobs.background.map { job -> job.kind }.distinct()) {
                runCatching { JobRepository.deleteKind(kind) }
            }
            reloadQueue()
        }
    }

    /** Takes a scene off the watchlist. */
    fun removeFromWatchlist(card: MediaCard) {
        val current = mutableWatchlist.value
        if (current is LoadState.Loaded) {
            mutableWatchlist.value = LoadState.Loaded(current.value.filter { other -> other.mediaId != card.mediaId })
        }
        viewModelScope.launch { runCatching { FeedRepository.removeFromWatchlist(card.mediaId) } }
    }
}
