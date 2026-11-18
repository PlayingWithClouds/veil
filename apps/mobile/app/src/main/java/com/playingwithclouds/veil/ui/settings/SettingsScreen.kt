package com.playingwithclouds.veil.ui.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.ServerForm
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Server choice, privacy, background work, the scene model, appearance and app updates. */
@Composable
fun SettingsScreen(navigator: AppNavigator) {
    DetailScreen("Settings", onBack = navigator::back) {
        Column(
            Modifier
                .verticalScroll(rememberScrollState())
                .padding(top = LocalTopBarInset.current)
                .padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.large),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        ) {
            SectionHeading("Server")
            VeilCard { ServerForm() }
            SectionHeading("Privacy")
            PrivacySection()
            SectionHeading("Background")
            BackgroundSection()
            SectionHeading("Scene model")
            SceneSection(navigator)
            SectionHeading("Cams")
            CamsSection()
            SectionHeading("Appearance")
            SecondaryButton("Design kit", onClick = navigator::openDesignKit)
            SectionHeading("Updates")
            UpdatesSection()
        }
    }
}

/** A label with a switch. */
@Composable
internal fun SwitchRow(label: String, checked: Boolean, onChange: (Boolean) -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text(label, Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge)
        VeilSwitch(checked = checked, onCheckedChange = onChange)
    }
}
