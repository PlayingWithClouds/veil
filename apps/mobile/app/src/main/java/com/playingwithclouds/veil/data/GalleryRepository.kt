package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.BrowseGalleriesMutation
import com.playingwithclouds.veil.graphql.FilteredGalleriesQuery
import com.playingwithclouds.veil.graphql.GalleryPluginsQuery
import com.playingwithclouds.veil.graphql.GallerySearchResultsSubscription
import com.playingwithclouds.veil.graphql.GalleryTagCountsQuery
import kotlinx.coroutines.flow.toList
import com.playingwithclouds.veil.graphql.OpenGalleryMutation
import com.playingwithclouds.veil.graphql.RateMutation
import com.playingwithclouds.veil.graphql.RecentlyViewedGalleriesQuery
import com.playingwithclouds.veil.graphql.RecommendedGalleriesQuery
import com.playingwithclouds.veil.graphql.RelatedGalleriesQuery
import com.playingwithclouds.veil.graphql.SetFavoriteMutation
import com.playingwithclouds.veil.graphql.type.FilterInput
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput

/** One photo of a gallery; its size is known for most, for laying out the grid. */
data class GalleryImage(val id: String, val url: String, val width: Int?, val height: Int?)

/** Everything the gallery page shows; [rating] is 0..10, [videoId] the video it belongs to. */
data class GalleryDetail(
    val summary: GallerySummary,
    val details: String?,
    val date: String?,
    val studio: PerformerSummary?,
    val performers: List<PerformerSummary>,
    val tags: List<TagSummary>,
    val images: List<GalleryImage>,
    val videoId: String?,
    val favorite: Boolean,
    val rating: Double?,
)

/** Galleries: the overview's rows, the endless feed and one gallery (`apps/web/src/lib/galleries.ts`). */
object GalleryRepository {

    /** Capability of the plugins that list galleries. */
    private const val GALLERY_LIST_CAPABILITY = "gallery:list"

    /** A page of galleries recommended to the user. */
    suspend fun recommended(limit: Int, offset: Int): List<GallerySummary> {
        val query = RecommendedGalleriesQuery(Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().recommendations.mapNotNull { match ->
            match.record?.onGallery?.galleryCardFields?.toSummary()
        }
    }

    /** Galleries the user opened, most recently viewed first. */
    suspend fun recentlyViewed(limit: Int): List<GallerySummary> {
        val data = VeilApi.client.query(RecentlyViewedGalleriesQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.history.mapNotNull { state -> state.record?.onGallery?.galleryCardFields?.toSummary() }
    }

    /** The favourite galleries, most recently added first. */
    suspend fun favorites(limit: Int): List<GallerySummary> {
        return filtered(FilterInput(favorite = Optional.present(true)), limit, 0)
    }

    /** The library's galleries, most recently added first. */
    suspend fun listing(limit: Int, offset: Int): List<GallerySummary> {
        return filtered(FilterInput(), limit, offset)
    }

    /** Names of the enabled gallery sites that can run here, as the web's `listerNames('GALLERY')`. */
    suspend fun siteNames(): List<String> {
        val plugins = VeilApi.client.query(GalleryPluginsQuery()).execute().dataOrThrow().plugins
        return plugins
            .filter { plugin -> plugin.enabled && plugin.available && GALLERY_LIST_CAPABILITY in plugin.capabilities }
            .map { plugin -> plugin.name }
    }

    /**
     * Every gallery one page (0 first) of a search of the gallery sites finds, once all answered;
     * the filter narrows them, and its sites are the ones asked.
     */
    suspend fun searchPage(query: String, page: Int, filter: SearchFilter): List<GallerySummary> {
        val subscription = GallerySearchResultsSubscription(query, Optional.present(page), Optional.present(filter.toInput()))
        val batches = VeilApi.client.subscription(subscription).toFlow().toList()
        return batches.flatMap { response ->
            response.dataOrThrow().searchResults.listed.galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
        }
    }

    /** The gallery sites the search filter offers. */
    suspend fun siteChoices(): List<SiteChoice> {
        return SearchRepository.siteChoices(GALLERY_LIST_CAPABILITY)
    }

    /** The tags most common among the galleries matching the filter, most common first, for hiding. */
    suspend fun tagCounts(filter: SearchFilter, limit: Int): List<Pair<TagSummary, Int>> {
        val query = GalleryTagCountsQuery(Optional.present(filter.toInput()), Optional.present(limit))
        return VeilApi.client.query(query).execute().dataOrThrow().mediaTagCounts.map { entry ->
            TagSummary(entry.tag.id, entry.tag.name) to entry.count
        }
    }

    /** One page of a gallery site's catalog; the backend stores what it lists. */
    suspend fun browse(plugin: String, limit: Int, offset: Int): List<GallerySummary> {
        val mutation = BrowseGalleriesMutation(plugin, Optional.present(limit), Optional.present(offset))
        val listed = VeilApi.client.mutation(mutation).execute().dataOrThrow().browse
        return listed.galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
    }

    /** Opens a gallery: reads its pages (they list the images); null when it's gone. */
    suspend fun open(id: String): GalleryDetail? {
        val gallery = VeilApi.client.mutation(OpenGalleryMutation(id)).execute().dataOrThrow().fetchGallery ?: return null
        return GalleryDetail(
            summary = gallery.galleryCardFields.toSummary(),
            details = gallery.details,
            date = gallery.date,
            studio = gallery.studio?.let { studio -> PerformerSummary(studio.id, studio.name, null) },
            performers = gallery.performers.map { performer -> PerformerSummary(performer.id, performer.name, performer.image?.url) },
            tags = gallery.tags.map { tag -> TagSummary(tag.id, tag.name) },
            images = gallery.images.sortedBy { image -> image.position }.map { image ->
                GalleryImage(image.id, image.asset.url, image.width, image.height)
            },
            videoId = gallery.video?.id,
            favorite = gallery.userState.favorite,
            rating = gallery.userState.rating,
        )
    }

    /** Galleries like one, most alike first. */
    suspend fun related(id: String, limit: Int): List<GallerySummary> {
        val data = VeilApi.client.query(RelatedGalleriesQuery(id, Optional.present(limit))).execute().dataOrThrow()
        return data.related.mapNotNull { match -> match.record?.onGallery?.galleryCardFields?.toSummary() }
    }

    /** Marks a gallery as a favourite, or not. */
    suspend fun setFavorite(id: String, favorite: Boolean) {
        VeilApi.client.mutation(SetFavoriteMutation(galleryRef(id), favorite)).execute().dataOrThrow()
    }

    /** Rates a gallery 0..10; null clears the rating. */
    suspend fun rate(id: String, rating: Double?) {
        VeilApi.client.mutation(RateMutation(galleryRef(id), Optional.present(rating))).execute().dataOrThrow()
    }

    /** Galleries matching a filter, newest first. */
    private suspend fun filtered(filter: FilterInput, limit: Int, offset: Int): List<GallerySummary> {
        val query = FilteredGalleriesQuery(Optional.present(filter), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
    }

    /** A reference to a gallery, as the user-state mutations take it. */
    private fun galleryRef(id: String): RecordRefInput {
        return RecordRefInput(RecordKind.gallery, id)
    }
}
