package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.FilteredGalleriesQuery
import com.playingwithclouds.veil.graphql.FilteredVideosQuery
import com.playingwithclouds.veil.graphql.PerformersQuery
import com.playingwithclouds.veil.graphql.ReadPerformerMutation
import com.playingwithclouds.veil.graphql.SetFavoriteMutation
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

    /** A studio as stored; null when gone. */
    suspend fun studio(id: String): PersonDetail? {
        return VeilApi.client.query(StudioQuery(id)).execute().dataOrThrow().studio?.studioDetailFields?.toDetail()
    }

    /** Newest videos of a performer or studio, narrowed to [text] when it isn't blank. */
    suspend fun videos(person: PersonDetail, text: String, limit: Int, offset: Int): List<VideoSummary> {
        val query = FilteredVideosQuery(Optional.present(filterOf(person, text)), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().videos.map { video -> video.videoCardFields.toSummary() }
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

    /** The filter of everything credited to a performer or studio. */
    private fun filterOf(person: PersonDetail, text: String = ""): FilterInput {
        val textFilter = Optional.presentIfNotNull(text.trim().ifEmpty { null })
        if (person.kind == RecordKind.studio) {
            return FilterInput(studioIds = Optional.present(listOf(person.id)), text = textFilter)
        }
        return FilterInput(performerIds = Optional.present(listOf(person.id)), text = textFilter)
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
