package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.BlockedRecordsQuery
import com.playingwithclouds.veil.graphql.FilteredGalleriesQuery
import com.playingwithclouds.veil.graphql.FilteredVideosQuery
import com.playingwithclouds.veil.graphql.FollowOfQuery
import com.playingwithclouds.veil.graphql.FollowRecordMutation
import com.playingwithclouds.veil.graphql.PerformersQuery
import com.playingwithclouds.veil.graphql.ReadPerformerMutation
import com.playingwithclouds.veil.graphql.RefreshPerformerMutation
import com.playingwithclouds.veil.graphql.SetBlockedMutation
import com.playingwithclouds.veil.graphql.SetFavoriteMutation
import com.playingwithclouds.veil.graphql.UnfollowMutation
import com.playingwithclouds.veil.graphql.VideoTagCountsQuery
import com.playingwithclouds.veil.graphql.StudioQuery
import com.playingwithclouds.veil.graphql.StudiosQuery
import com.playingwithclouds.veil.graphql.fragment.PerformerDetailFields
import com.playingwithclouds.veil.graphql.fragment.StudioDetailFields
import com.playingwithclouds.veil.graphql.type.FilterInput
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import java.time.LocalDate
import java.time.Period

/** A performer or studio in a list: picture, name and how much of theirs the library holds. */
data class MemberSummary(val id: String, val name: String, val imageUrl: String?, val videoCount: Int, val galleryCount: Int)

/**
 * A performer's or studio's page. [facts] are short lines like "29 years old" or "From Japan";
 * [kind] says which it is, for the user-state writes.
 */
data class PersonDetail(
    val id: String,
    val kind: RecordKind,
    val name: String,
    val imageUrl: String?,
    val videoCount: Int,
    val galleryCount: Int,
    val aliases: List<String>,
    val details: String?,
    val facts: List<String>,
    val tags: List<TagSummary>,
    val favorite: Boolean,
)

/** A blocked video, performer or studio in the Library's Blocked list; [video] is the card of a video. */
data class BlockedRecord(val kind: RecordKind, val id: String, val name: String, val imageUrl: String?, val video: VideoSummary? = null)

/** Performers and studios, their pages and their media (`apps/web/src/lib/performers.ts`, `library.ts`). */
object PeopleRepository {

