package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.HomeAudiosQuery
import com.playingwithclouds.veil.graphql.HomeFavoritePerformersQuery
import com.playingwithclouds.veil.graphql.HomeLiveQuery
import com.playingwithclouds.veil.graphql.HomePerformersQuery
import com.playingwithclouds.veil.graphql.HomeRecommendationsQuery
import com.playingwithclouds.veil.graphql.HomeStoriesQuery
import com.playingwithclouds.veil.graphql.HomeStoryHistoryQuery
import com.playingwithclouds.veil.graphql.HomeVideoHistoryQuery
import com.playingwithclouds.veil.graphql.fragment.AudioCardFields
import com.playingwithclouds.veil.graphql.fragment.CamCardFields
import com.playingwithclouds.veil.graphql.fragment.PerformerCardFields
import com.playingwithclouds.veil.graphql.fragment.StoryCardFields
import com.playingwithclouds.veil.graphql.fragment.VideoCardFields

/** The rows of Home, read the way the web home reads them (`apps/web/src/lib/home.ts`). */
object HomeRepository {

    /** Recommended videos with their first reason, best first. */
    suspend fun recommendedVideos(limit: Int): List<RecommendedVideo> {
        val data = VeilApi.client.query(HomeRecommendationsQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.recommendations.mapNotNull { match ->
            val video = match.record?.onVideo?.videoCardFields ?: return@mapNotNull null
            RecommendedVideo(videoOf(video), match.evidence.firstOrNull()?.detail)
        }
    }

    /**
     * Videos stopped partway, most recently watched first. History holds finished videos too, so
     * twice as many are asked for.
     */
    suspend fun continueWatching(limit: Int): List<VideoSummary> {
        val query = HomeVideoHistoryQuery(Optional.present(limit * 2))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        val started = data.history.filter { state -> state.progress != null }
        return started.mapNotNull { state -> state.record?.onVideo?.videoCardFields?.let(::videoOf) }.take(limit)
    }

    /** Stories stopped partway, most recently read first. */
    suspend fun continueReading(limit: Int): List<StorySummary> {
        val query = HomeStoryHistoryQuery(Optional.present(limit * 2))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        val started = data.history.filter { state -> state.progress != null }
        return started.mapNotNull { state -> state.record?.onStory?.storyCardFields?.let(::storyOf) }.take(limit)
    }

    /** Rooms being recorded now, then the rooms live now with the most viewers, each room once. */
    suspend fun live(limit: Int): List<CamSummary> {
        val data = VeilApi.client.query(HomeLiveQuery(Optional.present(limit))).execute().dataOrThrow()
        val recorded = data.activeRecordings.mapNotNull { recording -> recording.camModel?.camCardFields }
        val recordedIds = recorded.map { room -> room.id }.toSet()
        val live = data.camRooms.rooms.map { room -> room.camCardFields }.filter { room -> room.id !in recordedIds }
        return recorded.map { room -> camOf(room, recording = true) } + live.map { room -> camOf(room, recording = false) }
    }

    /** The newest stories. */
    suspend fun stories(limit: Int): List<StorySummary> {
        val data = VeilApi.client.query(HomeStoriesQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.stories.map { story -> storyOf(story.storyCardFields) }
    }

    /** The newest audio. */
    suspend fun audios(limit: Int): List<AudioSummary> {
        val data = VeilApi.client.query(HomeAudiosQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.audios.map { audio -> audioOf(audio.audioCardFields) }
    }

    /** The user's favourite performers, else the library's first ones. */
    suspend fun performers(limit: Int): List<PerformerSummary> {
        val favoritesQuery = HomeFavoritePerformersQuery(Optional.present(limit))
        val favorites = VeilApi.client.query(favoritesQuery).execute().dataOrThrow()
            .userStates.mapNotNull { state -> state.record?.onPerformer?.performerCardFields?.let(::performerOf) }
        if (favorites.isNotEmpty()) {
            return favorites
        }
        val data = VeilApi.client.query(HomePerformersQuery(Optional.present(limit))).execute().dataOrThrow()
        return data.performers.map { performer -> performerOf(performer.performerCardFields) }
    }

    /** A video card from its fragment: the studio is the byline, else the first performer. */
    private fun videoOf(video: VideoCardFields): VideoSummary {
        val firstPerformer = video.performers.firstOrNull()
        var byline = firstPerformer?.name
        var avatarUrl = video.performers.firstNotNullOfOrNull { performer -> performer.image?.url }
        val studio = video.studio
        if (studio != null) {
            byline = studio.name
            avatarUrl = studio.image?.url ?: avatarUrl
        }
        return VideoSummary(
            id = video.id,
            title = video.title,
            coverUrl = video.cover?.url,
            previewUrl = video.preview?.url,
            durationSeconds = video.durationSeconds,
            progressSeconds = video.userState.progress,
            byline = byline,
            avatarUrl = avatarUrl,
            date = video.date,
        )
    }

    /** A cam card from its fragment: the live picture while live, else the last one. */
    private fun camOf(room: CamCardFields, recording: Boolean): CamSummary {
        return CamSummary(
            id = room.id,
            name = room.displayName,
            site = room.plugin,
            pictureUrl = room.live?.thumbnail?.url ?: room.thumbnail?.url,
            viewers = room.live?.viewers,
            recording = recording,
        )
    }

    /** A story card from its fragment; the first performer is the author. */
    private fun storyOf(story: StoryCardFields): StorySummary {
        return StorySummary(story.id, story.title, story.cover?.url, story.performers.firstOrNull()?.name, story.chapterCount)
    }

    /** An audio card from its fragment; the first performer is the voice. */
    private fun audioOf(audio: AudioCardFields): AudioSummary {
        return AudioSummary(audio.id, audio.title, audio.cover?.url, audio.performers.firstOrNull()?.name, audio.durationSeconds)
    }

    /** A performer card from its fragment. */
    private fun performerOf(performer: PerformerCardFields): PerformerSummary {
        return PerformerSummary(performer.id, performer.name, performer.image?.url)
    }
}
