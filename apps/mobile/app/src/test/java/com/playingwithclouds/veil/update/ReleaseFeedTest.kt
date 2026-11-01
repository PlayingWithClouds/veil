package com.playingwithclouds.veil.update

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ReleaseFeedTest {

    private val release = """
        {
          "tag_name": "v0.0.3",
          "name": "Veil 0.0.3",
          "body": "- Self-updater",
          "assets": [
            {"name": "veil-0.0.3.apk", "size": 111, "browser_download_url": "https://example.test/veil-0.0.3.apk"},
            {"name": "veil.apk", "size": 222, "browser_download_url": "https://example.test/veil.apk"}
          ]
        }
    """.trimIndent()

    @Test
    fun readsTagNotesAndTheStableApk() {
        val parsed = ReleaseFeed.parse(release)
        assertEquals(
            Release(
                tagName = "v0.0.3",
                name = "Veil 0.0.3",
                notes = "- Self-updater",
                apkUrl = "https://example.test/veil.apk",
                apkSizeBytes = 222,
            ),
            parsed,
        )
    }

    @Test
    fun missingNameAndBodyFallBack() {
        val json = """
            {"tag_name": "v0.0.3", "name": null, "body": null,
             "assets": [{"name": "veil.apk", "size": 1, "browser_download_url": "https://example.test/veil.apk"}]}
        """.trimIndent()
        val parsed = ReleaseFeed.parse(json)
        assertEquals("v0.0.3", parsed?.name)
        assertEquals("", parsed?.notes)
    }

    @Test
    fun releaseWithoutTheApkIsNothing() {
        val json = """
            {"tag_name": "plugins", "assets": [{"name": "kvs.zip", "size": 1, "browser_download_url": "https://example.test/kvs.zip"}]}
        """.trimIndent()
        assertNull(ReleaseFeed.parse(json))
        assertNull(ReleaseFeed.parse("""{"tag_name": "v0.0.3"}"""))
    }

    @Test
    fun unreadableResponsesAreNothing() {
        assertNull(ReleaseFeed.parse("not json"))
        assertNull(ReleaseFeed.parse("""{"message": "Not Found"}"""))
    }
}
