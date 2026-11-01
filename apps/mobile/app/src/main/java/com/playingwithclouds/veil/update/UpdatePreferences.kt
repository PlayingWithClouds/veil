package com.playingwithclouds.veil.update

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** The update options of Settings → Updates and the checker's bookkeeping, stored on this device. */
object UpdatePreferences {

    private const val PREFERENCES = "updates"
    private const val KEY_AUTO_CHECK = "autoCheck"
    private const val KEY_LAST_CHECK = "lastCheckMillis"
    private const val KEY_DISMISSED_TAG = "dismissedTag"

    private lateinit var preferences: SharedPreferences

    private val mutableAutoCheck = MutableStateFlow(true)
    private val mutableDismissedTag = MutableStateFlow<String?>(null)

    /** Whether the app looks for a new version by itself, at most once a day; on by default. */
    val autoCheck: StateFlow<Boolean> = mutableAutoCheck

    /** The release tag the user chose to skip; it no longer marks the menu. */
    val dismissedTag: StateFlow<String?> = mutableDismissedTag

    /** When the last check succeeded, in epoch milliseconds; 0 if never. */
    val lastCheckMillis: Long
        get() = preferences.getLong(KEY_LAST_CHECK, 0L)

    /** Reads the stored options; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableAutoCheck.value = preferences.getBoolean(KEY_AUTO_CHECK, true)
        mutableDismissedTag.value = preferences.getString(KEY_DISMISSED_TAG, null)
    }

    /** Turns the automatic check on or off. */
    fun setAutoCheck(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_AUTO_CHECK, enabled) }
        mutableAutoCheck.value = enabled
    }

    /** Records when a check succeeded. */
    fun setLastCheckMillis(millis: Long) {
        preferences.edit { putLong(KEY_LAST_CHECK, millis) }
    }

    /** Remembers the release tag the user skipped. */
    fun setDismissedTag(tag: String) {
        preferences.edit { putString(KEY_DISMISSED_TAG, tag) }
        mutableDismissedTag.value = tag
    }
}
