package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.ui.components.InfoRow
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * Label and value pairs; entries without a value are left out. Bleeds past the grid's gutter
 * because [InfoRow] insets itself by the gutter.
 */
@Composable
fun ProfileInfo(rows: List<Pair<String, String?>>, details: String?) {
    Column(Modifier.bleed()) {
        for ((label, value) in rows) {
            if (!value.isNullOrBlank()) {
                InfoRow(label, value)
            }
        }
        if (!details.isNullOrBlank()) {
            Text(
                details,
                modifier = Modifier.gutterPadding().padding(vertical = VeilSpacing.small),
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.contentMuted,
            )
        }
    }
}
