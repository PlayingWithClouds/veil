package com.playingwithclouds.veil.data

/** A tag as a video lists it. */
data class TagSummary(val id: String, val name: String)

/** A site page the video was found on. */
data class SiteSource(val plugin: String, val url: String)

/**
 * Something the player can play: a local copy (download, cam recording) or a site's stream, which
 * has to be resolved first. [provider] is "local" or the site's plugin.
 */
data class PlayerSource(val id: String, val provider: String, val quality: String?, val streamUrl: String?, val local: PlayableStream?)

/** A source resolved to something the player can load. */
data class PlayableStream(val url: String, val mimeType: String) {

    /** Whether it's HLS rather than a progressive file. */
    val isHls: Boolean
        get() = mimeType.contains("mpegURL", ignoreCase = true) || url.contains(".m3u8")
}

/**
 * A stretch of a video with a name, e.g. an act; [endSeconds] is null when only its start is known.
 * [origin] is who made it (user, plugin, ai, as the API spells it); [tag] is its tag's name.
 */
data class VideoSegment(
    val id: String,
    val title: String,
    val seconds: Double,
    val endSeconds: Double?,
    val origin: String = "user",
    val tag: String? = null,
)

/**
 * Everything the video page shows. [sources] lists local copies first, then the sites' streams;
 * [rating] is 0..10; [progressSeconds] is the resume point; [previewFrames] are the site's stills
 * across the video, in time order.
 */
data class VideoDetail(
    val summary: VideoSummary,
    val details: String?,
    val width: Int?,
    val height: Int?,
    val studio: PerformerSummary?,
    val performers: List<PerformerSummary>,
    val tags: List<TagSummary>,
    val sites: List<SiteSource>,
    val sources: List<PlayerSource>,
    val segments: List<VideoSegment>,
    val favorite: Boolean,
    val rating: Double?,
    val oCount: Int,
    val galleryId: String?,
    val previewFrames: List<String> = emptyList(),
) {
    val id: String get() = summary.id
    val title: String get() = summary.title
}
