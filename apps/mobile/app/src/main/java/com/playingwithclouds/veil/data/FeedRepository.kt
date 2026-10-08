package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.DeleteWatchHistoryMutation
import com.playingwithclouds.veil.graphql.DownloadedScenesQuery
import com.playingwithclouds.veil.graphql.HomeFeedQuery
import com.playingwithclouds.veil.graphql.HomeRowsQuery
import com.playingwithclouds.veil.graphql.HomeScenesQuery
import com.playingwithclouds.veil.graphql.MediaCardsQuery
import com.playingwithclouds.veil.graphql.RandomScenesQuery
import com.playingwithclouds.veil.graphql.RemoveFromWatchlistMutation
import com.playingwithclouds.veil.graphql.WatchHistoryQuery
import com.playingwithclouds.veil.graphql.WatchlistQuery

/** One row of the watch history. */
data class WatchHistoryEntry(
    val mediaId: String,
    val progressSeconds: Int,
    val durationSeconds: Int?,
    val completed: Boolean,
    val scene: SceneSummary?,
)

/** A titled row of scenes the recommender grouped by reason. */
data class SceneRow(val key: String, val title: String, val scenes: List<SceneSummary>)

/** A runtime window in seconds; a null bound is open. */
data class RuntimeWindow(val minSeconds: Int?, val maxSeconds: Int?) {

    /** Whether the window narrows anything. */
    val isOpen: Boolean
        get() = minSeconds == null && maxSeconds == null

    companion object {
        /** Every runtime. */
        val ANY = RuntimeWindow(null, null)
    }
}

/** A saved record resolved to the card data of its media id. */
data class MediaCard(val mediaId: String, val mediaType: String, val title: String, val posterPath: String?)

/** Home feed, history, watchlist and downloaded scenes. */
object FeedRepository {

    /**
     * One page of the ranked recommendation feed. Offset 0 keeps the ranking for 10 minutes unless
     * refresh is set; later offsets continue that ranking so pages don't overlap. Non-empty
     * [sources] keep only scenes from those plugins; a [runtime] window keeps scenes with a known
     * length inside it.
     */
    suspend fun recommendations(
        limit: Int,
        offset: Int,
        refresh: Boolean,
        sources: List<String>,
        runtime: RuntimeWindow = RuntimeWindow.ANY,
    ): List<RecommendedScene> {
        val query = HomeFeedQuery(
            Optional.present(limit),
            Optional.present(offset),
            Optional.present(refresh),
            sources.orAbsentIfEmpty(),
            runtime.minSeconds.orAbsent(),
            runtime.maxSeconds.orAbsent(),
        )
        return VeilApi.client.query(query).execute().dataOrThrow().toRecommendedScenes()
    }

    /** One page of scenes filed under a tag, optionally narrowed by sites and runtime; newest first. */
    suspend fun filteredScenes(
        limit: Int,
        offset: Int,
        tagId: String?,
        sources: List<String>,
        runtime: RuntimeWindow,
    ): List<SceneSummary> {
        val query = HomeScenesQuery(
            Optional.present(limit),
            Optional.present(offset),
            tagId.orAbsent(),
            sources.orAbsentIfEmpty(),
            runtime.minSeconds.orAbsent(),
            runtime.maxSeconds.orAbsent(),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.scenes.map { scene -> scene.sceneCardFields.toSummary() }
    }

    /** The recommender's titled rows ("Because you watched X", "More from Y"), strongest first. */
    suspend fun recommendedRows(rowLimit: Int, perRow: Int): List<SceneRow> {
        val data = VeilApi.client.query(HomeRowsQuery(Optional.present(rowLimit), Optional.present(perRow))).execute().dataOrThrow()
        return data.recommendedRows.map { row ->
            SceneRow(row.key, row.title, row.items.map { item -> item.scene.sceneCardFields.toSummary() })
        }
    }


    /** A random pick of stored scenes (blocklist applied). */
    suspend fun randomScenes(limit: Int): List<SceneSummary> {
        val data = VeilApi.client.query(RandomScenesQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.randomScenes.map { scene -> scene.sceneCardFields.toSummary() }
    }

    /** The watch history, most recently watched first. */
    suspend fun watchHistory(limit: Int, offset: Int): List<WatchHistoryEntry> {
        val query = WatchHistoryQuery(Optional.present(limit), Optional.present(offset))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.watchHistory.map { row ->
            WatchHistoryEntry(
                mediaId = row.media,
                progressSeconds = row.progressSeconds,
                durationSeconds = row.durationSeconds,
                completed = row.completed,
                scene = row.scene?.sceneCardFields?.toSummary(),
            )
        }
    }

    /** Removes a scene from the watch history. */
    suspend fun deleteWatchHistory(mediaId: String) {
        VeilApi.client.mutation(DeleteWatchHistoryMutation(mediaId)).execute().dataOrThrow()
    }

    /** The saved-for-later list as cards. */
    suspend fun watchlist(): List<MediaCard> {
        val ids = VeilApi.client.query(WatchlistQuery()).execute().dataOrThrow().watchlist.map { item -> item.media }
        if (ids.isEmpty()) {
            return emptyList()
        }
        val cards = VeilApi.client.query(MediaCardsQuery(ids)).execute().dataOrThrow().mediaCards
        return cards.map { card -> MediaCard(card.mediaId, card.mediaType, card.title, card.posterPath) }
    }

    /** Takes a scene off the watchlist. */
    suspend fun removeFromWatchlist(mediaId: String) {
        VeilApi.client.mutation(RemoveFromWatchlistMutation(mediaId)).execute().dataOrThrow()
    }

    /** Scenes with a completed download, newest first. */
    suspend fun downloadedScenes(): List<SceneSummary> {
        val data = VeilApi.client.query(DownloadedScenesQuery()).execute().dataOrThrow()
        return data.downloadedScenes.map { scene -> scene.sceneCardFields.toSummary() }
    }
}

/** Maps the feed response to scenes with the source and reason they were served with. */
fun HomeFeedQuery.Data.toRecommendedScenes(): List<RecommendedScene> {
    return recommendations.map { row ->
        RecommendedScene(row.scene.sceneCardFields.toSummary(), row.source, row.reason.text)
    }
}
