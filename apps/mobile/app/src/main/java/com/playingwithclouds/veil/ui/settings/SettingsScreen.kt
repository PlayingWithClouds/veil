package com.playingwithclouds.veil.ui.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.Button
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.BackendSettings
import com.playingwithclouds.veil.data.BlockedEntity
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.ServerForm
import com.playingwithclouds.veil.ui.components.VeilTopBar

/** Server choice, job and download settings, and the blocklist. */
@Composable
fun SettingsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { SettingsViewModel() }
    val settings by viewModel.settings.collectAsStateWithLifecycle()
    val blocklist by viewModel.blocklist.collectAsStateWithLifecycle()
    val status by viewModel.status.collectAsStateWithLifecycle()

    Scaffold(topBar = { VeilTopBar("Settings", onBack = navigator::back) }) { padding ->
        Column(
            Modifier.padding(padding).verticalScroll(rememberScrollState()).padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            SectionHeading("Server")
            ServerForm()
            HorizontalDivider()
            when (val current = settings) {
                is LoadState.Loading -> Text("Loading backend settings...", color = MaterialTheme.colorScheme.onSurfaceVariant)
                is LoadState.Failed -> Text(current.message, color = MaterialTheme.colorScheme.error)
                is LoadState.Loaded -> BackendSettingsForm(current.value, status, viewModel)
            }
            HorizontalDivider()
            BlocklistSection(blocklist, viewModel)
        }
    }
}

/** A small heading above a block of settings. */
@Composable
private fun SectionHeading(text: String) {
    Text(text, style = MaterialTheme.typography.titleMedium)
}

/** Job and download settings with a save button. */
@Composable
private fun BackendSettingsForm(settings: BackendSettings, status: String?, viewModel: SettingsViewModel) {
    SectionHeading("Jobs and downloads")
    NumberField("Max concurrent jobs", settings.maxConcurrentJobs) { value -> viewModel.edit(settings.copy(maxConcurrentJobs = value)) }
    NumberField("Max job retries", settings.maxJobRetries) { value -> viewModel.edit(settings.copy(maxJobRetries = value)) }
    NumberField("Download speed limit (KB/s, 0 = unlimited)", settings.downloadSpeedLimitKBps) { value ->
        viewModel.edit(settings.copy(downloadSpeedLimitKBps = value))
    }
    SwitchRow("Require VPN for downloads", settings.requireVpn) { value -> viewModel.edit(settings.copy(requireVpn = value)) }
    SwitchRow("Allow downloads while streaming", settings.allowDownloadsWhileStreaming) { value ->
        viewModel.edit(settings.copy(allowDownloadsWhileStreaming = value))
    }
    SwitchRow("Auto-enrich after scrape", settings.autoEnrichAfterScrape) { value ->
        viewModel.edit(settings.copy(autoEnrichAfterScrape = value))
    }
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Button(onClick = viewModel::save) { Text("Save") }
        if (status != null) {
            Text(status, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

/** A whole-number text field; input that is not a number is ignored. */
@Composable
private fun NumberField(label: String, value: Int, onChange: (Int) -> Unit) {
    OutlinedTextField(
        value = value.toString(),
        onValueChange = { text ->
            val number = text.toIntOrNull()
            if (number != null) {
                onChange(number)
            }
        },
        label = { Text(label) },
        singleLine = true,
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
        modifier = Modifier.fillMaxWidth(),
    )
}

/** A label with a switch. */
@Composable
private fun SwitchRow(label: String, checked: Boolean, onChange: (Boolean) -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, Modifier.weight(1f))
        Switch(checked = checked, onCheckedChange = onChange)
    }
}

/** Blocked tags, performers and studios, each with an unblock button. */
@Composable
private fun BlocklistSection(blocklist: List<BlockedEntity>, viewModel: SettingsViewModel) {
    SectionHeading("Blocklist")
    if (blocklist.isEmpty()) {
        Text("Nothing is blocked.", color = MaterialTheme.colorScheme.onSurfaceVariant)
        return
    }
    for (entry in blocklist) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f)) {
                Text(entry.label ?: entry.targetId)
                Text(entry.kind, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            IconButton(onClick = { viewModel.unblock(entry) }) { Icon(Icons.Filled.Close, contentDescription = "Unblock") }
        }
    }
}
