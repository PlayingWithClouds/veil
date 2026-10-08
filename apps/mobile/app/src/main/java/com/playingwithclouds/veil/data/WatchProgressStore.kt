package com.playingwithclouds.veil.data

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** How far into a scene the user got. */
data class SceneProgress(val progressSeconds: Int, val durationSeconds: Int?) {

    /** Fraction watched in 0..1, or null when the runtime is unknown. */
    val fraction: Float?
        get() {
            val total = durationSeconds
            if (total == null || total <= 0) {
                return null
            }
            return (progressSeconds.toFloat() / total).coerceIn(0f, 1f)
        }
}

/**
 * In-progress playback positions by scene id, loaded from the watch history so any scene card can
 * draw a resume bar by looking itself up.
 */
object WatchProgressStore {

    private const val HISTORY_LIMIT = 500

    private val mutableProgress = MutableStateFlow<Map<String, SceneProgress>>(emptyMap())

    /** The unfinished watches. */
    val progress: StateFlow<Map<String, SceneProgress>> = mutableProgress

    /** Reloads the positions; keeps the previous ones when the backend cannot be reached. */
    suspend fun refresh() {
        try {
            mutableProgress.value = unfinished(FeedRepository.watchHistory(HISTORY_LIMIT, 0))
        } catch (error: Exception) {
            // Best-effort: leave the previous map in place on failure.
        }
    }

    /** The in-progress watches by media id; completed and zero-progress rows would draw an empty or full bar. */
    fun unfinished(entries: List<WatchHistoryEntry>): Map<String, SceneProgress> {
        val result = mutableMapOf<String, SceneProgress>()
        for (entry in entries) {
            if (entry.completed || entry.progressSeconds <= 0) {
                continue
            }
            result[entry.mediaId] = SceneProgress(entry.progressSeconds, entry.durationSeconds)
        }
        return result
    }
}
