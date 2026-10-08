package com.playingwithclouds.veil.privacy

import android.content.Context
import androidx.biometric.BiometricManager
import androidx.biometric.BiometricPrompt
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity

/** The system biometric prompt used to unlock the app instead of typing the PIN. */
object BiometricUnlock {

    private const val AUTHENTICATORS = BiometricManager.Authenticators.BIOMETRIC_WEAK

    /** Whether the device has a fingerprint or face enrolled. */
    fun isAvailable(context: Context): Boolean {
        return BiometricManager.from(context).canAuthenticate(AUTHENTICATORS) == BiometricManager.BIOMETRIC_SUCCESS
    }

    /** Shows the prompt; [onSuccess] runs after a match, cancelling leaves the PIN pad to the user. */
    fun prompt(activity: FragmentActivity, onSuccess: () -> Unit) {
        val callback = object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                onSuccess()
            }
        }
        val info = BiometricPrompt.PromptInfo.Builder()
            .setTitle("Unlock")
            .setNegativeButtonText("Use PIN")
            .setAllowedAuthenticators(AUTHENTICATORS)
            .build()
        BiometricPrompt(activity, ContextCompat.getMainExecutor(activity), callback).authenticate(info)
    }
}
