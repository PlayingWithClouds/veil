package com.playingwithclouds.veil.ui.settings

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.update.InstallState
import com.playingwithclouds.veil.update.Release
import com.playingwithclouds.veil.update.UpdateChecker
import com.playingwithclouds.veil.update.UpdateInstaller
import com.playingwithclouds.veil.update.UpdatePreferences
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** The outcome of the last manual update check. */
sealed interface CheckStatus {
    data object Idle : CheckStatus
    data object Checking : CheckStatus
    data object UpToDate : CheckStatus
    data class Failed(val message: String) : CheckStatus
}

/** Settings → Updates: the automatic check option, manual checks, and installing an available release. */
class UpdatesViewModel : ViewModel() {

    private val mutableCheckStatus = MutableStateFlow<CheckStatus>(CheckStatus.Idle)

    /** How the last "Check now" went. */
    val checkStatus: StateFlow<CheckStatus> = mutableCheckStatus

    /** Whether the app checks by itself once a day. */
    val autoCheck: StateFlow<Boolean> = UpdatePreferences.autoCheck

    /** A release newer than the installed app, if a check found one. */
    val available: StateFlow<Release?> = UpdateChecker.available

    /** The release tag the user skipped. */
    val dismissedTag: StateFlow<String?> = UpdatePreferences.dismissedTag

    /** Download and install progress of the update. */
    val installState: StateFlow<InstallState> = UpdateInstaller.state

    /** Looks for a newer release now; ignored while a check is running. */
    fun checkNow() {
        if (mutableCheckStatus.value == CheckStatus.Checking) {
            return
        }
        mutableCheckStatus.value = CheckStatus.Checking
        viewModelScope.launch {
            try {
                val newer = UpdateChecker.check()
                mutableCheckStatus.value = statusAfterCheck(newer)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableCheckStatus.value = CheckStatus.Failed(error.displayMessage())
            }
        }
    }

    /** "Up to date" when nothing newer was found; otherwise the update card says it all. */
    private fun statusAfterCheck(newer: Release?): CheckStatus {
        if (newer == null) {
            return CheckStatus.UpToDate
        }
        return CheckStatus.Idle
    }

    /** Turns the automatic check on or off. */
    fun setAutoCheck(enabled: Boolean) {
        UpdatePreferences.setAutoCheck(enabled)
    }

    /** Downloads and installs the available release. */
    fun update(context: Context) {
        val release = available.value ?: return
        UpdateInstaller.start(context, release)
    }

    /** Stops marking the available release in the menu. */
    fun skip() {
        val release = available.value ?: return
        UpdatePreferences.setDismissedTag(release.tagName)
    }
}
