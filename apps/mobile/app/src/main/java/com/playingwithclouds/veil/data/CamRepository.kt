package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.graphql.ActiveRecordingsQuery
import com.playingwithclouds.veil.graphql.AutoRecordedRoomsQuery
import com.playingwithclouds.veil.graphql.CamRecommendationsQuery
import com.playingwithclouds.veil.graphql.CamRoomQuery
import com.playingwithclouds.veil.graphql.CamRoomRecordingsQuery
import com.playingwithclouds.veil.graphql.CamRoomsQuery
import com.playingwithclouds.veil.graphql.CamSitesQuery
import com.playingwithclouds.veil.graphql.FollowRecordMutation
import com.playingwithclouds.veil.graphql.FollowsQuery
import com.playingwithclouds.veil.graphql.SetCamAutoRecordMutation
import com.playingwithclouds.veil.graphql.StartRecordingMutation
import com.playingwithclouds.veil.graphql.StopRecordingMutation
import com.playingwithclouds.veil.graphql.UnfollowMutation
import com.playingwithclouds.veil.graphql.WatchCamMutation
import com.playingwithclouds.veil.graphql.fragment.CamRoomFields
import com.playingwithclouds.veil.graphql.type.CamQuery
import com.playingwithclouds.veil.graphql.type.CamSort
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope

/**
 * A cam room with the user's follow and auto-record choices; [online] while it's live. [reason]
 * says why a room ranked by taste is listed ("Because you like #asian").
 */
data class CamModel(
    val id: String,
    val pluginName: String,
    val username: String,
    val displayName: String,
    val thumbnailUrl: String?,
    val tags: List<String>,
    val subject: String?,
    val viewers: Int,
    val online: Boolean,
    val lastSeenOnlineAt: String?,
    val followed: Boolean,
    val autoRecord: Boolean,
    /** "source" or a target height like "480". */
    val recordQuality: String,
    val reason: String? = null,
)

/** A site with cam rooms. */
data class CamSource(val name: String, val displayName: String, val iconUrl: String?)

/** The orders of the live rooms, as the web offers them. */
enum class CamOrder(val label: String, val sort: CamSort) {
    VIEWERS("Most viewers", CamSort.viewers),
    FOR_YOU("For you", CamSort.forYou),
    STARTED("Just started", CamSort.started),
    FEWEST("Fewest viewers", CamSort.fewest),
}

/** What narrows the live rooms and how they're ordered; empty values don't narrow. */
data class CamFilter(
    val sources: List<String> = emptyList(),
    val tag: String? = null,
    val query: String? = null,
    /** female | male | couple | trans, or null for all. */
    val gender: String? = null,
    val order: CamOrder = CamOrder.VIEWERS,
)

/** A recording running now. */
data class CamRecording(
    val camModelId: String,
    val displayName: String,
    val thumbnailUrl: String?,
    /** The rendition captured, like 720p. */
    val quality: String?,
    val startedAt: String,
    val durationSeconds: Double,
    val sizeBytes: Double,
)

/** Cam rooms: listing, recommendations, following, the live stream and recordings (`apps/web/src/lib/cams.ts`). */
object CamRepository {

    /** The quality of a recording that takes the best rendition. */
    const val SOURCE_QUALITY = "source"

    /** Capability of the sites that list cam rooms. */
    private const val CAM_LIST_CAPABILITY = "cam:list"

