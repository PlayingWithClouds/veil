package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.RecentSearchesQuery
import com.playingwithclouds.veil.graphql.SearchResultsSubscription
import com.playingwithclouds.veil.graphql.TagsNamedQuery
import com.playingwithclouds.veil.graphql.VideoSitesQuery
import com.playingwithclouds.veil.graphql.VideoTagCountsQuery
import com.playingwithclouds.veil.graphql.type.EventKind
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.toList

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

    /** Capability of the plugins that list videos. */
    private const val VIDEO_LIST_CAPABILITY = "scene:list"

    /** Tags matching "pov" read to find the one named exactly that. */
    private const val POV_TAG_LOOKUP_LIMIT = 20

    /**
     * One page (0 first) of a video search: each site's answer as it comes; completes when all
     * answered. The filter narrows the results, and its sites are the ones asked.
     */
    fun page(query: String, page: Int, filter: SearchFilter, content: String? = null): Flow<SearchBatch> {
        val subscription = SearchResultsSubscription(query, Optional.present(page), Optional.present(filter.toInput()), Optional.presentIfNotNull(content))
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

    /**
     * Every video one unfiltered search page (0 first) finds, once all sites answered, as the web's
     * `collectSearchPage`; [content] asks only the sites carrying that kind ("cam-shows").
     */
    suspend fun collectPage(query: String, page: Int, content: String? = null): List<VideoSummary> {
        return page(query, page, SearchFilter(), content).toList().flatMap { batch -> batch.videos }
    }

    /**
     * The sites the filter offers, by name: each site of a plugin serving several ("plugin/host"),
     * the other plugins listing videos (or what [capability] names, e.g. "gallery:list") whole. Only enabled plugins that can run here; sites needing a
     * Cloudflare solver the phone doesn't have are left out.
     */
    suspend fun siteChoices(capability: String = VIDEO_LIST_CAPABILITY): List<SiteChoice> {
        val data = VeilApi.client.query(VideoSitesQuery()).execute().dataOrThrow()
        val listers = data.plugins.filter { plugin -> plugin.enabled && plugin.available && capability in plugin.capabilities }
        return listers.flatMap { plugin ->
            if (plugin.sites.isEmpty()) {
                return@flatMap listOf(SiteChoice(plugin.name, plugin.displayName ?: plugin.name, plugin.iconUrl))
            }
            plugin.sites.map { site -> SiteChoice("${plugin.name}/${site.host}", site.name, site.iconUrl) }
        }.sortedBy { choice -> choice.label.lowercase() }
    }

    /** The tags most common among the videos matching the filter, most common first, for hiding. */
    suspend fun tagCounts(filter: SearchFilter, limit: Int): List<Pair<TagSummary, Int>> {
        val query = VideoTagCountsQuery(Optional.present(filter.toInput()), Optional.present(limit))
        return VeilApi.client.query(query).execute().dataOrThrow().videoTagCounts.map { entry ->
            TagSummary(entry.tag.id, entry.tag.name) to entry.count
        }
    }

    /**
     * The library's POV tag for the Camera filter: the tag named "POV" (any case); null while no site
     * has tagged anything POV. Once the scene model labels videos, the filter reads its labels instead.
     */
    suspend fun povTag(): TagSummary? {
        val tags = VeilApi.client.query(TagsNamedQuery(Optional.present("pov"), Optional.present(POV_TAG_LOOKUP_LIMIT))).execute().dataOrThrow().tags
        val tag = tags.firstOrNull { candidate -> candidate.name.trim().equals("pov", ignoreCase = true) } ?: return null
        return TagSummary(tag.id, tag.name)
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
