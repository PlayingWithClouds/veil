package com.playingwithclouds.veil

import android.app.Application
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.backend.EmbeddedBackend

/** Starts the on-device backend once per app process, before any activity. */
class VeilApplication : Application() {

    /** Opens the stored server choice and starts the embedded backend. */
    override fun onCreate() {
        super.onCreate()
        ServerSettings.initialize(this)
        EmbeddedBackend(this).start()
    }
}
