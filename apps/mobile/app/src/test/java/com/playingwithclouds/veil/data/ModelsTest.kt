package com.playingwithclouds.veil.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ModelsTest {

    private fun scene(studio: String? = null, performers: List<String> = emptyList()): SceneSummary {
        var studioRef: EntityRef? = null
        if (studio != null) {
            studioRef = EntityRef("studio:1", studio, null)
        }
        return SceneSummary(
            id = "scene:1",
            title = "T",
            sourceUrl = "https://x/1",
            posterPath = null,
            previewVideo = null,
            previewImages = emptyList(),
            durationSeconds = null,
            date = null,
            studio = studioRef,
            performers = performers.map { name -> EntityRef("performer:$name", name, null) },
            tags = emptyList(),
        )
    }

    @Test
    fun bylinePrefersTheStudio() {
        assertEquals("Brazzers", scene(studio = "Brazzers", performers = listOf("A")).byline)
    }

    @Test
    fun bylineFallsBackToAtMostTwoPerformers() {
        assertEquals("A, B", scene(performers = listOf("A", "B", "C")).byline)
        assertNull(scene().byline)
    }

    @Test
    fun memberTypesAreMatchedIgnoringCase() {
        assertEquals(MemberType.SCENE, MemberType.fromBackend("scene"))
        assertEquals(MemberType.PERFORMER, MemberType.fromBackend("Performer"))
        assertNull(MemberType.fromBackend("collection"))
    }

    @Test
    fun verdictFollowsTheRatingAroundFive() {
        assertEquals(Verdict.UP, Verdict.fromRating(10.0))
        assertEquals(Verdict.DOWN, Verdict.fromRating(1.0))
        assertNull(Verdict.fromRating(5.0))
        assertNull(Verdict.fromRating(null))
    }

    @Test
    fun markerTitleJoinsTagAndLabel() {
        assertEquals("Doggy - nice", SceneMarker("1", "Doggy", "nice", 3.0, null, true).title)
        assertEquals("Doggy", SceneMarker("1", "Doggy", null, 3.0, null, true).title)
        assertEquals("nice", SceneMarker("1", null, "nice", 3.0, null, true).title)
    }

    @Test
    fun subscriptionNameIsTheTargetsCurrentName() {
        val followed = subscription(query = "old name", targetName = "New Name")
        assertEquals("New Name", followed.displayName)
        assertEquals("a search", subscription(query = "a search", targetName = null).displayName)
    }

    private fun subscription(query: String, targetName: String?): SubscriptionSummary {
        return SubscriptionSummary("s", "SEARCH", query, null, targetName, null, emptyList(), 6, true, null, null, null, 0, 0)
    }

    @Test
    fun downloadStatesSeparateActiveFailedAndDone() {
        fun job(status: String) = DownloadJob("1", status, "T", null, null, null, null, null)
        assertTrue(job("running").isActive)
        assertTrue(job("pending").isActive)
        assertFalse(job("completed").isActive)
        assertFalse(job("completed").isFailed)
        assertTrue(job("failed").isFailed)
        assertTrue(job("cancelled").isFailed)
    }

    @Test
    fun downloadTitleFallsBackFromSceneToJobToPlaceholder() {
        assertEquals("T", JobRepository.downloadTitle("job", scene()))
        assertEquals("job", JobRepository.downloadTitle("job", null))
        assertEquals("Download", JobRepository.downloadTitle(null, null))
        assertEquals("Download", JobRepository.downloadTitle("", null))
    }

    @Test
    fun progressFractionIsClampedAndNeedsARuntime() {
        assertEquals(0.5f, SceneProgress(30, 60).fraction!!, 0.001f)
        assertEquals(1f, SceneProgress(90, 60).fraction!!, 0.001f)
        assertNull(SceneProgress(30, null).fraction)
        assertNull(SceneProgress(30, 0).fraction)
    }

    @Test
    fun onlyUnfinishedWatchesDrawAResumeBar() {
        val entries = listOf(
            WatchHistoryEntry("scene:1", 30, 60, completed = false, scene = null),
            WatchHistoryEntry("scene:2", 60, 60, completed = true, scene = null),
            WatchHistoryEntry("scene:3", 0, 60, completed = false, scene = null),
        )
        assertEquals(setOf("scene:1"), WatchProgressStore.unfinished(entries).keys)
    }

    @Test
    fun sitesAreMatchedByDomainOrName() {
        val sites = listOf(
            SearchSite("eporner", "EPORNER", "icon", listOf("eporner.com")),
            SearchSite("resolver", null, null, listOf("*")),
        )
        assertEquals(SceneSite("EPORNER", "icon"), SceneSites.resolve("https://www.eporner.com/video-1", sites))
        assertEquals(SceneSite("cdn.example.org", null), SceneSites.resolve("https://cdn.example.org/a", sites))
        assertEquals("eporner.com", SceneSites.hostOf("https://www.eporner.com/x"))
        assertNull(SceneSites.resolve("not a url", sites))
    }
}
