package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.graphql.type.FilterInput
import com.playingwithclouds.veil.graphql.type.SortField
import com.playingwithclouds.veil.graphql.type.SortInput

/** A choice in one of the filter's single-pick groups. */
data class FilterOption(val value: String, val label: String)

/** A site the filter offers: a whole plugin, or one site of a plugin serving several ("plugin/host"). */
data class SiteChoice(val value: String, val label: String, val iconUrl: String?)

/** A style of video and the offered sites carrying it, picked all at once. */
data class SiteStyle(val key: String, val label: String, val sites: List<String>)

/**
 * The search filter, as the desktop's filter panel has it (`apps/web/src/lib/libraryFilter.ts`):
 * length, quality, the user's rating, unwatched and favourites, sort, the sites asked and the tags
 * hidden (by id, with their names for the chips). Empty values don't narrow.
 */
data class SearchFilter(
    val length: String? = null,
    val minHeight: String? = null,
    val minRating: String? = null,
    val unwatched: Boolean = false,
    val favorite: Boolean = false,
    val sort: String? = null,
    val sites: Set<String> = emptySet(),
    val hiddenTags: Map<String, String> = emptyMap(),
    /** Tags every result must carry (by id, with their names), e.g. POV from the Camera group. */
    val requiredTags: Map<String, String> = emptyMap(),
) {

    /** How many of the sheet's settings are on, for the Filters chip; hidden tags have their own chips. */
    val panelCount: Int
        get() = listOfNotNull(length, minHeight, minRating, sort).size + listOf(unwatched, favorite).count { on -> on } +
            sites.size.coerceAtMost(1) + requiredTags.size

    /** With a required tag added, or taken away when it was required. */
    fun toggledRequiredTag(tag: TagSummary): SearchFilter {
        if (tag.id in requiredTags) {
            return copy(requiredTags = requiredTags - tag.id)
        }
        return copy(requiredTags = requiredTags + (tag.id to tag.name))
    }

    /** Whether anything is set at all. */
    val isEmpty: Boolean
        get() = this == SearchFilter()

    /** The API filter this describes. */
    fun toInput(): FilterInput {
        val bounds = LENGTH_BOUNDS[length]
        val sortField = SORT_FIELDS[sort]
        return FilterInput(
            sources = Optional.presentIfNotNull(sites.toList().ifEmpty { null }),
            excludeTagIds = Optional.presentIfNotNull(hiddenTags.keys.toList().ifEmpty { null }),
            includeTagIds = Optional.presentIfNotNull(requiredTags.keys.toList().ifEmpty { null }),
            minDurationSeconds = Optional.presentIfNotNull(bounds?.first),
            maxDurationSeconds = Optional.presentIfNotNull(bounds?.second),
            minHeight = Optional.presentIfNotNull(minHeight?.toIntOrNull()),
            minRating = Optional.presentIfNotNull(minRating?.toDoubleOrNull()),
            unwatched = Optional.presentIfNotNull(unwatched.takeIf { on -> on }),
            favorite = Optional.presentIfNotNull(favorite.takeIf { on -> on }),
            sort = Optional.presentIfNotNull(sortField?.let { field -> SortInput(field, Optional.present(true)) }),
        )
    }

    /** With one site picked or unpicked. */
    fun toggledSite(site: String): SearchFilter {
        if (site in sites) {
            return copy(sites = sites - site)
        }
        return copy(sites = sites + site)
    }

    /** With every site of a style picked, or all of them unpicked when they all were. */
    fun toggledStyle(style: SiteStyle): SearchFilter {
        if (sites.containsAll(style.sites)) {
            return copy(sites = sites - style.sites.toSet())
        }
        return copy(sites = sites + style.sites)
    }

    /** Whether every site of a style is picked. */
    fun styleOn(style: SiteStyle): Boolean {
        return style.sites.isNotEmpty() && sites.containsAll(style.sites)
    }

    companion object {
        val LENGTH_OPTIONS = listOf(
            FilterOption("short", "Under 10 min"),
            FilterOption("medium", "10–30 min"),
            FilterOption("long", "30–60 min"),
            FilterOption("feature", "Over 1 h"),
        )

        val RESOLUTION_OPTIONS = listOf(
            FilterOption("720", "720p+"),
            FilterOption("1080", "1080p+"),
            FilterOption("2160", "4K"),
        )

        val RATING_OPTIONS = listOf(
            FilterOption("6", "6+"),
            FilterOption("8", "8+"),
            FilterOption("9", "9+"),
        )

        val SORT_OPTIONS = listOf(
            FilterOption("date", "Release date"),
            FilterOption("duration", "Longest"),
            FilterOption("rating", "Your rating"),
            FilterOption("viewCount", "Most watched"),
            FilterOption("random", "Shuffle"),
        )

        /** Seconds each length option spans; null leaves that end open. */
        private val LENGTH_BOUNDS: Map<String?, Pair<Int?, Int?>> = mapOf(
            "short" to (null to 600),
            "medium" to (600 to 1800),
            "long" to (1800 to 3600),
            "feature" to (3600 to null),
        )

        /** Sort values as the API's fields. */
        private val SORT_FIELDS: Map<String?, SortField> = mapOf(
            "date" to SortField.date,
            "duration" to SortField.duration,
            "rating" to SortField.rating,
            "viewCount" to SortField.viewCount,
            "random" to SortField.random,
        )
    }
}

/**
 * The styles the filter picks all sites of at once (`apps/web/src/lib/siteStyles.ts`); a site is
 * named by its filter value. "Tubes" is every offered site no other style claims.
 */
object SiteStyles {

    private val STYLES = listOf(
        SiteStyle("cams", "Cams", listOf("showcamrips", "kvs/cloudbate.com", "kvs/camstreams.tv", "kvs/stream-leak.com", "kvs/camgirlsleak.com")),
        SiteStyle(
            "professional",
            "Professional",
            listOf("aylo", "hqporner", "porn-star", "kvs/pornhat.com", "kvs/perfectgirls.xxx", "kvs/xxxfiles.com"),
        ),
        SiteStyle(
            "amateur",
            "Amateur",
            listOf("kvs/xgroovy.com", "kvs/fapnfuck.com", "kvs/fuqster.com", "kvs/fapgoat.com", "kvs/w1mp.com", "kvs/xozilla.com"),
        ),
        // Ebony, Latina, Asian and the like; only JAV sites are installed so far.
        SiteStyle("exotic", "Exotic", listOf("missav", "kvs/javbangers.com")),
        SiteStyle("fetish", "Fetish", listOf("kvs/severeporn.com", "kvs/analdin.com", "kvs/whoreshub.com")),
    )

    /** The styles among the offered sites, each with the sites it picks, then "Tubes" for the rest; empty styles left out. */
    fun of(choices: List<SiteChoice>): List<SiteStyle> {
        val offered = choices.map { choice -> choice.value }.toSet()
        val claimed = STYLES.flatMap { style -> style.sites }.toSet()
        val styles = STYLES.map { style -> style.copy(sites = style.sites.filter { site -> site in offered }) }
        val tubes = SiteStyle("tubes", "Tubes", choices.map { choice -> choice.value }.filter { value -> value !in claimed })
        return (styles + tubes).filter { style -> style.sites.isNotEmpty() }
    }
}
