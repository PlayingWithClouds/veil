package com.playingwithclouds.veil.api

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class BackendUrlsTest {

    private val phoneBackend = "http://127.0.0.1:47831"
    private val remoteBackend = "http://192.168.1.10:8080"

    @Test
    fun normalizeTrimsWhitespaceAndTrailingSlashes() {
        assertEquals("http://nas:8080", BackendUrls.normalizeServerUrl("  http://nas:8080// "))
    }

    @Test
    fun normalizeDefaultsToHttp() {
        assertEquals("http://nas.local:8080", BackendUrls.normalizeServerUrl("nas.local:8080"))
    }

    @Test
    fun normalizeKeepsHttps() {
        assertEquals("https://veil.example.com", BackendUrls.normalizeServerUrl("https://veil.example.com/"))
    }

    @Test
    fun rebaseMovesLoopbackApiUrlsOntoTheActiveServer() {
        val rebased = BackendUrls.rebase("http://localhost:8080/api/blob/posters/a.jpg?v=2", remoteBackend)
        assertEquals("http://192.168.1.10:8080/api/blob/posters/a.jpg?v=2", rebased)
    }

    @Test
    fun rebaseHandlesEveryLoopbackSpelling() {
        for (host in listOf("localhost", "127.0.0.1", "[::1]")) {
            val rebased = BackendUrls.rebase("http://$host:8080/api/stream/x.m3u8", remoteBackend)
            assertEquals("http://192.168.1.10:8080/api/stream/x.m3u8", rebased)
        }
    }

    @Test
    fun rebaseLeavesUrlsAlreadyOnTheActiveServer() {
        val url = "http://127.0.0.1:47831/api/stream/x.m3u8"
        assertEquals(url, BackendUrls.rebase(url, phoneBackend))
    }

    @Test
    fun rebaseTreatsDefaultPortsAsEqual() {
        val url = "http://localhost/api/blob/a.jpg"
        assertEquals(url, BackendUrls.rebase(url, "http://localhost:80"))
    }

    @Test
    fun rebaseLeavesForeignHostsAlone() {
        val url = "https://cdn.example.com/api/blob/a.jpg"
        assertEquals(url, BackendUrls.rebase(url, remoteBackend))
    }

    @Test
    fun rebaseLeavesNonApiPathsAlone() {
        val url = "http://localhost:8080/graphql"
        assertEquals(url, BackendUrls.rebase(url, remoteBackend))
    }

    @Test
    fun rebaseLeavesRelativePathsAlone() {
        assertEquals("/api/blob/a.jpg", BackendUrls.rebase("/api/blob/a.jpg", remoteBackend))
    }

    @Test
    fun imageUrlIsNullForMissingImages() {
        assertNull(BackendUrls.imageUrl(null, phoneBackend))
        assertNull(BackendUrls.imageUrl("", phoneBackend))
    }

    @Test
    fun imageUrlRoutesRemoteImagesThroughTheImageCache() {
        val url = BackendUrls.imageUrl("https://cdn.example.com/p/1.jpg?x=1&y=2", phoneBackend)
        assertEquals("http://127.0.0.1:47831/api/img?url=https%3A%2F%2Fcdn.example.com%2Fp%2F1.jpg%3Fx%3D1%26y%3D2", url)
    }

    @Test
    fun imageUrlRebasesBackendBlobUrlsInsteadOfProxyingThem() {
        val url = BackendUrls.imageUrl("http://localhost:8080/api/blob/images/a.jpg", remoteBackend)
        assertEquals("http://192.168.1.10:8080/api/blob/images/a.jpg", url)
    }

    @Test
    fun imageUrlPassesRelativePathsThrough() {
        assertEquals("/api/blob/a.jpg", BackendUrls.imageUrl("/api/blob/a.jpg", remoteBackend))
    }

    @Test
    fun webSocketUrlSwitchesTheScheme() {
        assertEquals("ws://192.168.1.10:8080/graphql", BackendUrls.webSocketUrl(remoteBackend))
        assertEquals("wss://veil.example.com/graphql", BackendUrls.webSocketUrl("https://veil.example.com"))
    }
}
