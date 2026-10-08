package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.EnsureGalleryImagesMutation
import com.playingwithclouds.veil.graphql.GalleryDetailQuery
import com.playingwithclouds.veil.graphql.PluginCategoriesQuery
import com.playingwithclouds.veil.graphql.fragment.GalleryDetailFields

/** One image of a gallery. */
data class GalleryImage(val id: String, val filePath: String, val title: String?, val width: Int?, val height: Int?)

/** A gallery with its images. */
data class GalleryDetail(
    val id: String,
    val title: String,
    val details: String?,
    val date: String?,
    val imageCount: Int,
    val tags: List<EntityRef>,
    val images: List<GalleryImage>,
)

/** A browsable category of a gallery site, with a preview thumbnail. */
data class GalleryCategory(val id: String, val name: String, val poster: String?)

/** Image sets. */
object GalleryRepository {

    /** The plugin that supplies galleries. */
    const val GALLERY_PLUGIN = "pornpics"

    /** The gallery site's category index. */
    suspend fun categories(): List<GalleryCategory> {
        val data = VeilApi.client.query(PluginCategoriesQuery(GALLERY_PLUGIN, Optional.Absent)).execute().dataOrThrow()
        return data.pluginCategories.map { category -> GalleryCategory(category.id, category.name, category.poster) }
    }

    /** A stored gallery, or null when it does not exist. */
    suspend fun gallery(id: String): GalleryDetail? {
        return VeilApi.client.query(GalleryDetailQuery(id)).execute().dataOrThrow().gallery?.galleryDetailFields?.toDetail()
    }

    /**
     * Records a gallery visit: a stub gallery (from search/browse) has its page fetched from its
     * site so the images land. Returns the gallery as stored afterwards.
     */
    suspend fun ensureImages(id: String): GalleryDetail? {
        val data = VeilApi.client.mutation(EnsureGalleryImagesMutation(id)).execute().dataOrThrow()
        return data.ensureGalleryImages?.galleryDetailFields?.toDetail()
    }

    /** Maps the gallery fragment to the gallery model. */
    private fun GalleryDetailFields.toDetail(): GalleryDetail {
        return GalleryDetail(
            id = id,
            title = title,
            details = details,
            date = date,
            imageCount = imageCount,
            tags = tags.map { tag -> EntityRef(tag.id, tag.name, null) },
            images = images
                .sortedBy { image -> image.position ?: Int.MAX_VALUE }
                .map { image -> GalleryImage(image.id, image.filePath, image.title, image.width, image.height) },
        )
    }
}
