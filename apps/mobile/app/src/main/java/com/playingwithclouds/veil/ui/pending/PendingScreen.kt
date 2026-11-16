package com.playingwithclouds.veil.ui.pending

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.TabTopBar
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** A screen not rebuilt for the new backend yet, opened like the real one would be. */
@Composable
fun PendingScreen(title: String, navigator: AppNavigator) {
    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }) { padding ->
        PendingNotice(title, Modifier.padding(padding))
    }
}

/** A tab whose screen isn't rebuilt yet: its top bar over the notice. */
@Composable
fun PendingTab(title: String, navigator: AppNavigator) {
    Column(Modifier.fillMaxSize()) {
        TabTopBar(navigator, title)
        PendingNotice(title, Modifier.weight(1f))
    }
}

/** Says, in the middle of the page, that this screen comes back later. */
@Composable
private fun PendingNotice(title: String, modifier: Modifier = Modifier) {
    Box(modifier.fillMaxSize().padding(VeilSpacing.huge), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            Text(title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            Text(
                "Not rebuilt for the new backend yet.",
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.contentMuted,
                textAlign = TextAlign.Center,
            )
        }
    }
}
