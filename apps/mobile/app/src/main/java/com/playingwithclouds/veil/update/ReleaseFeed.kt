package com.playingwithclouds.veil.update

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONException
import org.json.JSONObject
import java.io.IOException
import java.util.concurrent.TimeUnit

/** A published app release: its tag, title, notes and the APK to install. */
data class Release(
    val tagName: String,
    val name: String,
    val notes: String,
    val apkUrl: String,
    val apkSizeBytes: Long,
)

/** A failed update check or download, with a message fit to show as is. */
class UpdateException(message: String, cause: Throwable? = null) : IOException(message, cause)

/** The newest app release on GitHub. */
object ReleaseFeed {

    private const val LATEST_RELEASE_URL = "https://api.github.com/repos/PlayingWithClouds/veil/releases/latest"
    private const val APK_ASSET_NAME = "veil.apk"
    private const val HTTP_NOT_FOUND = 404
    private const val TIMEOUT_SECONDS = 15L

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    /** Fetches the latest release; null when there is none or it carries no APK. */
    suspend fun fetchLatest(): Release? = withContext(Dispatchers.IO) {
        val body = requestLatest() ?: return@withContext null
        parse(body)
    }

    /** The latest release's JSON, or null when the repository has no release yet. */
    private fun requestLatest(): String? {
        val request = Request.Builder()
            .url(LATEST_RELEASE_URL)
            .header("Accept", "application/vnd.github+json")
            .build()
        try {
            httpClient.newCall(request).execute().use { response ->
                if (response.code == HTTP_NOT_FOUND) {
                    return null
                }
                if (!response.isSuccessful) {
                    throw UpdateException("GitHub answered with error ${response.code}. Try again later.")
                }
                return response.body.string()
            }
        } catch (error: UpdateException) {
            throw error
        } catch (error: IOException) {
            throw UpdateException("Couldn't reach GitHub. Check the connection and try again.", error)
        }
    }

    /** Reads a GitHub release response; null when it is unreadable or has no `veil.apk` asset. */
    fun parse(json: String): Release? {
        val release = try {
            JSONObject(json)
        } catch (error: JSONException) {
            return null
        }
        val tagName = release.optString("tag_name")
        if (tagName.isEmpty()) {
            return null
        }
        val asset = findApkAsset(release) ?: return null
        val apkUrl = asset.optString("browser_download_url")
        if (apkUrl.isEmpty()) {
            return null
        }
        return Release(
            tagName = tagName,
            name = optText(release, "name", tagName),
            notes = optText(release, "body", ""),
            apkUrl = apkUrl,
            apkSizeBytes = asset.optLong("size", 0L),
        )
    }

    /** The asset named `veil.apk`; the versioned copy beside it is the same file. */
    private fun findApkAsset(release: JSONObject): JSONObject? {
        val assets = release.optJSONArray("assets") ?: return null
        for (index in 0 until assets.length()) {
            val asset = assets.optJSONObject(index) ?: continue
            if (asset.optString("name") == APK_ASSET_NAME) {
                return asset
            }
        }
        return null
    }

    /** A string field, or [fallback] when it is missing, null or blank. */
    private fun optText(json: JSONObject, key: String, fallback: String): String {
        // optString turns a JSON null into the text "null".
        if (json.isNull(key)) {
            return fallback
        }
        val text = json.optString(key)
        if (text.isBlank()) {
            return fallback
        }
        return text
    }
}
