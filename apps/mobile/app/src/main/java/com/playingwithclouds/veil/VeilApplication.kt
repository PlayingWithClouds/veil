package com.playingwithclouds.veil

import android.app.Application
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.backend.EmbeddedBackend
import com.playingwithclouds.veil.habits.HabitPreferences
import com.playingwithclouds.veil.habits.Habits
import com.playingwithclouds.veil.privacy.AppLock
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.update.UpdateChecker
import com.playingwithclouds.veil.update.UpdatePreferences

/** Starts the on-device backend once per app process, before any activity. */
class VeilApplication : Application() {

    /**
     * Opens the stored settings, starts the embedded backend, a background update check if due, and
     * the habit tracking (guessed finishes, the subscription check, the usual-time reminder).
     */
    override fun onCreate() {
        super.onCreate()
        ServerSettings.initialize(this)
        PrivacyPreferences.initialize(this)
        AppLock.initialize()
        UpdatePreferences.initialize(this)
        HabitPreferences.initialize(this)
        EmbeddedBackend(this).start()
        UpdateChecker.checkInBackground()
        Habits.start(this)
    }
}
