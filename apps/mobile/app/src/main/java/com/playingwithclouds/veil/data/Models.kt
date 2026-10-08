package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.fragment.CollectionFields
import com.playingwithclouds.veil.graphql.fragment.GalleryCardFields
import com.playingwithclouds.veil.graphql.fragment.SceneCardFields
import com.playingwithclouds.veil.graphql.fragment.SubscriptionFields

/** A studio, performer or tag reference as listed on cards. */
data class EntityRef(val id: String, val name: String, val imagePath: String?)

/** What a scene card shows. Image fields are raw backend paths; the UI routes them through the image cache. */
data class SceneSummary(
    val id: String,
    val title: String,
    val sourceUrl: String,
    val posterPath: String?,
    val previewVideo: String?,
    val previewImages: List<String>,
    val durationSeconds: Int?,
    val date: String?,
    val studio: EntityRef?,
    val performers: List<EntityRef>,
    val tags: List<EntityRef>,
) {

    /** Caption byline: the channel/studio, else up to two performers. */
    val byline: String?
        get() {
            if (studio != null) {
                return studio.name
            }
            if (performers.isEmpty()) {
                return null
            }
            return performers.take(2).joinToString(", ") { performer -> performer.name }
        }
}

/** A recommended scene with the candidate source the engine served it from. */
data class RecommendedScene(val scene: SceneSummary, val source: String, val reason: String)

/** A scene found by a subscription, with when and whether it is new. */
data class SubscriptionFeedEntry(val scene: SceneSummary, val foundAt: String, val subscriptionId: String, val isNew: Boolean)

/** What a gallery card shows. */
data class GallerySummary(val id: String, val title: String, val coverPath: String?, val imageCount: Int, val sourceUrl: String?)

/** A performer as listed in grids. */
data class PerformerSummary(
    val id: String,
    val name: String,
    val imagePath: String?,
    val birthdate: String?,
    val sceneCount: Int,
    val favorite: Boolean,
)

/** A studio as listed in grids. */
data class StudioSummary(val id: String, val name: String, val imagePath: String?, val sceneCount: Int)

/** A tag as listed in the index. */
data class TagSummary(val id: String, val name: String, val category: String?, val sceneCount: Int)

/** A collection as listed in the index. */
data class CollectionSummary(
    val id: String,
    val name: String,
    val details: String?,
    val itemCount: Int,
    val coverPath: String?,
    val isUserCreated: Boolean,
    val tags: List<EntityRef>,
)

/** An entity type a collection can contain. */
enum class MemberType {
    SCENE,
    GALLERY,
    IMAGE,
    PERFORMER,
    STUDIO,
    ;

    companion object {

        /** The type named by the backend ("scene", "gallery", ...), or null for unknown values. */
        fun fromBackend(name: String): MemberType? {
            return entries.firstOrNull { type -> type.name.equals(name, ignoreCase = true) }
        }
    }
}

/** One member of a collection, shaped as a card. */
data class CollectionMember(val mediaId: String, val type: MemberType, val title: String, val posterPath: String?)

/** A subscription: a saved search, or a followed studio, performer or tag. */
data class SubscriptionSummary(
    val id: String,
    val kind: String,
    val query: String,
    val targetId: String?,
    val targetName: String?,
    val targetImageUrl: String?,
    val sources: List<String>,
    val intervalHours: Int,
    val enabled: Boolean,
    val lastRunAt: String?,
    val nextRunAt: String?,
    val lastError: String?,
    val newCount: Int,
    val totalCount: Int,
) {

    /** Display name: the followed target's current name, else the stored query. */
    val displayName: String
        get() = targetName ?: query
}

/** Maps a card fragment to the scene model. */
fun SceneCardFields.toSummary(): SceneSummary {
    val studioRef = studio
    var studioEntity: EntityRef? = null
    if (studioRef != null) {
        studioEntity = EntityRef(studioRef.id, studioRef.name, studioRef.imagePath)
    }
    return SceneSummary(
        id = id,
        title = title,
        sourceUrl = sourceUrl,
        posterPath = posterPath,
        previewVideo = previewVideo,
        previewImages = previewImages,
        durationSeconds = durationSeconds,
        date = date,
        studio = studioEntity,
        performers = performers.map { performer -> EntityRef(performer.id, performer.name, performer.imagePath) },
        tags = tags.map { tag -> EntityRef(tag.id, tag.name, null) },
    )
}

/** Maps a gallery card fragment to the gallery model. */
fun GalleryCardFields.toSummary(): GallerySummary {
    return GallerySummary(id, title, coverPath, imageCount, sourceUrl)
}

/** Maps a collection fragment to the collection model. */
fun CollectionFields.toSummary(): CollectionSummary {
    return CollectionSummary(
        id = id,
        name = name,
        details = details,
        itemCount = itemCount,
        coverPath = coverPath,
        isUserCreated = origin == "user",
        tags = tags.map { tag -> EntityRef(tag.id, tag.name, null) },
    )
}

/** Maps a subscription fragment to the subscription model. */
fun SubscriptionFields.toSummary(): SubscriptionSummary {
    return SubscriptionSummary(
        id = id,
        kind = kind.rawValue,
        query = query,
        targetId = target?.id,
        targetName = target?.name,
        targetImageUrl = target?.imageUrl,
        sources = sources,
        intervalHours = intervalHours,
        enabled = enabled,
        lastRunAt = lastRunAt,
        nextRunAt = nextRunAt,
        lastError = lastError,
        newCount = newCount,
        totalCount = totalCount,
    )
}
