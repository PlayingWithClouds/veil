package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.PreviewFeedQuery
import com.playingwithclouds.veil.graphql.TasteProfileQuery
import com.playingwithclouds.veil.graphql.fragment.TasteEntryFields

/** One entity of the taste dashboard; [affinity] is in -1..1, negative for what the user avoids. */
data class TasteEntry(val id: String, val name: String, val imagePath: String?, val affinity: Double)

/** What the recommender has learned: the entities it ranks up and down, per kind. */
data class TasteProfile(
    val signalCount: Int,
    val tags: List<TasteEntry>,
    val performers: List<TasteEntry>,
    val studios: List<TasteEntry>,
    val sites: List<TasteEntry>,
)

/** The vertical preview feed and the taste dashboard. */
object DiscoveryRepository {

    /**
     * One page of the ranked feed without the scenes carrying [excludedTagIds]. The backend keeps
     * the ranking for a while, so later offsets continue the first page.
     */
    suspend fun previewFeed(limit: Int, offset: Int, refresh: Boolean, excludedTagIds: List<String>): List<RecommendedScene> {
        val query = PreviewFeedQuery(
            Optional.present(limit),
            Optional.present(offset),
            Optional.present(refresh),
            excludedTagIds.orAbsentIfEmpty(),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.recommendations.map { row -> RecommendedScene(row.scene.sceneCardFields.toSummary(), row.source, "") }
    }

    /** The taste profile, with at most [limit] liked and [limit] disliked entries per kind. */
    suspend fun tasteProfile(limit: Int): TasteProfile {
        val profile = VeilApi.client.query(TasteProfileQuery(Optional.present(limit))).execute().dataOrThrow().tasteProfile
        return TasteProfile(
            signalCount = profile.signalCount,
            tags = profile.tags.map { entry -> entry.tasteEntryFields.toEntry() },
            performers = profile.performers.map { entry -> entry.tasteEntryFields.toEntry() },
            studios = profile.studios.map { entry -> entry.tasteEntryFields.toEntry() },
            sites = profile.sites.map { entry -> entry.tasteEntryFields.toEntry() },
        )
    }
}

/** Maps a taste fragment to the dashboard model. */
private fun TasteEntryFields.toEntry(): TasteEntry {
    return TasteEntry(id, name, imagePath, affinity)
}
