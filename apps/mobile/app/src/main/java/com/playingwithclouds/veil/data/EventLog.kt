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
 * One event for the log the recommendation engine learns from (see the backend's
 * internal/recommend/events.go for what [data] each [kind] carries). It's about a record unless
 * it's a search, whose subject is null. [sessionId] ties the events of one playback together.
 */
data class LoggedEvent(
    val kind: EventKind,
    val subjectKind: RecordKind?,
    val subjectId: String?,
    val data: Map<String, Any> = emptyMap(),
    val sessionId: String? = null,
)

/**
 * The event log, sent as the web sends it: in batches after a short quiet period, at once when a
 * batch fills up or a caller asks to flush (a choice, leaving a video). Nothing is queued while
 * [enabled] says no (incognito). Sending is best-effort; failures are dropped.
 */
class EventLog(
    private val scope: CoroutineScope,
    private val send: suspend (List<LoggedEvent>) -> Unit,
    private val enabled: () -> Boolean = { true },
    private val flushDelayMilliseconds: Long = FLUSH_DELAY_MILLISECONDS,
    private val batchLimit: Int = BATCH_LIMIT,
) {

    private var pending = mutableListOf<LoggedEvent>()
    private var flushJob: Job? = null

    /**
     * Queues an event, sending the batch right away when [flush] is set. Events with an empty
     * subject id are dropped: one invalid id would fail the whole batch.
     */
    @Synchronized
    fun log(event: LoggedEvent, flush: Boolean = false) {
        if (!enabled() || event.subjectId?.isEmpty() == true) {
            return
        }
        pending.add(event)
        if (flush || pending.size >= batchLimit) {
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

    /** Sends every queued event in batches of at most the batch limit. */
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
                    // The log is a ranking and resume signal only; never surface failures.
                }
            }
        }
    }

    companion object {
        const val FLUSH_DELAY_MILLISECONDS = 3000L

        /** The most events the backend takes in one call, as the web sends them. */
        const val BATCH_LIMIT = 200

        /** Sends a batch to the backend's event log. */
        suspend fun sendToBackend(batch: List<LoggedEvent>) {
            val events = batch.map { event ->
                EventInput(
                    kind = event.kind,
                    subject = Optional.presentIfNotNull(subjectOf(event)),
                    data = Optional.present(event.data),
                    sessionId = Optional.presentIfNotNull(event.sessionId),
                )
            }
            VeilApi.client.mutation(RecordEventsMutation(events)).execute().dataOrThrow()
        }

        /** The record an event is about; null for a search. */
        private fun subjectOf(event: LoggedEvent): RecordRefInput? {
            val kind = event.subjectKind ?: return null
            val id = event.subjectId ?: return null
            return RecordRefInput(kind, id)
        }
    }
}
