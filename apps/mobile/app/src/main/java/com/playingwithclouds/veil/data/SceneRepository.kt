package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.ReplaceAiSegmentsMutation
import com.playingwithclouds.veil.graphql.SceneFetchSourcesMutation
import com.playingwithclouds.veil.graphql.SceneMatchesQuery
import com.playingwithclouds.veil.graphql.ScenePreviewPlanQuery
import com.playingwithclouds.veil.graphql.SceneQueueQuery
import com.playingwithclouds.veil.graphql.SceneScanCandidatesQuery
import com.playingwithclouds.veil.graphql.SceneSettingsQuery
import com.playingwithclouds.veil.graphql.SceneVideoSourcesQuery
import com.playingwithclouds.veil.graphql.ShortsQuery
import com.playingwithclouds.veil.graphql.UpdateSceneSettingsMutation
import com.playingwithclouds.veil.graphql.fragment.SceneSegmentFields
import com.playingwithclouds.veil.graphql.fragment.SceneSourceFields
import com.playingwithclouds.veil.graphql.type.AiSegmentInput
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import com.playingwithclouds.veil.graphql.type.SegmentOrigin
import com.playingwithclouds.veil.scenes.TimedSegment
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** A video ready to play from: what the player loads and the runtime, when known. */
data class PlayableVideo(val stream: PlayableStream, val durationSeconds: Double?)

/** What a card's moment preview can show: the ai segments, else the site's stills across the video. */
data class PreviewPlan(val durationSeconds: Double?, val segments: List<TimedSegment>, val frames: List<String>)

/** A video whose ai segments match the liked scenes, with those segments. */
data class SceneMatchVideo(val video: VideoSummary, val segments: List<TimedSegment>)

/**
 * One entry of the shorts feed: the stretch of [video] to play and why it's there. [itemKind] and
 * [itemId] are what the feed recommended (a clip, a segment or the video), which a skip is about.
 */
data class ShortItem(
    val itemKind: RecordKind,
    val itemId: String,
    val video: VideoSummary,
    val startSeconds: Double,
    val endSeconds: Double,
    val detail: String?,
) {
    /** Unique within the feed. */
    val key: String get() = "${itemKind.rawValue}:$itemId"
}

/**
 * The scene model's reads and writes against the backend: what to scan, playable streams without
 * opening the video page, the ai segments, the liked scenes (the global setting
 * `likedSceneTags`, shared with the web), compilations and the shorts feed.
 */
object SceneRepository {

    private const val LIKED_SETTING = "likedSceneTags"

    private val mutableLikedTags = MutableStateFlow<Set<String>>(emptySet())

    /** The tag names of the scenes the user likes, as last read or saved. */
    val likedTags: StateFlow<Set<String>> = mutableLikedTags

    private var likedTagsRead = false

    /** Reads the liked scenes unless they were read (or saved) in this process already. */
    suspend fun ensureLikedTags() {
        if (likedTagsRead) {
            return
        }
        runCatching { refreshLikedTags() }
    }

    /** Reads the liked scenes from the backend. */
    suspend fun refreshLikedTags(): Set<String> {
        val settings = VeilApi.client.query(SceneSettingsQuery()).execute().dataOrThrow().settings as? Map<*, *>
        val liked = (settings?.get(LIKED_SETTING) as? List<*>).orEmpty().filterIsInstance<String>().toSet()
        mutableLikedTags.value = liked
        likedTagsRead = true
        return liked
    }

    /** Saves the liked scenes. */
    suspend fun setLikedTags(tags: Set<String>) {
        mutableLikedTags.value = tags
        likedTagsRead = true
        VeilApi.client.mutation(UpdateSceneSettingsMutation(mapOf(LIKED_SETTING to tags.sorted()))).execute().dataOrThrow()
    }

    /**
     * Video ids worth scanning, in priority order without repeats: favourites, collections
     * (watchlist, queue and the user's own), recently watched, then recommendations.
     */
    suspend fun scanCandidates(limit: Int): List<String> {
        val data = VeilApi.client.query(SceneScanCandidatesQuery(Optional.present(limit))).execute().dataOrThrow()
        val ids = data.favorites.map { video -> video.id } +
            data.collections.flatMap { collection -> collection.items.mapNotNull { item -> item.record?.onVideo?.id } } +
            data.history.mapNotNull { state -> state.record?.onVideo?.id } +
            data.recommendations.map { match -> match.item.id }
        return ids.distinct()
    }

