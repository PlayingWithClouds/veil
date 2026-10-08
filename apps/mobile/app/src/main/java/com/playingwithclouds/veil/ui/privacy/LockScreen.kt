package com.playingwithclouds.veil.ui.privacy

import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.foundation.background
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.fragment.app.FragmentActivity
import com.playingwithclouds.veil.privacy.AppLock
import com.playingwithclouds.veil.privacy.BiometricUnlock
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.findActivity
import kotlinx.coroutines.delay

/**
 * Covers the whole app until the PIN or a biometric is accepted. Opaque, so nothing behind it
 * shows; back sends the app to the background instead of past the lock.
 */
@Composable
fun LockScreen() {
    val activity = LocalContext.current.findActivity() as? FragmentActivity
    val biometricEnabled = PrivacyPreferences.biometric.value && activity != null && BiometricUnlock.isAvailable(activity)
    var pin by remember { mutableStateOf("") }
    var message by remember { mutableStateOf<String?>(null) }
    var clockMilliseconds by remember { mutableStateOf(SystemClock.elapsedRealtime()) }

    BackHandler { activity?.moveTaskToBack(true) }
    LaunchedEffect(biometricEnabled) {
        if (biometricEnabled && activity != null) {
            BiometricUnlock.prompt(activity, onSuccess = AppLock::unlock)
        }
    }
    LaunchedEffect(Unit) {
        while (true) {
            clockMilliseconds = SystemClock.elapsedRealtime()
            delay(CLOCK_TICK_MILLISECONDS)
        }
    }

    val lockoutSeconds = (AppLock.lockoutRemaining(clockMilliseconds) + MILLISECONDS_PER_SECOND - 1) / MILLISECONDS_PER_SECOND
    val submit = {
        val accepted = AppLock.unlockWithPin(pin, SystemClock.elapsedRealtime())
        if (!accepted) {
            message = "Wrong PIN"
            pin = ""
        }
    }
    Column(
        Modifier.fillMaxSize().background(VeilColors.canvas).systemBarsPadding().padding(VeilSpacing.gutter),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraLarge, Alignment.CenterVertically),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Enter PIN", style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
        PinPad(
            pin = pin,
            onPinChange = { next ->
                pin = next
                message = null
            },
            trailingKey = if (biometricEnabled && activity != null) {
                { TextAction("Scan", onClick = { BiometricUnlock.prompt(activity, onSuccess = AppLock::unlock) }) }
            } else {
                null
            },
        )
        val note = if (lockoutSeconds > 0) "Too many tries. Wait ${lockoutSeconds}s." else message
        Text(note ?: " ", color = VeilColors.error, style = MaterialTheme.typography.bodyMedium)
        PrimaryButton("Unlock", onClick = submit, enabled = pin.length >= PIN_MIN_LENGTH && lockoutSeconds == 0L)
    }
}

private const val CLOCK_TICK_MILLISECONDS = 500L
private const val MILLISECONDS_PER_SECOND = 1000L
