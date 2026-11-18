package com.playingwithclouds.veil.solver

import android.os.Bundle
import android.view.ViewGroup
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import com.playingwithclouds.veil.ui.solver.ChallengeScreen
import com.playingwithclouds.veil.ui.theme.VeilTheme

/**
 * Shows a site's browser check when it wants a tap ("Verify you are human"): the solver's WebView,
 * moved in here until the check clears, then the activity closes by itself and the page is read
 * in the background as usual.
 */
class ChallengeActivity : ComponentActivity() {

    /** Hosts the WebView the solver is waiting on; closes when there's none (cleared or timed out). */
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WebSolver.attachTo(this)
        setContent {
            VeilTheme {
                val webView by WebSolver.shown.collectAsState()
                LaunchedEffect(webView) {
                    if (webView == null) {
                        finish()
                    }
                }
                val shown = webView
                if (shown != null) {
                    ChallengeScreen(shown, onClose = ::finish)
                }
            }
        }
    }

    /** Takes the WebView out of this screen and gives it back to the application. */
    override fun onDestroy() {
        val shown = WebSolver.shown.value
        (shown?.parent as? ViewGroup)?.removeView(shown)
        WebSolver.attachTo(null)
        super.onDestroy()
    }
}
