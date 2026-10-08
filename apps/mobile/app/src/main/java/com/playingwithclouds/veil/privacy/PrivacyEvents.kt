package com.playingwithclouds.veil.privacy

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow

/** One-shot privacy events that screens react to without depending on who raised them. */
object PrivacyEvents {

    private val mutablePauseRequests = MutableSharedFlow<Unit>(extraBufferCapacity = 1)

    /** Emits when playback must stop: the panic action or the app locking. */
    val pauseRequests: SharedFlow<Unit> = mutablePauseRequests

    /** Asks every player to pause. */
    fun requestPause() {
        mutablePauseRequests.tryEmit(Unit)
    }
}
