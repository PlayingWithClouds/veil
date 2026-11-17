package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.RecordKind
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

/**
 * Records the user put out of sight this session (not interested, blocked), as the web's
 * `hiddenRecords` store: cards already on screen leave at once, without reloading their lists.
 * Lists loaded later leave them out on the backend's side.
 */
object HiddenRecords {

    private val hiddenKeys = MutableStateFlow(emptySet<String>())

    /** The hidden records as "kind:id" keys. */
    val keys: StateFlow<Set<String>> = hiddenKeys.asStateFlow()

    /** Hides a record from the cards on screen. */
    fun hide(kind: RecordKind, id: String) {
        hiddenKeys.update { current -> current + keyOf(kind, id) }
    }

    /** Whether a video card should leave: the video, a performer or its studio is hidden. */
    fun hidesVideo(keys: Set<String>, video: VideoSummary): Boolean {
        if (keyOf(RecordKind.video, video.id) in keys) {
            return true
        }
        if (video.performers.any { performer -> keyOf(RecordKind.performer, performer.id) in keys }) {
            return true
        }
        val studio = video.studio ?: return false
        return keyOf(RecordKind.studio, studio.id) in keys
    }

    /** A record's key in the set. */
    private fun keyOf(kind: RecordKind, id: String): String {
        return "${kind.rawValue}:$id"
    }
}
