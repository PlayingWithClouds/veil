package com.playingwithclouds.veil.ui.plugins

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.PluginInfo
import com.playingwithclouds.veil.data.PluginPackage
import com.playingwithclouds.veil.data.PluginRepository
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.loadInto
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** Installed plugins, their settings and the published plugin catalog. */
class PluginsViewModel : ViewModel() {

    private val mutablePlugins = MutableStateFlow<LoadState<List<PluginInfo>>>(LoadState.Loading)
    private val mutableCatalog = MutableStateFlow<LoadState<List<PluginPackage>>>(LoadState.Loading)
    private val mutableMessage = MutableStateFlow<String?>(null)
    private val mutableBusy = MutableStateFlow(false)

    /** The installed plugins. */
    val plugins: StateFlow<LoadState<List<PluginInfo>>> = mutablePlugins

    /** Packages published in the plugin index. */
    val catalog: StateFlow<LoadState<List<PluginPackage>>> = mutableCatalog

    /** A one-off note for a snackbar. */
    val message: StateFlow<String?> = mutableMessage

    /** Whether an install, update or uninstall is running. */
    val busy: StateFlow<Boolean> = mutableBusy

    init {
        load()
    }

    /** Loads the installed plugins. */
    fun load() {
        viewModelScope.loadInto(mutablePlugins) { PluginRepository.plugins() }
    }

    /** Marks the snackbar note as shown. */
    fun messageShown() {
        mutableMessage.value = null
    }

    /** Enables or disables a plugin. */
    fun toggle(plugin: PluginInfo) {
        runAction(null) { PluginRepository.toggle(plugin.name, !plugin.enabled) }
    }

    /** Saves a plugin's settings. */
    fun saveSettings(plugin: PluginInfo, values: Map<String, String>) {
        runAction("Settings saved") { PluginRepository.saveSettings(plugin.name, values) }
    }

    /** Removes a plugin. */
    fun uninstall(plugin: PluginInfo) {
        runAction("Removed ${plugin.label}") {
            PluginRepository.uninstall(plugin.name)
            waitForReload()
        }
    }

    /** Installs newer versions of the installed plugins. */
    fun checkForUpdates() {
        viewModelScope.launch {
            mutableBusy.value = true
            try {
                val updated = PluginRepository.updateAll()
                if (updated.isEmpty()) {
                    mutableMessage.value = "All plugins are up to date"
                } else {
                    mutableMessage.value = "Updated ${updated.joinToString(", ")}"
                    waitForReload()
                }
                load()
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableMessage.value = error.displayMessage()
            }
            mutableBusy.value = false
        }
    }

    /** Loads the published plugin catalog for a search text. */
    fun searchCatalog(query: String) {
        viewModelScope.loadInto(mutableCatalog) { PluginRepository.packages(query) }
    }

    /** Installs a package from the catalog. */
    fun install(pluginPackage: PluginPackage) {
        runAction("Installed ${pluginPackage.name}") {
            PluginRepository.install(pluginPackage.name)
            waitForReload()
        }
    }

    /** Runs a change, reloads the list and reports the outcome. */
    private fun runAction(successMessage: String?, action: suspend () -> Unit) {
        viewModelScope.launch {
            mutableBusy.value = true
            try {
                action()
                if (successMessage != null) {
                    mutableMessage.value = successMessage
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableMessage.value = error.displayMessage()
            }
            mutableBusy.value = false
            load()
        }
    }

    /** The backend loads a changed plugin folder within a second of the change. */
    private suspend fun waitForReload() {
        delay(PLUGIN_RELOAD_DELAY_MILLISECONDS)
    }

    companion object {
        private const val PLUGIN_RELOAD_DELAY_MILLISECONDS = 1200L
    }
}
