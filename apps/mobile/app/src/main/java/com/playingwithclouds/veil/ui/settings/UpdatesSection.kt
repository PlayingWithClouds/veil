package com.playingwithclouds.veil.ui.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.BuildConfig
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.update.InstallState
import com.playingwithclouds.veil.update.Release
import com.playingwithclouds.veil.util.formatBytes
import kotlin.math.roundToInt

/** The installed version, the automatic check option, "Check now" and, when there is one, the available update. */
@Composable
fun UpdatesSection() {
    val viewModel = viewModel { UpdatesViewModel() }
    val autoCheck by viewModel.autoCheck.collectAsStateWithLifecycle()
    val checkStatus by viewModel.checkStatus.collectAsStateWithLifecycle()
    val available by viewModel.available.collectAsStateWithLifecycle()
    val dismissedTag by viewModel.dismissedTag.collectAsStateWithLifecycle()
    val installState by viewModel.installState.collectAsStateWithLifecycle()
    val context = LocalContext.current

    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text("Installed version", Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge)
            Text(BuildConfig.VERSION_NAME, color = VeilColors.contentMuted)
        }
        SwitchRow("Check for updates automatically", autoCheck, viewModel::setAutoCheck)
        Hint("Looks for a new release on GitHub at most once a day.")
        CheckNowRow(checkStatus, viewModel::checkNow)
    }
    val release = available
    if (release != null) {
        AvailableUpdateCard(
            release = release,
            installState = installState,
            skipped = release.tagName == dismissedTag,
            onUpdate = { viewModel.update(context) },
            onSkip = viewModel::skip,
        )
    }
}

/** The "Check now" button with the last check's outcome beside it. */
@Composable
private fun CheckNowRow(status: CheckStatus, onCheck: () -> Unit) {
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        SecondaryButton("Check now", onClick = onCheck, enabled = status != CheckStatus.Checking)
        when (status) {
            is CheckStatus.Idle -> Unit
            is CheckStatus.Checking -> Text("Checking...", Modifier.weight(1f), color = VeilColors.contentMuted)
            is CheckStatus.UpToDate -> Text("You have the latest version.", Modifier.weight(1f), color = VeilColors.contentMuted)
            is CheckStatus.Failed -> Text(status.message, Modifier.weight(1f), color = VeilColors.error)
        }
    }
}

/** A release newer than the installed app: version, size, notes, progress, and the Update and Skip actions. */
@Composable
private fun AvailableUpdateCard(
    release: Release,
    installState: InstallState,
    skipped: Boolean,
    onUpdate: () -> Unit,
    onSkip: () -> Unit,
) {
    VeilCard(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Text("Version ${release.tagName.removePrefix("v")} is available", style = MaterialTheme.typography.titleSmall)
        if (release.apkSizeBytes > 0) {
            Text(formatBytes(release.apkSizeBytes.toDouble()), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        ReleaseNotes(release.notes)
        InstallProgress(installState)
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            PrimaryButton("Update", onClick = onUpdate, enabled = installState !is InstallState.Downloading)
            if (!skipped) {
                SecondaryButton("Skip this version", onClick = onSkip)
            }
        }
    }
}

/** The release notes behind a show/hide toggle; nothing when the release has none. */
@Composable
private fun ReleaseNotes(notes: String) {
    if (notes.isBlank()) {
        return
    }
    var expanded by rememberSaveable { mutableStateOf(false) }
    if (expanded) {
        TextAction("Hide release notes", onClick = { expanded = false })
        Text(notes.trim(), style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
        return
    }
    TextAction("Show release notes", onClick = { expanded = true })
}

/** The download bar, the installing note or the failure, by [state]. */
@Composable
private fun InstallProgress(state: InstallState) {
    when (state) {
        is InstallState.Idle -> Unit
        is InstallState.Downloading -> DownloadProgress(state.progress)
        is InstallState.Installing -> Hint("Installing. Android may ask you to confirm.")
        is InstallState.Failed -> Text(state.message, color = VeilColors.error, style = MaterialTheme.typography.bodyMedium)
    }
}

/** A bar filled to [progress] with the percentage, or a sweeping bar while the size is unknown. */
@Composable
private fun DownloadProgress(progress: Float?) {
    if (progress == null) {
        LoadingBar()
        Hint("Downloading...")
        return
    }
    ProgressBar(progress)
    Hint("Downloading... ${(progress * 100).roundToInt()}%")
}
