package com.playingwithclouds.veil.data

import java.net.URI
import java.net.URISyntaxException

/** The site a scene was scraped from, with the matching plugin's icon when known. */
data class SceneSite(val name: String, val iconUrl: String?)

/** Works out which plugin site a scene's source URL belongs to. */
object SceneSites {

    /** Hostname of a URL without a leading "www.", or null when it does not parse. */
    fun hostOf(url: String): String? {
        return try {
            URI(url).host?.removePrefix("www.")
        } catch (error: URISyntaxException) {
            null
        }
    }

    /** Whether a plugin serves a host: by a listed domain, or by its name in the host. */
    fun servesHost(site: SearchSite, host: String): Boolean {
        val byDomain = site.domains.any { domain -> domain != "*" && (host == domain || host.endsWith(".$domain")) }
        if (byDomain) {
            return true
        }
        return host.contains(site.name)
    }

    /** The site of a source URL: the serving plugin's name and icon, else the bare host. */
    fun resolve(sourceUrl: String, sites: List<SearchSite>): SceneSite? {
        val host = hostOf(sourceUrl) ?: return null
        val match = sites.firstOrNull { site -> servesHost(site, host) } ?: return SceneSite(host, null)
        return SceneSite(match.label, match.iconUrl)
    }
}
