package com.playingwithclouds.veil.ui.settings

import android.content.Intent
import android.provider.Settings
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import com.playingwithclouds.veil.cams.CamRuleStore
import com.playingwithclouds.veil.cams.CamWatch
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.cams.CamRulesSheet
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * Cams → scene rules: how many there are and the sheet with all of them, background watching and,
 * when Android refused to start it, the way to let the app run unrestricted.
 */
@Composable
fun CamsSection() {
    val context = LocalContext.current
    val rules by CamRuleStore.rules.collectAsState()
    val background by CamRuleStore.background.collectAsState()
    val blocked by CamRuleStore.backgroundBlocked.collectAsState()
    val watching by CamWatch.watching.collectAsState()
    val model by SceneModelStore.state.collectAsState()
    var showingRules by remember { mutableStateOf(false) }
    if (model.ready == null) {
        VeilCard { Hint("Scene rules for cam rooms need the scene model.") }
        return
    }
    if (showingRules) {
        CamRulesSheet(roomId = null, roomName = null, onDismiss = { showingRules = false })
    }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        var summary = "${rules.size} scene rules"
        if (watching.isNotEmpty()) {
            summary += " · watching ${watching.size} rooms"
        }
        Text(summary, style = MaterialTheme.typography.bodyLarge, color = VeilColors.content)
        SecondaryButton("Scene rules", onClick = { showingRules = true })
        SwitchRow("Watch followed rooms in the background", background, CamRuleStore::setBackground)
        if (blocked) {
            Text(
                "Android stopped background watching from starting. Let the app run unrestricted in its battery settings.",
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.warning,
            )
            SecondaryButton("Battery settings", onClick = { context.startActivity(Intent(Settings.ACTION_IGNORE_BATTERY_OPTIMIZATION_SETTINGS)) })
        }
        Hint(
            "While a room is open, a frame every 5 seconds goes through the scene model. In the background, up to 3 live " +
                "followed or auto-recorded rooms with a rule are watched the same way, on their lowest quality, while the " +
                "battery is above 20 % or charging, on Wi-Fi when Background asks for it. A silent \"Syncing\" notification " +
                "shows while it runs; alerts read only \"Reminder\" and stay off the lock screen.",
        )
    }
}
