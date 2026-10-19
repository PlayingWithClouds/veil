package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.CamModelQuery
import com.playingwithclouds.veil.graphql.CamRecordingsQuery
import com.playingwithclouds.veil.graphql.CamRoomsQuery
import com.playingwithclouds.veil.graphql.CamSourcesQuery
import com.playingwithclouds.veil.graphql.CamStreamQuery
import com.playingwithclouds.veil.graphql.FollowCamModelMutation
import com.playingwithclouds.veil.graphql.FollowedCamModelsQuery
import com.playingwithclouds.veil.graphql.SetCamAutoRecordMutation
import com.playingwithclouds.veil.graphql.StartCamRecordingMutation
import com.playingwithclouds.veil.graphql.StopCamRecordingMutation
import com.playingwithclouds.veil.graphql.UnfollowCamModelMutation
import com.playingwithclouds.veil.graphql.fragment.CamModelFields
import com.playingwithclouds.veil.graphql.fragment.CamRecordingFields

/** A live cam room as a listing last showed it, with the user's follow and auto-record choices. */
data class CamModel(
    val id: String,
    val pluginName: String,
    val username: String,
    val displayName: String,
    val sourceUrl: String,
    val thumbnailUrl: String?,
    val tags: List<String>,
    val gender: String?,
    val subject: String?,
    val viewers: Int,
    val online: Boolean,
    val lastSeenOnlineAt: String?,
    val followed: Boolean,
    val autoRecord: Boolean,
    /** "source" or a target height like "480". */
    val recordQuality: String,
)

/** One page of live rooms and how many the sites have in all. */
data class CamRoomPage(val rooms: List<CamModel>, val totalCount: Int)

/** A site with cam rooms. */
data class CamSource(val name: String, val displayName: String, val iconUrl: String?)

/** What narrows the live room listing. */
data class CamFilter(
    val sources: List<String> = emptyList(),
    val tag: String? = null,
    val query: String? = null,
    /** female | male | couple | trans, or null for all. */
    val gender: String? = null,
)

/** A recording of a live room, running or finished. */
data class CamRecording(
    val id: String,
    val camModelId: String,
    val displayName: String,
    val thumbnailUrl: String?,
    /** The rendition captured, like 720p. */
    val quality: String?,
    val status: String,
    val stopReason: String?,
    val startedAt: String,
    val endedAt: String?,
    val durationSeconds: Double,
    val sizeBytes: Double,
    /** The scene a finished recording became. */
    val sceneId: String?,
) {

    /** Whether the backend is still capturing it. */
    val isRecording: Boolean
        get() = status == STATUS_RECORDING

    companion object {
        const val STATUS_RECORDING = "recording"
    }
}

/** Live cam rooms: listing, following, the live stream and recordings. */
object CamRepository {

    /** The quality of a recording that takes the best rendition. */
    const val SOURCE_QUALITY = "source"

