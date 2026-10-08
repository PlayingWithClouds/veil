package com.playingwithclouds.veil.ui.privacy

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.glassControl
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** A small label under the status bar on every screen while incognito is on; tapping it ends the session. */
@Composable
fun IncognitoBadge(modifier: Modifier = Modifier) {
    val incognito by PrivacyPreferences.incognito.collectAsState()
    if (!incognito) {
        return
    }
    Box(modifier.fillMaxWidth().statusBarsPadding(), contentAlignment = Alignment.TopCenter) {
        Text(
            "You're incognito · tap to end",
            modifier = Modifier
                .padding(top = VeilSpacing.hairline)
                .pressClickable { PrivacyPreferences.setIncognito(false) }
                .glassControl(VeilShapes.capsule)
                .padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.extraSmall),
            style = MaterialTheme.typography.labelSmall,
            color = VeilColors.contentMuted,
        )
    }
}
