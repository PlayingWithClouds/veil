package com.playingwithclouds.veil.scenes

import android.content.Context
import android.util.LruCache
import coil3.SingletonImageLoader
import coil3.request.ImageRequest
import coil3.request.SuccessResult
import coil3.request.allowHardware
import coil3.toBitmap
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.data.CoverTagSource
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Deferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.io.File
import java.security.MessageDigest
import java.util.concurrent.ConcurrentHashMap

/**
 * The scene model's reading of every cover a card shows: the tag names it is fairly sure of
 * ([MIN_PROBABILITY] or more) with their probability. Covers come from Coil's cache (the card just
 * loaded them); readings are kept per cover URL and model version, in memory and in the app's
 * cache directory, so a cover is read once. The impression logger sends them with the impression
 * and choice events, from which the engine learns which scenes draw the user in.
 */
object CoverLabels : CoverTagSource {

    /** Labels below this probability are left out. */
    const val MIN_PROBABILITY = 0.5f

    private const val MEMORY_ENTRIES = 500
    private const val DIRECTORY = "cover-labels"

    private lateinit var appContext: Context

    /** The cover URL each record's card showed last. */
    private val coverByRecord = ConcurrentHashMap<String, String>()

    /** Readings in progress or done, by cover URL and model version. */
    private val readings = LruCache<String, Deferred<Map<String, Double>?>>(MEMORY_ENTRIES)

    /** One cover at a time, so reading covers never competes with scrolling. */
    @OptIn(ExperimentalCoroutinesApi::class)
    private val inference = Dispatchers.Default.limitedParallelism(1)

    /** Keeps the context for Coil and the cache directory; called once from the application. */
    fun initialize(context: Context) {
        appContext = context.applicationContext
    }

    /** A card showed [coverPath] for [recordId]: reads it in the background unless it was read before. Nothing without a model. */
    fun read(recordId: String, coverPath: String?) {
        val url = ServerSettings.imageUrl(coverPath) ?: return
        val version = SceneClassifier.manifest()?.version ?: return
        coverByRecord[recordId] = url
        reading(url, version)
    }

    /** Whether covers are being read at all. */
    override fun active(): Boolean {
        return SceneClassifier.available()
    }

    /** The reading of [recordId]'s cover if it's done; null otherwise. */
    @OptIn(ExperimentalCoroutinesApi::class)
    override fun cached(recordId: String): Map<String, Double>? {
        val url = coverByRecord[recordId] ?: return null
        val version = SceneClassifier.manifest()?.version ?: return null
        val deferred = synchronized(readings) { readings.get(key(url, version)) } ?: return null
        if (!deferred.isCompleted) {
            return null
        }
        return deferred.getCompleted()
    }

    /** Waits for the reading of [recordId]'s cover; null when its card never showed or the cover can't be read. */
    override suspend fun await(recordId: String): Map<String, Double>? {
        val url = coverByRecord[recordId] ?: return null
        val version = SceneClassifier.manifest()?.version ?: return null
        return reading(url, version).await()
    }

    /** The reading of [url], started when it isn't known yet. */
    private fun reading(url: String, version: String): Deferred<Map<String, Double>?> {
        val key = key(url, version)
        synchronized(readings) {
            val known = readings.get(key)
            if (known != null) {
                return known
            }
            val deferred = CompletableDeferred<Map<String, Double>?>()
            readings.put(key, deferred)
            AppScope.launch { deferred.complete(runCatching { readCover(url, key) }.getOrNull()) }
            return deferred
        }
    }

    /** The stored reading, else the model's reading of the cover (then stored). */
    private suspend fun readCover(url: String, key: String): Map<String, Double>? {
        val file = File(File(appContext.cacheDir, DIRECTORY), key)
        val stored = withContext(Dispatchers.IO) { readStored(file) }
        if (stored != null) {
            return stored
        }
        val tags = classify(url) ?: return null
        withContext(Dispatchers.IO) { store(file, tags) }
        return tags
    }

    /** Loads the cover through Coil (its cache, usually) and reads it with the frame graph. */
    private suspend fun classify(url: String): Map<String, Double>? {
        val request = ImageRequest.Builder(appContext).data(url).allowHardware(false).build()
        val result = SingletonImageLoader.get(appContext).execute(request) as? SuccessResult ?: return null
        val bitmap = result.image.toBitmap()
        return withContext(inference) {
            val manifest = SceneClassifier.manifest() ?: return@withContext null
            val outputs = SceneClassifier.frames(listOf(bitmap)) ?: return@withContext null
            val probabilities = perFrame(outputs.logits, manifest).firstOrNull() ?: return@withContext null
            FrameProbabilities.confidentTags(probabilities, manifest, MIN_PROBABILITY)
        }
    }

    /** A stored reading; null when there is none or it can't be read. */
    private fun readStored(file: File): Map<String, Double>? {
        if (!file.isFile) {
            return null
        }
        return runCatching {
            val json = JSONObject(file.readText())
            json.keys().asSequence().associateWith { tag -> json.getDouble(tag) }
        }.getOrNull()
    }

    /** Keeps a reading for next time. */
    private fun store(file: File, tags: Map<String, Double>) {
        file.parentFile?.mkdirs()
        file.writeText(JSONObject(tags).toString())
    }

    /** The cache key of a cover under a model version: a hash, so it's a safe file name. */
    private fun key(url: String, version: String): String {
        val digest = MessageDigest.getInstance("SHA-1").digest("$version|$url".toByteArray())
        return digest.joinToString("") { byte -> "%02x".format(byte) }
    }
}
