package com.playingwithclouds.veil.scenes

import com.playingwithclouds.veil.data.PlayableStream
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.OkHttpClient
import okhttp3.Request
import java.util.concurrent.TimeUnit

/** One `#EXT-X-STREAM-INF` entry of a master playlist: its bandwidth and its playlist's URI as written. */
data class HlsVariant(val bandwidth: Long, val uri: String, val hasVideo: Boolean)

/**
 * The lowest-bandwidth rendition of an HLS stream, for the scene scan: the model reads 224 px
 * pictures, so the smallest rendition is enough and reads far less (and the frame extractor has no
 * track selection of its own, so given the master it would pick by its bandwidth guess).
 */
object HlsVariants {

    private val bandwidthAttribute = Regex("""(?:^|,)BANDWIDTH=(\d+)""")
    private val codecsAttribute = Regex("""(?:^|,)CODECS="([^"]*)"""")
    private val videoCodecPrefixes = listOf("avc", "hvc", "hev", "vp0", "vp8", "vp9", "av01", "mp4v", "dvh", "dva")

    private const val STREAM_INF = "#EXT-X-STREAM-INF:"
    private const val TIMEOUT_SECONDS = 10L

    private val httpClient by lazy {
        OkHttpClient.Builder().callTimeout(TIMEOUT_SECONDS, TimeUnit.SECONDS).build()
    }

    /** The variants of a master playlist in file order; empty for a media playlist. */
    fun parse(playlist: String): List<HlsVariant> {
        val out = mutableListOf<HlsVariant>()
        var pending: String? = null
        for (raw in playlist.lines()) {
            val line = raw.trim()
            if (line.startsWith(STREAM_INF)) {
                pending = line.removePrefix(STREAM_INF)
                continue
            }
            val attributes = pending
            if (attributes == null || line.isEmpty() || line.startsWith("#")) {
                continue
            }
            out.add(variantOf(attributes, line))
            pending = null
        }
        return out
    }

    /**
     * The URL of the lowest-bandwidth variant with video in [playlist], resolved against
     * [playlistUrl] (the backend's proxy writes root-relative URIs); null for a media playlist.
     */
    fun lowestUrl(playlist: String, playlistUrl: String): String? {
        val lowest = parse(playlist).filter { variant -> variant.hasVideo }.minByOrNull { variant -> variant.bandwidth } ?: return null
        val base = playlistUrl.toHttpUrlOrNull() ?: return null
        return base.resolve(lowest.uri)?.toString()
    }

    /**
     * [stream] at its lowest rendition: for HLS, the master playlist is read and its smallest
     * variant's playlist returned; anything else, or a playlist that can't be read, stays as it is.
     */
    suspend fun lowest(stream: PlayableStream): PlayableStream {
        if (!stream.isHls) {
            return stream
        }
        try {
            val playlist = withContext(Dispatchers.IO) { fetch(stream.url) } ?: return stream
            val url = lowestUrl(playlist, stream.url) ?: return stream
            return stream.copy(url = url)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            return stream
        }
    }

    /** The playlist's text; null for a non-2xx answer or a non-HTTP URL. */
    private fun fetch(url: String): String? {
        val httpUrl = url.toHttpUrlOrNull() ?: return null
        httpClient.newCall(Request.Builder().url(httpUrl).build()).execute().use { response ->
            if (!response.isSuccessful) {
                return null
            }
            return response.body.string()
        }
    }

    /** One variant from its attribute list and URI line; without CODECS it's taken to have video. */
    private fun variantOf(attributes: String, uri: String): HlsVariant {
        val bandwidth = bandwidthAttribute.find(attributes)?.groupValues?.get(1)?.toLongOrNull() ?: Long.MAX_VALUE
        val codecs = codecsAttribute.find(attributes)?.groupValues?.get(1)
        var hasVideo = true
        if (codecs != null) {
            hasVideo = codecs.split(',').any { codec -> videoCodecPrefixes.any { prefix -> codec.trim().startsWith(prefix) } }
        }
        return HlsVariant(bandwidth, uri, hasVideo)
    }
}
