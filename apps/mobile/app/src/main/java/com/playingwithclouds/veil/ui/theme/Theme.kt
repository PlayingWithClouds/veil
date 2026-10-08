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
 * The accent marks selection and progress: the current tab, selected pills and toggles,
 * resume bars, why a scene is recommended. Main actions stay white ([content]).
 */
object VeilColors {
    val canvas = Color(0xFF000000)
    val elevated = Color(0xFF0B0B0D)
    val surface = Color(0xFF141416)
    val surfaceHigh = Color(0xFF1C1C1F)
    val content = Color(0xFFFAFAFA)
    val contentMuted = Color(0xFFA1A1AA)
    val contentFaint = Color(0xFF71717A)
    val accent = Color(0xFFA78BFA)
    val onAccent = Color(0xFF0B0B0D)
    val accentSoft = Color(0xFFA78BFA).copy(alpha = 0.18f)
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

    /** Hairline edge of glass surfaces. */
    val glassEdge = Color.White.copy(alpha = 0.10f)

    /** Tint laid over the blurred backdrop of glass. */
    val glassTint = Color(0xFF101013).copy(alpha = 0.62f)

    /** Glass where nothing behind can be blurred. */
    val glassFallback = surfaceHigh.copy(alpha = 0.92f)

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
