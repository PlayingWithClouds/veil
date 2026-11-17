package com.playingwithclouds.veil.ui.theme

import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.LocalRippleConfiguration
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontVariation
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import com.playingwithclouds.veil.R

/**
 * Surface ramp and accents of the web client's design tokens (`routes/layout.css`), plus the
 * translucent layers drawn over images and glass. The only place colours are defined.
 *
 * The accent marks selection and progress: the current tab, selected pills and toggles, resume
 * bars, the seek bar. It is white (the web's violet looked too loud on the phone); the only hue
 * is [live] pink for what's live.
 */
object VeilColors {
    val canvas = Color(0xFF000000)
    val elevated = Color(0xFF0B0B0D)
    val surface = Color(0xFF141416)
    val surfaceHigh = Color(0xFF1C1C1F)
    val content = Color(0xFFFAFAFA)
    val contentMuted = Color(0xFFA1A1AA)
    val contentFaint = Color(0xFF71717A)
    // No colour accent: selection and progress are white, set apart by fill rather than hue.
    val accent = Color(0xFFFAFAFA)
    val onAccent = Color(0xFF0B0B0D)
    val accentSoft = Color.White.copy(alpha = 0.14f)

    /** What's live (cams), as the web's `--color-live`. */
    val live = Color(0xFFF472B6)
    val onLive = Color(0xFF0B0B0D)
    val info = Color(0xFF60A5FA)
    val success = Color(0xFF4ADE80)
    val warning = Color(0xFFFBBF24)
    val error = Color(0xFFF87171)

    /** Behind labels drawn over images (runtime badge). */
    val imageLabel = Color.Black.copy(alpha = 0.75f)

    /** Track of a bar drawn over an image (resume progress). */
    val imageTrack = Color.Black.copy(alpha = 0.5f)

    /** Dims the page behind sheets and full-screen viewers. */
    val scrim = Color.Black.copy(alpha = 0.6f)

    /** Dims the video evenly while the player's controls show, and fills its round centre buttons. */
    val playerDim = Color.Black.copy(alpha = 0.38f)

    /** Shadow under the player's flat icons and text so they read on bright frames. */
    val playerShadow = Color.Black.copy(alpha = 0.55f)

    /** Hairline edge of glass surfaces. */
    val glassEdge = Color.White.copy(alpha = 0.10f)

    /** Tint laid over the blurred backdrop of glass. */
    val glassTint = Color(0xFF101013).copy(alpha = 0.62f)

    /** Glass where nothing behind can be blurred. */
    val glassFallback = surfaceHigh.copy(alpha = 0.92f)

    /** Top of the fill of glass controls sitting in the page: a light sheen... */
    val glassControlTop = Color.White.copy(alpha = 0.18f)

    /** ...fading to this at the bottom. */
    val glassControlBottom = Color.White.copy(alpha = 0.05f)

    /** The rim of glass controls where light catches it, at the top... */
    val glassRimTop = Color.White.copy(alpha = 0.42f)

    /** ...fading to this along the bottom. */
    val glassRimBottom = Color.White.copy(alpha = 0.06f)

    /** The selected slot inside a glass control. */
    val glassSelection = Color.White.copy(alpha = 0.14f)
}

private val veilColorScheme = darkColorScheme(
    primary = VeilColors.content,
    onPrimary = VeilColors.elevated,
    secondary = VeilColors.surfaceHigh,
    onSecondary = VeilColors.contentMuted,
    tertiary = VeilColors.accent,
    onTertiary = VeilColors.canvas,
    // Screens (Scaffold) are see-through so the shell's glow shows behind them; the app's root paints the canvas.
    background = Color.Transparent,
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

/** One weight of the bundled Inter variable font. */
private fun interWeight(weight: FontWeight): Font {
    return Font(R.font.inter, weight, variationSettings = FontVariation.Settings(FontVariation.weight(weight.weight)))
}

/** Inter, the app's only typeface. */
val Inter = FontFamily(
    interWeight(FontWeight.Normal),
    interWeight(FontWeight.Medium),
    interWeight(FontWeight.SemiBold),
    interWeight(FontWeight.Bold),
    interWeight(FontWeight.ExtraBold),
)

/** A text style in Inter; large sizes get tighter tracking, like display type. */
private fun interStyle(size: Int, weight: FontWeight, lineHeight: Int, tracking: Double = 0.0): TextStyle {
    return TextStyle(
        fontFamily = Inter,
        fontWeight = weight,
        fontSize = size.sp,
        lineHeight = lineHeight.sp,
        letterSpacing = tracking.em,
    )
}

private val veilTypography = Typography(
    displayLarge = interStyle(40, FontWeight.ExtraBold, 44, -0.03),
    displayMedium = interStyle(34, FontWeight.ExtraBold, 38, -0.03),
    displaySmall = interStyle(30, FontWeight.Bold, 34, -0.025),
    headlineLarge = interStyle(30, FontWeight.Bold, 34, -0.025),
    headlineMedium = interStyle(26, FontWeight.Bold, 30, -0.02),
    headlineSmall = interStyle(22, FontWeight.Bold, 28, -0.02),
    titleLarge = interStyle(20, FontWeight.Bold, 26, -0.015),
    titleMedium = interStyle(17, FontWeight.SemiBold, 22, -0.01),
    titleSmall = interStyle(15, FontWeight.SemiBold, 20, -0.005),
    bodyLarge = interStyle(16, FontWeight.Normal, 22),
    bodyMedium = interStyle(15, FontWeight.Normal, 20),
    bodySmall = interStyle(13, FontWeight.Normal, 18),
    labelLarge = interStyle(15, FontWeight.SemiBold, 20),
    labelMedium = interStyle(13, FontWeight.Medium, 16),
    labelSmall = interStyle(11, FontWeight.Medium, 14, 0.01),
)

/** Text styles beyond the Material scale, for dense feeds. */
object VeilType {

    /** A scene card's title. */
    val cardTitle = interStyle(14, FontWeight.Medium, 19)

    /** The muted meta line under a scene card's title. */
    val cardMeta = interStyle(12, FontWeight.Normal, 16)

    /** The wordmark in the top bar. */
    val wordmark = interStyle(22, FontWeight.ExtraBold, 26, -0.03)
}

private val veilShapes = Shapes(
    extraSmall = VeilShapes.extraSmall,
    small = VeilShapes.small,
    medium = VeilShapes.medium,
    large = VeilShapes.large,
    extraLarge = VeilShapes.extraLarge,
)

/**
 * The app's theme: Material 3 underneath for its components, restyled with Inter, large radii and
 * no ripples; presses answer with a scale instead (`pressClickable`).
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VeilTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = veilColorScheme, typography = veilTypography, shapes = veilShapes) {
        CompositionLocalProvider(LocalRippleConfiguration provides null, content = content)
    }
}
