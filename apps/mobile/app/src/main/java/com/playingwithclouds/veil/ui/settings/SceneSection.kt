package com.playingwithclouds.veil.ui.settings

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.scenes.SceneModelState
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.SceneScans
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.scenes.LikedScenesSheet
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.launch

/**
 * The on-phone scene model: where it comes from (a URL or a zip from storage), the background
 * scan, the scenes the user likes and the compilation built from them. Everything past the
 * install card stays hidden until a model is installed.
 */
@Composable
fun SceneSection(navigator: AppNavigator) {
    val state by SceneModelStore.state.collectAsState()
    ModelCard(state)
    if (state.ready == null) {
        return
    }
    ScanCard()
    LikedScenesCard(navigator)
}

/** The model's state, its URL with Download, Pick a zip and Remove. */
@Composable
private fun ModelCard(state: SceneModelState) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var url by remember { mutableStateOf(SceneModelStore.modelUrl) }
    val pickZip = rememberLauncherForActivityResult(ActivityResultContracts.OpenDocument()) { uri ->
        if (uri != null) {
            scope.launch { SceneModelStore.installFromZip(context, uri) }
        }
    }
    val installing = state is SceneModelState.Installing
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text(modelSummary(state), style = MaterialTheme.typography.bodyLarge, color = VeilColors.content)
        VeilTextField(url, onValueChange = { text -> url = text }, label = "Model URL", placeholder = "https://…/model.json or …/model.zip")
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            PrimaryButton("Download", onClick = { scope.launch { SceneModelStore.installFromUrl(url) } }, enabled = url.isNotBlank() && !installing)
            SecondaryButton("Pick a zip", onClick = { pickZip.launch(arrayOf("application/zip", "application/octet-stream")) }, enabled = !installing)
        }
        if (state.ready != null) {
            TextAction("Remove the model", onClick = { scope.launch { SceneModelStore.remove() } })
        }
        Hint(
            "The scene model recognises acts and positions on the phone: it names sections of a video, picks preview " +
                "moments, plays compilations and learns which covers you tap. Nothing leaves the phone. A model is " +
                "model.json with frame.onnx and temporal.onnx beside it, or the three in a zip.",
        )
    }
}

/** What the model card says about the model. */
private fun modelSummary(state: SceneModelState): String {
    return when (state) {
        SceneModelState.Missing -> "No scene model installed."
        SceneModelState.Installing -> "Installing the scene model…"
        is SceneModelState.Ready -> "Scene model ${state.manifest.version}, ${state.manifest.heads.sumOf { head -> head.labels.size }} labels."
        is SceneModelState.Failed -> "Couldn't install: ${state.message}"
    }
}

/** The background scan's switch and progress. */
@Composable
private fun ScanCard() {
    val context = LocalContext.current
    val enabled by SceneScans.enabled.collectAsState()
    val status by SceneScans.status.collectAsState()
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        SwitchRow("Scan videos while charging", enabled, SceneScans::setEnabled)
        Text(scanSummary(status.running, status.progress, status.scannedCount), style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
        if (enabled && !status.running) {
            SecondaryButton("Scan at the next charge", onClick = { SceneScans.scanSoon(context) })
        }
        Hint(
            "While the phone charges (on Wi-Fi if Background asks for it), favourites, collections, the watchlist, the " +
                "queue and recently watched videos are scanned first, then recommendations: a frame every few seconds, " +
                "read straight from the stream. Found scenes are saved as segments; your own and the sites' stay.",
        )
    }
}

/** "Scanning… 40 %" or how many videos are done. */
private fun scanSummary(running: Boolean, progress: Float, scannedCount: Int): String {
    if (running) {
        return "Scanning a video… ${(progress * 100).toInt()} %"
    }
    return "$scannedCount videos scanned with this model."
}

/** The liked scenes with the picker, and the compilation and shorties they play. */
@Composable
private fun LikedScenesCard(navigator: AppNavigator) {
    val liked by SceneRepository.likedTags.collectAsState()
    var picking by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) { runCatching { SceneRepository.refreshLikedTags() } }
    if (picking) {
        LikedScenesSheet(onDismiss = { picking = false })
    }
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        var summary = "No scenes picked yet."
        if (liked.isNotEmpty()) {
            summary = liked.sorted().joinToString(", ")
        }
        Text(summary, style = MaterialTheme.typography.bodyMedium, color = VeilColors.content)
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            SecondaryButton("Pick scenes", onClick = { picking = true })
            PrimaryButton("Play compilation", onClick = navigator::openCompilation, enabled = liked.isNotEmpty())
        }
        Hint("The scenes you like make the compilation, lead the shorties and choose the moments card previews jump through.")
    }
}
