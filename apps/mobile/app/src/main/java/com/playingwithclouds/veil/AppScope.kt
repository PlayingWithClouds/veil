package com.playingwithclouds.veil

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

/** Work that outlives any screen: impression batches, last-moment progress saves. */
object AppScope : CoroutineScope by CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
