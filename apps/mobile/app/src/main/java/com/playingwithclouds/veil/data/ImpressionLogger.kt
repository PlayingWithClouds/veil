package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.RecordEventsMutation
import com.playingwithclouds.veil.graphql.type.EventInput
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/**
 * The surfaces impressions are logged under, named as the web logs them so the engine reads both
 * clients alike.
 */
object ImpressionSurface {

    /** Home's recommendations: the top picks, then More for you. */
    const val HOME = "home"

    /** Home's Continue watching row. */
    const val CONTINUE = "continue"
}

/**
 * One record shown (or opened) on a surface. [position] is the 0-based place in the list the user
 * sees.
 */
data class Impression(
    val subjectId: String,
    val subjectKind: RecordKind,
    val surface: String,
    val position: Int,
    val clicked: Boolean,
)

/**
 * Impression logging into the event log the recommendation engine learns from: what was shown
 * ("impression") and what was opened ("choice"), on which surface and at which position. Events go
 * out in batches: after a short quiet period, or at once when the batch fills up or something was
 * opened. Sending is best-effort; failures are dropped.
 */
class ImpressionLogger(
    private val scope: CoroutineScope,
    private val send: suspend (List<Impression>) -> Unit,
    private val enabled: () -> Boolean = { true },
    private val flushDelayMilliseconds: Long = FLUSH_DELAY_MILLISECONDS,
    private val batchLimit: Int = BATCH_LIMIT,
) {

    private var pending = mutableListOf<Impression>()
    private var flushJob: Job? = null

    /** Surface and record pairs already logged as shown, so re-renders don't count twice. */
    private val shownKeys = mutableSetOf<String>()

    /** Queues a "shown" impression, once per record and surface until [resetShown]. */
    @Synchronized
    fun recordShown(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        val key = "$surface|${subjectKind.rawValue}|$subjectId"
        if (!shownKeys.add(key)) {
            return
        }
        enqueue(Impression(subjectId, subjectKind, surface, position, clicked = false))
    }

    /** Logs that the user opened a record, sending the batch right away. */
    @Synchronized
    fun recordClicked(subjectId: String, subjectKind: RecordKind, surface: String, position: Int) {
        enqueue(Impression(subjectId, subjectKind, surface, position, clicked = true))
        flushNow()
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

    /**
     * Adds an impression to the batch and schedules or triggers the send. Impressions without an
     * id are dropped: one invalid id would fail the whole batch. Nothing is queued while [enabled]
     * says no (incognito).
     */
    private fun enqueue(impression: Impression) {
        if (!enabled()) {
            return
        }
        if (impression.subjectId.isEmpty()) {
            return
        }
        pending.add(impression)
        if (pending.size >= batchLimit) {
            flushNow()
            return
        }
        if (flushJob == null) {
            flushJob = scope.launch {
                delay(flushDelayMilliseconds)
                flushNow()
            }
        }
    }

    /** Sends every queued impression in batches of at most the batch limit. */
    @Synchronized
    fun flushNow() {
        flushJob?.cancel()
        flushJob = null
        if (pending.isEmpty()) {
            return
        }
        val batches = pending.chunked(batchLimit)
        pending = mutableListOf()
        scope.launch {
            for (batch in batches) {
                try {
                    send(batch)
                } catch (error: Exception) {
                    // Impressions are a ranking signal only; never surface failures.
                }
            }
        }
    }

    companion object {
        const val FLUSH_DELAY_MILLISECONDS = 3000L

        /** The most events the backend takes in one call, as the web sends them. */
        const val BATCH_LIMIT = 200

        /** Sends a batch to the backend's event log. */
        suspend fun sendToBackend(batch: List<Impression>) {
            val events = batch.map { impression -> eventOf(impression) }
            VeilApi.client.mutation(RecordEventsMutation(events)).execute().dataOrThrow()
        }

        /** An impression as the event log takes it: a choice when opened, else an impression. */
        private fun eventOf(impression: Impression): EventInput {
            var kind = EventKind.impression
            if (impression.clicked) {
                kind = EventKind.choice
            }
            return EventInput(
                kind = kind,
                subject = Optional.present(RecordRefInput(impression.subjectKind, impression.subjectId)),
                data = Optional.present(mapOf("surface" to impression.surface, "position" to impression.position)),
            )
        }
    }
}
