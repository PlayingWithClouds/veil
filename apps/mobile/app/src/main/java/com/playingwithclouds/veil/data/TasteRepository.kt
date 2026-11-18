package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.TasteAnswerMutation
import com.playingwithclouds.veil.graphql.TasteQuery
import com.playingwithclouds.veil.graphql.fragment.TasteEntryFields
import com.playingwithclouds.veil.graphql.type.EventInput
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput

/**
 * One liked (weight > 0) or disliked (< 0) tag, performer, studio or site; weight is in [-1, 1].
 * [confirmed] is the user's answer: true yes, false not really, null not asked.
 */
data class TasteEntry(val id: String, val name: String, val imageUrl: String?, val weight: Double, val confirmed: Boolean?)

/**
 * How much the user seems to like a scene label: [score] in [-1, 1], [certainty] in [0, 1] (low is
 * a guess worth asking about), [evidence] in words, the user's answer and how many videos have it.
 */
data class SceneLabelTaste(
    val tag: TagSummary,
    val score: Double,
    val certainty: Double,
    val evidence: List<String>,
    val confirmed: Boolean?,
    val videoCount: Int,
)

/** What the user likes and dislikes, per dimension, liked first; [signalCount] is how much it rests on. */
data class Taste(
    val signalCount: Int,
    val tags: List<TasteEntry>,
    val performers: List<TasteEntry>,
    val studios: List<TasteEntry>,
    val sites: List<TasteEntry>,
    val sceneLabels: List<SceneLabelTaste>,
)

/** What a taste answer is about: a record of [kind], or a site (by plugin name) when [kind] is null. */
data class TasteSubject(val kind: RecordKind?, val id: String)

/** The user's taste, from the event log (`apps/web/src/lib/taste.ts`), and their answers on it. */
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
            sceneLabels = taste.sceneLabels.map { label ->
                SceneLabelTaste(TagSummary(label.tag.id, label.tag.name), label.score, label.certainty, label.evidence, label.confirmed, label.videoCount)
            },
        )
    }

    /**
     * Answers about [subject]: 1 yes, -1 not really, 0 takes the answer back. It's a like event the
     * engine reads as a verdict, sent at once (not through the batched event log), so it lands
     * while incognito too, like likes on the video page.
     */
    suspend fun answer(subject: TasteSubject, value: Int) {
        val kind = subject.kind
        var event = EventInput(kind = EventKind.like, data = Optional.present(mapOf("site" to subject.id, "value" to value)))
        if (kind != null) {
            event = EventInput(kind = EventKind.like, subject = Optional.present(RecordRefInput(kind, subject.id)), data = Optional.present(mapOf("value" to value)))
        }
        VeilApi.client.mutation(TasteAnswerMutation(event)).execute().dataOrThrow()
    }

    /** An entry from its fragment. */
    private fun TasteEntryFields.toEntry(): TasteEntry {
        return TasteEntry(id, name, image?.url, weight, confirmed)
    }
}
