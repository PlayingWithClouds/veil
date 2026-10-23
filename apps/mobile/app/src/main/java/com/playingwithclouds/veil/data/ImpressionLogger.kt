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

/**
 * The surfaces impressions are logged under. The engine learns from [HOME], [ROW] and
 * [PREVIEW_FEED], whose sources are the candidate sources of a ranking; the others are logged for
 * analysis and never change a ranking.
 */
object ImpressionSurface {

    /** The Home feed. */
    const val HOME = "home"

    /** A titled shelf between the Home cards. */
    const val ROW = "row"

    /** The Shorts-style preview feed. */
    const val PREVIEW_FEED = "preview-feed"

    /** The related scenes of a scene page, inline and over the fullscreen video. */
    const val RELATED = "related"

    /** The scenes of a performer page. */
    const val PERFORMER = "performer"

    /** The scenes of a studio page. */
    const val STUDIO = "studio"

    /** The scenes of a tag page. */
    const val TAG = "tag"

    /** Search results. */
    const val SEARCH = "search"
}

/**
 * One shown (or opened) scene, as the recordImpressions mutation takes it. [position] is the
 * 0-based place in the list the user sees: the feed rank on Home, the index within a shelf or list
 * elsewhere.
 */
data class Impression(
    val mediaId: String,
    val source: String?,
    val surface: String,
    val position: Int,
    val clicked: Boolean,
)

/**
 * Impression logging: what was shown and what was opened, on which surface and at which position.
 * The engine learns from the recommendation surfaces (see [ImpressionSurface]): source priors,
 * position bias and the "shown but not clicked" fatigue. Impressions go out in batches:
 * after a short quiet period, or at once when the batch fills up or something was opened.
 * Sending is best-effort; failures are dropped.
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

    /** Surface and media id pairs already logged as shown, so re-renders don't count twice. */
    private val shownKeys = mutableSetOf<String>()

    /** Queues a "shown" impression, once per scene and surface until [resetShown]. */
    @Synchronized
    fun recordShown(mediaId: String, source: String?, surface: String, position: Int) {
        val key = "$surface|$mediaId"
        if (!shownKeys.add(key)) {
            return
        }
        enqueue(Impression(mediaId, source, surface, position, clicked = false))
    }

    /** Logs that the user opened a recommendation, sending the batch right away. */
    @Synchronized
    fun recordClicked(mediaId: String, source: String?, surface: String, position: Int) {
        enqueue(Impression(mediaId, source, surface, position, clicked = true))
        flushNow()
    }

    /**
     * Forgets which scenes were logged as shown on [surface], or on every surface without one,
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
     * Adds an impression to the batch and schedules or triggers the send. Impressions without a
     * media id are dropped: one invalid id would fail the whole batch. Nothing is queued while
     * [enabled] says no (incognito).
     */
    private fun enqueue(impression: Impression) {
        if (!enabled()) {
            return
        }
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
                    position = Optional.present(impression.position),
                    clicked = Optional.present(impression.clicked),
                )
            }
            VeilApi.client.mutation(RecordImpressionsMutation(inputs)).execute().dataOrThrow()
        }
    }
}
