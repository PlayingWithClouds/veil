package com.playingwithclouds.veil.update

import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.BuildConfig
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

/** Looks for a newer release than the installed app and keeps the answer for the whole app. */
object UpdateChecker {

    private val mutableAvailable = MutableStateFlow<Release?>(null)

    /** The newest release when it is newer than the installed app, from the last check. */
    val available: StateFlow<Release?> = mutableAvailable

    /** Whether an available release has not been skipped, so the menu marks it. */
    val hasUnseenUpdate: StateFlow<Boolean> = combine(mutableAvailable, UpdatePreferences.dismissedTag) { release, dismissedTag ->
        release != null && release.tagName != dismissedTag
    }.stateIn(AppScope, SharingStarted.Eagerly, false)

    /** Starts an automatic check without waiting for it, when it is on and the last one is a day old. */
    fun checkInBackground() {
        if (!UpdatePreferences.autoCheck.value) {
            return
        }
        if (!UpdatePolicy.isCheckDue(UpdatePreferences.lastCheckMillis, System.currentTimeMillis())) {
            return
        }
        AppScope.launch {
            try {
                check()
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                // An automatic check fails quietly and runs again at the next start.
            }
        }
    }

    /** Fetches the latest release and returns it if it is newer than the installed app. */
    suspend fun check(): Release? {
        val latest = ReleaseFeed.fetchLatest()
        UpdatePreferences.setLastCheckMillis(System.currentTimeMillis())
        var newer: Release? = null
        if (latest != null && UpdatePolicy.isNewer(latest.tagName, BuildConfig.VERSION_CODE)) {
            newer = latest
        }
        mutableAvailable.value = newer
        return newer
    }
}
