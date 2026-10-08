package com.playingwithclouds.veil.api

import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import okhttp3.Call
import okhttp3.Callback
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import java.io.IOException
import java.util.concurrent.TimeUnit

/**
 * Streaming plugin search over the `/api/search` server-sent events: library matches arrive
 * first, then each source plugin's results as it finishes, so a grid fills incrementally.
 */
object LiveSearch {

    private const val READ_TIMEOUT_SECONDS = 120L

    private val httpClient = OkHttpClient.Builder()
        .readTimeout(READ_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    /**
     * Emits every result of the search and completes when all sources are exhausted. An empty
     * source list searches every enabled plugin; an explicit offset makes the plugin listings
     * paginate (infinite-scroll feeds).
     */
    fun results(
        query: String,
        sources: List<String>,
        offset: Int? = null,
        limit: Int? = null,
    ): Flow<LiveSearchHit> = callbackFlow {
        val call = httpClient.newCall(Request.Builder().url(searchUrl(query, sources, offset, limit)).build())
        call.enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                close()
            }

            override fun onResponse(call: Call, response: Response) {
                response.use { openResponse -> readEvents(openResponse, ::trySend) }
                close()
            }
        })
        awaitClose { call.cancel() }
    }

    /** Builds the search address with its optional parameters. */
    private fun searchUrl(query: String, sources: List<String>, offset: Int?, limit: Int?): okhttp3.HttpUrl {
        val builder = "${ServerSettings.baseUrl}/api/search".toHttpUrl().newBuilder()
        builder.addQueryParameter("q", query)
        if (sources.isNotEmpty()) {
            builder.addQueryParameter("sources", sources.joinToString(","))
        }
        if (offset != null) {
            builder.addQueryParameter("offset", offset.toString())
        }
        if (limit != null) {
            builder.addQueryParameter("limit", limit.toString())
        }
        return builder.build()
    }

    /** Reads the event stream until `done` or the end of the body, handing each hit on. */
    private fun readEvents(response: Response, emit: (LiveSearchHit) -> Unit) {
        val parser = ServerSentEventParser()
        try {
            val reader = response.body.source()
            while (true) {
                val line = reader.readUtf8Line() ?: return
                val event = parser.feed(line) ?: continue
                if (event.type == "done") {
                    return
                }
                if (event.type == "result") {
                    LiveSearchEvents.parseResult(event.data)?.let(emit)
                }
            }
        } catch (error: IOException) {
            // A dropped connection ends the search rather than failing it.
        }
    }
}
