package com.playingwithclouds.veil.ui.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.BackendSettings
import com.playingwithclouds.veil.data.BlockedEntity
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.ServerForm
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Server choice, job and download settings, and the blocklist. */
@Composable
fun SettingsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { SettingsViewModel() }
    val settings by viewModel.settings.collectAsStateWithLifecycle()
    val blocklist by viewModel.blocklist.collectAsStateWithLifecycle()
    val status by viewModel.status.collectAsStateWithLifecycle()

    Scaffold(topBar = { VeilTopBar("Settings", onBack = navigator::back) }) { padding ->
        Column(
            Modifier.padding(padding).verticalScroll(rememberScrollState()).padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.large),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        ) {
            SectionHeading("Server")
            VeilCard { ServerForm() }
            SectionHeading("Privacy")
            PrivacySection()
            SectionHeading("Jobs and downloads")
            when (val current = settings) {
                is LoadState.Loading -> Text("Loading backend settings...", color = VeilColors.contentMuted)
                is LoadState.Failed -> Text(current.message, color = VeilColors.error)
                is LoadState.Loaded -> BackendSettingsForm(current.value, status, viewModel)
            }
            SectionHeading("Blocklist")
            BlocklistSection(blocklist, viewModel)
            SectionHeading("Appearance")
            SecondaryButton("Design kit", onClick = navigator::openDesignKit)
        }
    }
}

/** Job and download settings with a save button. */
@Composable
private fun BackendSettingsForm(settings: BackendSettings, status: String?, viewModel: SettingsViewModel) {
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        NumberField("Max concurrent jobs", settings.maxConcurrentJobs) { value -> viewModel.edit(settings.copy(maxConcurrentJobs = value)) }
        NumberField("Max job retries", settings.maxJobRetries) { value -> viewModel.edit(settings.copy(maxJobRetries = value)) }
        NumberField("Download speed limit (KB/s, 0 = unlimited)", settings.downloadSpeedLimitKBps) { value ->
            viewModel.edit(settings.copy(downloadSpeedLimitKBps = value))
        }
    }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        SwitchRow("Require VPN for downloads", settings.requireVpn) { value -> viewModel.edit(settings.copy(requireVpn = value)) }
        SwitchRow("Allow downloads while streaming", settings.allowDownloadsWhileStreaming) { value ->
            viewModel.edit(settings.copy(allowDownloadsWhileStreaming = value))
        }
        SwitchRow("Auto-enrich after scrape", settings.autoEnrichAfterScrape) { value ->
            viewModel.edit(settings.copy(autoEnrichAfterScrape = value))
        }
    }
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        PrimaryButton("Save", onClick = viewModel::save)
        if (status != null) {
            Text(status, color = VeilColors.contentMuted)
        }
    }
}

/** A whole-number text field; input that is not a number is ignored. */
@Composable
private fun NumberField(label: String, value: Int, onChange: (Int) -> Unit) {
    VeilTextField(
        value = value.toString(),
        onValueChange = { text ->
            val number = text.toIntOrNull()
            if (number != null) {
                onChange(number)
            }
        },
        label = label,
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
        modifier = Modifier.fillMaxWidth(),
    )
}

/** A label with a switch. */
@Composable
internal fun SwitchRow(label: String, checked: Boolean, onChange: (Boolean) -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text(label, Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge)
        VeilSwitch(checked = checked, onCheckedChange = onChange)
    }
}

/** Blocked tags, performers and studios, each with an unblock button. */
@Composable
private fun BlocklistSection(blocklist: List<BlockedEntity>, viewModel: SettingsViewModel) {
    if (blocklist.isEmpty()) {
        Text("Nothing is blocked.", color = VeilColors.contentMuted)
        return
    }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
        for (entry in blocklist) {
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f)) {
                    Text(entry.label ?: entry.targetId, style = MaterialTheme.typography.bodyLarge)
                    Text(entry.kind, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
                }
                IconTap(VeilIcons.Close, contentDescription = "Unblock", onClick = { viewModel.unblock(entry) })
            }
        }
    }
}