    /** Live rooms across the cam sites, merged by viewers. */
    suspend fun rooms(filter: CamFilter, limit: Int, offset: Int): CamRoomPage {
        val query = CamRoomsQuery(
            sources = filter.sources.orAbsentIfEmpty(),
            tag = filter.tag.orAbsent(),
            query = filter.query.orAbsent(),
            gender = filter.gender.orAbsent(),
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val page = VeilApi.client.query(query).execute().dataOrThrow().camRooms
        return CamRoomPage(page.rooms.map { room -> room.camModelFields.toModel() }, page.totalCount)
    }

    /** The cam sites, for the site filter. */
    suspend fun sources(): List<CamSource> {
        val data = VeilApi.client.query(CamSourcesQuery()).execute().dataOrThrow()
        return data.camSources.map { source -> CamSource(source.name, source.displayName, source.iconUrl) }
    }

    /** One room with its online status checked against the site; null when unknown. */
    suspend fun model(id: String): CamModel? {
        return VeilApi.client.query(CamModelQuery(id)).execute().dataOrThrow().camModel?.camModelFields?.toModel()
    }

    /** Followed rooms, online ones first. */
    suspend fun followed(): List<CamModel> {
        val data = VeilApi.client.query(FollowedCamModelsQuery()).execute().dataOrThrow()
        return data.followedCamModels.map { model -> model.camModelFields.toModel() }
    }

    /** Follows or unfollows a room; returns it as stored afterwards. */
    suspend fun setFollowed(id: String, followed: Boolean): CamModel {
        if (followed) {
            return VeilApi.client.mutation(FollowCamModelMutation(id)).execute().dataOrThrow().followCamModel.camModelFields.toModel()
        }
        return VeilApi.client.mutation(UnfollowCamModelMutation(id)).execute().dataOrThrow().unfollowCamModel.camModelFields.toModel()
    }

    /**
     * The room's live HLS stream through the backend proxy, rebased onto the server in use. The
     * site's token expires within about a minute, so ask right before loading it.
     */
    suspend fun stream(id: String): PlayableStream {
        val stream = VeilApi.client.query(CamStreamQuery(id)).execute().dataOrThrow().camStream
        return PlayableStream(ServerSettings.backendUrl(stream.url), stream.mimeType, emptyMap())
    }

    /** Running recordings first, then finished ones; of one room when [camModelId] is set. */
    suspend fun recordings(camModelId: String? = null, activeOnly: Boolean = false): List<CamRecording> {
        val query = CamRecordingsQuery(camModelId.orAbsent(), Optional.present(activeOnly))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.camRecordings.map { recording -> recording.camRecordingFields.toRecording() }
    }

    /**
     * Starts recording a live room at [quality] ("source" or a height); with [fromSecondsAgo] it
     * begins that far back in the room's live buffer. Returns the running recording.
     */
    suspend fun startRecording(id: String, quality: String, fromSecondsAgo: Int? = null): CamRecording {
        val mutation = StartCamRecordingMutation(id, Optional.present(quality), fromSecondsAgo.orAbsent())
        return VeilApi.client.mutation(mutation).execute().dataOrThrow().startCamRecording.camRecordingFields.toRecording()
    }

    /** Stops the room's recording and waits for its scene; null when none was running. */
    suspend fun stopRecording(id: String): CamRecording? {
        val data = VeilApi.client.mutation(StopCamRecordingMutation(id)).execute().dataOrThrow()
        return data.stopCamRecording?.camRecordingFields?.toRecording()
    }

    /** Turns automatic recording on or off (on also follows); [quality] stays as it was when null. */
    suspend fun setAutoRecord(id: String, enabled: Boolean, quality: String?): CamModel {
        val mutation = SetCamAutoRecordMutation(id, enabled, quality.orAbsent())
        return VeilApi.client.mutation(mutation).execute().dataOrThrow().setCamAutoRecord.camModelFields.toModel()
    }

    /** Maps the cam model fragment to the model. */
    private fun CamModelFields.toModel(): CamModel {
        return CamModel(
            id = id,
            pluginName = pluginName,
            username = username,
            displayName = displayName,
            sourceUrl = sourceUrl,
            thumbnailUrl = thumbnailUrl,
            tags = tags,
            gender = gender,
            subject = subject,
            viewers = viewers,
            online = online,
            lastSeenOnlineAt = lastSeenOnlineAt,
            followed = followed,
            autoRecord = autoRecord,
            recordQuality = recordQuality,
        )
    }

    /** Maps the recording fragment to the model. */
    private fun CamRecordingFields.toRecording(): CamRecording {
        return CamRecording(
            id = id,
            camModelId = camModelId,
            displayName = displayName,
            thumbnailUrl = thumbnailUrl,
            quality = quality,
            status = status,
            stopReason = stopReason,
            startedAt = startedAt,
            endedAt = endedAt,
            durationSeconds = durationSeconds,
            sizeBytes = sizeBytes,
            sceneId = sceneId,
        )
    }
}
