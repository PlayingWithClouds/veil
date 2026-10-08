package com.playingwithclouds.veil.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.playingwithclouds.veil.data.BackendSettings
import com.playingwithclouds.veil.data.BlockedEntity
import com.playingwithclouds.veil.data.SettingsRepository
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.loadInto
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** The backend's job settings and the blocklist. */
class SettingsViewModel : ViewModel() {

    private val mutableSettings = MutableStateFlow<LoadState<BackendSettings>>(LoadState.Loading)
    private val mutableBlocklist = MutableStateFlow<List<BlockedEntity>>(emptyList())
    private val mutableStatus = MutableStateFlow<String?>(null)

    /** The settings as last loaded or edited. */
    val settings: StateFlow<LoadState<BackendSettings>> = mutableSettings

    /** Everything the user blocked. */
    val blocklist: StateFlow<List<BlockedEntity>> = mutableBlocklist

    /** The outcome of the last save, for a short note. */
    val status: StateFlow<String?> = mutableStatus

    init {
        load()
    }

    /** Loads settings and blocklist. */
    fun load() {
        viewModelScope.loadInto(mutableSettings) { SettingsRepository.settings() }
        viewModelScope.launch {
            mutableBlocklist.value = runCatching { SettingsRepository.blocklist() }.getOrDefault(emptyList())
        }
    }

    /** Keeps an edit in the form until it is saved. */
    fun edit(settings: BackendSettings) {
        mutableSettings.value = LoadState.Loaded(settings)
        mutableStatus.value = null
    }

    /** Saves the edited settings. */
    fun save() {
        val settings = (mutableSettings.value as? LoadState.Loaded)?.value ?: return
        viewModelScope.launch {
            try {
                SettingsRepository.save(settings)
                mutableStatus.value = "Saved"
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableStatus.value = error.displayMessage()
            }
        }
    }

    /** Lifts a block. */
    fun unblock(entry: BlockedEntity) {
        mutableBlocklist.value = mutableBlocklist.value - entry
        viewModelScope.launch { runCatching { SettingsRepository.unblock(entry.targetId) } }
    }
}
