package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.LiveSearchHit
import com.playingwithclouds.veil.ui.search.SearchState
import com.playingwithclouds.veil.ui.search.withLiveResult
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class LiveResultTest {

    private fun hit(mediaType: String, recordId: String?): LiveSearchHit {
        return LiveSearchHit("eporner", recordId, "Title", mediaType, "https://x/1", "1", "2024-01-01", "poster", listOf("a"), null)
    }

    @Test
    fun storedScenesBecomeSceneCards() {
        val result = SearchRepository.toLiveResult(hit("scene", "scene:1")) as LiveResultItem.SceneResult
        assertEquals("scene:1", result.scene.id)
        assertEquals("Title", result.scene.title)
        assertEquals("poster", result.scene.posterPath)
        assertEquals(listOf("a"), result.scene.previewImages)
    }

    @Test
    fun storedGalleriesBecomeGalleryCards() {
        val result = SearchRepository.toLiveResult(hit("gallery", "gallery:1")) as LiveResultItem.GalleryResult
        assertEquals("gallery:1", result.gallery.id)
        assertEquals("poster", result.gallery.coverPath)
    }

    @Test
    fun hitsThatWereNotStoredOrAreOtherTypesAreSkipped() {
        assertNull(SearchRepository.toLiveResult(hit("scene", null)))
        assertNull(SearchRepository.toLiveResult(hit("performer", "performer:1")))
    }

    @Test
    fun liveResultsAreAddedOnlyOnce() {
        val result = SearchRepository.toLiveResult(hit("scene", "scene:1"))!!
        val once = SearchState().withLiveResult(result)
        val twice = once.withLiveResult(result)
        assertEquals(1, twice.scenes.size)
        assertTrue(twice.galleries.isEmpty())
    }
}
