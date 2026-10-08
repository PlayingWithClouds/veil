package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.util.restartApp
import kotlinx.coroutines.launch

/** Picks which backend the app talks to: the one on this phone, or another server such as a NAS. */
@Composable
fun ServerForm() {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var address by rememberSaveable { mutableStateOf(ServerSettings.storedUrl.orEmpty()) }
    var checking by remember { mutableStateOf(false) }
    var problem by remember { mutableStateOf<String?>(null) }

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(
            currentServerDescription(),
            style = MaterialTheme.typography.bodyMedium,
            color = VeilColors.contentMuted,
        )
        VeilTextField(
            value = address,
            onValueChange = { newAddress -> address = newAddress },
            label = "Server address",
            placeholder = "http://192.168.1.10:8080",
            modifier = Modifier.fillMaxWidth(),
        )
        if (problem != null) {
            Text(problem.orEmpty(), color = VeilColors.error, style = MaterialTheme.typography.bodySmall)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            PrimaryButton(
                "Connect",
                enabled = address.isNotBlank() && !checking,
                onClick = {
                    scope.launch {
                        checking = true
                        problem = null
                        if (VeilApi.isHealthy(address)) {
                            ServerSettings.save(address)
                            context.restartApp()
                        } else {
                            problem = "No Veil backend answers at that address."
                        }
                        checking = false
                    }
                },
            )
            if (!ServerSettings.usesEmbeddedBackend) {
                SecondaryButton(
                    "Use this phone",
                    onClick = {
                        ServerSettings.clear()
                        context.restartApp()
                    },
                )
            }
        }
    }
}

/** Says which backend is in use. */
private fun currentServerDescription(): String {
    if (ServerSettings.usesEmbeddedBackend) {
        return "Using the backend running on this phone."
    }
    return "Using the server at ${ServerSettings.baseUrl}."
}