    /**
     * Something the scanner, a preview or a compilation can play: a local copy first, else the
     * first site stream that resolves. [readPages] reads the video's unread pages first (a stub has
     * no streams yet); without it the stored streams are used. Null when nothing plays.
     */
    suspend fun playable(videoId: String, readPages: Boolean): PlayableVideo? {
        val sources = if (readPages) {
            VeilApi.client.mutation(SceneFetchSourcesMutation(videoId)).execute().dataOrThrow().fetchVideo?.sceneSourceFields
        } else {
            VeilApi.client.query(SceneVideoSourcesQuery(videoId)).execute().dataOrThrow().video?.sceneSourceFields
        }
        return sources?.let { video -> firstPlayable(video) }
    }

    /** The first source of [video] that plays. */
    private suspend fun firstPlayable(video: SceneSourceFields): PlayableVideo? {
        val file = video.files.firstOrNull { file -> !file.missing }
        if (file != null) {
            return PlayableVideo(PlayableStream(ServerSettings.backendUrl(file.url), file.mimeType), video.durationSeconds)
        }
        for (stream in video.streams) {
            val playable = runCatching { VideoRepository.play(stream.url) }.getOrNull()
            if (playable != null) {
                return PlayableVideo(playable, video.durationSeconds)
            }
        }
        return null
    }

    /** Replaces the video's ai segments with what the model found. */
    suspend fun replaceSegments(videoId: String, segments: List<TimedSegment>) {
        val inputs = segments.map { segment -> AiSegmentInput(segment.startSeconds, segment.endSeconds, segment.tag) }
        VeilApi.client.mutation(ReplaceAiSegmentsMutation(RecordRefInput(RecordKind.video, videoId), inputs)).execute().dataOrThrow()
    }

    /** What a card's moment preview can show; null when the video is gone. */
    suspend fun previewPlan(videoId: String): PreviewPlan? {
        val video = VeilApi.client.query(ScenePreviewPlanQuery(videoId)).execute().dataOrThrow().video ?: return null
        val segments = video.segments.map { segment -> segment.sceneSegmentFields }.mapNotNull(::timedOf)
        return PreviewPlan(video.durationSeconds, segments, video.previewFrames.orEmpty().map { frame -> frame.url })
    }

    /** A page of videos with ai segments of [tags], most matching first, or only [videoIds] in their order. */
    suspend fun sceneMatches(tags: Set<String>, videoIds: List<String>?, limit: Int, offset: Int): List<SceneMatchVideo> {
        val query = SceneMatchesQuery(tags.toList(), Optional.presentIfNotNull(videoIds), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().sceneMatches.mapNotNull { match ->
            val video = match.video ?: return@mapNotNull null
            SceneMatchVideo(video.videoCardFields.toSummary(), match.segments.map { segment -> segment.sceneSegmentFields }.mapNotNull(::timedOf))
        }
    }

    /** The queue's video ids, in order. */
    suspend fun queueIds(): List<String> {
        val data = VeilApi.client.query(SceneQueueQuery()).execute().dataOrThrow()
        return data.builtInCollection.items.mapNotNull { item -> item.record?.onVideo?.id }
    }

    /** The shorts feed, best first. */
    suspend fun shorts(limit: Int): List<ShortItem> {
        val data = VeilApi.client.query(ShortsQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.shorts.mapNotNull { short ->
            val video = short.video ?: return@mapNotNull null
            val start = short.startSeconds ?: 0.0
            val end = short.endSeconds ?: (start + DEFAULT_SHORT_SECONDS)
            val evidence = short.match.evidence.firstOrNull()
            ShortItem(short.match.item.kind, short.match.item.id, video.videoCardFields.toSummary(), start, end, evidence?.detail)
        }
    }

    /** How long a short without an end plays. */
    private const val DEFAULT_SHORT_SECONDS = 30.0

    /** An ai segment with an end and a tag as a timed one; others are left out. */
    private fun timedOf(segment: SceneSegmentFields): TimedSegment? {
        if (segment.origin != SegmentOrigin.ai) {
            return null
        }
        val end = segment.endSeconds ?: return null
        val tag = segment.tag?.name ?: return null
        return TimedSegment(tag, segment.startSeconds, end)
    }
}
