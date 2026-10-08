package com.playingwithclouds.veil

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.playingwithclouds.veil.ui.VeilApp

/** The single activity; every screen is Compose. */
class MainActivity : ComponentActivity() {

    /** Draws edge to edge and hands the window to the Compose app. */
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent { VeilApp() }
    }
}
