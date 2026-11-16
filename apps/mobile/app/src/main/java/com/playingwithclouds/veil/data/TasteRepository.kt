package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.TasteQuery
import com.playingwithclouds.veil.graphql.fragment.TasteEntryFields

/** One liked (weight > 0) or disliked (< 0) tag, performer, studio or site; weight is in [-1, 1]. */
data class TasteEntry(val id: String, val name: String, val imageUrl: String?, val weight: Double)

/** What the user likes and dislikes, per dimension, liked first; [signalCount] is how much it rests on. */
data class Taste(
    val signalCount: Int,
    val tags: List<TasteEntry>,
    val performers: List<TasteEntry>,
    val studios: List<TasteEntry>,
    val sites: List<TasteEntry>,
)

/** The user's taste, from the event log (`apps/web/src/lib/taste.ts`). */
object TasteRepository {

    /** What the recommendations think the user likes and dislikes. */
    suspend fun taste(): Taste {
        val taste = VeilApi.client.query(TasteQuery()).execute().dataOrThrow().taste
        return Taste(
            signalCount = taste.signalCount,
            tags = taste.tags.map { entry -> entry.tasteEntryFields.toEntry() },
            performers = taste.performers.map { entry -> entry.tasteEntryFields.toEntry() },
            studios = taste.studios.map { entry -> entry.tasteEntryFields.toEntry() },
            sites = taste.sites.map { entry -> entry.tasteEntryFields.toEntry() },
        )
    }

    /** An entry from its fragment. */
    private fun TasteEntryFields.toEntry(): TasteEntry {
        return TasteEntry(id, name, image?.url, weight)
    }
}
