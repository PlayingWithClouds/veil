package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull

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

/** What the scene model read off the cover a record's card showed: tag names with their probability. */
interface CoverTagSource {

    /** Whether covers are read at all (a model is installed). */
    fun active(): Boolean

    /** The record's cover tags if they're read already; null otherwise. */
    fun cached(recordId: String): Map<String, Double>?

    /** Waits for the record's cover tags; null when there are none. */
    suspend fun await(recordId: String): Map<String, Double>?
}

/** How long an impression waits for its cover's tags before it's logged without them. */
private const val COVER_TAG_WAIT_MILLISECONDS = 2_000L

/** The event data key of the cover tags (see the backend's internal/recommend/events.go). */
private const val COVER_TAGS = "coverTags"

/**
 * What was shown ("impression") and what was opened ("choice"), on which surface and at which
 * 0-based position, into the [events] log. A record counts as shown once per surface until
 * [resetShown]; a choice is sent right away. With [coverTags] (a scene model is installed), each
 * carries the tags read off the card's cover: an impression waits briefly for them in [scope], a
 * choice takes them when they're read.
 */
class ImpressionLogger(
    private val events: EventLog,
    private val coverTags: CoverTagSource? = null,
    private val scope: CoroutineScope? = null,
) {

    /** Surface and record pairs already logged as shown, so re-renders don't count twice. */
    private val shownKeys = mutableSetOf<String>()

    /** Logs a record as shown, once per record and surface until [resetShown]. */
    fun recordShown(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        val key = "$surface|${subjectKind.rawValue}|$subjectId"
        val firstTime = synchronized(this) { shownKeys.add(key) }
        if (!firstTime) {
            return
        }
        val source = coverTags
        val launcher = scope
        if (source == null || launcher == null || !source.active()) {
            events.log(eventOf(EventKind.impression, subjectId, subjectKind, surface, position, null))
            return
        }
        launcher.launch {
            val tags = withTimeoutOrNull(COVER_TAG_WAIT_MILLISECONDS) { source.await(subjectId) }
            events.log(eventOf(EventKind.impression, subjectId, subjectKind, surface, position, tags))
        }
    }

    /** Logs that the user opened a record, sending it right away. */
    fun recordClicked(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        val tags = coverTags?.cached(subjectId)
        events.log(eventOf(EventKind.choice, subjectId, subjectKind, surface, position, tags), flush = true)
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

    /** An impression or choice event with where it happened and, when read, the cover's tags. */
    private fun eventOf(kind: EventKind, subjectId: String, subjectKind: RecordKind, surface: String, position: Int, tags: Map<String, Double>?): LoggedEvent {
        val data = mutableMapOf<String, Any>("surface" to surface, "position" to position)
        if (!tags.isNullOrEmpty()) {
            data[COVER_TAGS] = tags
        }
        return LoggedEvent(kind, subjectKind, subjectId, data)
    }
}
