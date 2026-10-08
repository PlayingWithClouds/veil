package com.playingwithclouds.veil.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

/** Surface ramp and accents of the web client's design tokens (`routes/layout.css`). */
object VeilColors {
    val canvas = Color(0xFF000000)
    val elevated = Color(0xFF0B0B0D)
    val surface = Color(0xFF141416)
    val surfaceHigh = Color(0xFF1C1C1F)
    val content = Color(0xFFFAFAFA)
    val contentMuted = Color(0xFFA1A1AA)
    val contentFaint = Color(0xFF71717A)
    val accent = Color(0xFFA78BFA)
    val info = Color(0xFF60A5FA)
    val success = Color(0xFF4ADE80)
    val warning = Color(0xFFFBBF24)
    val error = Color(0xFFF87171)
}

private val veilColorScheme = darkColorScheme(
    primary = VeilColors.content,
    onPrimary = VeilColors.elevated,
    secondary = VeilColors.surfaceHigh,
    onSecondary = VeilColors.contentMuted,
    tertiary = VeilColors.accent,
    onTertiary = VeilColors.canvas,
    background = VeilColors.canvas,
    onBackground = VeilColors.content,
    surface = VeilColors.elevated,
    onSurface = VeilColors.content,
    surfaceVariant = VeilColors.surface,
    onSurfaceVariant = VeilColors.contentMuted,
    surfaceContainer = VeilColors.surface,
    surfaceContainerHigh = VeilColors.surfaceHigh,
    surfaceContainerHighest = VeilColors.surfaceHigh,
    outline = VeilColors.contentFaint,
    outlineVariant = VeilColors.surfaceHigh,
    error = VeilColors.error,
    onError = VeilColors.canvas,
)

/** The app's dark Material 3 theme. */
@Composable
fun VeilTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = veilColorScheme, content = content)
}
