package com.playingwithclouds.veil.data

import kotlin.random.Random
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update

/** A scene waiting in the queue: just what the queue list shows. */
data class QueueEntry(val sceneId: String, val title: String, val posterPath: String?)

/** What is queued, in order, and whether the next scene is picked at random. */
data class QueueState(val entries: List<QueueEntry> = emptyList(), val shuffle: Boolean = false)

/** The scene summary as a queue entry. */
fun SceneSummary.toQueueEntry(): QueueEntry {
    return QueueEntry(id, title, posterPath)
}

/** [entries] with [entry] first, moved up when it was already queued. */
fun queueWithPlayNext(entries: List<QueueEntry>, entry: QueueEntry): List<QueueEntry> {
    return listOf(entry) + entries.filter { queued -> queued.sceneId != entry.sceneId }
}

/** [entries] with [entry] last; a scene already queued keeps its place. */
fun queueWithAdded(entries: List<QueueEntry>, entry: QueueEntry): List<QueueEntry> {
    if (entries.any { queued -> queued.sceneId == entry.sceneId }) {
        return entries
    }
    return entries + entry
}

/** Which of [size] queued scenes plays next: the first, or a random one when shuffling; -1 when empty. */
fun nextQueueIndex(size: Int, shuffle: Boolean, random: Random): Int {
    if (size == 0) {
        return -1
    }
    if (shuffle) {
        return random.nextInt(size)
    }
    return 0
}

/**
 * The scenes of a playlist [entries] in the order to play them: as they are, or shuffled. Scenes
 * listed twice play once.
 */
fun playlistOrder(entries: List<QueueEntry>, shuffled: Boolean, random: Random): List<QueueEntry> {
    val distinct = entries.distinctBy { entry -> entry.sceneId }
    if (shuffled) {
        return distinct.shuffled(random)
    }
    return distinct
}

/**
 * The session's queue of scenes to play after the current one: "play next" and "add to queue" from
 * anywhere in the app, whole playlists (optionally shuffled), and the scene page's auto-advance.
 * Kept in memory only; a queue is for one sitting.
 */
object PlaybackQueue {

    private val mutableState = MutableStateFlow(QueueState())

    /** The queue as it stands. */
    val state: StateFlow<QueueState> = mutableState

    /** Puts [entry] at the front of the queue. */
    fun playNext(entry: QueueEntry) {
        mutableState.update { current -> current.copy(entries = queueWithPlayNext(current.entries, entry)) }
    }

    /** Appends [entry] to the queue. */
    fun add(entry: QueueEntry) {
        mutableState.update { current -> current.copy(entries = queueWithAdded(current.entries, entry)) }
    }

    /** Takes a scene out of the queue. */
    fun remove(sceneId: String) {
        mutableState.update { current -> current.copy(entries = current.entries.filter { entry -> entry.sceneId != sceneId }) }
    }

    /** Empties the queue. */
    fun clear() {
        mutableState.update { current -> current.copy(entries = emptyList()) }
    }

    /** Turns random picking of the next scene on or off. */
    fun setShuffle(enabled: Boolean) {
        mutableState.update { current -> current.copy(shuffle = enabled) }
    }

    /** Takes the scene to play next off the queue and returns it; null when the queue is empty. */
    fun takeNext(random: Random = Random.Default): QueueEntry? {
        val current = mutableState.value
        val index = nextQueueIndex(current.entries.size, current.shuffle, random)
        if (index < 0) {
            return null
        }
        val next = current.entries[index]
        remove(next.sceneId)
        return next
    }

    /**
     * Replaces the queue with a playlist and returns the scene to start with. With [shuffled] the
     * scenes are in random order and the queue keeps picking at random.
     */
    fun startPlaylist(entries: List<QueueEntry>, shuffled: Boolean, random: Random = Random.Default): QueueEntry? {
        val ordered = playlistOrder(entries, shuffled, random)
        val first = ordered.firstOrNull() ?: return null
        mutableState.value = QueueState(ordered.drop(1), shuffled)
        return first
    }
}
