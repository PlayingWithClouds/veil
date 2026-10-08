package com.playingwithclouds.veil.ui

import com.playingwithclouds.veil.data.SearchRepository
import com.playingwithclouds.veil.data.SearchSite
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/**
 * The site badges' state for a list: the sites that list [offers]' kind of content and the ones
 * picked. [onChange] runs after every change of the pick so the owner can reload its list.
 */
class SiteFilter(scope: CoroutineScope, private val offers: (SearchSite) -> Boolean, private val onChange: () -> Unit) {

    private val mutableSites = MutableStateFlow<List<SearchSite>>(emptyList())

    /** The sites offered as badges. */
    val sites: StateFlow<List<SearchSite>> = mutableSites

    private val mutableSelected = MutableStateFlow<Set<String>>(emptySet())

    /** The plugin names the list is narrowed to; empty means every site. */
    val selected: StateFlow<Set<String>> = mutableSelected

    init {
        scope.launch { load() }
    }

    /** The picked plugin names as a query argument. */
    fun selectedSources(): List<String> {
        return mutableSelected.value.toList()
    }

    /** Adds or removes a site. */
    fun toggle(name: String) {
        val current = mutableSelected.value
        if (current.contains(name)) {
            mutableSelected.value = current - name
        } else {
            mutableSelected.value = current + name
        }
        onChange()
    }

    /** Back to every site. */
    fun clear() {
        if (mutableSelected.value.isEmpty()) {
            return
        }
        mutableSelected.value = emptySet()
        onChange()
    }

    /** Loads the sites; without them the badges simply stay hidden. */
    private suspend fun load() {
        try {
            mutableSites.value = SearchRepository.searchSites().filter(offers)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            // Badges are optional.
        }
    }
}
