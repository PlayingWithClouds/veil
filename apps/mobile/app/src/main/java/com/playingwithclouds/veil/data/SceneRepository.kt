package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.graphql.AddBlockMutation
import com.playingwithclouds.veil.graphql.AddToCollectionMutation
import com.playingwithclouds.veil.graphql.AddToWatchlistMutation
import com.playingwithclouds.veil.graphql.AttachAlikeSourceMutation
import com.playingwithclouds.veil.graphql.CreateSceneMarkerMutation
import com.playingwithclouds.veil.graphql.DecrementOCountMutation
import com.playingwithclouds.veil.graphql.DeleteSceneMarkerMutation
import com.playingwithclouds.veil.graphql.DeleteUserRatingMutation
import com.playingwithclouds.veil.graphql.EnsureSceneStreamsMutation
import com.playingwithclouds.veil.graphql.FindAlikeSourcesQuery
import com.playingwithclouds.veil.graphql.IncrementOCountMutation
import com.playingwithclouds.veil.graphql.OCountQuery
import com.playingwithclouds.veil.graphql.QueueDownloadMutation
import com.playingwithclouds.veil.graphql.RelatedChangedSubscription
import com.playingwithclouds.veil.graphql.RemoveFromCollectionMutation
import com.playingwithclouds.veil.graphql.ResolveStreamQuery
import com.playingwithclouds.veil.graphql.SceneCollectionIdsQuery
import com.playingwithclouds.veil.graphql.SceneDetailQuery
import com.playingwithclouds.veil.graphql.SceneMarkersQuery
import com.playingwithclouds.veil.graphql.SceneOnWatchlistQuery
import com.playingwithclouds.veil.graphql.SceneRelatedQuery
import com.playingwithclouds.veil.graphql.StreamsChangedSubscription
import com.playingwithclouds.veil.graphql.UpsertUserRatingMutation
import com.playingwithclouds.veil.graphql.UpsertWatchHistoryMutation
import com.playingwithclouds.veil.graphql.UserRatingQuery
import com.playingwithclouds.veil.graphql.WatchHistoryEntryQuery
import com.playingwithclouds.veil.graphql.RemoveFromWatchlistMutation
import com.playingwithclouds.veil.graphql.type.UpsertUserRatingInput
import com.playingwithclouds.veil.graphql.type.UpsertWatchHistoryInput
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.retryWhen

/** A scene's full detail page data. */
data class SceneDetail(
    val id: String,
    val title: String,
    val sourceUrl: String,
    val details: String?,
    val date: String?,
    val durationSeconds: Int?,
    val viewCount: Int,
    val posterPath: String?,
    val studio: StudioInfo?,
    val performers: List<EntityRef>,
    val tags: List<EntityRef>,
)

/** The studio line of a scene page. */
data class StudioInfo(val id: String, val name: String, val imagePath: String?, val sceneCount: Int)

/** A marker at a point or span of a scene. */
data class SceneMarker(
    val id: String,
    val tagName: String?,
    val label: String?,
    val seconds: Double,
    val endSeconds: Double?,
    val personal: Boolean,
) {

    /** What the marker says: tag and label joined, or whichever is set. */
    val title: String
        get() {
            if (tagName != null && label != null) {
                return "$tagName - $label"
            }
            return tagName ?: label.orEmpty()
        }
}

/** Where playback of a scene last stopped. */
data class ResumePoint(val progressSeconds: Int, val durationSeconds: Int?, val completed: Boolean)

/** A like or dislike, stored on the 1-10 rating scale. */
enum class Verdict {
    UP,
    DOWN,
    ;

    companion object {
        /** Thumbs map to the extremes so the taste signal (centered at 5) gets a strong pull. */
        const val RATING_UP = 10.0
        const val RATING_DOWN = 1.0
        private const val NEUTRAL_RATING = 5.0

        /** Interprets a stored numeric rating as a thumbs verdict. */
        fun fromRating(rating: Double?): Verdict? {
            if (rating == null) {
                return null
            }
            if (rating > NEUTRAL_RATING) {
                return UP
            }
            if (rating < NEUTRAL_RATING) {
                return DOWN
            }
            return null
        }
    }
}

