package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind

/**
 * The surfaces impressions are logged under, named as the web logs them so the engine reads both
 * clients alike.
 */
object ImpressionSurface {

    /** Home's recommendations: the top picks, then More for you. */
    const val HOME = "home"

    /** Home's Continue watching row. */
    const val CONTINUE = "continue"

    /** A video page's Up next list and the related panel over the fullscreen video. */
    const val RELATED = "related"

    /** Search results. */
    const val SEARCH = "search"
}

/**
 * What was shown ("impression") and what was opened ("choice"), on which surface and at which
 * 0-based position, into the [events] log. A record counts as shown once per surface until
 * [resetShown]; a choice is sent right away.
 */
class ImpressionLogger(private val events: EventLog) {

    /** Surface and record pairs already logged as shown, so re-renders don't count twice. */
    private val shownKeys = mutableSetOf<String>()

    /** Logs a record as shown, once per record and surface until [resetShown]. */
    @Synchronized
    fun recordShown(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        val key = "$surface|${subjectKind.rawValue}|$subjectId"
        if (!shownKeys.add(key)) {
            return
        }
        events.log(eventOf(EventKind.impression, subjectId, subjectKind, surface, position))
    }

    /** Logs that the user opened a record, sending it right away. */
    fun recordClicked(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        events.log(eventOf(EventKind.choice, subjectId, subjectKind, surface, position), flush = true)
    }

    /**
     * Forgets which records were logged as shown on [surface], or on every surface without one,
     * e.g. when the list is reloaded or the screen is opened again.
     */
    @Synchronized
    fun resetShown(surface: String? = null) {
        if (surface == null) {
            shownKeys.clear()
            return
        }
        shownKeys.removeAll { key -> key.startsWith("$surface|") }
    }

    /** An impression or choice event with where it happened. */
    private fun eventOf(kind: EventKind, subjectId: String, subjectKind: RecordKind, surface: String, position: Int): LoggedEvent {
        return LoggedEvent(kind, subjectKind, subjectId, mapOf("surface" to surface, "position" to position))
    }
}
