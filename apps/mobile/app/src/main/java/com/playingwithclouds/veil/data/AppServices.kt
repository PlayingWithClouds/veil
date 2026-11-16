package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.privacy.PrivacyPreferences

/** Process-wide services that screens share. */
object AppServices {

    /** The event log the recommendation engine learns from; silent while incognito. */
    val events = EventLog(
        AppScope,
        { batch -> EventLog.sendToBackend(batch) },
        enabled = { !PrivacyPreferences.incognito.value },
    )

    /** Logs which recommendations were shown and opened. */
    val impressions = ImpressionLogger(events)
}
