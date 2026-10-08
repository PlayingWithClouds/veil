package com.playingwithclouds.veil.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.ui.components.ServerForm
import com.playingwithclouds.veil.ui.theme.VeilTheme
import androidx.compose.material3.Button

/** Whether the backend answers yet. */
private enum class BootState {
    WAITING,
    READY,
    UNREACHABLE,
}

/** The app: theme, then the backend wait, then the screens. */
@Composable
fun VeilApp() {
    VeilTheme {
        // Sets the default content color; bare Text would otherwise be black.
        Surface(color = MaterialTheme.colorScheme.background, modifier = Modifier.fillMaxSize()) {
            BootGate { VeilShell() }
        }
    }
}

/** Waits for the backend at launch, which takes a moment after the app starts. */
@Composable
private fun BootGate(content: @Composable () -> Unit) {
    var attempt by remember { mutableIntStateOf(0) }
    var bootState by remember { mutableStateOf(BootState.WAITING) }
    LaunchedEffect(attempt) {
        bootState = BootState.WAITING
        if (VeilApi.awaitBackend()) {
            bootState = BootState.READY
            return@LaunchedEffect
        }
        bootState = BootState.UNREACHABLE
    }
    when (bootState) {
        BootState.READY -> content()
        BootState.WAITING -> BootWaiting()
        BootState.UNREACHABLE -> BootUnreachable(onRetry = { attempt++ })
    }
}

/** Shown while the backend is starting. */
@Composable
private fun BootWaiting() {
    Column(
        Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        CircularProgressIndicator()
        Text("Starting Veil...", Modifier.padding(top = 16.dp), color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

/** Shown when the backend does not come up: retry, or pick another server. */
@Composable
private fun BootUnreachable(onRetry: () -> Unit) {
    Box(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(24.dp)) {
        Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
            Text("The backend is not reachable", style = MaterialTheme.typography.titleLarge)
            Text(
                "Check the server address below, or retry if the on-device backend is still starting.",
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Button(onClick = onRetry) { Text("Retry") }
            ServerForm()
        }
    }
}
