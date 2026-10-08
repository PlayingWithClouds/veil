package com.playingwithclouds.veil.util

import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.content.Intent
import com.playingwithclouds.veil.MainActivity

/** The activity a (possibly wrapped) context belongs to, or null. */
fun Context.findActivity(): Activity? {
    var current: Context = this
    while (current is ContextWrapper) {
        if (current is Activity) {
            return current
        }
        current = current.baseContext
    }
    return null
}

/**
 * Starts the app over in a fresh task so screens and view models drop what they loaded from the
 * previous server. The process, and with it the embedded backend, keeps running.
 */
fun Context.restartApp() {
    val activity = findActivity() ?: return
    val intent = Intent(activity, MainActivity::class.java)
    intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
    activity.startActivity(intent)
    activity.finish()
}