    /** Performers by name (all of them with an empty [text]). */
    suspend fun performers(text: String, limit: Int, offset: Int): List<MemberSummary> {
        val query = PerformersQuery(Optional.present(text), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().performers.map { performer ->
            val member = performer.memberFields
            MemberSummary(member.id, member.name, member.image?.url, member.videoCount, member.galleryCount)
        }
    }

    /** Studios by name (all of them with an empty [text]). */
    suspend fun studios(text: String, limit: Int, offset: Int): List<MemberSummary> {
        val query = StudiosQuery(Optional.present(text), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().studios.map { studio ->
            MemberSummary(studio.id, studio.name, studio.image?.url, studio.videoCount, studio.galleryCount)
        }
    }

    /** A performer, after reading their pages on the sites not read yet (bio, picture, facts); null when gone. */
    suspend fun readPerformer(id: String): PersonDetail? {
        val data = VeilApi.client.mutation(ReadPerformerMutation(id)).execute().dataOrThrow()
        return data.fetchPerformer?.performerDetailFields?.toDetail()
    }

    /** A performer after reading every one of their site pages again, which can bring new videos in; null when gone. */
    suspend fun refreshPerformer(id: String): PersonDetail? {
        val data = VeilApi.client.mutation(RefreshPerformerMutation(id)).execute().dataOrThrow()
        return data.fetchPerformer?.performerDetailFields?.toDetail()
    }

    /** A studio as stored; null when gone. */
    suspend fun studio(id: String): PersonDetail? {
        return VeilApi.client.query(StudioQuery(id)).execute().dataOrThrow().studio?.studioDetailFields?.toDetail()
    }

    /** Videos of a performer or studio as [filter] narrows and sorts them (newest first unsorted), narrowed to [text] when it isn't blank. */
    suspend fun videos(person: PersonDetail, text: String, filter: SearchFilter, limit: Int, offset: Int): List<VideoSummary> {
        val query = FilteredVideosQuery(Optional.present(filterOf(person, text, filter)), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().videos.map { video -> video.videoCardFields.toSummary() }
    }

    /** The tags most common among their videos matching [filter], most common first, for hiding. */
    suspend fun tagCounts(person: PersonDetail, filter: SearchFilter, limit: Int): List<Pair<TagSummary, Int>> {
        val query = VideoTagCountsQuery(Optional.present(filterOf(person, "", filter)), Optional.present(limit))
        return VeilApi.client.query(query).execute().dataOrThrow().videoTagCounts.map { entry ->
            TagSummary(entry.tag.id, entry.tag.name) to entry.count
        }
    }

    /** The id of the follow on a performer or studio; null when they aren't followed. */
    suspend fun followId(person: PersonDetail): String? {
        return VeilApi.client.query(FollowOfQuery(RecordRefInput(person.kind, person.id))).execute().dataOrThrow().followOf?.id
    }

    /** Follows a performer or studio: their new videos come in on their own, with a notification. Returns the follow's id. */
    suspend fun follow(person: PersonDetail): String {
        return VeilApi.client.mutation(FollowRecordMutation(RecordRefInput(person.kind, person.id))).execute().dataOrThrow().followRecord.id
    }

    /** Stops a follow. */
    suspend fun unfollow(followId: String) {
        VeilApi.client.mutation(UnfollowMutation(followId)).execute().dataOrThrow()
    }

    /** Blocks a performer or studio, or lets them back: blocked, they and what credits them leave every list. */
    suspend fun setBlocked(kind: RecordKind, id: String, blocked: Boolean) {
        VeilApi.client.mutation(SetBlockedMutation(RecordRefInput(kind, id), blocked)).execute().dataOrThrow()
    }

    /** A page of blocked records of [kind] (videos, performers or studios), most recently blocked first. */
    suspend fun blocked(kind: RecordKind, limit: Int, offset: Int): List<BlockedRecord> {
        val query = BlockedRecordsQuery(kind, Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().userStates.mapNotNull { state -> blockedRecordOf(kind, state.record) }
    }

    /** One blocked record as the list shows it; null when it's gone from the library. */
    private fun blockedRecordOf(kind: RecordKind, record: BlockedRecordsQuery.Record?): BlockedRecord? {
        val video = record?.onVideo?.videoCardFields
        if (video != null) {
            return BlockedRecord(kind, video.id, video.title, video.cover?.url, video = video.toSummary())
        }
        val performer = record?.onPerformer?.memberFields
        if (performer != null) {
            return BlockedRecord(kind, performer.id, performer.name, performer.image?.url)
        }
        val studio = record?.onStudio ?: return null
        return BlockedRecord(kind, studio.id, studio.name, studio.image?.url)
    }

    /** Newest galleries of a performer or studio. */
    suspend fun galleries(person: PersonDetail, limit: Int): List<GallerySummary> {
        val query = FilteredGalleriesQuery(Optional.present(filterOf(person)), Optional.present(limit), Optional.absent())
        return VeilApi.client.query(query).execute().dataOrThrow().galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
    }

    /** Marks a performer or studio as a favourite, or not; a favourite performer's sites list all their videos. */
    suspend fun setFavorite(person: PersonDetail, favorite: Boolean) {
        VeilApi.client.mutation(SetFavoriteMutation(RecordRefInput(person.kind, person.id), favorite)).execute().dataOrThrow()
    }

    /** The filter of everything credited to a performer or studio, narrowed by [text] and [filter]. */
    private fun filterOf(person: PersonDetail, text: String = "", filter: SearchFilter = SearchFilter()): FilterInput {
        val narrowed = filter.toInput().copy(text = Optional.presentIfNotNull(text.trim().ifEmpty { null }))
        if (person.kind == RecordKind.studio) {
            return narrowed.copy(studioIds = Optional.present(listOf(person.id)))
        }
        return narrowed.copy(performerIds = Optional.present(listOf(person.id)))
    }

    /** A performer's page from its fragment. */
    private fun PerformerDetailFields.toDetail(): PersonDetail {
        val member = memberFields
        return PersonDetail(
            id = member.id,
            kind = RecordKind.performer,
            name = member.name,
            imageUrl = member.image?.url,
            videoCount = member.videoCount,
            galleryCount = member.galleryCount,
            aliases = aliases,
            details = details,
            facts = performerFacts(attributes),
            tags = tags.map { tag -> TagSummary(tag.id, tag.name) },
            favorite = userState.favorite,
        )
    }

    /** A studio's page from its fragment; its parent network is its one fact. */
    private fun StudioDetailFields.toDetail(): PersonDetail {
        return PersonDetail(
            id = id,
            kind = RecordKind.studio,
            name = name,
            imageUrl = image?.url,
            videoCount = videoCount,
            galleryCount = galleryCount,
            aliases = aliases,
            details = details,
            facts = listOfNotNull(parent?.name?.let { network -> "Part of $network" }),
            tags = tags.map { tag -> TagSummary(tag.id, tag.name) },
            favorite = userState.favorite,
        )
    }

    /** "29 years old", "From Japan", "Asian", "Blonde", "165 cm", "86-58-88", "Active 2015–2021"; unknown ones left out. */
    private fun performerFacts(attributes: PerformerDetailFields.Attributes): List<String> {
        return listOfNotNull(
            attributes.birthdate?.let(::ageLabel),
            attributes.country?.let { country -> "From $country" },
            attributes.ethnicity,
            attributes.hairColor?.let { hair -> "$hair hair" },
            attributes.heightCm?.let { height -> "$height cm" },
            attributes.measurements,
            careerLabel(attributes.careerStart, attributes.careerEnd),
        )
    }

    /** "29 years old" from a birthdate like 1995-04-12; null when it can't be read. */
    private fun ageLabel(birthdate: String): String? {
        val born = runCatching { LocalDate.parse(birthdate.take(ISO_DATE_LENGTH)) }.getOrNull() ?: return null
        return "${Period.between(born, LocalDate.now()).years} years old"
    }

    /** "Active 2015–2021", "Active since 2015"; null without a start. */
    private fun careerLabel(start: Int?, end: Int?): String? {
        if (start == null) {
            return null
        }
        if (end == null) {
            return "Active since $start"
        }
        return "Active $start–$end"
    }

    private const val ISO_DATE_LENGTH = 10
}
