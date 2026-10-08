package com.playingwithclouds.veil.privacy

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import androidx.media3.common.MediaMetadata

/**
 * What the system may show outside the app: media controls and notifications. Neither carries a
 * title, name, thumbnail or anything else from the library.
 */
object NeutralMedia {

    const val PLAYING_TITLE = "Playing"
    const val APP_NAME = "Veil"

    /** Metadata for every media item: "Playing" and the app name, no artwork. */
    fun metadata(): MediaMetadata {
        return MediaMetadata.Builder().setTitle(PLAYING_TITLE).setArtist(APP_NAME).build()
    }
}

/** Notification wording and flags that keep the lock screen blank. */
object NeutralNotifications {

    const val CHANNEL_ID = "neutral"

    /** The only texts a notification may use. */
    enum class Message(val text: String) {
        DOWNLOAD_FINISHED("Download finished"),
        LIBRARY_UPDATED("Library updated"),
    }

    /** Text for new items in Following, by count. */
    fun newItemsText(count: Int): String {
        if (count == 1) {
            return "1 new item in Following"
        }
        return "$count new items in Following"
    }

    /** Creates the quiet channel notifications go through. */
    fun createChannel(context: Context) {
        val channel = NotificationChannel(CHANNEL_ID, "Updates", NotificationManager.IMPORTANCE_LOW)
        channel.lockscreenVisibility = Notification.VISIBILITY_SECRET
        context.getSystemService(NotificationManager::class.java).createNotificationChannel(channel)
    }

    /** A notification with [text] only, hidden on the lock screen. */
    fun build(context: Context, text: String, smallIcon: Int): Notification {
        return Notification.Builder(context, CHANNEL_ID)
            .setSmallIcon(smallIcon)
            .setContentTitle(NeutralMedia.APP_NAME)
            .setContentText(text)
            .setVisibility(Notification.VISIBILITY_SECRET)
            .setAutoCancel(true)
            .build()
    }
}
