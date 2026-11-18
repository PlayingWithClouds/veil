package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.FetchUploaderMutation
import com.playingwithclouds.veil.graphql.VideoUploaderQuery

/** The account that uploaded a video to [plugin]'s site; [url] is their channel page, null when the site gives none. */
data class VideoUploader(val plugin: String, val name: String, val url: String?)

/** Uploaders: who a video came from, and the videos on their channel pages, as the web's uploader page reads them. */
object UploaderRepository {

    /** The uploader one of the video's sources names; null when no site did. */
    suspend fun ofVideo(videoId: String): VideoUploader? {
        val video = VeilApi.client.query(VideoUploaderQuery(videoId)).execute().dataOrThrow().video ?: return null
        for (source in video.sources) {
            val uploader = source.uploader
            if (uploader != null) {
                return VideoUploader(source.plugin, uploader.name, uploader.url)
            }
        }
        return null
    }

    /** One page of the videos on the uploader's channel page, read from the site and stored. */
    suspend fun videos(uploader: VideoUploader, limit: Int, offset: Int): List<VideoSummary> {
        val url = uploader.url ?: return emptyList()
        val mutation = FetchUploaderMutation(uploader.plugin, url, uploader.name, Optional.present(limit), Optional.present(offset))
        return VeilApi.client.mutation(mutation).execute().dataOrThrow().fetchUploader.map { video -> video.videoCardFields.toSummary() }
    }
}
