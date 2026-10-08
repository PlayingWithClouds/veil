package com.playingwithclouds.veil

import android.app.Application
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.backend.EmbeddedBackend
import com.playingwithclouds.veil.privacy.AppLock
import com.playingwithclouds.veil.privacy.PrivacyPreferences

/** Starts the on-device backend once per app process, before any activity. */
class VeilApplication : Application() {

    /** Opens the stored server choice and starts the embedded backend. */
    override fun onCreate() {
        super.onCreate()
        ServerSettings.initialize(this)
        PrivacyPreferences.initialize(this)
        AppLock.initialize()
        EmbeddedBackend(this).start()
    }
}
