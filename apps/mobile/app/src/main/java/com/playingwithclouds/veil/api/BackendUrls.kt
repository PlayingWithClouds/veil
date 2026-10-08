package com.playingwithclouds.veil.api

import java.net.URI
import java.net.URISyntaxException
import java.net.URLEncoder

/** Pure URL rules for reaching the backend; mirror `lib/server.ts` and `lib/img.ts` of the web client. */
object BackendUrls {

    /** Hosts that only mean "this machine", so their URLs are useless to other devices. */
    private val loopbackHosts = listOf("localhost", "127.0.0.1", "[::1]", "::1")

    /** Trims whitespace and trailing slashes and defaults the scheme to http. */
    fun normalizeServerUrl(url: String): String {
        var normalized = url.trim().trimEnd('/')
        if (!normalized.startsWith("http://") && !normalized.startsWith("https://")) {
            normalized = "http://$normalized"
        }
        return normalized
    }

    /**
     * Points a backend-built URL at the address the client reaches the backend on. The backend
     * stamps blob and stream URLs with its PUBLIC_URL, which defaults to localhost; from another
     * device that host is unreachable. URLs of other hosts and non-`/api/` paths stay as they are.
     */
    fun rebase(url: String, baseUrl: String): String {
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            return url
        }
        val parsed = parse(url) ?: return url
        if (parsed.host !in loopbackHosts) {
            return url
        }
        if (!parsed.rawPath.startsWith("/api/")) {
            return url
        }
        val base = parse(baseUrl) ?: return url
        if (originOf(parsed) == originOf(base)) {
            return url
        }
        return baseUrl + parsed.rawPath + queryOf(parsed)
    }

    /**
     * Routes a remote image through the backend image cache: adult-source CDNs hotlink-protect
     * their images, so the app can't load them directly. Relative paths pass through and the
     * backend's own blob URLs are only pointed at its reachable address.
     */
    fun imageUrl(url: String?, baseUrl: String): String? {
        if (url.isNullOrEmpty()) {
            return null
        }
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            return url
        }
        val rebased = rebase(url, baseUrl)
        if (rebased != url) {
            return rebased
        }
        return "$baseUrl/api/img?url=" + URLEncoder.encode(url, "UTF-8")
    }

    /** The GraphQL-over-WebSocket address that belongs to an HTTP backend address. */
    fun webSocketUrl(baseUrl: String): String {
        return baseUrl.replaceFirst("http", "ws") + "/graphql"
    }

    /** Parses a URL, or null when it is malformed. */
    private fun parse(url: String): URI? {
        return try {
            URI(url)
        } catch (error: URISyntaxException) {
            null
        }
    }

    /** Scheme, host and effective port, which together identify who serves a URL. */
    private fun originOf(uri: URI): String {
        var port = uri.port
        if (port == -1) {
            port = defaultPort(uri.scheme)
        }
        return "${uri.scheme}://${uri.host}:$port"
    }

    /** The port a scheme uses when the URL names none. */
    private fun defaultPort(scheme: String): Int {
        if (scheme == "https") {
            return 443
        }
        return 80
    }

    /** The URL's query string including the leading question mark, or empty. */
    private fun queryOf(uri: URI): String {
        val query = uri.rawQuery ?: return ""
        return "?$query"
    }
}
