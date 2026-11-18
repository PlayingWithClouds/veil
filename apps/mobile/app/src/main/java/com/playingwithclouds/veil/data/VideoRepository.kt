package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.AddToCollectionMutation
import com.playingwithclouds.veil.graphql.BuiltInCollectionIdQuery
import com.playingwithclouds.veil.graphql.DownloadMutation
import com.playingwithclouds.veil.graphql.OpenVideoMutation
import com.playingwithclouds.veil.graphql.PlayQuery
import com.playingwithclouds.veil.graphql.RelatedVideosQuery
import com.playingwithclouds.veil.graphql.SetBlockedMutation
import com.playingwithclouds.veil.graphql.SetFavoriteMutation
import com.playingwithclouds.veil.graphql.SiteRelatedVideosQuery
import com.playingwithclouds.veil.graphql.VideoCardQuery
import com.playingwithclouds.veil.graphql.fragment.VideoCardFields
import com.playingwithclouds.veil.graphql.fragment.VideoDetailFields
import com.playingwithclouds.veil.graphql.type.CollectionKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput

/** The video page's reads and writes, as the web's `library.ts`, `curation.ts` and `userState.ts` do them. */
object VideoRepository {

    /** Opens a video: reads the pages behind its sources not read yet (credits, streams); null when it's gone. */
    suspend fun open(id: String): VideoDetail? {
        val data = VeilApi.client.mutation(OpenVideoMutation(id)).execute().dataOrThrow()
        return data.fetchVideo?.videoDetailFields?.let(::detailOf)
    }

    /**
     * A video as its card shows it, from the library alone (no site is read, nothing is logged):
     * fast enough to show the cover and preview while [open] reads the pages. Null when it's gone.
     */
    suspend fun card(id: String): VideoSummary? {
        return VeilApi.client.query(VideoCardQuery(id)).execute().dataOrThrow().video?.videoCardFields?.toSummary()
    }

    /** Resolves a site's stream to a URL the player can load; null when the site gives none. */
    suspend fun play(streamUrl: String): PlayableStream? {
        val playable = VeilApi.client.query(PlayQuery(streamUrl)).execute().dataOrThrow().play ?: return null
        return PlayableStream(ServerSettings.backendUrl(playable.url), playable.mimeType)
    }

    /** Videos in the library like this one, most alike first. */
    suspend fun related(id: String, limit: Int): List<VideoSummary> {
        val data = VeilApi.client.query(RelatedVideosQuery(id, Optional.present(limit))).execute().dataOrThrow()
        return data.related.mapNotNull { match -> match.record?.onVideo?.videoCardFields?.toSummary() }
    }

    /** A page of what the sites list as related on this video's pages. */
    suspend fun siteRelated(id: String, limit: Int, offset: Int): List<VideoSummary> {
        val query = SiteRelatedVideosQuery(id, Optional.present(limit), Optional.present(offset))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.video?.siteRelated.orEmpty().map { video -> video.videoCardFields.toSummary() }
    }

    /** Marks a video as a favourite, or not. */
    suspend fun setFavorite(id: String, favorite: Boolean) {
        VeilApi.client.mutation(SetFavoriteMutation(videoRef(id), favorite)).execute().dataOrThrow()
    }

