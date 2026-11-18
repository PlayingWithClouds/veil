package com.playingwithclouds.veil.cams

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.MainActivity
import com.playingwithclouds.veil.R
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.scenes.FrameProbabilities
import com.playingwithclouds.veil.scenes.SceneClassifier
import com.playingwithclouds.veil.scenes.perFrame
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

/** Who wants a room watched: its open page, or background watching. */
enum class WatchOrigin { IN_APP, BACKGROUND }

/**
 * Watches live rooms for the scene rules: per room, a [LiveFrameSampler] on the lowest rendition,
 * a frame every [SAMPLE_MILLISECONDS] through the scene model's frame graph (the picture's heads
 * only: frames this far apart make no temporal window), the tags it's fairly sure of into
 * [CamRuleEngine], whose actions are carried out here: a discreet notification, starting or
 * stopping a recording. A room stays watched while its page or background watching wants it; it
 * stops when the stream can't be had again or the model is gone. Main thread only.
 */
object CamWatch {

    /** One frame every this often per room. */
    const val SAMPLE_MILLISECONDS = 5_000L

    /** Tags read at this probability or more count as seen. */
    private const val MIN_PROBABILITY = 0.6f

    /** Fresh stream URLs asked for in a row before a room counts as gone. */
    private const val MAX_STREAM_RETRIES = 3

    private const val ALERT_CHANNEL = "alerts"
    private const val ALERT_NOTIFICATION_ID = 7302

    private lateinit var appContext: Context
    private val engine = CamRuleEngine()

    /** A watched room: who wants it and the loop reading its frames. */
    private class RoomWatch(val origins: MutableSet<WatchOrigin>, var job: Job? = null)

    private val rooms = mutableMapOf<String, RoomWatch>()
    private val mutableWatching = MutableStateFlow<Set<String>>(emptySet())

    /** The rooms being watched now, for the room page and Settings. */
    val watching: StateFlow<Set<String>> = mutableWatching

    /** Keeps the application context; called once from the application. */
    fun initialize(context: Context) {
        appContext = context.applicationContext
    }

    /** Whether rules would look at [roomId] if it were watched. */
    fun hasRulesFor(roomId: String): Boolean {
        return CamRuleStore.rules.value.any { rule -> rule.appliesTo(roomId) }
    }

    /** Starts watching [roomId] for [origin], unless it's watched already. */
    fun watch(roomId: String, origin: WatchOrigin) {
        val existing = rooms[roomId]
        if (existing != null) {
            existing.origins.add(origin)
            return
        }
        val watch = RoomWatch(mutableSetOf(origin))
        rooms[roomId] = watch
        watch.job = AppScope.launch { run(roomId) }
        publish()
    }

    /** [origin] no longer wants [roomId]; it stops when nobody does. */
    fun unwatch(roomId: String, origin: WatchOrigin) {
        val watch = rooms[roomId] ?: return
        watch.origins.remove(origin)
        if (watch.origins.isEmpty()) {
            stop(roomId)
        }
    }

    /** Background watching wants exactly [roomIds] now. */
    fun keepBackground(roomIds: Set<String>) {
        for (roomId in rooms.filterValues { watch -> WatchOrigin.BACKGROUND in watch.origins }.keys.toList()) {
            if (roomId !in roomIds) {
                unwatch(roomId, WatchOrigin.BACKGROUND)
            }
        }
        for (roomId in roomIds) {
            watch(roomId, WatchOrigin.BACKGROUND)
        }
    }

    /** Ends the room's loop and forgets it. */
    private fun stop(roomId: String) {
        val watch = rooms.remove(roomId) ?: return
        watch.job?.cancel()
        engine.forget(roomId)
        publish()
    }

    /** Publishes the watched rooms. */
    private fun publish() {
        mutableWatching.value = rooms.keys.toSet()
    }

    /** Reads the room's frames until it's unwatched, its stream is gone or there's no model. */
    private suspend fun run(roomId: String) {
        val failures = Channel<Unit>(Channel.CONFLATED)
        val sampler = LiveFrameSampler(appContext) { failures.trySend(Unit) }
        try {
            sampler.play(CamRepository.stream(roomId))
            sample(roomId, sampler, failures)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // The room went offline or its stream can't be had: nothing to watch.
        } finally {
            sampler.release()
        }
        // Ended by itself (not unwatched): forget the room so a later watch starts afresh.
        stop(roomId)
    }

    /** The sampling loop: a frame, its tags, the engine's actions; a fresh stream after errors. */
    private suspend fun sample(roomId: String, sampler: LiveFrameSampler, failures: Channel<Unit>) {
        var retries = 0
        while (true) {
            delay(SAMPLE_MILLISECONDS)
            if (failures.tryReceive().isSuccess) {
                retries++
                if (retries > MAX_STREAM_RETRIES) {
                    return
                }
                sampler.play(CamRepository.stream(roomId))
                continue
            }
            val bitmap = sampler.grab() ?: continue
            retries = 0
            val seen = withContext(Dispatchers.Default) { tagsOf(bitmap) } ?: return
            val actions = engine.observe(CamRuleStore.rules.value, roomId, System.currentTimeMillis(), seen)
            for (action in actions) {
                perform(action)
            }
        }
    }

    /** The tags the picture shows fairly surely; null without a model. */
    private fun tagsOf(bitmap: Bitmap): Set<String>? {
        val manifest = SceneClassifier.manifest() ?: return null
        val outputs = SceneClassifier.frames(listOf(bitmap)) ?: return null
        val probabilities = perFrame(outputs.logits, manifest).firstOrNull() ?: return emptySet()
        return FrameProbabilities.confidentTags(probabilities, manifest, MIN_PROBABILITY).keys
    }

    /** Carries out one action; a failed recording call is let go (the room may be recorded already). */
    private suspend fun perform(action: CamRuleAction) {
        try {
            when (action) {
                is CamRuleAction.Notify -> showAlert(appContext)
                is CamRuleAction.StartRecording -> {
                    val quality = CamRepository.model(action.roomId)?.recordQuality ?: CamRepository.SOURCE_QUALITY
                    CamRepository.startRecording(action.roomId, quality, action.fromSecondsAgo)
                }
                is CamRuleAction.StopRecording -> CamRepository.stopRecording(action.roomId)
            }
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Nothing to undo: the next sighting may try again after the cooldown.
        }
    }

    /**
     * The rule's alert, as discreet as the usual-time reminder: neutral words, no room, no picture,
     * hidden on the lock screen; opens the app. Nothing without the notification permission.
     */
    private fun showAlert(context: Context) {
        val granted = ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED
        if (!granted) {
            return
        }
        val manager = context.getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(NotificationChannel(ALERT_CHANNEL, "Alerts", NotificationManager.IMPORTANCE_DEFAULT))
        val open = PendingIntent.getActivity(
            context,
            0,
            Intent(context, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
            PendingIntent.FLAG_IMMUTABLE,
        )
        val notification = NotificationCompat.Builder(context, ALERT_CHANNEL)
            .setSmallIcon(R.drawable.ic_notification_dot)
            .setContentTitle("Reminder")
            .setContentText("Something you asked about is on.")
            .setVisibility(NotificationCompat.VISIBILITY_SECRET)
            .setCategory(NotificationCompat.CATEGORY_REMINDER)
            .setContentIntent(open)
            .setAutoCancel(true)
            .build()
        NotificationManagerCompat.from(context).notify(ALERT_NOTIFICATION_ID, notification)
    }
}
