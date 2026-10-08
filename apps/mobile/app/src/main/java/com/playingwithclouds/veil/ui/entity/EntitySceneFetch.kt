package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/**
 * Asks the backend, once per page open, to search every site for a studio's or performer's
 * videos, and calls [onFound] when that credited any to it so the page can reload its scenes.
 */
class EntitySceneFetch(
    private val scope: CoroutineScope,
    private val entityId: String,
    private val onFound: () -> Unit,
) {

    private val mutableFetching = MutableStateFlow(false)

    /** Whether the sites are being searched right now. */
    val fetching: StateFlow<Boolean> = mutableFetching

    /** Starts the search; the backend skips it when it ran within the last day. */
    fun start() {
        mutableFetching.value = true
        scope.launch {
            try {
                val found = EntityRepository.ensureEntityScenes(entityId)
                if (found > 0) {
                    onFound()
                }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                // Best-effort: the page keeps the videos it already has.
            } finally {
                mutableFetching.value = false
            }
        }
    }
}

/** A thin sweeping bar while [fetch] searches the sites, nothing otherwise. */
@Composable
fun EntitySceneFetchBar(fetch: EntitySceneFetch) {
    val fetching by fetch.fetching.collectAsStateWithLifecycle()
    if (fetching) {
        LoadingBar(Modifier.padding(vertical = VeilSpacing.extraSmall))
    }
}
