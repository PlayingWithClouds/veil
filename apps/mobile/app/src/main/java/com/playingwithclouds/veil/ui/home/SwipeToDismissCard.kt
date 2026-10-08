package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SwipeToDismissBox
import androidx.compose.material3.SwipeToDismissBoxValue
import androidx.compose.material3.Text
import androidx.compose.material3.rememberSwipeToDismissBoxState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** How far a card has to travel, as a share of its width, before letting go dismisses it. */
private const val DISMISS_THRESHOLD = 0.4f

/**
 * Lets the card be swiped away to the right with "Not interested" showing underneath, then calls
 * [onDismiss]. Only that direction dismisses: swiping left on Home belongs to the tab pager.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SwipeToDismissCard(onDismiss: () -> Unit, content: @Composable () -> Unit) {
    val state = rememberSwipeToDismissBoxState(
        confirmValueChange = { target -> target == SwipeToDismissBoxValue.StartToEnd },
        positionalThreshold = { distance -> distance * DISMISS_THRESHOLD },
    )
    LaunchedEffect(state.currentValue) {
        if (state.currentValue == SwipeToDismissBoxValue.StartToEnd) {
            onDismiss()
        }
    }
    SwipeToDismissBox(
        state = state,
        backgroundContent = { NotInterestedBackground() },
        enableDismissFromEndToStart = false,
    ) {
        Box(Modifier.background(VeilColors.canvas)) { content() }
    }
}

/** What shows under a card being swiped away. */
@Composable
private fun NotInterestedBackground() {
    Row(
        Modifier.fillMaxSize().background(VeilColors.surfaceHigh).padding(horizontal = VeilSpacing.gutter),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(VeilIcons.Dislike, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(24.dp))
        Text("Not interested", style = MaterialTheme.typography.labelLarge, color = VeilColors.content)
    }
}
