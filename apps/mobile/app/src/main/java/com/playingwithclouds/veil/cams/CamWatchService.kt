package com.playingwithclouds.veil.cams

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.net.ConnectivityManager
import android.os.BatteryManager
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import androidx.core.app.ServiceCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ProcessLifecycleOwner
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.MainActivity
import com.playingwithclouds.veil.R
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.habits.HabitPreferences
import com.playingwithclouds.veil.scenes.SceneClassifier
import com.playingwithclouds.veil.scenes.SceneModelStore
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch
import java.util.concurrent.TimeUnit

/**
 * When background watching runs: only with a scene rule, background watching on, a model, enough
 * battery (or charging), Wi-Fi when Settings → Background asks for it, and a covered room live.
 * WorkManager looks every 15 minutes (on Wi-Fi when asked, battery not low) and the app looks
 * each time it comes to the front; either starts [CamWatchService]. From the background Android
 * only allows that when the app isn't battery-optimised; a refusal is noted for Settings.
 */
object CamWatchBackground {

    private const val PERIODIC_WORK = "camWatchCheck"
    private const val CHECK_INTERVAL_MINUTES = 15L

    /** Keeps the check scheduled as rules and switches change, and checks whenever the app comes to the front; called from the application. */
    fun start(context: Context) {
        val appContext = context.applicationContext
        AppScope.launch {
            combine(CamRuleStore.rules, CamRuleStore.background, HabitPreferences.wifiOnly, SceneModelStore.state) { rules, background, _, _ ->
                rules.any { rule -> rule.enabled } && background && SceneClassifier.available()
            }.collect { wanted -> schedule(appContext, wanted) }
        }
        ProcessLifecycleOwner.get().lifecycle.addObserver(object : DefaultLifecycleObserver {
            override fun onStart(owner: LifecycleOwner) {
                AppScope.launch { startIfNeeded(appContext) }
            }
        })
    }

    /** The rooms to watch now; empty when background watching may not run or no covered room is live. */
    suspend fun targets(context: Context): List<String> {
        val rules = CamRuleStore.rules.value
        val hasRules = rules.any { rule -> rule.enabled } && CamRuleStore.background.value && SceneClassifier.available()
        if (!CamWatchPlan.allowed(hasRules, batteryPercent(context), charging(context), HabitPreferences.wifiOnly.value, metered(context))) {
            return emptyList()
        }
        val rooms = CamRepository.yourRooms().map { room ->
            WatchCandidate(room.id, room.online, room.followed, room.autoRecord, room.viewers)
        }
        return CamWatchPlan.targets(rules, rooms)
    }

    /** Starts [CamWatchService] when there's something to watch; notes when Android refuses. */
    suspend fun startIfNeeded(context: Context) {
        val targets = runCatching { targets(context) }.getOrDefault(emptyList())
        if (targets.isEmpty()) {
            return
        }
        try {
            ContextCompat.startForegroundService(context, Intent(context, CamWatchService::class.java))
            CamRuleStore.setBackgroundBlocked(false)
        } catch (error: IllegalStateException) {
            // ForegroundServiceStartNotAllowedException: the app is battery-optimised and in the background.
            CamRuleStore.setBackgroundBlocked(true)
        }
    }

    /** Schedules the periodic check when wanted (updating its network rule), else cancels it. */
    private fun schedule(context: Context, wanted: Boolean) {
        val workManager = WorkManager.getInstance(context)
        if (!wanted) {
            workManager.cancelUniqueWork(PERIODIC_WORK)
            context.stopService(Intent(context, CamWatchService::class.java))
            return
        }
        var network = NetworkType.CONNECTED
        if (HabitPreferences.wifiOnly.value) {
            network = NetworkType.UNMETERED
        }
        val constraints = Constraints.Builder().setRequiredNetworkType(network).setRequiresBatteryNotLow(true).build()
        val request = PeriodicWorkRequestBuilder<CamWatchWorker>(CHECK_INTERVAL_MINUTES, TimeUnit.MINUTES).setConstraints(constraints).build()
        workManager.enqueueUniquePeriodicWork(PERIODIC_WORK, ExistingPeriodicWorkPolicy.UPDATE, request)
    }

    /** The battery's charge, 0..100. */
    private fun batteryPercent(context: Context): Int {
        return context.getSystemService(BatteryManager::class.java)?.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY) ?: FULL_BATTERY
    }

    /** Whether the phone is charging. */
    private fun charging(context: Context): Boolean {
        return context.getSystemService(BatteryManager::class.java)?.isCharging ?: false
    }

    /** Whether the network in use costs data. */
    private fun metered(context: Context): Boolean {
        return context.getSystemService(ConnectivityManager::class.java)?.isActiveNetworkMetered ?: true
    }

    private const val FULL_BATTERY = 100
}

/** The 15-minute look: starts background watching when a covered room is live. */
class CamWatchWorker(context: Context, parameters: WorkerParameters) : CoroutineWorker(context, parameters) {

    /** Starts the service if there's something to watch. */
    override suspend fun doWork(): Result {
        CamWatchBackground.startIfNeeded(applicationContext)
        return Result.success()
    }
}

/**
 * Background watching: a foreground service (its notification as discreet as the usual-time
 * reminder: "Syncing", no picture, silent, hidden on the lock screen) that watches up to
 * [CamWatchPlan.MAX_ROOMS] covered live rooms and looks again every few minutes; it stops itself
 * when none is left, the battery runs low, Wi-Fi only meets mobile data, or the rules are gone.
 */
class CamWatchService : Service() {

    private var loop: Job? = null

    /** Goes to the foreground and starts the loop once. */
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        var type = 0
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            type = ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE
        }
        ServiceCompat.startForeground(this, NOTIFICATION_ID, notification(), type)
        if (loop == null) {
            loop = AppScope.launch { watchLoop() }
        }
        return START_NOT_STICKY
    }

    /** Keeps the watched rooms in step with what's live until nothing is left to watch. */
    private suspend fun watchLoop() {
        try {
            while (true) {
                val targets = runCatching { CamWatchBackground.targets(this) }.getOrDefault(emptyList())
                if (targets.isEmpty()) {
                    break
                }
                CamWatch.keepBackground(targets.toSet())
                delay(REPLAN_MILLISECONDS)
            }
        } finally {
            CamWatch.keepBackground(emptySet())
            loop = null
            stopSelf()
        }
    }

    /** Lets the rooms go. */
    override fun onDestroy() {
        loop?.cancel()
        CamWatch.keepBackground(emptySet())
        super.onDestroy()
    }

    /** Not bound. */
    override fun onBind(intent: Intent?): IBinder? = null

    /** The persistent notification: neutral, silent, secret. */
    private fun notification(): android.app.Notification {
        val manager = getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(NotificationChannel(CHANNEL, "Background work", NotificationManager.IMPORTANCE_MIN))
        val open = PendingIntent.getActivity(
            this,
            0,
            Intent(this, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
            PendingIntent.FLAG_IMMUTABLE,
        )
        return NotificationCompat.Builder(this, CHANNEL)
            .setSmallIcon(R.drawable.ic_notification_dot)
            .setContentTitle("Syncing")
            .setVisibility(NotificationCompat.VISIBILITY_SECRET)
            .setPriority(NotificationCompat.PRIORITY_MIN)
            .setSilent(true)
            .setOngoing(true)
            .setContentIntent(open)
            .build()
    }

    private companion object {
        const val CHANNEL = "background"
        const val NOTIFICATION_ID = 7303

        /** How often the live rooms, the battery and the network are looked at again. */
        const val REPLAN_MILLISECONDS = 2 * 60_000L
    }
}
