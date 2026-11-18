package com.playingwithclouds.veil.habits

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** The options of Settings → Background and the nudge's bookkeeping, stored on this device. */
object HabitPreferences {

    private const val PREFERENCES = "habits"
    private const val KEY_NUDGE = "nudge"
    private const val KEY_WIFI_ONLY = "wifiOnly"
    private const val KEY_LAST_OPEN = "lastOpenMillis"
    private const val KEY_LAST_NUDGE = "lastNudgeMillis"

    private lateinit var preferences: SharedPreferences

    private val mutableNudge = MutableStateFlow(true)
    private val mutableWifiOnly = MutableStateFlow(false)

    /** Whether one discreet notification comes 15 minutes before the usual time; on by default. */
    val nudge: StateFlow<Boolean> = mutableNudge

    /** Whether background work (subscriptions, scanning videos) waits for Wi-Fi; off by default. */
    val wifiOnly: StateFlow<Boolean> = mutableWifiOnly

    /** When the app was last opened, in epoch milliseconds; 0 if never. */
    val lastOpenMillis: Long
        get() = preferences.getLong(KEY_LAST_OPEN, 0L)

    /** When the last nudge was shown, in epoch milliseconds; 0 if never. */
    val lastNudgeMillis: Long
        get() = preferences.getLong(KEY_LAST_NUDGE, 0L)

    /** Reads the stored options; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableNudge.value = preferences.getBoolean(KEY_NUDGE, true)
        mutableWifiOnly.value = preferences.getBoolean(KEY_WIFI_ONLY, false)
    }

    /** Turns the nudge on or off. */
    fun setNudge(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_NUDGE, enabled) }
        mutableNudge.value = enabled
    }

    /** Makes background work wait for Wi-Fi, or not. */
    fun setWifiOnly(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_WIFI_ONLY, enabled) }
        mutableWifiOnly.value = enabled
    }

    /** Records that the app came to the front. */
    fun setLastOpenMillis(millis: Long) {
        preferences.edit { putLong(KEY_LAST_OPEN, millis) }
    }

    /** Records that a nudge was shown. */
    fun setLastNudgeMillis(millis: Long) {
        preferences.edit { putLong(KEY_LAST_NUDGE, millis) }
    }
}
