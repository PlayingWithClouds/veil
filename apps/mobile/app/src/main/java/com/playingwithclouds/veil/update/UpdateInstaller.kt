package com.playingwithclouds.veil.update

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageInstaller
import android.os.Build
import androidx.core.content.IntentCompat
import com.playingwithclouds.veil.AppScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.ensureActive
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.File
import java.io.IOException
import java.io.InputStream
import java.io.OutputStream
import java.util.concurrent.TimeUnit

/** Where an update installation stands. */
sealed interface InstallState {
    data object Idle : InstallState

    /** Downloading the APK; [progress] runs from 0 to 1, null while the size is unknown. */
    data class Downloading(val progress: Float?) : InstallState

    /** Handed to the system installer, waiting for it to finish or for the user to confirm. */
    data object Installing : InstallState

    data class Failed(val message: String) : InstallState
}

/**
 * Downloads a release's APK into the app cache and installs it through a [PackageInstaller]
 * session. Once this app is its own installer of record (Android 12+), updates need no
 * confirmation; otherwise the system asks the user.
 */
object UpdateInstaller {

    private const val CACHE_DIRECTORY = "updates"
    private const val APK_FILE = "update.apk"
    private const val SESSION_FILE = "veil.apk"
    private const val CONNECT_TIMEOUT_SECONDS = 15L
    private const val READ_TIMEOUT_SECONDS = 60L
    private const val BUFFER_BYTES = 64 * 1024

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(CONNECT_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(READ_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    private val mutableState = MutableStateFlow<InstallState>(InstallState.Idle)
    private var job: Job? = null

    /** The current download and install progress. */
    val state: StateFlow<InstallState> = mutableState

    /** Downloads and installs [release]; ignored while a download is already running. */
    fun start(context: Context, release: Release) {
        if (job?.isActive == true) {
            return
        }
        val appContext = context.applicationContext
        mutableState.value = InstallState.Downloading(null)
        job = AppScope.launch {
            try {
                downloadAndCommit(appContext, release)
            } catch (error: CancellationException) {
                mutableState.value = InstallState.Idle
                throw error
            } catch (error: UpdateException) {
                mutableState.value = InstallState.Failed(error.message.orEmpty())
            } catch (error: Exception) {
                mutableState.value = InstallState.Failed("The update couldn't be installed. Try again.")
            }
        }
    }

    /** Fetches the APK, writes it into an installer session and commits it; the cached copy is removed either way. */
    private suspend fun downloadAndCommit(context: Context, release: Release) = withContext(Dispatchers.IO) {
        val directory = File(context.cacheDir, CACHE_DIRECTORY)
        // Leftovers of an interrupted earlier attempt.
        directory.deleteRecursively()
        directory.mkdirs()
        val apk = File(directory, APK_FILE)
        try {
            download(release, apk)
            // Set before committing: the installer may report back before commit returns.
            mutableState.value = InstallState.Installing
            commitSession(context, apk)
        } finally {
            directory.deleteRecursively()
        }
    }

    /** Streams the release's APK into [target], reporting progress, and checks its size. */
    private suspend fun download(release: Release, target: File) {
        val request = Request.Builder().url(release.apkUrl).build()
        try {
            httpClient.newCall(request).execute().use { response ->
                if (!response.isSuccessful) {
                    throw UpdateException("The download failed with error ${response.code}. Try again later.")
                }
                var totalBytes = response.body.contentLength()
                if (totalBytes <= 0) {
                    totalBytes = release.apkSizeBytes
                }
                response.body.byteStream().use { input ->
                    target.outputStream().use { output -> copyWithProgress(input, output, totalBytes) }
                }
            }
        } catch (error: UpdateException) {
            throw error
        } catch (error: IOException) {
            throw UpdateException("The download was interrupted. Check the connection and try again.", error)
        }
        if (release.apkSizeBytes > 0 && target.length() != release.apkSizeBytes) {
            throw UpdateException("The download was incomplete. Try again.")
        }
    }

    /** Copies [input] to [output] and publishes the share of [totalBytes] done (unknown when not positive). */
    private suspend fun copyWithProgress(input: InputStream, output: OutputStream, totalBytes: Long) {
        val buffer = ByteArray(BUFFER_BYTES)
        var copiedBytes = 0L
        var lastPercent = -1
        while (true) {
            currentCoroutineContext().ensureActive()
            val read = input.read(buffer)
            if (read < 0) {
                return
            }
            output.write(buffer, 0, read)
            copiedBytes += read
            if (totalBytes <= 0) {
                continue
            }
            val percent = (copiedBytes * 100 / totalBytes).toInt()
            if (percent != lastPercent) {
                lastPercent = percent
                mutableState.value = InstallState.Downloading(percent / 100f)
            }
        }
    }

    /** Writes [apk] into a new installer session and commits it; the result arrives at [UpdateStatusReceiver]. */
    private fun commitSession(context: Context, apk: File) {
        val installer = context.packageManager.packageInstaller
        val sessionId = installer.createSession(sessionParams(context, apk.length()))
        val session = installer.openSession(sessionId)
        try {
            session.openWrite(SESSION_FILE, 0, apk.length()).use { output ->
                apk.inputStream().use { input -> input.copyTo(output, BUFFER_BYTES) }
                session.fsync(output)
            }
            session.commit(statusIntent(context, sessionId).intentSender)
        } catch (error: Exception) {
            session.abandon()
            throw error
        } finally {
            session.close()
        }
    }

    /** A full install of this app's own package, silent where Android allows it. */
    private fun sessionParams(context: Context, sizeBytes: Long): PackageInstaller.SessionParams {
        val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL)
        params.setAppPackageName(context.packageName)
        params.setSize(sizeBytes)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            params.setRequireUserAction(PackageInstaller.SessionParams.USER_ACTION_NOT_REQUIRED)
        }
        return params
    }

