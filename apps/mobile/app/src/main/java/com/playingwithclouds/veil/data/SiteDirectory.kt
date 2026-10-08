package com.playingwithclouds.veil.data

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.util.concurrent.atomic.AtomicBoolean

/** The installed plugin sites, loaded once, so any scene card can show where it came from. */
object SiteDirectory {

    private val mutableSites = MutableStateFlow<List<SearchSite>>(emptyList())

    private val loading = AtomicBoolean(false)

    /** Every installed plugin as a site; empty until loaded. */
    val sites: StateFlow<List<SearchSite>> = mutableSites

    /** Loads the sites unless they are loaded or loading; a failure allows the next call to retry. */
    suspend fun ensureLoaded() {
        if (mutableSites.value.isNotEmpty() || !loading.compareAndSet(false, true)) {
            return
        }
        try {
            mutableSites.value = SearchRepository.installedSites()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Best-effort: cards fall back to the bare host.
        } finally {
            loading.set(false)
        }
    }
}
