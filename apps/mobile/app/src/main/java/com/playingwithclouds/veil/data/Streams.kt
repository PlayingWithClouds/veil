package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.graphql.fragment.StreamFields

/** A playback source of a scene. */
data class StreamOption(
    val id: String,
    val url: String,
    val kind: String,
    val label: String?,
    val provider: String?,
    val resolution: String?,
    val height: Int?,
    val format: String?,
    val mimeType: String?,
    val expectedSpeedBps: Double?,
    val verified: Boolean,
    val pluginName: String?,
)

/** A scene's sources from one site, qualities sorted best-first. */
data class StreamGroup(val provider: String, val streams: List<StreamOption>)

/** A stream resolved to something the player can load. */
data class PlayableStream(val url: String, val mimeType: String, val headers: Map<String, String>) {

    /** Whether the source is HLS rather than a progressive file. */
    val isHls: Boolean
        get() = mimeType.contains("mpegURL", ignoreCase = true) || url.contains(".m3u8")
}

/** Maps the stream fragment to the stream model. */
fun StreamFields.toOption(): StreamOption {
    return StreamOption(
        id = id,
        url = url,
        kind = kind,
        label = label,
        provider = provider,
        resolution = resolution,
        height = height,
        format = format,
        mimeType = mimeType,
        expectedSpeedBps = expectedSpeedBps,
        verified = verified,
        pluginName = pluginName,
    )
}

/** The site a stream came from, used to group qualities under one entry. */
fun StreamOption.providerName(): String {
    val plugin = pluginName
    if (!plugin.isNullOrEmpty()) {
        return plugin
    }
    val hostingProvider = provider
    if (!hostingProvider.isNullOrEmpty()) {
        return hostingProvider
    }
    return "Source"
}

/** Approximate vertical resolution, for ranking qualities best-first. */
fun StreamOption.verticalResolution(): Int {
    val pixels = height
    if (pixels != null && pixels > 0) {
        return pixels
    }
    val digits = resolution.orEmpty().takeWhile { character -> character.isDigit() }
    return digits.toIntOrNull() ?: 0
}

/**
 * The per-quality label shown next to a provider's name. "Auto" when the site doesn't say:
 * the player picks whatever the page or playlist serves.
 */
fun StreamOption.qualityLabel(): String {
    if (!resolution.isNullOrEmpty()) {
        return resolution
    }
    val pixels = height
    if (pixels != null && pixels > 0) {
        return "${pixels}p"
    }
    // A label that just repeats the site name says nothing next to it.
    if (!label.isNullOrEmpty() && !label.equals(providerName(), ignoreCase = true)) {
        return label
    }
    if (!format.isNullOrEmpty()) {
        return format.uppercase()
    }
    return "Auto"
}

/** Groups streams by source site, each group's qualities sorted best-first. */
fun groupStreams(streams: List<StreamOption>): List<StreamGroup> {
    return streams
        .groupBy { stream -> stream.providerName() }
        .map { (provider, members) ->
            StreamGroup(provider, members.sortedByDescending { stream -> stream.verticalResolution() })
        }
}
