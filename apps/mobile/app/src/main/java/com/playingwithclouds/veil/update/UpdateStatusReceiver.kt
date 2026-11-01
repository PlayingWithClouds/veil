package com.playingwithclouds.veil.update

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/** Receives the installer session's status for [UpdateInstaller]; not exported, only the app's own PendingIntent reaches it. */
class UpdateStatusReceiver : BroadcastReceiver() {

    /** Passes the status on to [UpdateInstaller]. */
    override fun onReceive(context: Context, intent: Intent) {
        UpdateInstaller.onSessionStatus(context, intent)
    }
}
