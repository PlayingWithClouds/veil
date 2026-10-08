package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** A centered spinner. */
@Composable
fun LoadingIndicator(modifier: Modifier = Modifier) {
    Box(modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Spinner()
    }
}

/** A failure message with a retry button. */
@Composable
fun FailedMessage(message: String, onRetry: () -> Unit, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.fillMaxSize().padding(VeilSpacing.extraLarge),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.large, Alignment.CenterVertically),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(message, color = VeilColors.error, style = MaterialTheme.typography.bodyMedium, textAlign = TextAlign.Center)
        PrimaryButton("Retry", onClick = onRetry)
    }
}

/** A muted centered note for lists without entries. */
@Composable
fun EmptyMessage(text: String, modifier: Modifier = Modifier) {
    Box(modifier.fillMaxWidth().padding(VeilSpacing.huge), contentAlignment = Alignment.Center) {
        Text(text, color = VeilColors.contentMuted, style = MaterialTheme.typography.bodyMedium, textAlign = TextAlign.Center)
    }
}

/** Shows the loading spinner, the failure with retry, or the loaded content. */
@Composable
fun <T> LoadStateContent(
    state: LoadState<T>,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable (T) -> Unit,
) {
    when (state) {
        is LoadState.Loading -> LoadingIndicator(modifier)
        is LoadState.Failed -> FailedMessage(state.message, onRetry, modifier)
        is LoadState.Loaded -> content(state.value)
    }
}
