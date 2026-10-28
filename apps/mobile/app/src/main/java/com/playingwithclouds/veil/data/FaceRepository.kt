package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.graphql.CreditPerformerByNameMutation
import com.playingwithclouds.veil.graphql.IdentifyFaceMutation
import com.playingwithclouds.veil.graphql.ResolvePluginResultMutation

/** A performer who looks like the person in the picture. */
data class FaceCandidate(
    /** The performer's stage name. */
    val name: String,
    /** 0 to 1, the best any service gave. */
    val similarity: Double,
    val imageUrl: String?,
    val profileUrls: List<String>,
    val sources: List<String>,
    /** The library's performer of that name, when there is one. */
    val performerId: String?,
)

/** A tube-site video with a similar face. It names nobody and is never credited. */
data class FaceVideoMatch(
    val title: String,
    val similarity: Double,
    val url: String,
    val thumbnailUrl: String?,
    /** The scene plugin that serves the page, so it can open as a scene; null when none does. */
    val scenePlugin: String?,
)

/** A post on an identification forum (NameThatPorn) whose picture looks like the searched one. */
data class FacePostMatch(
    /** The asker's title. */
    val title: String,
    val url: String,
    val thumbnailUrl: String?,
    /** Whether the post has an accepted answer. */
    val solved: Boolean,
    /** Stage names the accepted answer gives. */
    val names: List<String>,
)

/** A performer the accepted answers of matching posts name. */
data class FacePostPerformer(
    val name: String,
    /** How many matching posts name the performer. */
    val posts: Int,
    /** The thumbnail of the best post naming the performer. */
    val imageUrl: String?,
    val postUrl: String,
    /** The library's performer of that name, when there is one. */
    val performerId: String?,
)

/** What the face-recognition plugins answered for one picture. */
data class FaceIdentification(
    val candidates: List<FaceCandidate>,
    val videos: List<FaceVideoMatch>,
    /** Matching forum posts, in the sites' order. */
    val posts: List<FacePostMatch>,
    /** The performers the posts' answers name, most often named first. */
    val postPerformers: List<FacePostPerformer>,
    /** Plugins that answered. */
    val sources: List<String>,
    /** One line per plugin that failed or timed out. */
    val failures: List<String>,
    /** The stored picture that was sent; null when a frame was. */
    val pictureUrl: String?,
)

/** "Who is this?": performers who look like the person in a picture. */
object FaceRepository {

    /**
     * Sends a picture of [mediaId] to every face-recognition plugin: [frameBase64] (a JPEG, e.g.
     * the frame on screen) when given, else the gallery image [imageId], else the media's own
     * picture (scene thumbnail or poster, the gallery's first image).
     */
    suspend fun identify(mediaId: String, imageId: String? = null, frameBase64: String? = null): FaceIdentification {
        val mutation = IdentifyFaceMutation(mediaId, imageId.orAbsent(), frameBase64.orAbsent())
        val result = VeilApi.client.mutation(mutation).execute().dataOrThrow().identifyFace
        return FaceIdentification(
            candidates = result.candidates.map { candidate ->
                FaceCandidate(
                    name = candidate.name,
                    similarity = candidate.similarity,
                    imageUrl = candidate.imageUrl,
                    profileUrls = candidate.profileUrls,
                    sources = candidate.sources,
                    performerId = candidate.performerId,
                )
            },
            videos = result.videos.map { video ->
                FaceVideoMatch(video.title, video.similarity, video.url, video.thumbnailUrl, video.scenePlugin)
            },
            posts = result.posts.map { post ->
                FacePostMatch(post.title, post.url, post.thumbnailUrl, post.solved, post.names)
            },
            postPerformers = result.postPerformers.map { performer ->
                FacePostPerformer(performer.name, performer.posts, performer.imageUrl, performer.postUrl, performer.performerId)
            },
            sources = result.sources,
            failures = result.failures,
            pictureUrl = result.pictureUrl,
        )
    }

    /** Credits the performer called [name] to the scene or gallery, creating them when needed; returns their id. */
    suspend fun creditPerformer(mediaId: String, name: String): String {
        val data = VeilApi.client.mutation(CreditPerformerByNameMutation(mediaId, name)).execute().dataOrThrow()
        return data.creditPerformerByName.id
    }

    /** Stores a similar video's page as a scene through its plugin; returns the scene id. */
    suspend fun openVideoAsScene(video: FaceVideoMatch): String {
        val plugin = requireNotNull(video.scenePlugin) { "No plugin serves this video" }
        val mutation = ResolvePluginResultMutation(plugin, video.url, video.thumbnailUrl.orAbsent())
        return VeilApi.client.mutation(mutation).execute().dataOrThrow().resolvePluginResult
    }
}
