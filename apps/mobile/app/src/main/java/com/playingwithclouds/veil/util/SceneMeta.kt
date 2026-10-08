package com.playingwithclouds.veil.util

import com.playingwithclouds.veil.data.SceneSummary
import java.time.Instant
import kotlin.math.roundToInt

/** The site's rating (0 to 10) as a percentage like "92%"; null when there is none. */
fun formatRatingPercent(rating: Double?): String? {
    if (rating == null || rating <= 0.0) {
        return null
    }
    return "${(rating * 10).roundToInt()}%"
}

/** "12K views"; null when the site sent no counter. */
fun formatViews(viewCount: Int): String? {
    if (viewCount <= 0) {
        return null
    }
    if (viewCount == 1) {
        return "1 view"
    }
    return "${formatCount(viewCount)} views"
}

/**
 * The muted line under a scene title: studio, first performer, site rating, site views, then the
 * video's age ("Studio · Performer · 92% · 12K views · 3 days ago"). Parts the scene lacks are
 * left out; [fallbackSource] (the site's name) stands in when neither studio nor performer is known.
 */
fun sceneMetaLine(scene: SceneSummary, fallbackSource: String?, now: Instant = Instant.now()): String? {
    val credits = mutableListOf<String>()
    scene.studio?.let { studio -> credits.add(studio.name) }
    scene.performers.firstOrNull()?.let { performer ->
        if (!credits.contains(performer.name)) {
            credits.add(performer.name)
        }
    }
    if (credits.isEmpty() && fallbackSource != null) {
        credits.add(fallbackSource)
    }
    val parts = credits + listOfNotNull(
        formatRatingPercent(scene.rating),
        formatViews(scene.viewCount),
        formatReleaseDate(scene.date, now),
    )
    if (parts.isEmpty()) {
        return null
    }
    return parts.joinToString(" · ")
}
