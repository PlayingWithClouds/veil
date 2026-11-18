package com.playingwithclouds.veil.habits

import android.content.Context
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ProcessLifecycleOwner
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import kotlinx.coroutines.flow.drop
import kotlinx.coroutines.launch
import java.time.ZonedDateTime
import java.util.concurrent.TimeUnit

/** How often subscriptions are checked in the background. */
private const val CHECK_INTERVAL_HOURS = 6L

/** How long before the usual time the nudge comes. */
private const val NUDGE_LEAD_MINUTES = 15L

private const val CHECK_WORK = "subscriptionCheck"
private const val NUDGE_WORK = "usualTimeNudge"

/**
 * What the app learns from habits, wired up once per process: an O guessed when the app is left
 * right after watching ([FinishGuesser]), when it was last opened, the background subscription
 * check every few hours and the nudge before the usual time.
 */
object Habits {

    private val guesser = FinishGuesser { System.currentTimeMillis() }

    /** Watches the app coming and going and schedules the background check; called from the application. */
    fun start(context: Context) {
        val appContext = context.applicationContext
        ProcessLifecycleOwner.get().lifecycle.addObserver(object : DefaultLifecycleObserver {
            override fun onStart(owner: LifecycleOwner) {
                HabitPreferences.setLastOpenMillis(System.currentTimeMillis())
            }

            override fun onStop(owner: LifecycleOwner) {
                logGuessedFinish()
                AppScope.launch { runCatching { planNudge(appContext) } }
            }
        })
        scheduleCheck(appContext, ExistingPeriodicWorkPolicy.KEEP)
        AppScope.launch {
            HabitPreferences.wifiOnly.drop(1).collect { scheduleCheck(appContext, ExistingPeriodicWorkPolicy.UPDATE) }
        }
    }

    /** The player of [videoId] reported where it is, for the finish guess. */
    fun progress(videoId: String, positionSeconds: Double) {
        guesser.progress(videoId, positionSeconds)
    }

    /** Logs the O the app being left suggests, marked as inferred so the engine trusts it less than a tap. */
    private fun logGuessedFinish() {
        val guess = guesser.appLeft() ?: return
        val data = mapOf("position" to guess.positionSeconds, "inferred" to true)
        AppServices.events.log(LoggedEvent(EventKind.o, RecordKind.video, guess.videoId, data), flush = true)
    }

    /** The background work's network rule: Wi-Fi only when the user asked for it. */
    fun constraints(): Constraints {
        var network = NetworkType.CONNECTED
        if (HabitPreferences.wifiOnly.value) {
            network = NetworkType.UNMETERED
        }
        return Constraints.Builder().setRequiredNetworkType(network).setRequiresBatteryNotLow(true).build()
    }

    /** Schedules the subscription check every few hours. */
    private fun scheduleCheck(context: Context, policy: ExistingPeriodicWorkPolicy) {
        val request = PeriodicWorkRequestBuilder<SubscriptionCheckWorker>(CHECK_INTERVAL_HOURS, TimeUnit.HOURS)
            .setConstraints(constraints())
            .build()
        WorkManager.getInstance(context).enqueueUniquePeriodicWork(CHECK_WORK, policy, request)
    }

    /** Schedules the nudge for [NUDGE_LEAD_MINUTES] before [usualStart], replacing an earlier plan. */
    fun scheduleNudge(context: Context, usualStart: ZonedDateTime) {
        val at = usualStart.minusMinutes(NUDGE_LEAD_MINUTES)
        val delayMillis = at.toInstant().toEpochMilli() - System.currentTimeMillis()
        if (delayMillis <= 0) {
            return
        }
        val request = OneTimeWorkRequestBuilder<NudgeWorker>().setInitialDelay(delayMillis, TimeUnit.MILLISECONDS).build()
        WorkManager.getInstance(context).enqueueUniqueWork(NUDGE_WORK, ExistingWorkPolicy.REPLACE, request)
    }
}
