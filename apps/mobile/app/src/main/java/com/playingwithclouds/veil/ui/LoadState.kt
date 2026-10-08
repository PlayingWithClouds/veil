package com.playingwithclouds.veil.ui

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch

/** The lifecycle of data loaded once: loading, loaded or failed. */
sealed interface LoadState<out T> {
    data object Loading : LoadState<Nothing>
    data class Loaded<T>(val value: T) : LoadState<T>
    data class Failed(val message: String) : LoadState<Nothing>
}

/** The text to show for a failure. */
fun Throwable.displayMessage(): String {
    val text = message
    if (text.isNullOrEmpty()) {
        return "Something went wrong"
    }
    return text
}

/** Loads a value into the state: Loading first, then Loaded or Failed. */
fun <T> CoroutineScope.loadInto(state: MutableStateFlow<LoadState<T>>, block: suspend () -> T) {
    state.value = LoadState.Loading
    launch {
        try {
            state.value = LoadState.Loaded(block())
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            state.value = LoadState.Failed(error.displayMessage())
        }
    }
}