/** A copy of the scene found on another site. */
data class AlikeCandidate(
    val title: String,
    val plugin: String,
    val sourceUrl: String,
    val posterUrl: String?,
    val durationSeconds: Int?,
    val matchScore: Double,
)

/** Everything the scene page reads and writes. */
object SceneRepository {

    /** Fraction of the runtime after which a watch counts as completed. */
    const val COMPLETE_FRACTION = 0.9

    private const val SUBSCRIPTION_RETRY_DELAY_MILLISECONDS = 2000L
    private const val SUBSCRIPTION_MAX_RETRIES = 5L

    /** Loads a scene and its stored playback sources; the scene is null when it does not exist. */
    suspend fun load(id: String): Pair<SceneDetail?, List<StreamOption>> {
        return VeilApi.client.query(SceneDetailQuery(id)).execute().dataOrThrow().toSceneAndStreams()
    }


    /**
     * Populates a stub scene's sources and details on demand (a discovered-but-never-visited scene
     * has none). Returns the freshly ingested, ranked streams.
     */
    suspend fun ensureStreams(sceneId: String): List<StreamOption> {
        val data = VeilApi.client.mutation(EnsureSceneStreamsMutation(sceneId)).execute().dataOrThrow()
        return data.ensureSceneStreams.map { stream -> stream.streamFields.toOption() }
    }

    /** Live updates of a scene's sources as plugins resolve them. */
    fun streamsChanged(sceneId: String): Flow<List<StreamOption>> {
        return VeilApi.client.subscription(StreamsChangedSubscription(sceneId))
            .toFlow()
            .map { response -> response.dataOrThrow().streamsChanged.map { stream -> stream.streamFields.toOption() } }
            .retryWhen { _, attempt -> retryAfterDelay(attempt) }
    }

    /** The scene's related scenes now, then again whenever a visit links more. */
    fun relatedChanged(sceneId: String): Flow<List<SceneSummary>> {
        return VeilApi.client.subscription(RelatedChangedSubscription(sceneId))
            .toFlow()
            .map { response -> response.dataOrThrow().relatedChanged.map { scene -> scene.sceneCardFields.toSummary() } }
            .retryWhen { _, attempt -> retryAfterDelay(attempt) }
    }

    /** Waits, then reports whether a dropped subscription should be opened again. */
    private suspend fun retryAfterDelay(attempt: Long): Boolean {
        delay(SUBSCRIPTION_RETRY_DELAY_MILLISECONDS)
        return attempt < SUBSCRIPTION_MAX_RETRIES
    }

    /** The scene's related scenes as stored so far. */
    suspend fun related(sceneId: String, limit: Int): List<SceneSummary> {
        val data = VeilApi.client.query(SceneRelatedQuery(sceneId, Optional.present(limit))).execute().dataOrThrow()
        val related = data.scene?.related ?: return emptyList()
        return related.map { scene -> scene.sceneCardFields.toSummary() }
    }

    /** Resolves a source to a playable URL (rebased onto the active server) plus the headers it needs. */
    suspend fun resolve(stream: StreamOption): PlayableStream {
        val query = ResolveStreamQuery(stream.url, stream.pluginName.orAbsent())
        val resolved = VeilApi.client.query(query).execute().dataOrThrow().stream
            ?: throw IllegalStateException("No playable stream")
        return PlayableStream(
            url = ServerSettings.backendUrl(resolved.url),
            mimeType = resolved.mimeType,
            headers = resolved.headers.associate { header -> header.name to header.value },
        )
    }

