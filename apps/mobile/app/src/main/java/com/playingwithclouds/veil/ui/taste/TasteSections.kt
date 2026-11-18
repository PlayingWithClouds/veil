package com.playingwithclouds.veil.ui.taste

import com.playingwithclouds.veil.data.SceneLabelTaste
import com.playingwithclouds.veil.data.Taste
import com.playingwithclouds.veil.data.TasteEntry

/** Labels shown under "What you're into". */
private const val INTO_LIMIT = 12

/** Entries shown per section of performers, studios and sites. */
private const val ENTRY_LIMIT = 8

/** Questions asked under "Not sure yet". */
private const val QUESTION_LIMIT = 6

/** A label with less certainty than this is a guess: it's asked about rather than stated. */
const val SURE_CERTAINTY = 0.3

/** A label scored at or below this, with enough certainty, is something the user avoids. */
private const val AVOID_SCORE = -0.2

/** Whose answer a disliked entry of "Things you avoid" is: a scene label or an entry of a dimension. */
sealed interface AvoidedItem {
    /** The list key. */
    val key: String

    data class Label(val label: SceneLabelTaste) : AvoidedItem {
        override val key: String get() = "label-${label.tag.id}"
    }

    data class Entry(val dimension: TasteDimension, val entry: TasteEntry) : AvoidedItem {
        override val key: String get() = "${dimension.name}-${entry.id}"
    }
}

/** A dimension of the taste with entries the user can answer about. */
enum class TasteDimension(val title: String) {
    PERFORMERS("Performers"),
    STUDIOS("Studios"),
    SITES("Sites"),
    TAGS("Tags"),
}

/**
 * The "Your taste" page split into its sections: the labels the user is into (said yes, or the
 * evidence is sure enough), the liked performers, studios and sites, what they avoid (said not
 * really, or disliked), and labels with little evidence to ask about. A label said no to is never
 * suggested again.
 */
data class TasteSections(
    val into: List<SceneLabelTaste>,
    val liked: Map<TasteDimension, List<TasteEntry>>,
    val avoided: List<AvoidedItem>,
    val questions: List<SceneLabelTaste>,
) {
    companion object {

        /** Splits a taste into the page's sections. */
        fun of(taste: Taste): TasteSections {
            val into = taste.sceneLabels.filter(::isInto).take(INTO_LIMIT)
            val avoidedLabels = taste.sceneLabels.filter(::isAvoided)
            val shown = (into + avoidedLabels).map { label -> label.tag.id }.toSet()
            val questions = taste.sceneLabels
                .filter { label -> label.confirmed == null && label.tag.id !in shown }
                .sortedWith(compareBy<SceneLabelTaste> { label -> label.certainty }.thenByDescending { label -> label.videoCount })
                .take(QUESTION_LIMIT)
            val dimensions = mapOf(
                TasteDimension.PERFORMERS to taste.performers,
                TasteDimension.STUDIOS to taste.studios,
                TasteDimension.SITES to taste.sites,
            )
            val liked = dimensions.mapValues { (_, entries) -> entries.filter(::isLikedEntry).take(ENTRY_LIMIT) }
            val avoidedEntries = (dimensions + (TasteDimension.TAGS to taste.tags.filter { tag -> tag.id !in shown })).flatMap { (dimension, entries) ->
                entries.filter(::isAvoidedEntry).take(ENTRY_LIMIT).map { entry -> AvoidedItem.Entry(dimension, entry) }
            }
            return TasteSections(into, liked, avoidedLabels.map { label -> AvoidedItem.Label(label) } + avoidedEntries, questions)
        }

        /** A label said yes to, or one liked with enough evidence and not said no to. */
        fun isInto(label: SceneLabelTaste): Boolean {
            if (label.confirmed != null) {
                return label.confirmed
            }
            return label.score > 0 && label.certainty >= SURE_CERTAINTY
        }

        /** A label said no to, or one disliked with enough evidence and not said yes to. */
        fun isAvoided(label: SceneLabelTaste): Boolean {
            if (label.confirmed != null) {
                return !label.confirmed
            }
            return label.score <= AVOID_SCORE && label.certainty >= SURE_CERTAINTY
        }

        /** An entry liked and not said no to. */
        private fun isLikedEntry(entry: TasteEntry): Boolean {
            if (entry.confirmed != null) {
                return entry.confirmed
            }
            return entry.weight > 0
        }

        /** An entry said no to, or disliked and not said yes to. */
        private fun isAvoidedEntry(entry: TasteEntry): Boolean {
            if (entry.confirmed != null) {
                return !entry.confirmed
            }
            return entry.weight < 0
        }
    }
}
