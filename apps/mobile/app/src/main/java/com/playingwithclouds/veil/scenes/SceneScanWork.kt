package com.playingwithclouds.veil.scenes

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.graphql.SceneSettingsQuery
import com.playingwithclouds.veil.habits.HabitPreferences
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch
import java.util.concurrent.TimeUnit

/** What the background scan is doing, for Settings. */
data class ScanStatus(val running: Boolean = false, val videoId: String? = null, val progress: Float = 0f, val scannedCount: Int = 0)

/**
 * The background scan: which videos were scanned with which model, the switch, and the
 * WorkManager job that runs it only while charging with the battery not low (on Wi-Fi when
 * Settings → Background asks for it).
 */
object SceneScans {

    private const val PREFERENCES = "sceneScans"
    private const val KEY_ENABLED = "enabled"
    private const val KEY_VERSION = "modelVersion"
    private const val KEY_SCANNED = "scanned"
    private const val PERIODIC_WORK = "sceneScan"
    private const val NOW_WORK = "sceneScanNow"

    /** How often the job is offered to WorkManager; it only runs when its constraints hold. */
    private const val INTERVAL_HOURS = 3L

    private lateinit var preferences: SharedPreferences

    private val mutableEnabled = MutableStateFlow(true)
    private val mutableStatus = MutableStateFlow(ScanStatus())

    /** Whether videos are scanned in the background; on by default (nothing runs without a model). */
    val enabled: StateFlow<Boolean> = mutableEnabled

    /** The running scan, for Settings. */
    val status: StateFlow<ScanStatus> = mutableStatus

    /** Reads the stored state and keeps the job scheduled as the model, switch and Wi-Fi rule change; called from the application. */
    fun start(context: Context) {
        val appContext = context.applicationContext
        preferences = appContext.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableEnabled.value = preferences.getBoolean(KEY_ENABLED, true)
        mutableStatus.value = ScanStatus(scannedCount = scannedIds().size)
        AppScope.launch {
            combine(SceneModelStore.state, enabled, HabitPreferences.wifiOnly) { state, on, _ -> state.ready != null && on }
                .collect { wanted -> schedule(appContext, wanted) }
        }
    }

    /** Turns the background scan on or off. */
    fun setEnabled(on: Boolean) {
        preferences.edit { putBoolean(KEY_ENABLED, on) }
        mutableEnabled.value = on
    }

    /** Asks for a scan as soon as the phone is charging (the same constraints, no waiting for the period). */
    fun scanSoon(context: Context) {
        val request = OneTimeWorkRequestBuilder<SceneScanWorker>().setConstraints(constraints()).build()
        WorkManager.getInstance(context).enqueueUniqueWork(NOW_WORK, ExistingWorkPolicy.KEEP, request)
    }

    /** The video ids scanned with the installed model. */
    fun scannedIds(): Set<String> {
        val version = SceneClassifier.manifest()?.version ?: return emptySet()
        if (preferences.getString(KEY_VERSION, null) != version) {
            return emptySet()
        }
        return preferences.getStringSet(KEY_SCANNED, emptySet()).orEmpty()
    }

    /** Records [videoId] as scanned with the installed model; a new model starts the list over. */
    fun markScanned(videoId: String) {
        val version = SceneClassifier.manifest()?.version ?: return
        val scanned = scannedIds() + videoId
        preferences.edit {
            putString(KEY_VERSION, version)
            putStringSet(KEY_SCANNED, scanned)
        }
        mutableStatus.value = mutableStatus.value.copy(scannedCount = scanned.size)
    }

    /** Updates what Settings shows. */
    internal fun report(status: ScanStatus) {
        mutableStatus.value = status.copy(scannedCount = scannedIds().size)
    }

    /** Charging, battery not low, and Wi-Fi when the user asked for it. */
    fun constraints(): Constraints {
        var network = NetworkType.CONNECTED
        if (HabitPreferences.wifiOnly.value) {
            network = NetworkType.UNMETERED
        }
        return Constraints.Builder()
            .setRequiredNetworkType(network)
            .setRequiresCharging(true)
            .setRequiresBatteryNotLow(true)
            .build()
    }

    /** Schedules the periodic scan when it's wanted (updating its constraints when the Wi-Fi rule changed), else cancels it. */
    private fun schedule(context: Context, wanted: Boolean) {
        val workManager = WorkManager.getInstance(context)
        if (!wanted) {
            workManager.cancelUniqueWork(PERIODIC_WORK)
            return
        }
        val request = PeriodicWorkRequestBuilder<SceneScanWorker>(INTERVAL_HOURS, TimeUnit.HOURS).setConstraints(constraints()).build()
        workManager.enqueueUniquePeriodicWork(PERIODIC_WORK, ExistingPeriodicWorkPolicy.UPDATE, request)
    }
}

/**
 * Scans videos not scanned with the installed model yet, favourites, collections, the watchlist,
 * the queue and recently watched first, then recommendations, until [VIDEOS_PER_RUN] are done or
 * WorkManager stops it (unplugged, battery low): what was scanned stays scanned.
 */
class SceneScanWorker(context: Context, parameters: WorkerParameters) : CoroutineWorker(context, parameters) {

    /** Runs one round of scanning. */
    override suspend fun doWork(): Result {
        if (!SceneClassifier.available() || !SceneScans.enabled.value) {
            return Result.success()
        }
        if (!awaitBackend()) {
            return Result.retry()
        }
        val pending = runCatching { SceneRepository.scanCandidates(CANDIDATE_LIMIT) }.getOrDefault(emptyList()).filter { id -> id !in SceneScans.scannedIds() }
        val scanner = SceneScanner(applicationContext)
        try {
            for (videoId in pending.take(VIDEOS_PER_RUN)) {
                scanOne(scanner, videoId)
            }
        } finally {
            SceneScans.report(ScanStatus())
        }
        return Result.success()
    }

    /** Scans one video; a failure marks it done for this model so one bad video doesn't block the rest. */
    private suspend fun scanOne(scanner: SceneScanner, videoId: String) {
        SceneScans.report(ScanStatus(running = true, videoId = videoId))
        try {
            val outcome = scanner.scan(videoId) { share -> SceneScans.report(ScanStatus(running = true, videoId = videoId, progress = share)) }
            if (outcome == ScanOutcome.NO_MODEL) {
                return
            }
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Unplayable streams and decoder failures are expected for some videos; move on.
        }
        SceneScans.markScanned(videoId)
    }

    /** Waits until the embedded backend answers; false when it never does. */
    private suspend fun awaitBackend(): Boolean {
        repeat(BACKEND_TRIES) {
            val answered = runCatching { VeilApi.client.query(SceneSettingsQuery()).execute().dataOrThrow() }.isSuccess
            if (answered) {
                return true
            }
            delay(BACKEND_POLL_MILLISECONDS)
        }
        return false
    }

    companion object {
        /** Videos one run scans at most; the next run goes on. */
        const val VIDEOS_PER_RUN = 25

        /** How many of each priority list are read. */
        const val CANDIDATE_LIMIT = 60

        private const val BACKEND_TRIES = 20
        private const val BACKEND_POLL_MILLISECONDS = 3_000L
    }
}
