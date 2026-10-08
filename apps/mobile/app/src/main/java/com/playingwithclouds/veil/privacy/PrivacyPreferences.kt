package com.playingwithclouds.veil.privacy

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/**
 * The privacy options of Settings → Privacy, stored on this device. Each option is a [StateFlow]
 * so the UI and the activity follow changes at once. Incognito is the exception: it lives in
 * memory only, so a session ends when the app process does.
 */
object PrivacyPreferences {

    private const val PREFERENCES = "privacy"
    private const val KEY_HIDE_IN_RECENTS = "hideInRecents"
    private const val KEY_LOCK_ENABLED = "lockEnabled"
    private const val KEY_LOCK_DELAY = "lockDelay"
    private const val KEY_BIOMETRIC = "biometric"
    private const val KEY_PIN_HASH = "pinHash"
    private const val KEY_DISGUISE = "disguise"
    private const val KEY_PANIC_FLIP = "panicFlip"
    private const val KEY_PANIC_TAP = "panicTap"
    private const val KEY_ENCRYPT_DOWNLOADS = "encryptDownloads"

    private lateinit var preferences: SharedPreferences

    private val mutableHideInRecents = MutableStateFlow(true)
    private val mutableLockEnabled = MutableStateFlow(false)
    private val mutableLockDelay = MutableStateFlow(LockDelay.IMMEDIATELY)
    private val mutableBiometric = MutableStateFlow(false)
    private val mutableDisguise = MutableStateFlow(Disguise.NONE)
    private val mutablePanicFlip = MutableStateFlow(false)
    private val mutablePanicTap = MutableStateFlow(false)
    private val mutableIncognito = MutableStateFlow(false)
    private val mutableEncryptDownloads = MutableStateFlow(true)

    /** Blank recents card and black screenshots and recordings (FLAG_SECURE); on by default. */
    val hideInRecents: StateFlow<Boolean> = mutableHideInRecents

    /** Whether opening the app asks for the PIN or a biometric. */
    val lockEnabled: StateFlow<Boolean> = mutableLockEnabled

    /** How long the app may stay in the background before it locks. */
    val lockDelay: StateFlow<LockDelay> = mutableLockDelay

    /** Whether a fingerprint or face may unlock instead of the PIN. */
    val biometric: StateFlow<Boolean> = mutableBiometric

    /** The launcher entry the app currently shows. */
    val disguise: StateFlow<Disguise> = mutableDisguise

    /** Whether turning the phone face-down triggers the panic action. */
    val panicFlip: StateFlow<Boolean> = mutablePanicFlip

    /** Whether tapping twice with two fingers triggers the panic action. */
    val panicTap: StateFlow<Boolean> = mutablePanicTap

    /**
     * Whether the embedded backend encrypts new downloads and cached streams at rest. Read when
     * the backend starts, so a change applies after the app restarts.
     */
    val encryptDownloads: StateFlow<Boolean> = mutableEncryptDownloads

    /** Whether this session records no history and no recommendation signals. */
    val incognito: StateFlow<Boolean> = mutableIncognito

    /** The stored PIN hash (see [PinHasher]), if a PIN is set. */
    val pinHash: String?
        get() = preferences.getString(KEY_PIN_HASH, null)

    /** Reads the stored options; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableHideInRecents.value = preferences.getBoolean(KEY_HIDE_IN_RECENTS, true)
        mutableLockEnabled.value = preferences.getBoolean(KEY_LOCK_ENABLED, false) && pinHash != null
        mutableLockDelay.value = LockDelay.fromName(preferences.getString(KEY_LOCK_DELAY, null))
        mutableBiometric.value = preferences.getBoolean(KEY_BIOMETRIC, false)
        mutableDisguise.value = Disguise.fromName(preferences.getString(KEY_DISGUISE, null))
        mutablePanicFlip.value = preferences.getBoolean(KEY_PANIC_FLIP, false)
        mutablePanicTap.value = preferences.getBoolean(KEY_PANIC_TAP, false)
        mutableEncryptDownloads.value = preferences.getBoolean(KEY_ENCRYPT_DOWNLOADS, true)
    }

    /** Turns the recents and screenshot protection on or off. */
    fun setHideInRecents(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_HIDE_IN_RECENTS, enabled) }
        mutableHideInRecents.value = enabled
    }

    /** Turns the app lock on with a new PIN hash. */
    fun enableLock(pinHash: String) {
        preferences.edit {
            putString(KEY_PIN_HASH, pinHash)
            putBoolean(KEY_LOCK_ENABLED, true)
        }
        mutableLockEnabled.value = true
    }

    /** Replaces the PIN while the lock stays on. */
    fun changePin(pinHash: String) {
        preferences.edit { putString(KEY_PIN_HASH, pinHash) }
    }

    /** Turns the app lock off and forgets the PIN and biometric choice. */
    fun disableLock() {
        preferences.edit {
            remove(KEY_PIN_HASH)
            putBoolean(KEY_LOCK_ENABLED, false)
            putBoolean(KEY_BIOMETRIC, false)
        }
        mutableLockEnabled.value = false
        mutableBiometric.value = false
    }

    /** Sets how long the app may stay in the background before it locks. */
    fun setLockDelay(delay: LockDelay) {
        preferences.edit { putString(KEY_LOCK_DELAY, delay.name) }
        mutableLockDelay.value = delay
    }

    /** Allows or forbids biometric unlock. */
    fun setBiometric(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_BIOMETRIC, enabled) }
        mutableBiometric.value = enabled
    }

    /** Remembers which launcher entry is enabled; [Disguises] does the switching. */
    fun setDisguise(disguise: Disguise) {
        preferences.edit { putString(KEY_DISGUISE, disguise.name) }
        mutableDisguise.value = disguise
    }

    /** Turns the face-down trigger on or off. */
    fun setPanicFlip(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_PANIC_FLIP, enabled) }
        mutablePanicFlip.value = enabled
    }

    /** Turns the two-finger double-tap trigger on or off. */
    fun setPanicTap(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_PANIC_TAP, enabled) }
        mutablePanicTap.value = enabled
    }

    /** Turns encryption of new downloads and cached streams on or off (applies at the next start). */
    fun setEncryptDownloads(enabled: Boolean) {
        preferences.edit { putBoolean(KEY_ENCRYPT_DOWNLOADS, enabled) }
        mutableEncryptDownloads.value = enabled
    }

    /** Starts or ends the incognito session. */
    fun setIncognito(enabled: Boolean) {
        mutableIncognito.value = enabled
    }
}