    /** Puts a video on the watchlist. */
    suspend fun addToWatchlist(id: String) {
        val collection = VeilApi.client.query(BuiltInCollectionIdQuery(CollectionKind.watchlist)).execute().dataOrThrow()
        val mutation = AddToCollectionMutation(collection.builtInCollection.id, listOf(videoRef(id)))
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Puts a video first in the queue, as the web's `playNext`. */
    suspend fun playNext(id: String) {
        val collection = VeilApi.client.query(BuiltInCollectionIdQuery(CollectionKind.queue)).execute().dataOrThrow()
        val mutation = AddToCollectionMutation(collection.builtInCollection.id, listOf(videoRef(id)), Optional.present(0))
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Adds a video to one of the user's collections, at its end. */
    suspend fun addToCollection(collectionId: String, id: String) {
        VeilApi.client.mutation(AddToCollectionMutation(collectionId, listOf(videoRef(id)))).execute().dataOrThrow()
    }

    /** Blocks a video, performer or studio: it, and what credits it, leaves every list. */
    suspend fun block(kind: RecordKind, id: String) {
        VeilApi.client.mutation(SetBlockedMutation(RecordRefInput(kind, id), true)).execute().dataOrThrow()
    }

    /** Downloads a video's best stream for keeps, as a background job. */
    suspend fun download(id: String) {
        VeilApi.client.mutation(DownloadMutation(videoRef(id))).execute().dataOrThrow()
    }

    /** A reference to a video, as the user-state mutations take it. */
    private fun videoRef(id: String): RecordRefInput {
        return RecordRefInput(RecordKind.video, id)
    }

    /** The page's model from its fragment: local copies on disk come before the sites' streams. */
    private fun detailOf(video: VideoDetailFields): VideoDetail {
        val card = video.videoCardFields
        val files = video.files.filter { file -> !file.missing }.map { file ->
            PlayerSource(file.id, "local", file.quality, null, PlayableStream(ServerSettings.backendUrl(file.url), file.mimeType))
        }
        val streams = video.streams.map { stream ->
            PlayerSource(stream.id, stream.source?.plugin ?: "site", stream.quality, stream.url, null)
        }
        // Copies from other sites come after this record's own streams, as fallbacks.
        val copyStreams = video.copies.flatMap { copy -> copy.streams }.map { stream ->
            PlayerSource(stream.id, stream.source?.plugin ?: "site", stream.quality, stream.url, null)
        }
        val copySites = video.copies.flatMap { copy -> copy.sources }.map { source -> SiteSource(source.plugin, source.url) }
        val studio = card.studio?.let { studio -> PerformerSummary(studio.id, studio.name, studio.image?.url) }
        return VideoDetail(
            summary = card.toSummary(),
            details = video.details,
            width = video.width,
            height = video.height,
            studio = studio,
            performers = card.performers.map { performer -> PerformerSummary(performer.id, performer.name, performer.image?.url) },
            tags = video.tags.map { tag -> TagSummary(tag.id, tag.name) },
            sites = video.sources.map { source -> SiteSource(source.plugin, source.url) } + copySites,
            sources = files + streams + copyStreams,
            segments = video.segments.map { segment ->
                VideoSegment(
                    segment.id,
                    segment.title ?: segment.tag?.name.orEmpty(),
                    segment.startSeconds,
                    segment.endSeconds,
                    segment.origin.rawValue,
                    segment.tag?.name,
                )
            },
            favorite = video.userState.favorite,
            rating = video.userState.rating,
            oCount = video.oCount,
            galleryId = video.gallery?.id,
        )
    }
}

/** A video card from its fragment: the studio is the byline, else the first performer. */
internal fun VideoCardFields.toSummary(): VideoSummary {
    var byline = performers.firstOrNull()?.name
    var avatarUrl = performers.firstNotNullOfOrNull { performer -> performer.image?.url }
    val creditedStudio = studio
    if (creditedStudio != null) {
        byline = creditedStudio.name
        avatarUrl = creditedStudio.image?.url ?: avatarUrl
    }
    return VideoSummary(
        id = id,
        title = title,
        coverUrl = cover?.url,
        previewUrl = preview?.url,
        durationSeconds = durationSeconds,
        progressSeconds = userState.progress,
        byline = byline,
        avatarUrl = avatarUrl,
        date = date,
        performers = performers.map { performer -> PerformerSummary(performer.id, performer.name, performer.image?.url) },
        studio = creditedStudio?.let { credited -> PerformerSummary(credited.id, credited.name, credited.image?.url) },
        favorite = userState.favorite,
        sourceUrl = sources.firstOrNull()?.url,
        seen = userState.viewCount > 0,
    )
}
