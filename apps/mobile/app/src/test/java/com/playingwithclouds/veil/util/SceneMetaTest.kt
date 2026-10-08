package com.playingwithclouds.veil.util

import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.SceneSummary
import java.time.Instant
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SceneMetaTest {

    private val now = Instant.parse("2026-10-08T12:00:00Z")

    private fun scene(
        studio: String? = null,
        performer: String? = null,
        rating: Double? = null,
        viewCount: Int = 0,
        date: String? = null,
    ): SceneSummary {
        var studioRef: EntityRef? = null
        if (studio != null) {
            studioRef = EntityRef("studio:1", studio, null)
        }
        val performers = listOfNotNull(performer).map { name -> EntityRef("performer:$name", name, null) }
        return SceneSummary(
            id = "scene:1",
            title = "T",
            sourceUrl = "https://x/1",
            posterPath = null,
            previewVideo = null,
            previewImages = emptyList(),
            durationSeconds = null,
            date = date,
            studio = studioRef,
            performers = performers,
            tags = emptyList(),
            rating = rating,
            viewCount = viewCount,
        )
    }

    @Test
    fun metaLineJoinsCreditsRatingViewsAndAge() {
        val line = sceneMetaLine(scene("Vixen", "Riley", 9.2, 12_400, "2026-10-05T12:00:00Z"), "eporner.com", now)
        assertEquals("Vixen · Riley · 92% · 12.4K views · 3 days ago", line)
    }

    @Test
    fun missingPartsAreLeftOut() {
        assertEquals("Vixen · 3 days ago", sceneMetaLine(scene("Vixen", date = "2026-10-05T12:00:00Z"), null, now))
    }

    @Test
    fun theSiteStandsInWhenNobodyIsCredited() {
        assertEquals("eporner.com", sceneMetaLine(scene(), "eporner.com", now))
    }

    @Test
    fun nothingKnownGivesNoLine() {
        assertNull(sceneMetaLine(scene(), null, now))
    }

    @Test
    fun performerSameAsStudioIsShownOnce() {
        assertEquals("Riley", sceneMetaLine(scene("Riley", "Riley"), null, now))
    }

    @Test
    fun ratingAndViewsNeedARealNumber() {
        assertNull(formatRatingPercent(0.0))
        assertNull(formatRatingPercent(null))
        assertNull(formatViews(0))
        assertEquals("1 view", formatViews(1))
    }
}
