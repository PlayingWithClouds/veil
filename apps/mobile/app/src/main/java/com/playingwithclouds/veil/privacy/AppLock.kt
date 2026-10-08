package com.playingwithclouds.veil.privacy

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** Whether the app is locked right now, and the PIN check that unlocks it. */
object AppLock {

    private val mutableLocked = MutableStateFlow(false)
    private var backgroundedAtMilliseconds: Long? = null
    private var failedAttempts = 0
    private var lockedOutUntilMilliseconds = 0L

    /** True while the lock screen covers the app. */
    val locked: StateFlow<Boolean> = mutableLocked

    /** Starts locked when the lock is on, since a cold start is a return to the app. */
    fun initialize() {
        mutableLocked.value = PrivacyPreferences.lockEnabled.value
    }

    /** Notes when the app left the foreground. */
    fun onBackgrounded(nowMilliseconds: Long) {
        backgroundedAtMilliseconds = nowMilliseconds
    }

    /** Locks again when the app was away long enough. */
    fun onForegrounded(nowMilliseconds: Long) {
        val lock = LockPolicy.shouldLockOnReturn(
            PrivacyPreferences.lockEnabled.value,
            PrivacyPreferences.lockDelay.value,
            backgroundedAtMilliseconds,
            nowMilliseconds,
        )
        backgroundedAtMilliseconds = null
        if (lock) {
            lockNow()
        }
    }

    /** Locks at once and silences playback; does nothing while the lock is off. */
    fun lockNow() {
        if (!PrivacyPreferences.lockEnabled.value) {
            return
        }
        PrivacyEvents.requestPause()
        mutableLocked.value = true
    }

    /** Opens the app after a successful biometric prompt. */
    fun unlock() {
        failedAttempts = 0
        mutableLocked.value = false
    }

    /** Milliseconds until PIN entry is accepted again; 0 when it is open. */
    fun lockoutRemaining(nowMilliseconds: Long): Long {
        return maxOf(0, lockedOutUntilMilliseconds - nowMilliseconds)
    }

    /** Checks [pin] and unlocks on a match; wrong tries are throttled. */
    fun unlockWithPin(pin: String, nowMilliseconds: Long): Boolean {
        val stored = PrivacyPreferences.pinHash
        if (lockoutRemaining(nowMilliseconds) > 0 || stored == null) {
            return false
        }
        if (PinHasher.verify(pin, stored)) {
            unlock()
            return true
        }
        failedAttempts++
        lockedOutUntilMilliseconds = nowMilliseconds + LockPolicy.lockoutMilliseconds(failedAttempts)
        return false
    }
}
