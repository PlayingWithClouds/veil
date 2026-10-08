package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.AppScope

/** Process-wide services that screens share. */
object AppServices {

    /** Logs which recommendations were shown and opened. */
    val impressions = ImpressionLogger(AppScope, { batch -> ImpressionLogger.sendToBackend(batch) })
}
