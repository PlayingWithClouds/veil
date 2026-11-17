package com.playingwithclouds.veil.data

/**
 * A video as a card shows it. [byline] and [avatarUrl] are who it's from: the studio, else the
 * first performer. [progressSeconds] is the resume point while it's being watched. [performers],
 * [studio] and [favorite] are for the card's long-press actions.
 */
data class VideoSummary(
    val id: String,
    val title: String,
    val coverUrl: String?,
    val previewUrl: String?,
    val durationSeconds: Double?,
    val progressSeconds: Double?,
    val byline: String?,
    val avatarUrl: String?,
    val date: String?,
    val performers: List<PerformerSummary> = emptyList(),
    val studio: PerformerSummary? = null,
    val favorite: Boolean = false,
)

/** A recommended video and why it's picked, e.g. "Because you watched …". */
data class RecommendedVideo(val video: VideoSummary, val reason: String?)

/** A cam room as a card shows it; [viewers] is null while it's offline. */
data class CamSummary(
    val id: String,
    val name: String,
    val site: String,
    val pictureUrl: String?,
    val viewers: Int?,
    val recording: Boolean,
)

/** A story as a card shows it; [author] is its first performer. */
data class StorySummary(
    val id: String,
    val title: String,
    val coverUrl: String?,
    val author: String?,
    val chapterCount: Int,
)

/** An audio as a card shows it; [voice] is its first performer. */
data class AudioSummary(
    val id: String,
    val title: String,
    val coverUrl: String?,
    val voice: String?,
    val durationSeconds: Double?,
)

/** A performer as a round picture with a name. */
data class PerformerSummary(val id: String, val name: String, val imageUrl: String?)
