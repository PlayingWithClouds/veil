package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.BuiltInCollectionQuery
import com.playingwithclouds.veil.graphql.CollectionQuery
import com.playingwithclouds.veil.graphql.CollectionsQuery
import com.playingwithclouds.veil.graphql.CreateCollectionMutation
import com.playingwithclouds.veil.graphql.DeleteDownloadMutation
import com.playingwithclouds.veil.graphql.DownloadsQuery
import com.playingwithclouds.veil.graphql.RemoveFromCollectionMutation
import com.playingwithclouds.veil.graphql.VideoCardQuery
import com.playingwithclouds.veil.graphql.fragment.CollectionPageFields
import com.playingwithclouds.veil.graphql.fragment.GalleryCardFields
import com.playingwithclouds.veil.graphql.type.CollectionKind
import com.playingwithclouds.veil.graphql.type.JobStatus
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope

/** A gallery as a card shows it; [byline] is its first performer. */
data class GallerySummary(val id: String, val title: String, val coverUrl: String?, val imageCount: Int, val byline: String?)

/** One video or gallery in a collection; other kinds of records aren't shown. */
sealed interface CollectionEntry {
    val id: String

    data class Video(val video: VideoSummary) : CollectionEntry {
        override val id: String get() = video.id
    }

    data class Gallery(val gallery: GallerySummary) : CollectionEntry {
        override val id: String get() = gallery.id
    }
}

/** A collection with its videos and galleries, in its order. [builtIn] is the watchlist or the queue. */
data class CollectionPage(
    val id: String,
    val name: String,
    val builtIn: Boolean,
    val details: String?,
    val itemCount: Int,
    val entries: List<CollectionEntry>,
)

/** One of the user's own collections as a tile: up to three covers of its first items. */
data class CollectionTile(val id: String, val name: String, val itemCount: Int, val coverUrls: List<String>)

/** A download of a video: its job's state, 0..1 progress, and the file's size once there. */
data class DownloadItem(
    val jobId: String,
    val status: JobStatus,
    val progress: Double,
    val error: String?,
    val sizeBytes: Double?,
    val video: VideoSummary,
)

/** What the user keeps (`apps/web/src/lib/curation.ts`, `downloads.ts`). */
object LibraryRepository {

    /** The watchlist or the queue, with its items. */
    suspend fun builtInCollection(kind: CollectionKind): CollectionPage {
        val data = VeilApi.client.query(BuiltInCollectionQuery(kind)).execute().dataOrThrow()
        return data.builtInCollection.collectionPageFields.toPage()
    }

    /** A collection with its items; null when it's gone. */
    suspend fun collection(id: String): CollectionPage? {
        val data = VeilApi.client.query(CollectionQuery(id)).execute().dataOrThrow()
        return data.collection?.collectionPageFields?.toPage()
    }

    /** The user's own collections, with their first covers. */
    suspend fun collections(): List<CollectionTile> {
        val data = VeilApi.client.query(CollectionsQuery()).execute().dataOrThrow()
        return data.collections.filter { collection -> collection.kind == CollectionKind.user }.map { collection ->
            val itemCovers = collection.items.mapNotNull { item ->
                item.record?.onVideo?.cover?.url ?: item.record?.onGallery?.cover?.url
            }
            val coverUrls = listOfNotNull(collection.cover?.url).ifEmpty { itemCovers }
            CollectionTile(collection.id, collection.name, collection.itemCount, coverUrls)
        }
    }

    /** Creates an empty collection and returns its id. */
    suspend fun createCollection(name: String): String {
        return VeilApi.client.mutation(CreateCollectionMutation(name)).execute().dataOrThrow().createCollection.id
    }

    /** Takes one video or gallery out of a collection. */
    suspend fun remove(collectionId: String, entry: CollectionEntry) {
        var kind = RecordKind.video
        if (entry is CollectionEntry.Gallery) {
            kind = RecordKind.gallery
        }
        val mutation = RemoveFromCollectionMutation(collectionId, listOf(RecordRefInput(kind, entry.id)))
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** The downloads of videos, newest first; downloads of videos since gone are left out. */
    suspend fun downloads(): List<DownloadItem> = coroutineScope {
        val data = VeilApi.client.query(DownloadsQuery()).execute().dataOrThrow()
        val ofVideos = data.downloads
            .filter { download -> download.job.subject?.kind == RecordKind.video }
            .sortedByDescending { download -> download.job.createdAt }
        val videos = ofVideos.map { download -> async { videoCard(download.job.subject?.id.orEmpty()) } }.awaitAll()
        ofVideos.zip(videos).mapNotNull { (download, video) ->
            if (video == null) {
                return@mapNotNull null
            }
            DownloadItem(
                jobId = download.job.id,
                status = download.job.status,
                progress = download.job.progress,
                error = download.job.error,
                sizeBytes = download.file?.sizeBytes?.toDouble(),
                video = video,
            )
        }
    }

    /** Stops a download and deletes it with its file. */
    suspend fun deleteDownload(jobId: String) {
        VeilApi.client.mutation(DeleteDownloadMutation(jobId)).execute().dataOrThrow()
    }

    /** A video as its card shows it; null when it's gone. */
    private suspend fun videoCard(id: String): VideoSummary? {
        return VeilApi.client.query(VideoCardQuery(id)).execute().dataOrThrow().video?.videoCardFields?.toSummary()
    }

    /** A collection's page from its fragment: only its videos and galleries. */
    private fun CollectionPageFields.toPage(): CollectionPage {
        val entries = items.mapNotNull { item ->
            val video = item.record?.onVideo?.videoCardFields
            val gallery = item.record?.onGallery?.galleryCardFields
            if (video != null) {
                return@mapNotNull CollectionEntry.Video(video.toSummary())
            }
            if (gallery != null) {
                return@mapNotNull CollectionEntry.Gallery(gallery.toSummary())
            }
            null
        }
        return CollectionPage(id, name, kind != CollectionKind.user, details, itemCount, entries)
    }
}

/** A gallery card from its fragment. */
internal fun GalleryCardFields.toSummary(): GallerySummary {
    return GallerySummary(id, title, cover?.url, imageCount, performers.firstOrNull()?.name)
}
