package com.playingwithclouds.veil.scenes

import android.content.Context
import android.content.SharedPreferences
import android.net.Uri
import androidx.core.content.edit
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.withContext
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.File
import java.io.IOException
import java.io.InputStream
import java.util.concurrent.TimeUnit
import java.util.zip.ZipInputStream

/** Where the scene model stands on this phone. */
sealed interface SceneModelState {
    /** No model installed: the features it drives stay hidden. */
    data object Missing : SceneModelState

    /** A download or import is running. */
    data object Installing : SceneModelState

    /** Installed and readable. */
    data class Ready(val manifest: SceneModelManifest, val directory: File) : SceneModelState

    /** The last install failed, or the installed files can't be read; [installed] is what still works. */
    data class Failed(val message: String, val installed: Ready?) : SceneModelState
}

/** The model this state can run, if any. */
val SceneModelState.ready: SceneModelState.Ready?
    get() = when (this) {
        is SceneModelState.Ready -> this
        is SceneModelState.Failed -> installed
        else -> null
    }

/**
 * The scene model's files in app storage (`files/scene-model`: `model.json`, `frame.onnx`,
 * `temporal.onnx`), installed from a URL (the manifest, whose files sit beside it, or a zip of
 * the three) or a zip picked from storage. An install is unpacked into a fresh directory and only
 * replaces the old model once its manifest reads and its files are there.
 */
object SceneModelStore {

    private const val DIRECTORY = "scene-model"
    private const val INCOMING = "scene-model-incoming"
    private const val MANIFEST = "model.json"

    /** The manifest export.py writes, read as well (`SceneModelManifest.parse` takes both formats). */
    private const val EXPORT_MANIFEST = "manifest.json"
    private const val PREFERENCES = "scenes"
    private const val KEY_URL = "modelUrl"

    private lateinit var root: File
    private lateinit var preferences: SharedPreferences

    private val mutableState = MutableStateFlow<SceneModelState>(SceneModelState.Missing)

    /** The installed model, or why there is none. */
    val state: StateFlow<SceneModelState> = mutableState

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(120, TimeUnit.SECONDS)
        .build()

    /** The URL the model was last installed from, for Settings to show and update from. */
    val modelUrl: String
        get() = preferences.getString(KEY_URL, "").orEmpty()

    /** Reads the installed model, if any; called once from the application. */
    fun initialize(context: Context) {
        root = context.filesDir
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableState.value = readInstalled()
    }

    /** Downloads and installs the model at [url]: a `.zip` of the three files, or the manifest itself. */
    suspend fun installFromUrl(url: String) {
        preferences.edit { putString(KEY_URL, url.trim()) }
        install {
            val incoming = freshIncoming()
            if (url.trim().lowercase().substringBefore('?').endsWith(".zip")) {
                download(url.trim()).use { stream -> unzip(stream, incoming) }
                return@install incoming
            }
            downloadManifestAndFiles(url.trim(), incoming)
            incoming
        }
    }

    /** Installs the model from a zip the user picked. */
    suspend fun installFromZip(context: Context, uri: Uri) {
        install {
            val incoming = freshIncoming()
            val stream = context.contentResolver.openInputStream(uri) ?: throw IOException("The file can't be read")
            stream.use { input -> unzip(input, incoming) }
            incoming
        }
    }

    /** Deletes the model; the features it drives hide again. */
    suspend fun remove() {
        withContext(Dispatchers.IO) {
            SceneClassifier.close()
            File(root, DIRECTORY).deleteRecursively()
        }
        mutableState.value = SceneModelState.Missing
    }

    /** Runs an install that unpacks into a directory, checks it and swaps it in; failures keep the old model. */
    private suspend fun install(unpack: suspend () -> File) {
        val previous = mutableState.value.ready
        mutableState.value = SceneModelState.Installing
        try {
            val ready = withContext(Dispatchers.IO) { swapIn(checked(unpack())) }
            mutableState.value = ready
        } catch (error: Exception) {
            File(root, INCOMING).deleteRecursively()
            mutableState.value = SceneModelState.Failed(error.message ?: "The model couldn't be installed", previous)
        }
    }

