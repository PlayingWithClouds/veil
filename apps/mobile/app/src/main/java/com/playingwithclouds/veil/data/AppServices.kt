package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.scenes.CoverLabels

/** Process-wide services that screens share. */
object AppServices {

    /** The event log the recommendation engine learns from; silent while incognito. */
    val events = EventLog(
        AppScope,
        { batch -> EventLog.sendToBackend(batch) },
        enabled = { !PrivacyPreferences.incognito.value },
    )

    /** Logs which recommendations were shown and opened, with what the scene model read off their covers. */
    val impressions = ImpressionLogger(events, CoverLabels, AppScope)
}
