package com.playingwithclouds.veil.api

import org.json.JSONException
import org.json.JSONObject

/** One complete server-sent event: its type and the joined data lines. */
data class ServerSentEvent(val type: String, val data: String)

/** Assembles server-sent events from the lines of a text/event-stream body. */
class ServerSentEventParser {

    private var type = ""
    private val dataLines = mutableListOf<String>()

    /** Takes the next line of the stream; returns the event a blank line completes, else null. */
    fun feed(line: String): ServerSentEvent? {
        if (line.isEmpty()) {
            return completeEvent()
        }
        if (line.startsWith(":")) {
            return null
        }
        val separator = line.indexOf(':')
        if (separator == -1) {
            return null
        }
        val field = line.substring(0, separator)
        val value = line.substring(separator + 1).removePrefix(" ")
        if (field == "event") {
            type = value
        }
        if (field == "data") {
            dataLines.add(value)
        }
        return null
    }

    /** Returns the collected event, if it carries data, and starts over. */
    private fun completeEvent(): ServerSentEvent? {
        var event: ServerSentEvent? = null
        if (dataLines.isNotEmpty()) {
            event = ServerSentEvent(type, dataLines.joinToString("\n"))
        }
        type = ""
        dataLines.clear()
        return event
    }
}

/** One item of the live search stream (`/api/search`): a library match or a plugin's result. */
data class LiveSearchHit(
    /** The plugin that found it, or "db" for library matches. */
    val source: String,
    /** The stored record's id, when the backend ingested the item. */
    val recordId: String?,
    val title: String,
    val mediaType: String,
    val sourceUrl: String,
    val externalId: String,
    val date: String?,
    val posterPath: String?,
    val previewImages: List<String>,
    val previewVideo: String?,
)

/** Parses the payload of a live search `result` event. */
object LiveSearchEvents {

    /** The hit a `result` event carries, or null when the payload is malformed. */
    fun parseResult(data: String): LiveSearchHit? {
        return try {
            toHit(JSONObject(data))
        } catch (error: JSONException) {
            null
        }
    }

    /** Reads the hit's fields from the event payload. */
    private fun toHit(payload: JSONObject): LiveSearchHit {
        val item = payload.getJSONObject("item")
        val previewImages = mutableListOf<String>()
        val previewArray = item.optJSONArray("preview_images")
        if (previewArray != null) {
            for (index in 0 until previewArray.length()) {
                previewImages.add(previewArray.getString(index))
            }
        }
        return LiveSearchHit(
            source = payload.optString("source"),
            recordId = textOrNull(payload, "id"),
            title = item.optString("title"),
            mediaType = item.optString("media_type"),
            sourceUrl = item.optString("source_url"),
            externalId = item.optString("external_id"),
            date = textOrNull(item, "date"),
            posterPath = textOrNull(item, "poster_path"),
            previewImages = previewImages,
            previewVideo = textOrNull(item, "preview_video"),
        )
    }

    /** The string field, or null when it is missing, null or empty. */
    private fun textOrNull(container: JSONObject, name: String): String? {
        if (container.isNull(name)) {
            return null
        }
        val text = container.optString(name)
        if (text.isEmpty()) {
            return null
        }
        return text
    }
}
