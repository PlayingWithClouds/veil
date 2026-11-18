package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.graphql.type.SortField
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class SearchFilterTest {

    @Test
    fun anEmptyFilterNarrowsNothing() {
        val input = SearchFilter().toInput()
        assertEquals(Optional.Absent, input.sources)
        assertEquals(Optional.Absent, input.minDurationSeconds)
        assertEquals(Optional.Absent, input.unwatched)
        assertEquals(Optional.Absent, input.sort)
    }

    @Test
    fun aLengthBecomesItsDurationBounds() {
        val medium = SearchFilter(length = "medium").toInput()
        assertEquals(Optional.present(600), medium.minDurationSeconds)
        assertEquals(Optional.present(1800), medium.maxDurationSeconds)
        val feature = SearchFilter(length = "feature").toInput()
        assertEquals(Optional.present(3600), feature.minDurationSeconds)
        assertEquals(Optional.Absent, feature.maxDurationSeconds)
    }

    @Test
    fun theOtherSettingsMapToTheirFields() {
        val input = SearchFilter(
            minHeight = "1080",
            minRating = "8",
            unwatched = true,
            sort = "duration",
            sites = setOf("xvideos"),
            hiddenTags = mapOf("t1" to "Solo"),
        ).toInput()
        assertEquals(Optional.present(1080), input.minHeight)
        assertEquals(Optional.present(8.0), input.minRating)
        assertEquals(Optional.present(true), input.unwatched)
        assertEquals(SortField.duration, input.sort.getOrNull()?.field)
        assertEquals(Optional.present(listOf("xvideos")), input.sources)
        assertEquals(Optional.present(listOf("t1")), input.excludeTagIds)
    }

    @Test
    fun aStylePicksAllItsSitesAndUnpicksThemWhenAllWere() {
        val style = SiteStyle("cams", "Cams", listOf("showcamrips", "kvs/cloudbate.com"))
        val picked = SearchFilter(sites = setOf("showcamrips")).toggledStyle(style)
        assertTrue(picked.styleOn(style))
        assertFalse(picked.toggledStyle(style).styleOn(style))
        assertEquals(emptySet<String>(), picked.toggledStyle(style).sites)
    }

    @Test
    fun stylesKeepOnlyOfferedSitesAndTubesTakeTheRest() {
        val choices = listOf(
            SiteChoice("showcamrips", "ShowCamRips", null),
            SiteChoice("xvideos", "XVideos", null),
            SiteChoice("kvs/xgroovy.com", "XGroovy", null),
        )
        val styles = SiteStyles.of(choices).associate { style -> style.key to style.sites }
        assertEquals(listOf("showcamrips"), styles["cams"])
        assertEquals(listOf("kvs/xgroovy.com"), styles["amateur"])
        assertEquals(listOf("xvideos"), styles["tubes"])
        assertFalse("exotic" in styles)
    }

    @Test
    fun theFiltersChipCountsSitesAsOneSetting() {
        assertEquals(3, SearchFilter(length = "short", unwatched = true, sites = setOf("a", "b")).panelCount)
    }

    @Test
    fun sceneLabelsToggleAndBecomeSceneTagIds() {
        val pov = TagSummary("t1", "POV")
        val filter = SearchFilter().toggledSceneLabel(pov)
        assertEquals(Optional.present(listOf("t1")), filter.toInput().sceneTagIds)
        assertTrue(filter.toggledSceneLabel(pov).isEmpty)
    }
}
