package com.playingwithclouds.veil.ui.plugins

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.PluginInfo
import com.playingwithclouds.veil.data.PluginPackage
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.ConfirmDialog
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SearchField
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Installed plugins: enable, configure, update, remove and install more. */
@Composable
fun PluginsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { PluginsViewModel() }
    val plugins by viewModel.plugins.collectAsStateWithLifecycle()
    val message by viewModel.message.collectAsStateWithLifecycle()
    val busy by viewModel.busy.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    var showingCatalog by remember { mutableStateOf(false) }

    LaunchedEffect(message) {
        val note = message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    if (showingCatalog) {
        CatalogSheet(viewModel, onDismiss = { showingCatalog = false })
    }

    Scaffold(
        topBar = {
            VeilTopBar(
                title = "Plugins",
                onBack = navigator::back,
                actions = {
                    RoundIconButton(
                        VeilIcons.Refresh,
                        contentDescription = "Check for updates",
                        onClick = viewModel::checkForUpdates,
                        enabled = !busy,
                    )
                    RoundIconButton(VeilIcons.Plus, contentDescription = "Install plugins", onClick = {
                        viewModel.searchCatalog("")
                        showingCatalog = true
                    })
                },
            )
        },
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            if (busy) {
                LoadingBar(Modifier.padding(horizontal = 16.dp, vertical = 4.dp))
            }
            LoadStateContent(plugins, onRetry = viewModel::load) { list ->
                if (list.isEmpty()) {
                    EmptyMessage("No plugins installed.")
                    return@LoadStateContent
                }
                LazyColumn(contentPadding = PaddingValues(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(list, key = { plugin -> plugin.name }) { plugin -> PluginCard(plugin, viewModel) }
                }
            }
        }
    }
}

/** One plugin: icon, name, state and its actions. */
@Composable
private fun PluginCard(plugin: PluginInfo, viewModel: PluginsViewModel) {
    var editingSettings by remember { mutableStateOf(false) }
    var confirmingUninstall by remember { mutableStateOf(false) }

    if (editingSettings) {
        PluginSettingsDialog(plugin, onSave = { values ->
            editingSettings = false
            viewModel.saveSettings(plugin, values)
        }, onDismiss = { editingSettings = false })
    }
    if (confirmingUninstall) {
        ConfirmDialog(
            title = "Remove ${plugin.label}?",
            message = "Its folder is deleted. You can install it again from the catalog.",
            confirmLabel = "Remove",
            onConfirm = {
                confirmingUninstall = false
                viewModel.uninstall(plugin)
            },
            onDismiss = { confirmingUninstall = false },
        )
    }

    VeilCard {
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            PluginIcon(plugin)
            Column(Modifier.weight(1f)) {
                Text(plugin.label, style = MaterialTheme.typography.titleMedium)
                Text("v${plugin.version}", style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
            }
            VeilSwitch(checked = plugin.enabled, onCheckedChange = { viewModel.toggle(plugin) })
        }
        val description = plugin.description
        if (!description.isNullOrBlank()) {
            Text(description, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        Text(
            plugin.capabilities.joinToString(" · "),
            style = MaterialTheme.typography.labelSmall,
            color = VeilColors.contentFaint,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        if (!plugin.available) {
            Text("Unavailable: needs FlareSolverr, which this device cannot reach.", color = VeilColors.error, style = MaterialTheme.typography.bodySmall)
        }
        Row(Modifier.padding(top = 4.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            if (plugin.settings.isNotEmpty()) {
                SecondaryButton("Settings", onClick = { editingSettings = true }, icon = VeilIcons.Settings)
            }
            SecondaryButton("Remove", onClick = { confirmingUninstall = true }, icon = VeilIcons.Delete)
        }
    }
}

/** The plugin's icon in a rounded square: an image when it is a URL, else the emoji or a placeholder letter. */
@Composable
private fun PluginIcon(plugin: PluginInfo) {
    Box(
        Modifier.size(44.dp).clip(RoundedCornerShape(12.dp)).background(VeilColors.surfaceHigh),
        contentAlignment = Alignment.Center,
    ) {
        val icon = plugin.iconUrl
        if (icon != null && (icon.startsWith("http") || icon.startsWith("/"))) {
            RemoteImage(icon, Modifier.matchParentSize())
            return@Box
        }
        Text(icon ?: plugin.label.take(1).uppercase(), style = MaterialTheme.typography.titleLarge)
    }
}

/** Edit form for the settings a plugin declares. */
@Composable
private fun PluginSettingsDialog(plugin: PluginInfo, onSave: (Map<String, String>) -> Unit, onDismiss: () -> Unit) {
    val values = remember {
        mutableStateOf(
            plugin.settings.associate { field -> field.key to (plugin.settingValues[field.key] ?: field.defaultValue.orEmpty()) },
        )
    }
    VeilDialog(
        title = "${plugin.label} settings",
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("Cancel", onClick = onDismiss)
            PrimaryButton("Save", onClick = { onSave(values.value) })
        },
    ) {
        Column(Modifier.verticalScroll(rememberScrollState()), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            for (field in plugin.settings) {
                VeilTextField(
                    value = values.value[field.key].orEmpty(),
                    onValueChange = { text -> values.value = values.value + (field.key to text) },
                    label = field.label,
                    supportingText = field.description,
                    modifier = Modifier.fillMaxWidth(),
                )
            }
        }
    }
}

/** The published plugin catalog as a bottom sheet with search and install buttons. */
@Composable
private fun CatalogSheet(viewModel: PluginsViewModel, onDismiss: () -> Unit) {
    val catalog by viewModel.catalog.collectAsStateWithLifecycle()
    var query by remember { mutableStateOf("") }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        SearchField(query, { text -> query = text }, "Search plugins", onSearch = { viewModel.searchCatalog(query) })
        LoadStateContent(catalog, onRetry = { viewModel.searchCatalog(query) }) { packages ->
            if (packages.isEmpty()) {
                EmptyMessage("No plugins found.")
                return@LoadStateContent
            }
            LazyColumn(contentPadding = PaddingValues(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                items(packages, key = { entry -> entry.name }) { entry -> CatalogRow(entry, onInstall = { viewModel.install(entry) }) }
            }
        }
    }
}

/** A catalog package with its version and an install or update button. */
@Composable
private fun CatalogRow(entry: PluginPackage, onInstall: () -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Column(Modifier.weight(1f)) {
            Text(entry.name, style = MaterialTheme.typography.titleSmall)
            Text(
                entry.description ?: "v${entry.version}",
                style = MaterialTheme.typography.bodySmall,
                color = VeilColors.contentMuted,
                maxLines = 2,
            )
        }
        val installed = entry.installedVersion
        if (installed == null) {
            PrimaryButton("Install", onClick = onInstall)
            return@Row
        }
        if (installed != entry.version) {
            PrimaryButton("Update", onClick = onInstall)
            return@Row
        }
        Text("Installed", style = MaterialTheme.typography.labelLarge, color = VeilColors.contentMuted)
    }
}
