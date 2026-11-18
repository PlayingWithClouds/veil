package com.playingwithclouds.veil.ui.taste

import com.playingwithclouds.veil.data.SceneLabelTaste
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.data.Taste
import com.playingwithclouds.veil.data.TasteEntry
import org.junit.Assert.assertEquals
import org.junit.Test

class TasteSectionsTest {

    /** A scene label with an id from its name. */
    private fun label(name: String, score: Double, certainty: Double, confirmed: Boolean? = null): SceneLabelTaste {
        return SceneLabelTaste(TagSummary(name.lowercase(), name), score, certainty, emptyList(), confirmed, 3)
    }

    /** A performer, studio or site entry. */
    private fun entry(id: String, weight: Double, confirmed: Boolean? = null): TasteEntry {
        return TasteEntry(id, id, null, weight, confirmed)
    }

    @Test
    fun answersOverruleTheEvidence() {
        val taste = Taste(
            signalCount = 10,
            tags = emptyList(),
            performers = listOf(entry("liked", 0.8), entry("refused", 0.5, confirmed = false), entry("disliked", -0.4)),
            studios = emptyList(),
            sites = listOf(entry("tube", -0.2, confirmed = true)),
            sceneLabels = listOf(
                label("Doggystyle", 0.6, 0.9),
                label("Cowgirl", 0.4, 0.9, confirmed = false),
                label("POV", -0.5, 0.1, confirmed = true),
                label("Kissing", -0.6, 0.8),
                label("Anal", 0.2, 0.1),
            ),
        )
        val sections = TasteSections.of(taste)
        assertEquals(listOf("Doggystyle", "POV"), sections.into.map { label -> label.tag.name })
        assertEquals(listOf("Anal"), sections.questions.map { label -> label.tag.name })
        assertEquals(listOf("liked"), sections.liked.getValue(TasteDimension.PERFORMERS).map { entry -> entry.id })
        assertEquals(listOf("tube"), sections.liked.getValue(TasteDimension.SITES).map { entry -> entry.id })
        assertEquals(listOf("label-cowgirl", "label-kissing", "PERFORMERS-refused", "PERFORMERS-disliked"), sections.avoided.map { item -> item.key })
    }
}
