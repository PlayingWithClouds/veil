package com.playingwithclouds.veil.habits

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.MainActivity
import com.playingwithclouds.veil.R
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.FollowIdsQuery
import com.playingwithclouds.veil.graphql.HabitEventsQuery
import com.playingwithclouds.veil.graphql.JobStatusQuery
import com.playingwithclouds.veil.graphql.RunFollowMutation
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.JobStatus
import com.playingwithclouds.veil.util.parseTimestamp
import kotlinx.coroutines.delay
import java.time.Instant
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.temporal.ChronoUnit

/** How far back habits are learned from. */
private const val HABIT_WINDOW_DAYS = 42L

/** Events read to learn habits. */
private const val HABIT_EVENT_LIMIT = 2000

/** How long the check waits for the embedded backend to answer after the process starts. */
private const val BACKEND_WAIT_MILLISECONDS = 60_000L

/** How long the check waits for the follows' jobs before letting the process go. */
private const val JOBS_WAIT_MILLISECONDS = 8 * 60_000L

private const val POLL_MILLISECONDS = 3_000L

private const val NUDGE_CHANNEL = "suggestions"
private const val NUDGE_NOTIFICATION_ID = 7301

/**
 * Every few hours: runs every subscription so new videos are there before they're looked for,
 * then plans the nudge before the next usual time. The embedded backend starts with the process,
 * so the worker waits for it to answer first.
 */
class SubscriptionCheckWorker(context: Context, parameters: WorkerParameters) : CoroutineWorker(context, parameters) {

    /** Runs the follows, waits for their jobs, plans the nudge. */
    override suspend fun doWork(): Result {
        if (!awaitBackend()) {
            return Result.retry()
        }
        runFollows()
        planNudge(applicationContext)
        return Result.success()
    }

    /** Runs each enabled follow and waits (bounded) for the jobs to finish. */
    private suspend fun runFollows() {
        val follows = runCatching { VeilApi.client.query(FollowIdsQuery()).execute().dataOrThrow().follows }.getOrDefault(emptyList())
        val jobIds = follows.filter { follow -> follow.enabled }.mapNotNull { follow ->
            runCatching { VeilApi.client.mutation(RunFollowMutation(follow.id)).execute().dataOrThrow().runFollow.id }.getOrNull()
        }
        val deadline = System.currentTimeMillis() + JOBS_WAIT_MILLISECONDS
        var open = jobIds
        while (open.isNotEmpty() && System.currentTimeMillis() < deadline) {
            delay(POLL_MILLISECONDS)
            open = open.filter { id -> isRunning(id) }
        }
    }

    /** Whether a job is still queued or running. */
    private suspend fun isRunning(jobId: String): Boolean {
        val status = runCatching { VeilApi.client.query(JobStatusQuery(jobId)).execute().dataOrThrow().job?.status }.getOrNull()
        return status == JobStatus.queued || status == JobStatus.running
    }
}

/** Waits until the backend answers, up to [BACKEND_WAIT_MILLISECONDS]; false when it never does. */
private suspend fun awaitBackend(): Boolean {
    val deadline = System.currentTimeMillis() + BACKEND_WAIT_MILLISECONDS
    while (System.currentTimeMillis() < deadline) {
        val answered = runCatching { VeilApi.client.query(FollowIdsQuery()).execute().dataOrThrow() }.isSuccess
        if (answered) {
            return true
        }
        delay(POLL_MILLISECONDS)
    }
    return false
}

/** Learns the usual times from the last weeks and schedules the nudge before the next one. */
suspend fun planNudge(context: Context) {
    if (!HabitPreferences.nudge.value) {
        return
    }
    val zone = ZoneId.systemDefault()
    val since = Instant.now().minus(HABIT_WINDOW_DAYS, ChronoUnit.DAYS).toString()
    val query = HabitEventsQuery(since, Optional.present(HABIT_EVENT_LIMIT))
    val events = runCatching { VeilApi.client.query(query).execute().dataOrThrow().events }.getOrNull() ?: return
    val habits = events.mapNotNull { event ->
        val at = parseTimestamp(event.at) ?: return@mapNotNull null
        HabitEvent(at.atZone(zone), event.kind == EventKind.o)
    }
    val next = UsualTimes(habits).nextStart(ZonedDateTime.now(zone)) ?: return
    Habits.scheduleNudge(context, next)
}

/**
 * Shortly before the usual time: one discreet notification, unless the user turned it off, already
 * had one today, or already opened the app today. It says nothing about what the app is: a
 * neutral line, no picture, hidden on the lock screen.
 */
class NudgeWorker(context: Context, parameters: WorkerParameters) : CoroutineWorker(context, parameters) {

    /** Shows the nudge when it's still wanted. */
    override suspend fun doWork(): Result {
        if (shouldNudge()) {
            showNudge(applicationContext)
            HabitPreferences.setLastNudgeMillis(System.currentTimeMillis())
        }
        return Result.success()
    }

    /** Whether the nudge is on, and neither a nudge nor an open happened today. */
    private fun shouldNudge(): Boolean {
        if (!HabitPreferences.nudge.value) {
            return false
        }
        val startOfDay = ZonedDateTime.now().truncatedTo(ChronoUnit.DAYS).toInstant().toEpochMilli()
        return HabitPreferences.lastNudgeMillis < startOfDay && HabitPreferences.lastOpenMillis < startOfDay
    }
}

/** Posts the discreet notification; does nothing without the notification permission. */
private fun showNudge(context: Context) {
    val granted = ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED
    if (!granted) {
        return
    }
    val manager = context.getSystemService(NotificationManager::class.java)
    manager.createNotificationChannel(NotificationChannel(NUDGE_CHANNEL, "Reminders", NotificationManager.IMPORTANCE_DEFAULT))
    val open = PendingIntent.getActivity(
        context,
        0,
        Intent(context, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
        PendingIntent.FLAG_IMMUTABLE,
    )
    val notification = NotificationCompat.Builder(context, NUDGE_CHANNEL)
        .setSmallIcon(R.drawable.ic_notification_dot)
        .setContentTitle("Reminder")
        .setContentText("Something new is waiting.")
        .setVisibility(NotificationCompat.VISIBILITY_SECRET)
        .setCategory(NotificationCompat.CATEGORY_REMINDER)
        .setContentIntent(open)
        .setAutoCancel(true)
        .build()
    NotificationManagerCompat.from(context).notify(NUDGE_NOTIFICATION_ID, notification)
}
