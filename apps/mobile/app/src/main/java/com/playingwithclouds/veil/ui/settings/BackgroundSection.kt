package com.playingwithclouds.veil.ui.settings

import android.Manifest
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import com.playingwithclouds.veil.habits.HabitPreferences
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** The reminder before the usual time and the Wi-Fi rule for background work. */
@Composable
fun BackgroundSection() {
    ReminderCard()
    WifiOnlyCard()
}

/** The usual-time reminder; switching it on asks for the notification permission where Android needs it. */
@Composable
private fun ReminderCard() {
    val nudge by HabitPreferences.nudge.collectAsState()
    val askPermission = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
        HabitPreferences.setNudge(granted)
    }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Reminder before your usual time", nudge) { turnOn ->
            if (turnOn && Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                askPermission.launch(Manifest.permission.POST_NOTIFICATIONS)
                return@SwitchRow
            }
            HabitPreferences.setNudge(turnOn)
        }
        Hint(
            "One notification 15 minutes before the time you usually watch, learned from when you watch and finish. " +
                "It reads only \"Reminder\", has no picture and stays hidden on the lock screen. At most once a day, " +
                "and not when you've already opened the app that day.",
        )
    }
}

/** Whether subscriptions and video scanning wait for Wi-Fi. */
@Composable
private fun WifiOnlyCard() {
    val wifiOnly by HabitPreferences.wifiOnly.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Background work on Wi-Fi only", wifiOnly, HabitPreferences::setWifiOnly)
        Hint("Subscriptions are checked about every 6 hours and videos scanned while charging. With this on, both wait for Wi-Fi to save mobile data.")
    }
}