    /** Global markers plus the user's personal markers. */
    suspend fun markers(sceneId: String): List<SceneMarker> {
        val data = VeilApi.client.query(SceneMarkersQuery(sceneId)).execute().dataOrThrow()
        return data.sceneMarkers
            .map { marker ->
                SceneMarker(marker.id, marker.tag?.name, marker.label, marker.seconds, marker.endSeconds, marker.personal)
            }
            .sortedBy { marker -> marker.seconds }
    }

    /** Creates a personal marker; at least one of tag name and label must be set. */
    suspend fun createMarker(sceneId: String, seconds: Double, tagName: String?, label: String?) {
        val mutation = CreateSceneMarkerMutation(sceneId, seconds, tagName.orAbsent(), label.orAbsent())
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Deletes a personal marker. */
    suspend fun deleteMarker(markerId: String) {
        VeilApi.client.mutation(DeleteSceneMarkerMutation(markerId)).execute().dataOrThrow()
    }

    /** The user's like or dislike of a scene, if any. */
    suspend fun verdict(sceneId: String): Verdict? {
        val data = VeilApi.client.query(UserRatingQuery(sceneId)).execute().dataOrThrow()
        return Verdict.fromRating(data.userRating?.rating)
    }

    /** Stores a like or dislike. */
    suspend fun setVerdict(sceneId: String, verdict: Verdict) {
        var rating = Verdict.RATING_UP
        if (verdict == Verdict.DOWN) {
            rating = Verdict.RATING_DOWN
        }
        val mutation = UpsertUserRatingMutation(UpsertUserRatingInput(media = sceneId, rating = rating))
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Removes the like or dislike. */
    suspend fun clearVerdict(sceneId: String) {
        VeilApi.client.mutation(DeleteUserRatingMutation(sceneId)).execute().dataOrThrow()
    }

    /** How often the user marked the scene with the O-counter. */
    suspend fun oCount(sceneId: String): Int {
        return VeilApi.client.query(OCountQuery(sceneId)).execute().dataOrThrow().oCount
    }

    /** Adds an O-counter event; returns the new total. */
    suspend fun incrementOCount(sceneId: String): Int {
        return VeilApi.client.mutation(IncrementOCountMutation(sceneId)).execute().dataOrThrow().incrementOCount
    }

    /** Removes the latest O-counter event; returns the new total. */
    suspend fun decrementOCount(sceneId: String): Int {
        return VeilApi.client.mutation(DecrementOCountMutation(sceneId)).execute().dataOrThrow().decrementOCount
    }

    /** Whether the scene is on the watchlist. */
    suspend fun isOnWatchlist(sceneId: String): Boolean {
        val data = VeilApi.client.query(SceneOnWatchlistQuery()).execute().dataOrThrow()
        return data.watchlist.any { item -> item.media == sceneId }
    }

    /** Puts the scene on or takes it off the watchlist. */
    suspend fun setOnWatchlist(sceneId: String, onWatchlist: Boolean) {
        if (onWatchlist) {
            VeilApi.client.mutation(AddToWatchlistMutation(sceneId)).execute().dataOrThrow()
            return
        }
        VeilApi.client.mutation(RemoveFromWatchlistMutation(sceneId)).execute().dataOrThrow()
    }

    /** Ids of the user's own collections that contain the scene. */
    suspend fun collectionIds(sceneId: String): Set<String> {
        val data = VeilApi.client.query(SceneCollectionIdsQuery(sceneId)).execute().dataOrThrow()
        return data.collectionIdsForMedia.toSet()
    }

    /** Adds the scene to a collection or removes it from there. */
    suspend fun setInCollection(collectionId: String, sceneId: String, included: Boolean) {
        if (included) {
            VeilApi.client.mutation(AddToCollectionMutation(collectionId, sceneId)).execute().dataOrThrow()
            return
        }
        VeilApi.client.mutation(RemoveFromCollectionMutation(collectionId, sceneId)).execute().dataOrThrow()
    }

    /** Queues a source for download on the backend; returns the job id. */
    suspend fun queueDownload(stream: StreamOption, title: String): String {
        val mutation = QueueDownloadMutation(stream.url, title, stream.pluginName.orAbsent())
        return VeilApi.client.mutation(mutation).execute().dataOrThrow().queueDownload
    }

    /** Blocks a tag, performer or studio from recommendations. */
    suspend fun block(kind: String, targetId: String, label: String) {
        VeilApi.client.mutation(AddBlockMutation(kind, targetId, Optional.present(label))).execute().dataOrThrow()
    }

    /** Alternate copies of the scene on other sites, best match first. */
    suspend fun findAlikeSources(sceneId: String): List<AlikeCandidate> {
        val data = VeilApi.client.query(FindAlikeSourcesQuery(sceneId)).execute().dataOrThrow()
        return data.findAlikeSources.map { candidate ->
            AlikeCandidate(
                title = candidate.title,
                plugin = candidate.plugin,
                sourceUrl = candidate.sourceUrl,
                posterUrl = candidate.posterUrl,
                durationSeconds = candidate.durationSeconds,
                matchScore = candidate.matchScore,
            )
        }
    }

    /** Attaches a chosen alternate source to the scene; returns the scene's ranked streams. */
    suspend fun attachAlikeSource(sceneId: String, candidate: AlikeCandidate): List<StreamOption> {
        val data = VeilApi.client.mutation(AttachAlikeSourceMutation(sceneId, candidate.plugin, candidate.sourceUrl))
            .execute()
            .dataOrThrow()
        return data.attachAlikeSource.map { stream -> stream.streamFields.toOption() }
    }

    /** Where the user stopped watching the scene, if they started it. */
    suspend fun resumePoint(sceneId: String): ResumePoint? {
        val entry = VeilApi.client.query(WatchHistoryEntryQuery(sceneId)).execute().dataOrThrow().watchHistoryEntry
            ?: return null
        return ResumePoint(entry.progressSeconds, entry.durationSeconds, entry.completed)
    }

    /** Records playback progress for resume; whether the watch is completed follows from the position. */
    suspend fun saveProgress(sceneId: String, positionSeconds: Double, durationSeconds: Double?) {
        var completed = false
        if (durationSeconds != null && durationSeconds > 0) {
            completed = positionSeconds >= durationSeconds * COMPLETE_FRACTION
        }
        val input = UpsertWatchHistoryInput(
            media = sceneId,
            progressSeconds = positionSeconds.toInt(),
            durationSeconds = Optional.presentIfNotNull(durationSeconds?.toInt()),
            completed = Optional.present(completed),
        )
        VeilApi.client.mutation(UpsertWatchHistoryMutation(input)).execute().dataOrThrow()
    }
}

/** Maps the scene page response to the scene (null when missing) and its sources. */
fun SceneDetailQuery.Data.toSceneAndStreams(): Pair<SceneDetail?, List<StreamOption>> {
    val streams = mediaStreams.map { stream -> stream.streamFields.toOption() }
    val sceneData = scene ?: return Pair(null, streams)
    var studio: StudioInfo? = null
    val studioData = sceneData.studio
    if (studioData != null) {
        studio = StudioInfo(studioData.id, studioData.name, studioData.imagePath, studioData.sceneCount)
    }
    val detail = SceneDetail(
        id = sceneData.id,
        title = sceneData.title,
        sourceUrl = sceneData.sourceUrl,
        details = sceneData.details,
        date = sceneData.date,
        durationSeconds = sceneData.durationSeconds,
        viewCount = sceneData.viewCount,
        posterPath = sceneData.posterPath,
        studio = studio,
        performers = sceneData.performers.map { performer -> EntityRef(performer.id, performer.name, performer.imagePath) },
        tags = sceneData.tags.map { tag -> EntityRef(tag.id, tag.name, null) },
    )
    return Pair(detail, streams)
}
