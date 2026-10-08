package com.playingwithclouds.veil.ui

import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.FilterPreset
import com.playingwithclouds.veil.data.RuntimeWindow
import com.playingwithclouds.veil.data.SceneRow
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.home.HomeFilter
import com.playingwithclouds.veil.ui.home.HomeShelf
import com.playingwithclouds.veil.ui.home.RuntimeChip
import com.playingwithclouds.veil.ui.home.buildShelves
import com.playingwithclouds.veil.ui.home.suggestedPresetName
import com.playingwithclouds.veil.ui.home.topTags
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class HomeFilterTest {

    private fun scene(
        id: String,
        tags: List<String> = emptyList(),
        performers: List<String> = emptyList(),
        studio: String? = null,
    ): SceneSummary {
        var studioRef: EntityRef? = null
        if (studio != null) {
            studioRef = EntityRef("studio:$studio", studio, null)
        }
        return SceneSummary(
            id = id,
            title = id,
            sourceUrl = "https://x/$id",
            posterPath = null,
            previewVideo = null,
            previewImages = emptyList(),
            durationSeconds = null,
            date = null,
            studio = studioRef,
            performers = performers.map { name -> EntityRef("performer:$name", name, "https://img/$name.jpg") },
            tags = tags.map { name -> EntityRef("tag:$name", name, null) },
        )
    }

    @Test
    fun filterIsInactiveUntilSomethingIsPicked() {
        assertFalse(HomeFilter().isActive)
        assertTrue(HomeFilter().toggleSource("eporner").isActive)
        assertTrue(HomeFilter().toggleRuntime(RuntimeChip.SHORT).isActive)
    }

    @Test
    fun togglingTwiceRestoresTheFilter() {
        val filter = HomeFilter()
        assertEquals(filter, filter.toggleSource("eporner").toggleSource("eporner"))
        assertEquals(filter, filter.toggleTag("tag:a").toggleTag("tag:a"))
        assertEquals(filter, filter.toggleRuntime(RuntimeChip.LONG).toggleRuntime(RuntimeChip.LONG))
    }

    @Test
    fun runtimeChipsMapToWindows() {
        assertEquals(RuntimeChip.MEDIUM, RuntimeChip.matching(RuntimeWindow(600, 1800)))
        assertNull(RuntimeChip.matching(RuntimeWindow(1, 2)))
    }

    @Test
    fun presetRoundTripsThroughTheFilter() {
        val preset = FilterPreset("saved_filter:1", "Mid", "tag:a", listOf("eporner"), RuntimeChip.MEDIUM.window)
        val filter = HomeFilter.of(preset)
        assertTrue(filter.matches(preset))
        assertFalse(filter.toggleSource("xhamster").matches(preset))
    }

    @Test
    fun topTagsAreMostFrequentFirst() {
        val scenes = listOf(scene("1", tags = listOf("b", "a")), scene("2", tags = listOf("b")), scene("3", tags = listOf("c", "b")))
        assertEquals(listOf("b", "a", "c"), topTags(scenes, 5).map { tag -> tag.name })
        assertEquals(listOf("b"), topTags(scenes, 1).map { tag -> tag.name })
    }

    @Test
    fun presetNameListsTheParts() {
        val filter = HomeFilter(setOf("xhamster", "eporner"), "tag:a", RuntimeChip.MEDIUM.window)
        assertEquals("Anal · 10–30 min · eporner · xhamster", suggestedPresetName(filter, "Anal"))
        assertEquals("My filter", suggestedPresetName(HomeFilter(), null))
    }

    @Test
    fun shelvesAlternateSceneRowsWithPerformersAndStudios() {
        val feed = (1..6).map { number ->
            scene("s$number", performers = listOf("p${number % 5}", "q$number"), studio = "st$number")
        }
        val rows = listOf("one", "two", "three").map { title -> SceneRow("k-$title", title, feed.take(3)) }
        val shelves = buildShelves(rows, feed)
        assertEquals(
            listOf("k-one", "performers", "k-two", "studios", "k-three"),
            shelves.map { shelf -> shelf.key },
        )
        assertTrue(shelves[1] is HomeShelf.Performers)
    }

    @Test
    fun thinEntityRowsAreLeftOut() {
        val shelves = buildShelves(emptyList(), listOf(scene("1", performers = listOf("a"), studio = "x")))
        assertTrue(shelves.isEmpty())
    }
}