    /** A new, empty directory to unpack into. */
    private fun freshIncoming(): File {
        val incoming = File(root, INCOMING)
        incoming.deleteRecursively()
        incoming.mkdirs()
        return incoming
    }

    /**
     * Fails unless [directory] holds a readable manifest and the graphs it names. An export's own
     * `manifest.json` is renamed to [MANIFEST] first.
     */
    private fun checked(directory: File): File {
        val manifestFile = File(directory, MANIFEST)
        val exportManifest = File(directory, EXPORT_MANIFEST)
        if (!manifestFile.isFile && exportManifest.isFile) {
            exportManifest.renameTo(manifestFile)
        }
        if (!manifestFile.isFile) {
            throw IOException("No $MANIFEST in the model")
        }
        val manifest = SceneModelManifest.parse(manifestFile.readText())
        val files = listOfNotNull(manifest.frame.file, manifest.temporal?.file)
        val missing = files.filter { name -> !File(directory, name).isFile }
        if (missing.isNotEmpty()) {
            throw IOException("Missing ${missing.joinToString()}")
        }
        return directory
    }

    /** Replaces the installed model with [incoming] and returns it as ready. */
    private fun swapIn(incoming: File): SceneModelState.Ready {
        SceneClassifier.close()
        val target = File(root, DIRECTORY)
        target.deleteRecursively()
        if (!incoming.renameTo(target)) {
            throw IOException("The model couldn't be moved into place")
        }
        return SceneModelState.Ready(SceneModelManifest.parse(File(target, MANIFEST).readText()), target)
    }

    /** The installed model, or Missing/Failed. */
    private fun readInstalled(): SceneModelState {
        val directory = File(root, DIRECTORY)
        if (!File(directory, MANIFEST).isFile) {
            return SceneModelState.Missing
        }
        return try {
            SceneModelState.Ready(SceneModelManifest.parse(File(checked(directory), MANIFEST).readText()), directory)
        } catch (error: Exception) {
            SceneModelState.Failed(error.message ?: "The installed model can't be read", null)
        }
    }

    /** Fetches the manifest at [url] and the graph files it names, resolved beside it. */
    private fun downloadManifestAndFiles(url: String, incoming: File) {
        val manifestText = download(url).use { stream -> stream.readBytes().decodeToString() }
        val manifest = SceneModelManifest.parse(manifestText)
        File(incoming, MANIFEST).writeText(manifestText)
        val base = url.toHttpUrl()
        for (name in listOfNotNull(manifest.frame.file, manifest.temporal?.file)) {
            val fileUrl = base.resolve(name) ?: throw IOException("Bad file name $name")
            download(fileUrl.toString()).use { stream -> File(incoming, safeName(name)).outputStream().use { output -> stream.copyTo(output) } }
        }
    }

    /** The body of a GET, open for reading; throws on an HTTP error. */
    private fun download(url: String): InputStream {
        val response = httpClient.newCall(Request.Builder().url(url).build()).execute()
        if (!response.isSuccessful) {
            response.close()
            throw IOException("Download failed: HTTP ${response.code}")
        }
        return response.body.byteStream()
    }

    /** Unpacks a zip's files flat into [directory] (folders inside the zip are ignored). */
    private fun unzip(stream: InputStream, directory: File) {
        ZipInputStream(stream).use { zip ->
            var entry = zip.nextEntry
            while (entry != null) {
                if (!entry.isDirectory) {
                    File(directory, safeName(entry.name)).outputStream().use { output -> zip.copyTo(output) }
                }
                entry = zip.nextEntry
            }
        }
    }

    /** The last path part of a name, so nothing is written outside the model's directory. */
    private fun safeName(name: String): String {
        val last = name.substringAfterLast('/').substringAfterLast('\\')
        if (last.isEmpty() || last == "." || last == "..") {
            throw IOException("Bad file name $name")
        }
        return last
    }
}
