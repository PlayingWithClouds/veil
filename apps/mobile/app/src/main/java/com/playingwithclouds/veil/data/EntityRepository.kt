package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.graphql.EnrichPerformerMutation
import com.playingwithclouds.veil.graphql.FollowTargetMutation
import com.playingwithclouds.veil.graphql.GalleriesListQuery
import com.playingwithclouds.veil.graphql.PerformerDetailQuery
import com.playingwithclouds.veil.graphql.PerformersQuery
import com.playingwithclouds.veil.graphql.ScenesListQuery
import com.playingwithclouds.veil.graphql.SetPerformerFavoriteMutation
import com.playingwithclouds.veil.graphql.StudioDetailQuery
import com.playingwithclouds.veil.graphql.StudiosQuery
import com.playingwithclouds.veil.graphql.SubscriptionForTargetQuery
import com.playingwithclouds.veil.graphql.TagCollectionsQuery
import com.playingwithclouds.veil.graphql.TagDetailQuery
import com.playingwithclouds.veil.graphql.TagsQuery
import com.playingwithclouds.veil.graphql.type.SubscriptionKind

/** A performer's profile. */
data class PerformerDetail(
    val id: String,
    val name: String,
    val aliases: List<String>,
    val details: String?,
    val gender: String?,
    val birthdate: String?,
    val country: String?,
    val ethnicity: String?,
    val eyeColor: String?,
    val hairColor: String?,
    val heightCm: Int?,
    val weightKg: Int?,
    val measurements: String?,
    val careerLength: String?,
    val url: String?,
    val imagePath: String?,
    val favorite: Boolean,
    val sceneCount: Int,
    val tags: List<EntityRef>,
)

/** A studio's profile. */
data class StudioDetail(
    val id: String,
    val name: String,
    val aliases: List<String>,
    val url: String?,
    val details: String?,
    val imagePath: String?,
    val sceneCount: Int,
    val parent: EntityRef?,
    val tags: List<EntityRef>,
)

/** A tag's profile. */
data class TagDetail(val id: String, val name: String, val aliases: List<String>, val description: String?, val sceneCount: Int)

/** Scenes of a tag split by how they match it. */
data class TaggedScene(val scene: SceneSummary, val inherited: Boolean)

/** Which entity a scene or gallery listing is for. */
data class EntityFilter(val performerId: String? = null, val studioId: String? = null, val tagId: String? = null)

/** What a followable entity is, matching the subscription kinds. */
enum class FollowKind(val graphqlKind: SubscriptionKind) {
    STUDIO(SubscriptionKind.STUDIO),
    PERFORMER(SubscriptionKind.PERFORMER),
    TAG(SubscriptionKind.TAG),
}

/** Performers, studios, tags and the content filed under them. */
object EntityRepository {