    /** The broadcast the installer reports the session's outcome to; mutable so it can add the status extras. */
    private fun statusIntent(context: Context, sessionId: Int): PendingIntent {
        val intent = Intent(context, UpdateStatusReceiver::class.java)
        var flags = PendingIntent.FLAG_UPDATE_CURRENT
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            flags = flags or PendingIntent.FLAG_MUTABLE
        }
        return PendingIntent.getBroadcast(context, sessionId, intent, flags)
    }

    /** Handles a session status from [UpdateStatusReceiver]: asks the user to confirm, or records the outcome. */
    fun onSessionStatus(context: Context, intent: Intent) {
        val status = intent.getIntExtra(PackageInstaller.EXTRA_STATUS, PackageInstaller.STATUS_FAILURE)
        when (status) {
            PackageInstaller.STATUS_PENDING_USER_ACTION -> askUserToConfirm(context, intent)
            PackageInstaller.STATUS_SUCCESS -> mutableState.value = InstallState.Idle
            PackageInstaller.STATUS_FAILURE_ABORTED -> mutableState.value = InstallState.Failed("The update was cancelled.")
            else -> mutableState.value = InstallState.Failed(failureMessage(status))
        }
    }

    /** Opens the system's install confirmation that came with the status. */
    private fun askUserToConfirm(context: Context, intent: Intent) {
        val confirmation = IntentCompat.getParcelableExtra(intent, Intent.EXTRA_INTENT, Intent::class.java)
        if (confirmation == null) {
            mutableState.value = InstallState.Failed("The update couldn't be installed. Try again.")
            return
        }
        confirmation.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        try {
            context.startActivity(confirmation)
        } catch (error: RuntimeException) {
            // Android refuses to open it while the app is in the background.
            mutableState.value = InstallState.Failed("Couldn't ask for confirmation. Open the app and tap Update again.")
        }
    }

    /** A plain explanation of a failed installer status. */
    private fun failureMessage(status: Int): String {
        return when (status) {
            PackageInstaller.STATUS_FAILURE_STORAGE -> "Not enough storage to install the update."
            PackageInstaller.STATUS_FAILURE_CONFLICT,
            PackageInstaller.STATUS_FAILURE_INCOMPATIBLE,
            -> "The update doesn't match the installed app, so Android refused it."
            PackageInstaller.STATUS_FAILURE_BLOCKED -> "Android blocked the update."
            else -> "The update couldn't be installed. Try again."
        }
    }
}
