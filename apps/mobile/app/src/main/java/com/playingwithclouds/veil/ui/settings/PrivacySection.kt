package com.playingwithclouds.veil.ui.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import com.playingwithclouds.veil.privacy.BiometricUnlock
import com.playingwithclouds.veil.privacy.Disguise
import com.playingwithclouds.veil.privacy.Disguises
import com.playingwithclouds.veil.privacy.LockDelay
import com.playingwithclouds.veil.privacy.PinHasher
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.ui.components.ConfirmDialog
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.privacy.PIN_MIN_LENGTH
import com.playingwithclouds.veil.ui.privacy.PinPad
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Every privacy option together: recents, app lock, panic, launcher disguise, download encryption and incognito. */
@Composable
fun PrivacySection() {
    ScreenProtectionCard()
    AppLockCard()
    PanicCard()
    DisguiseCard()
    DownloadEncryptionCard()
    IncognitoCard()
}

/** The FLAG_SECURE switch. */
@Composable
private fun ScreenProtectionCard() {
    val hide by PrivacyPreferences.hideInRecents.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Hide in recents and screenshots", hide, PrivacyPreferences::setHideInRecents)
        Hint("The recents card stays blank and screenshots and screen recordings come out black.")
    }
}

/** Turning the lock on or off, the PIN, how soon it locks and biometric unlock. */
@Composable
private fun AppLockCard() {
    val context = LocalContext.current
    val enabled by PrivacyPreferences.lockEnabled.collectAsState()
    val delay by PrivacyPreferences.lockDelay.collectAsState()
    val biometric by PrivacyPreferences.biometric.collectAsState()
    var settingPin by remember { mutableStateOf(false) }
    var changingPin by remember { mutableStateOf(false) }

    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        SwitchRow("App lock", enabled) { turnOn ->
            if (turnOn) {
                settingPin = true
            } else {
                PrivacyPreferences.disableLock()
            }
        }
        Hint("Ask for a PIN when the app opens. A wrong PIN several times in a row makes you wait.")
        if (enabled) {
            Text("Lock", style = MaterialTheme.typography.labelMedium, color = VeilColors.contentMuted)
            FlowOptions(LockDelay.entries, { option -> option.label }, { option -> option == delay }, PrivacyPreferences::setLockDelay)
            if (BiometricUnlock.isAvailable(context)) {
                SwitchRow("Unlock with fingerprint or face", biometric, PrivacyPreferences::setBiometric)
            }
            SecondaryButton("Change PIN", onClick = { changingPin = true })
        }
    }
    if (settingPin) {
        NewPinDialog(
            onDismiss = { settingPin = false },
            onConfirm = { pin ->
                PrivacyPreferences.enableLock(PinHasher.hash(pin))
                settingPin = false
            },
        )
    }
    if (changingPin) {
        NewPinDialog(
            onDismiss = { changingPin = false },
            onConfirm = { pin ->
                PrivacyPreferences.changePin(PinHasher.hash(pin))
                changingPin = false
            },
        )
    }
}

/** Two entries of the PIN, the second confirming the first. */
@Composable
private fun NewPinDialog(onDismiss: () -> Unit, onConfirm: (String) -> Unit) {
    var first by remember { mutableStateOf<String?>(null) }
    var pin by remember { mutableStateOf("") }
    var mismatch by remember { mutableStateOf(false) }
    val confirming = first != null
    VeilDialog(
        title = if (confirming) "Repeat the PIN" else "Choose a PIN",
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("Cancel", onClick = onDismiss)
            PrimaryButton(
                if (confirming) "Save" else "Next",
                enabled = pin.length >= PIN_MIN_LENGTH,
                onClick = {
                    val firstEntry = first
                    if (firstEntry == null) {
                        first = pin
                        pin = ""
                    } else if (firstEntry == pin) {
                        onConfirm(pin)
                    } else {
                        first = null
                        pin = ""
                        mismatch = true
                    }
                },
            )
        },
    ) {
        if (mismatch && !confirming) {
            Text("The PINs differ. Start again.", color = VeilColors.error, style = MaterialTheme.typography.bodyMedium)
        }
        PinPad(pin = pin, onPinChange = { next -> pin = next })
    }
}

/** The two panic triggers. */
@Composable
private fun PanicCard() {
    val flip by PrivacyPreferences.panicFlip.collectAsState()
    val tap by PrivacyPreferences.panicTap.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        SwitchRow("Panic: lay the phone face-down", flip, PrivacyPreferences::setPanicFlip)
        SwitchRow("Panic: tap twice with two fingers", tap, PrivacyPreferences::setPanicTap)
        Hint("Pauses playback, locks the app if a lock is set and goes to the home screen.")
    }
}

/** The launcher name and icon choices. */
@Composable
private fun DisguiseCard() {
    val context = LocalContext.current
    val current by PrivacyPreferences.disguise.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text("App name and icon", style = MaterialTheme.typography.bodyLarge)
        FlowOptions(Disguise.entries, { option -> option.label }, { option -> option == current }) { option ->
            Disguises.apply(context, option)
            PrivacyPreferences.setDisguise(option)
        }
        Hint("Changes the launcher entry. Your launcher may take a moment to refresh; the icon is a placeholder for now.")
    }
}

/** The switch for encrypting new downloads and cached streams at rest. */
@Composable
private fun DownloadEncryptionCard() {
    val encrypt by PrivacyPreferences.encryptDownloads.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Encrypt downloads", encrypt, PrivacyPreferences::setEncryptDownloads)
        Hint(
            "Downloads and cached streams are stored encrypted with a key kept in the Android Keystore, " +
                "so the files are unreadable outside the app. Applies to new files after the app restarts; " +
                "this phone's own backend only.",
        )
    }
}

/** The incognito switch, with a short explanation when it turns on. */
@Composable
private fun IncognitoCard() {
    val incognito by PrivacyPreferences.incognito.collectAsState()
    var explaining by remember { mutableStateOf(false) }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Incognito", incognito) { turnOn ->
            if (turnOn) {
                explaining = true
            } else {
                PrivacyPreferences.setIncognito(false)
            }
        }
        Hint("No watch progress, searches or recommendation signals are saved. Ends when the app closes.")
    }
    if (explaining) {
        ConfirmDialog(
            title = "You've gone incognito",
            message = "Watch progress, searches and what you open or skip are not recorded, so recommendations don't learn from this session. " +
                "Likes, ratings, lists and downloads still count, and a scene's page is still fetched the first time you open it.",
            confirmLabel = "Got it",
            onConfirm = {
                PrivacyPreferences.setIncognito(true)
                explaining = false
            },
            onDismiss = { explaining = false },
        )
    }
}

/** Wrapping pills, one per option, the current one selected. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun <T> FlowOptions(options: List<T>, labelOf: (T) -> String, isSelected: (T) -> Boolean, onSelect: (T) -> Unit) {
    FlowRow(
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        for (option in options) {
            Pill(labelOf(option), onClick = { onSelect(option) }, selected = isSelected(option))
        }
    }
}

/** A muted explanation under a setting. */
@Composable
private fun Hint(text: String) {
    Text(text, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentFaint)
}
