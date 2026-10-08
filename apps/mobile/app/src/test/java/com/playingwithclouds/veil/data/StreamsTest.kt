package com.playingwithclouds.veil.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class StreamsTest {

    private fun stream(
        id: String,
        pluginName: String? = "eporner",
        provider: String? = null,
        resolution: String? = null,
        height: Int? = null,
        label: String? = null,
        format: String? = null,
    ): StreamOption {
        return StreamOption(id, "https://x/$id", "stream", label, provider, resolution, height, format, null, null, false, pluginName)
    }

    @Test
    fun providerPrefersThePluginThenTheHostingProvider() {
        assertEquals("eporner", stream("1").providerName())
        assertEquals("VOE", stream("2", pluginName = null, provider = "VOE").providerName())
        assertEquals("Source", stream("3", pluginName = null).providerName())
    }

    @Test
    fun qualityUsesResolutionThenHeightThenLabelThenFormat() {
        assertEquals("1080p", stream("1", resolution = "1080p").qualityLabel())
        assertEquals("720p", stream("2", height = 720).qualityLabel())
        assertEquals("HD", stream("3", label = "HD").qualityLabel())
        assertEquals("MP4", stream("4", format = "mp4").qualityLabel())
        assertEquals("Auto", stream("5").qualityLabel())
    }

    @Test
    fun aLabelRepeatingTheSiteNameSaysNothing() {
        assertEquals("Auto", stream("1", label = "Eporner").qualityLabel())
    }

    @Test
    fun verticalResolutionFallsBackToTheResolutionText() {
        assertEquals(480, stream("1", resolution = "480p").verticalResolution())
        assertEquals(1080, stream("2", height = 1080, resolution = "720p").verticalResolution())
        assertEquals(0, stream("3", resolution = "HD").verticalResolution())
    }

    @Test
    fun groupsHoldQualitiesBestFirst() {
        val groups = groupStreams(
            listOf(
                stream("a", pluginName = "eporner", height = 480),
                stream("b", pluginName = "xhamster", height = 720),
                stream("c", pluginName = "eporner", height = 1080),
            ),
        )
        assertEquals(listOf("eporner", "xhamster"), groups.map { group -> group.provider })
        assertEquals(listOf("c", "a"), groups[0].streams.map { option -> option.id })
    }

    @Test
    fun hlsIsRecognizedByMimeTypeOrPlaylistExtension() {
        assertTrue(PlayableStream("https://x/v", "application/vnd.apple.mpegURL", emptyMap()).isHls)
        assertTrue(PlayableStream("https://x/v.m3u8?t=1", "video/mp4", emptyMap()).isHls)
        assertFalse(PlayableStream("https://x/v.mp4", "video/mp4", emptyMap()).isHls)
    }
}
