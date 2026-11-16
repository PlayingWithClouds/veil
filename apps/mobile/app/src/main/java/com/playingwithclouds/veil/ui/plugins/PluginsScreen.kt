package com.playingwithclouds.veil.ui.plugins

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.InstalledPlugin
import com.playingwithclouds.veil.data.PluginPackage
import com.playingwithclouds.veil.data.PluginRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private val PluginIconSize = 32.dp

/** The installed plugins and the packages that could be installed. */
data class PluginsPage(val installed: List<InstalledPlugin>, val available: List<PluginPackage>)

/** Plugins: turning installed ones on and off, updating them, installing published ones. */
class PluginsViewModel : ViewModel() {

    private val pageState = MutableStateFlow<LoadState<PluginsPage>>(LoadState.Loading)
    private val messageState = MutableStateFlow<String?>(null)
    private val busyState = MutableStateFlow(false)

    val page: StateFlow<LoadState<PluginsPage>> = pageState.asStateFlow()
    val message: StateFlow<String?> = messageState.asStateFlow()

    /** An install or update is running. */
    val busy: StateFlow<Boolean> = busyState.asStateFlow()

    init {
        load()
    }

    /** Loads the installed plugins and the published packages side by side. */
    fun load() {
        viewModelScope.launch {
            try {
                val installed = async { PluginRepository.installed() }
                val packages = async { runCatching { PluginRepository.packages() }.getOrDefault(emptyList()) }
                val available = packages.await().filter { found -> found.installedVersion == null }
                // Sites that need a Cloudflare solver can't run on the phone; they're left out rather than shown broken.
                val runnable = installed.await().filter { plugin -> plugin.available }
                pageState.value = LoadState.Loaded(PluginsPage(runnable, available))
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                pageState.value = LoadState.Failed(error.displayMessage())
            }
        }
    }

    /** Turns a plugin on or off; the switch flips at once and back when that fails. */
    fun setEnabled(plugin: InstalledPlugin, enabled: Boolean) {
        replace(plugin.copy(enabled = enabled))
        viewModelScope.launch {
            val saved = runCatching { PluginRepository.setEnabled(plugin.name, enabled) }
            if (saved.isFailure) {
                replace(plugin)
                messageState.value = "Couldn't switch ${plugin.displayName}"
            }
        }
    }

    /** Installs a published package, then lists it among the installed ones. */
    fun install(found: PluginPackage) {
        runBusy("Installed ${found.name}") { PluginRepository.install(found.name) }
    }

    /** Installs newer versions of the installed plugins. */
    fun updateAll() {
        runBusy(null) {
            val updated = PluginRepository.updateAll()
            var note = "Updated ${updated.size} plugins."
            if (updated.isEmpty()) {
                note = "Every plugin is up to date."
            }
            messageState.value = note
        }
    }

    /** The snackbar showed the message. */
    fun messageShown() {
        messageState.value = null
    }

    /** Runs an install or update with the page busy, says [done] after, and reloads. */
    private fun runBusy(done: String?, action: suspend () -> Unit) {
        busyState.value = true
        viewModelScope.launch {
            try {
                action()
                if (done != null) {
                    messageState.value = done
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                messageState.value = error.displayMessage()
            }
            busyState.value = false
            load()
        }
    }

    /** Puts a changed plugin in place of the loaded one with its name. */
    private fun replace(plugin: InstalledPlugin) {
        pageState.update { current ->
            if (current !is LoadState.Loaded) {
                return@update current
            }
            val installed = current.value.installed.map { known ->
                if (known.name == plugin.name) {
                    return@map plugin
                }
                known
            }
            LoadState.Loaded(current.value.copy(installed = installed))
        }
    }
}

/** The installed plugins with a switch each and Update all, then the packages to install. Settings and logins are on the web. */
@Composable
fun PluginsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { PluginsViewModel() }
    val state by viewModel.page.collectAsStateWithLifecycle()
    val busy by viewModel.busy.collectAsStateWithLifecycle()
    val message by viewModel.message.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(message) {
        val note = message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    Scaffold(topBar = { VeilTopBar("Plugins", onBack = navigator::back) }, snackbarHost = { VeilSnackbarHost(snackbarHostState) }) { padding ->
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { page ->
            LazyColumn(contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
                item("update") {
                    Row(Modifier.gutterPadding().padding(top = VeilSpacing.medium)) {
                        SecondaryButton("Update all", onClick = viewModel::updateAll, enabled = !busy, icon = VeilIcons.Refresh)
                    }
                }
                item("installed-heading") { SectionHeading("Installed", modifier = Modifier.gutterPadding()) }
                items(page.installed, key = { plugin -> "installed-${plugin.name}" }) { plugin ->
                    PluginRow(plugin.displayName, pluginFacts(plugin), plugin.iconUrl) {
                        VeilSwitch(checked = plugin.enabled, onCheckedChange = { enabled -> viewModel.setEnabled(plugin, enabled) })
                    }
                }
                if (page.available.isNotEmpty()) {
                    item("available-heading") { SectionHeading("Available", modifier = Modifier.gutterPadding()) }
                    items(page.available, key = { found -> "package-${found.name}" }) { found ->
                        PluginRow(found.name, listOfNotNull(found.version, found.description).joinToString(" · "), null) {
                            SecondaryButton("Install", onClick = { viewModel.install(found) }, enabled = !busy)
                        }
                    }
                }
            }
        }
    }
}

/** "v1.2.0 · what it is": the version and the description. */
private fun pluginFacts(plugin: InstalledPlugin): String {
    val parts = mutableListOf("v${plugin.version}")
    val description = plugin.description
    if (description != null) {
        parts.add(description)
    }
    return parts.joinToString(" · ")
}

/** One plugin or package: icon, name and a line about it, and its control at the end. */
@Composable
private fun PluginRow(name: String, facts: String, iconUrl: String?, control: @Composable () -> Unit) {
    Row(
        Modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (iconUrl != null) {
            RemoteImage(iconUrl, Modifier.size(PluginIconSize).clip(VeilShapes.badge))
        }
        Column(Modifier.weight(1f)) {
            Text(name, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(facts, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted, maxLines = 2, overflow = TextOverflow.Ellipsis)
        }
        control()
    }
}
