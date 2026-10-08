package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.LiveSearch
import com.playingwithclouds.veil.api.LiveSearchHit
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.ForgetSearchMutation
import com.playingwithclouds.veil.graphql.RecordSearchMutation
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.graphql.SearchGalleriesQuery
import com.playingwithclouds.veil.graphql.SearchPerformersQuery
import com.playingwithclouds.veil.graphql.SearchPluginsQuery
import com.playingwithclouds.veil.graphql.SearchScenesQuery
import com.playingwithclouds.veil.graphql.SearchStudiosQuery
import com.playingwithclouds.veil.graphql.SearchSuggestionsQuery
import kotlinx.coroutines.flow.Flow

/** One entry of the search box's suggestion list. */
data class SearchSuggestion(
    val kind: String,
    val text: String,
    val entityId: String?,
    val imageUrl: String?,
    val detail: String?,
)

/** A site that can be searched. */
data class SearchSite(
    val name: String,
    val displayName: String?,
    val iconUrl: String?,
    val domains: List<String>,
    val capabilities: List<String> = emptyList(),
) {

    /** The name to show: the plugin's display name, else its identifier. */
    val label: String
        get() = displayName ?: name

    /** Whether the site lists scenes. */
    val listsScenes: Boolean
        get() = capabilities.contains("scene:list")

    /** Whether the site lists galleries. */
    val listsGalleries: Boolean
        get() = capabilities.contains("gallery:list")
}

/** A result of the live search that is not stored yet, as a card. */
sealed interface LiveResultItem {

    /** The stored record's id, so the item opens without scraping first. */
    val recordId: String

    /** A scene result. */
    data class SceneResult(val scene: SceneSummary) : LiveResultItem {
        override val recordId: String
            get() = scene.id
    }

    /** A gallery result. */
    data class GalleryResult(val gallery: GallerySummary) : LiveResultItem {
        override val recordId: String
            get() = gallery.id
    }
}

/** Keyword search over the library and the source plugins. */
object SearchRepository {

    /** Suggestions for a partially typed query; recent searches and taste-based picks when empty. */
    suspend fun suggestions(query: String, limit: Int): List<SearchSuggestion> {
        val data = VeilApi.client.query(SearchSuggestionsQuery(query, Optional.present(limit))).execute().dataOrThrow()
        return data.searchSuggestions.map { suggestion ->
            SearchSuggestion(suggestion.kind.rawValue, suggestion.text, suggestion.entityId, suggestion.imageUrl, suggestion.detail)
        }
    }

    /** Remembers a submitted search as a recent one, except in incognito. */
    suspend fun recordSearch(query: String) {
        if (PrivacyPreferences.incognito.value) {
            return
        }
        VeilApi.client.mutation(RecordSearchMutation(query)).execute().dataOrThrow()
    }

    /** Drops a query from the recent searches. */
    suspend fun forgetSearch(query: String) {
        VeilApi.client.mutation(ForgetSearchMutation(query)).execute().dataOrThrow()
    }

    /** Stored scenes matching the query, restricted to the given sites (empty = all). */
    suspend fun searchScenes(query: String, sources: List<String>, limit: Int, offset: Int): List<SceneSummary> {
        val searchQuery = SearchScenesQuery(
            search = Optional.present(query),
            sources = sources.orAbsentIfEmpty(),
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(searchQuery).execute().dataOrThrow()
        return data.scenes.map { scene -> scene.sceneCardFields.toSummary() }
    }

    /** Stored galleries matching the query, restricted to the given sites (empty = all). */
    suspend fun searchGalleries(query: String, sources: List<String>, limit: Int, offset: Int): List<GallerySummary> {
        val searchQuery = SearchGalleriesQuery(
            search = Optional.present(query),
            sources = sources.orAbsentIfEmpty(),
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(searchQuery).execute().dataOrThrow()
        return data.galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
    }

    /** Performers whose name matches the query. */
    suspend fun searchPerformers(query: String, limit: Int): List<PerformerSummary> {
        val data = VeilApi.client.query(SearchPerformersQuery(Optional.present(query), Optional.present(limit))).execute().dataOrThrow()
        return data.performers.map { performer ->
            PerformerSummary(performer.id, performer.name, performer.imagePath, null, performer.sceneCount, false)
        }
    }

    /** Studios whose name matches the query. */
    suspend fun searchStudios(query: String, limit: Int): List<StudioSummary> {
        val data = VeilApi.client.query(SearchStudiosQuery(Optional.present(query), Optional.present(limit))).execute().dataOrThrow()
        return data.studios.map { studio -> StudioSummary(studio.id, studio.name, studio.imagePath, studio.sceneCount) }
    }

    /** The installed sites that can list content, i.e. what the source filter offers. */
    suspend fun searchSites(): List<SearchSite> {
        val data = VeilApi.client.query(SearchPluginsQuery()).execute().dataOrThrow()
        return data.plugins
            .filter { plugin -> plugin.available && listsContent(plugin.capabilities) }
            .map { plugin -> SearchSite(plugin.name, plugin.displayName, plugin.iconUrl, plugin.domains, plugin.capabilities) }
    }

    /** Every installed plugin as a site, for labeling a scene with where it came from. */
    suspend fun installedSites(): List<SearchSite> {
        val data = VeilApi.client.query(SearchPluginsQuery()).execute().dataOrThrow()
        return data.plugins.map { plugin -> SearchSite(plugin.name, plugin.displayName, plugin.iconUrl, plugin.domains) }
    }

    /** Whether a plugin lists scenes or galleries (a searchable site, not a resolver). */
    private fun listsContent(capabilities: List<String>): Boolean {
        return capabilities.contains("scene:list") || capabilities.contains("gallery:list")
    }

    /** Streams library matches first, then each source plugin's results. */
    fun liveSearch(query: String, sources: List<String>, offset: Int? = null, limit: Int? = null): Flow<LiveSearchHit> {
        return LiveSearch.results(query, sources, offset, limit)
    }

    /**
     * Turns a live search hit into a card. Null for hits the backend did not store (no record id)
     * and for media types other than scenes and galleries.
     */
    fun toLiveResult(hit: LiveSearchHit): LiveResultItem? {
        val recordId = hit.recordId ?: return null
        if (hit.mediaType == "scene") {
            val scene = SceneSummary(
                id = recordId,
                title = hit.title,
                sourceUrl = hit.sourceUrl,
                posterPath = hit.posterPath,
                previewVideo = hit.previewVideo,
                previewImages = hit.previewImages,
                durationSeconds = null,
                date = hit.date,
                studio = null,
                performers = emptyList(),
                tags = emptyList(),
            )
            return LiveResultItem.SceneResult(scene)
        }
        if (hit.mediaType == "gallery") {
            return LiveResultItem.GalleryResult(GallerySummary(recordId, hit.title, hit.posterPath, 0, hit.sourceUrl))
        }
        return null
    }
}