    /** A page of the rooms live now that match the filter, in its order, each with why it's listed. */
    suspend fun rooms(filter: CamFilter, limit: Int, offset: Int): List<CamModel> {
        val query = CamQuery(
            plugins = Optional.presentIfNotNull(filter.sources.ifEmpty { null }),
            text = filter.query.orAbsent(),
            tag = filter.tag.orAbsent(),
            gender = filter.gender.orAbsent(),
            sort = Optional.present(filter.order.sort),
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val followed = followedRoomIds()
        val page = VeilApi.client.query(CamRoomsQuery(Optional.present(query))).execute().dataOrThrow().camRooms
        return page.rooms.mapIndexed { index, room -> room.camRoomFields.toModel(followed, page.reasons.getOrNull(index)) }
    }

    /** Live rooms picked by the user's taste, each with why. */
    suspend fun recommended(limit: Int): List<CamModel> {
        val followed = followedRoomIds()
        val page = VeilApi.client.query(CamRecommendationsQuery(Optional.present(limit))).execute().dataOrThrow().camRecommendations
        return page.rooms.mapIndexed { index, room -> room.camRoomFields.toModel(followed, page.reasons.getOrNull(index)) }
    }

    /** The enabled sites that list cam rooms, for the site filter. */
    suspend fun sources(): List<CamSource> {
        val data = VeilApi.client.query(CamSitesQuery()).execute().dataOrThrow()
        return data.plugins
            .filter { plugin -> plugin.enabled && plugin.available && CAM_LIST_CAPABILITY in plugin.capabilities }
            .map { plugin -> CamSource(plugin.name, plugin.displayName ?: plugin.name, plugin.iconUrl) }
    }

    /** One stored room with its live state; null when unknown. */
    suspend fun model(id: String): CamModel? {
        val room = VeilApi.client.query(CamRoomQuery(id)).execute().dataOrThrow().camModel ?: return null
        return room.camRoomFields.toModel(followedRoomIds())
    }

    /** The user's rooms: followed or recorded automatically, live ones first by viewers, then by name. */
    suspend fun yourRooms(): List<CamModel> = coroutineScope {
        val follows = followsByRoom()
        val autoRecorded = VeilApi.client.query(AutoRecordedRoomsQuery()).execute().dataOrThrow()
            .camModels.map { room -> room.camRoomFields.toModel(follows.keys) }
        val knownIds = autoRecorded.map { room -> room.id }.toSet()
        val followed = follows.keys.filter { id -> id !in knownIds }.map { id -> async { model(id) } }.awaitAll().filterNotNull()
        (autoRecorded + followed).sortedWith(compareByDescending<CamModel> { room -> room.online }
            .thenByDescending { room -> room.viewers }
            .thenBy { room -> room.displayName.lowercase() })
    }

    /** The recordings running now. */
    suspend fun activeRecordings(): List<CamRecording> {
        val data = VeilApi.client.query(ActiveRecordingsQuery()).execute().dataOrThrow()
        return data.activeRecordings.mapNotNull { recording ->
            val room = recording.camModel?.camRoomFields ?: return@mapNotNull null
            CamRecording(
                camModelId = room.id,
                displayName = room.displayName,
                thumbnailUrl = room.live?.thumbnail?.url ?: room.thumbnail?.url,
                quality = recording.quality,
                startedAt = recording.startedAt,
                durationSeconds = recording.durationSeconds,
                sizeBytes = recording.sizeBytes.toDouble(),
            )
        }
    }

    /** A room's finished recordings, as the videos they became, newest first. */
    suspend fun recordedVideos(id: String): List<VideoSummary> {
        val data = VeilApi.client.query(CamRoomRecordingsQuery(id)).execute().dataOrThrow()
        return data.camModel?.recordings.orEmpty().map { video -> video.videoCardFields.toSummary() }
    }

    /** Follows a room (going live notifies) or ends its follow. */
    suspend fun setFollowed(id: String, followed: Boolean) {
        if (followed) {
            VeilApi.client.mutation(FollowRecordMutation(RecordRefInput(RecordKind.camModel, id))).execute().dataOrThrow()
            return
        }
        val followId = followsByRoom()[id] ?: return
        VeilApi.client.mutation(UnfollowMutation(followId)).execute().dataOrThrow()
    }

    /**
     * The room's live stream through the backend, rebased onto the server in use. The site's token
     * runs out within about a minute, so ask right before loading it.
     */
    suspend fun stream(id: String): PlayableStream {
        val stream = VeilApi.client.mutation(WatchCamMutation(id)).execute().dataOrThrow().watchCam
        return PlayableStream(ServerSettings.backendUrl(stream.url), stream.mimeType)
    }

    /**
     * Starts recording a live room at [quality] ("source" or a height); with [fromSecondsAgo] it
     * begins that far back in the room's live buffer.
     */
    suspend fun startRecording(id: String, quality: String, fromSecondsAgo: Int? = null) {
        val mutation = StartRecordingMutation(id, Optional.present(quality), fromSecondsAgo.orAbsent())
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Stops the room's recording; the backend files it as a video. */
    suspend fun stopRecording(id: String) {
        VeilApi.client.mutation(StopRecordingMutation(id)).execute().dataOrThrow()
    }

    /** Turns automatic recording on or off; [quality] stays as it was when null. */
    suspend fun setAutoRecord(id: String, enabled: Boolean, quality: String?): CamModel {
        val mutation = SetCamAutoRecordMutation(id, enabled, quality.orAbsent())
        val room = VeilApi.client.mutation(mutation).execute().dataOrThrow().setCamAutoRecord
        return room.camRoomFields.toModel(followedRoomIds())
    }

    /** The ids of the followed rooms. */
    private suspend fun followedRoomIds(): Set<String> {
        return followsByRoom().keys
    }

    /** The enabled follows of rooms: room id to follow id. */
    private suspend fun followsByRoom(): Map<String, String> {
        val data = VeilApi.client.query(FollowsQuery()).execute().dataOrThrow()
        return data.follows
            .filter { follow -> follow.enabled && follow.targetKind == RecordKind.camModel && follow.targetId != null }
            .associate { follow -> follow.targetId.orEmpty() to follow.id }
    }

    /** The room from its fragment: the live picture while live, else the last one. */
    private fun CamRoomFields.toModel(followedIds: Set<String>, reason: String? = null): CamModel {
        val show = this.live
        return CamModel(
            id = id,
            pluginName = plugin,
            username = username,
            displayName = displayName,
            thumbnailUrl = show?.thumbnail?.url ?: thumbnail?.url,
            tags = tags,
            subject = show?.subject,
            viewers = show?.viewers ?: 0,
            online = show != null,
            lastSeenOnlineAt = lastSeenOnlineAt,
            followed = id in followedIds,
            autoRecord = autoRecord,
            recordQuality = recordQuality ?: SOURCE_QUALITY,
            reason = reason?.ifEmpty { null },
        )
    }
}