    /** Performers for the index grid. */
    suspend fun performers(search: String?, tagId: String?, limit: Int, offset: Int): List<PerformerSummary> {
        val query = PerformersQuery(
            search = search.orAbsent(),
            tagId = tagId.orAbsent(),
            sort = Optional.Absent,
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.performers.map { performer ->
            PerformerSummary(performer.id, performer.name, performer.imagePath, performer.birthdate, performer.sceneCount, performer.favorite)
        }
    }

    /** A performer's profile, or null when it does not exist. */
    suspend fun performer(id: String): PerformerDetail? {
        val performer = VeilApi.client.query(PerformerDetailQuery(id)).execute().dataOrThrow().performer ?: return null
        return PerformerDetail(
            id = performer.id,
            name = performer.name,
            aliases = performer.aliases,
            details = performer.details,
            gender = performer.gender,
            birthdate = performer.birthdate,
            country = performer.country,
            ethnicity = performer.ethnicity,
            eyeColor = performer.eyeColor,
            hairColor = performer.hairColor,
            heightCm = performer.heightCm,
            weightKg = performer.weightKg,
            measurements = performer.measurements,
            careerLength = performer.careerLength,
            url = performer.url,
            imagePath = performer.imagePath,
            favorite = performer.favorite,
            sceneCount = performer.sceneCount,
            tags = performer.tags.map { tag -> EntityRef(tag.id, tag.name, null) },
        )
    }

    /** Marks a performer as a favorite or not. */
    suspend fun setPerformerFavorite(id: String, favorite: Boolean) {
        VeilApi.client.mutation(SetPerformerFavoriteMutation(id, favorite)).execute().dataOrThrow()
    }

    /** Queues a job that fills in the performer's photo, measurements and bio. */
    suspend fun enrichPerformer(id: String) {
        VeilApi.client.mutation(EnrichPerformerMutation(id)).execute().dataOrThrow()
    }

    /** Studios for the index grid. */
    suspend fun studios(search: String?, tagId: String?, limit: Int, offset: Int): List<StudioSummary> {
        val query = StudiosQuery(
            search = search.orAbsent(),
            tagId = tagId.orAbsent(),
            sort = Optional.Absent,
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.studios.map { studio -> StudioSummary(studio.id, studio.name, studio.imagePath, studio.sceneCount) }
    }

    /** A studio's profile, or null when it does not exist. */
    suspend fun studio(id: String): StudioDetail? {
        val studio = VeilApi.client.query(StudioDetailQuery(id)).execute().dataOrThrow().studio ?: return null
        var parent: EntityRef? = null
        val parentData = studio.parent
        if (parentData != null) {
            parent = EntityRef(parentData.id, parentData.name, null)
        }
        return StudioDetail(
            id = studio.id,
            name = studio.name,
            aliases = studio.aliases,
            url = studio.url,
            details = studio.details,
            imagePath = studio.imagePath,
            sceneCount = studio.sceneCount,
            parent = parent,
            tags = studio.tags.map { tag -> EntityRef(tag.id, tag.name, null) },
        )
    }

    /** Tags for the index. */
    suspend fun tags(search: String?, limit: Int, offset: Int): List<TagSummary> {
        val query = TagsQuery(search.orAbsent(), Optional.present(limit), Optional.present(offset))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.tags.map { tag -> TagSummary(tag.id, tag.name, tag.category, tag.sceneCount) }
    }

    /** A tag's profile, or null when it does not exist. */
    suspend fun tag(id: String): TagDetail? {
        val tag = VeilApi.client.query(TagDetailQuery(id)).execute().dataOrThrow().tag ?: return null
        return TagDetail(tag.id, tag.name, tag.aliases, tag.description, tag.sceneCount)
    }

    /** Scenes of an entity, newest release first; for tags, direct matches rank first. */
    suspend fun scenes(filter: EntityFilter, limit: Int, offset: Int): List<TaggedScene> {
        var sort: String? = null
        if (filter.tagId == null) {
            sort = "date"
        }
        val query = ScenesListQuery(
            performerId = filter.performerId.orAbsent(),
            studioId = filter.studioId.orAbsent(),
            tagId = filter.tagId.orAbsent(),
            sort = sort.orAbsent(),
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.scenes.map { scene ->
            TaggedScene(scene.sceneCardFields.toSummary(), scene.tagMatch == "inherited")
        }
    }

    /** Galleries of an entity. */
    suspend fun galleries(filter: EntityFilter, limit: Int, offset: Int): List<GallerySummary> {
        val query = GalleriesListQuery(
            performerId = filter.performerId.orAbsent(),
            studioId = filter.studioId.orAbsent(),
            tagId = filter.tagId.orAbsent(),
            search = Optional.Absent,
            limit = Optional.present(limit),
            offset = Optional.present(offset),
        )
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.galleries.map { gallery -> gallery.galleryCardFields.toSummary() }
    }

    /** Collections that carry a tag. */
    suspend fun collectionsForTag(tagId: String, limit: Int): List<CollectionSummary> {
        val data = VeilApi.client.query(TagCollectionsQuery(Optional.present(tagId), Optional.present(limit))).execute().dataOrThrow()
        return data.collections.map { collection -> collection.collectionFields.toSummary() }
    }

    /** The subscription following a studio, performer or tag, if any. */
    suspend fun subscriptionFor(targetId: String): SubscriptionSummary? {
        val data = VeilApi.client.query(SubscriptionForTargetQuery(targetId)).execute().dataOrThrow()
        return data.subscriptionForTarget?.subscriptionFields?.toSummary()
    }

    /** Follows a studio, performer or tag; the backend runs the first update in the background. */
    suspend fun follow(kind: FollowKind, targetId: String): SubscriptionSummary {
        val data = VeilApi.client.mutation(FollowTargetMutation(kind.graphqlKind, targetId)).execute().dataOrThrow()
        return data.subscribe.subscriptionFields.toSummary()
    }
}
