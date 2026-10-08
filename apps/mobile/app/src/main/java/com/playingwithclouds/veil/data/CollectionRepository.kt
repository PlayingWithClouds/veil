package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.graphql.AddToCollectionMutation
import com.playingwithclouds.veil.graphql.CollectionDetailQuery
import com.playingwithclouds.veil.graphql.CollectionsQuery
import com.playingwithclouds.veil.graphql.CreateCollectionMutation
import com.playingwithclouds.veil.graphql.DeleteCollectionMutation
import com.playingwithclouds.veil.graphql.RemoveFromCollectionMutation
import com.playingwithclouds.veil.graphql.RenameCollectionMutation

/** A collection with its members in manual order. */
data class CollectionContents(val collection: CollectionSummary, val members: List<CollectionMember>)

/** User playlists and scraper-created collections. */
object CollectionRepository {

    /** Collections, optionally only user-created ("user") or scraper-created ("scraped") ones. */
    suspend fun collections(origin: String?, limit: Int, offset: Int): List<CollectionSummary> {
        val query = CollectionsQuery(origin.orAbsent(), Optional.Absent, Optional.present(limit), Optional.present(offset))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.collections.map { collection -> collection.collectionFields.toSummary() }
    }

    /** A collection and its members; null when it does not exist. */
    suspend fun contents(id: String): CollectionContents? {
        val data = VeilApi.client.query(CollectionDetailQuery(id)).execute().dataOrThrow()
        val collection = data.collection ?: return null
        val members = data.collectionMembers.mapNotNull { member ->
            val type = MemberType.fromBackend(member.mediaType) ?: return@mapNotNull null
            CollectionMember(member.mediaId, type, member.title, member.posterPath)
        }
        return CollectionContents(collection.collectionFields.toSummary(), members)
    }

    /** Creates a user collection. */
    suspend fun create(name: String): CollectionSummary {
        val data = VeilApi.client.mutation(CreateCollectionMutation(name)).execute().dataOrThrow()
        return data.createCollection.collectionFields.toSummary()
    }

    /** Renames a collection. */
    suspend fun rename(collectionId: String, name: String) {
        VeilApi.client.mutation(RenameCollectionMutation(collectionId, name)).execute().dataOrThrow()
    }

    /** Deletes a collection; its members stay where they are. */
    suspend fun delete(collectionId: String) {
        VeilApi.client.mutation(DeleteCollectionMutation(collectionId)).execute().dataOrThrow()
    }

    /** Adds a record to a collection. */
    suspend fun add(collectionId: String, mediaId: String) {
        VeilApi.client.mutation(AddToCollectionMutation(collectionId, mediaId)).execute().dataOrThrow()
    }

    /** Takes a record out of a collection. */
    suspend fun remove(collectionId: String, mediaId: String) {
        VeilApi.client.mutation(RemoveFromCollectionMutation(collectionId, mediaId)).execute().dataOrThrow()
    }
}
