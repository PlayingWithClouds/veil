package com.playingwithclouds.veil.ui.privacy

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.glassControl
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Shortest and longest PIN the app lock accepts. */
const val PIN_MIN_LENGTH = 4
const val PIN_MAX_LENGTH = 8

/**
 * Dots for the digits typed so far above a number keypad. The caller owns [pin]; [onPinChange]
 * gets every edit. [trailingKey] fills the bottom-left key (e.g. the biometric shortcut).
 */
@Composable
fun PinPad(
    pin: String,
    onPinChange: (String) -> Unit,
    modifier: Modifier = Modifier,
    trailingKey: (@Composable () -> Unit)? = null,
) {
    Column(modifier, horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraLarge)) {
        PinDots(pin.length)
        Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            for (rowDigits in listOf("123", "456", "789")) {
                Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
                    for (digit in rowDigits) {
                        PinKey(digit.toString(), onClick = { addDigit(pin, digit, onPinChange) })
                    }
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
                Box(Modifier.size(KEY_SIZE_DP.dp), contentAlignment = Alignment.Center) { trailingKey?.invoke() }
                PinKey("0", onClick = { addDigit(pin, '0', onPinChange) })
                PinKey("Del", onClick = { onPinChange(pin.dropLast(1)) })
            }
        }
    }
}

private const val KEY_SIZE_DP = 72

/** Appends a digit unless the PIN is already as long as allowed. */
private fun addDigit(pin: String, digit: Char, onPinChange: (String) -> Unit) {
    if (pin.length < PIN_MAX_LENGTH) {
        onPinChange(pin + digit)
    }
}

/** One dot per typed digit, with empty slots up to the minimum length. */
@Composable
private fun PinDots(length: Int) {
    Row(
        Modifier.size(width = 240.dp, height = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium, Alignment.CenterHorizontally),
    ) {
        val slots = maxOf(PIN_MIN_LENGTH, length)
        for (index in 0 until slots) {
            val color = if (index < length) VeilColors.content else VeilColors.surfaceHigh
            Box(Modifier.size(16.dp).clip(VeilShapes.capsule).background(color))
        }
    }
}

/** A round glass key. */
@Composable
private fun PinKey(label: String, onClick: () -> Unit) {
    Box(
        Modifier.size(KEY_SIZE_DP.dp).pressClickable(onClick).glassControl(VeilShapes.capsule),
        contentAlignment = Alignment.Center,
    ) {
        Text(label, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
    }
}
