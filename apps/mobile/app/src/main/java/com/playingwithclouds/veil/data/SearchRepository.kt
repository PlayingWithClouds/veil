package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.RecentSearchesQuery
import com.playingwithclouds.veil.graphql.SearchResultsSubscription
import com.playingwithclouds.veil.graphql.type.EventKind
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

/**
 * Part of one site's answer to a search page. [sites] is how many sites the page asked; [final]
 * marks the site's last part, [exhausted] that it has no further pages.
 */
data class SearchBatch(
    val plugin: String,
    val sites: Int,
    val final: Boolean,
    val exhausted: Boolean,
    val videos: List<VideoSummary>,
    val performers: List<PerformerSummary>,
)

/** Searching the sites and what the user searched before (`apps/web/src/lib/search.ts`). */
object SearchRepository {

    /** One page (0 first) of a video search: each site's answer as it comes; completes when all answered. */
    fun page(query: String, page: Int): Flow<SearchBatch> {
        val subscription = SearchResultsSubscription(query, Optional.present(page))
        return VeilApi.client.subscription(subscription).toFlow().map { response ->
            val batch = response.dataOrThrow().searchResults
            SearchBatch(
                plugin = batch.plugin,
                sites = batch.sites,
                final = batch.final,
                exhausted = batch.exhausted,
                videos = batch.listed.videos.map { video -> video.videoCardFields.toSummary() },
                performers = batch.listed.performers.map { performer ->
                    val card = performer.performerCardFields
                    PerformerSummary(card.id, card.name, card.image?.url)
                },
            )
        }
    }

    /** What the user searched lately, newest first. */
    suspend fun recent(limit: Int): List<String> {
        return VeilApi.client.query(RecentSearchesQuery(Optional.present(limit))).execute().dataOrThrow().recentSearches
    }

    /** Logs a search, so it's among the recent ones and the engine learns from it. */
    fun logSearch(text: String) {
        AppServices.events.log(LoggedEvent(EventKind.search, null, null, mapOf("text" to text)), flush = true)
    }
}
