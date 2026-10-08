package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.RecordImpressionsMutation
import com.playingwithclouds.veil.graphql.type.ImpressionInput
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** One shown (or opened) recommendation, as the recordImpressions mutation takes it. */
data class Impression(
    val mediaId: String,
    val source: String?,
    val surface: String,
    val position: Int?,
    val clicked: Boolean,
)

/**
 * Recommendation impression logging: what was shown and what was opened, which feeds the
 * engine's fatigue penalty and "shown but not clicked" signal. Impressions go out in batches:
 * after a short quiet period, or at once when the batch fills up or something was opened.
 * Sending is best-effort; failures are dropped.
 */
class ImpressionLogger(
    private val scope: CoroutineScope,
    private val send: suspend (List<Impression>) -> Unit,
    private val flushDelayMilliseconds: Long = FLUSH_DELAY_MILLISECONDS,
    private val batchLimit: Int = BATCH_LIMIT,
) {

    private var pending = mutableListOf<Impression>()
    private var flushJob: Job? = null

    /** Surface and media id pairs already logged as shown, so re-renders don't count twice. */
    private val shownKeys = mutableSetOf<String>()

    /** Queues a "shown" impression, once per scene and surface until [resetShown]. */
    @Synchronized
    fun recordShown(mediaId: String, source: String?, surface: String, position: Int?) {
        val key = "$surface|$mediaId"
        if (!shownKeys.add(key)) {
            return
        }
        enqueue(Impression(mediaId, source, surface, position, clicked = false))
    }

    /** Logs that the user opened a recommendation, sending the batch right away. */
    @Synchronized
    fun recordClicked(mediaId: String, source: String?, surface: String, position: Int?) {
        enqueue(Impression(mediaId, source, surface, position, clicked = true))
        flushNow()
    }

    /** Forgets which scenes were logged as shown, e.g. when the feed is reloaded. */
    @Synchronized
    fun resetShown() {
        shownKeys.clear()
    }

    /**
     * Adds an impression to the batch and schedules or triggers the send. Impressions without a
     * media id are dropped: one invalid id would fail the whole batch.
     */
    private fun enqueue(impression: Impression) {
        if (impression.mediaId.isEmpty()) {
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
        const val BATCH_LIMIT = 500

        /** Sends a batch to the backend. */
        suspend fun sendToBackend(batch: List<Impression>) {
            val inputs = batch.map { impression ->
                ImpressionInput(
                    mediaId = impression.mediaId,
                    source = Optional.presentIfNotNull(impression.source),
                    surface = Optional.present(impression.surface),
                    position = Optional.presentIfNotNull(impression.position),
                    clicked = Optional.present(impression.clicked),
                )
            }
            VeilApi.client.mutation(RecordImpressionsMutation(inputs)).execute().dataOrThrow()
        }
    }
}
