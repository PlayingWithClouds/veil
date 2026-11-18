package com.playingwithclouds.veil.ui.solver

import android.view.ViewGroup
import android.webkit.WebView
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * The site's browser check, shown only when it wants a tap: a short line on why, the page itself,
 * and Close (the page is then skipped). It closes by itself once the check clears.
 */
@Composable
fun ChallengeScreen(webView: WebView, onClose: () -> Unit) {
    Column(Modifier.fillMaxSize().background(VeilColors.canvas).statusBarsPadding().navigationBarsPadding()) {
        Column(Modifier.fillMaxWidth().padding(VeilSpacing.gutter), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            Text("Confirm you're human", style = MaterialTheme.typography.titleMedium, color = VeilColors.content)
            Text(
                "A site checks for a browser before it shows its pages. Tap the check below once; this closes by itself.",
                style = MaterialTheme.typography.bodySmall,
                color = VeilColors.contentMuted,
            )
            TextAction("Close", onClick = onClose)
        }
        Box(Modifier.fillMaxSize()) {
            AndroidView(
                factory = {
                    (webView.parent as? ViewGroup)?.removeView(webView)
                    webView
                },
                modifier = Modifier.fillMaxSize(),
            )
        }
    }
}
