package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.InfoRow

/** Label and value pairs; entries without a value are left out. */
@Composable
fun ProfileInfo(rows: List<Pair<String, String?>>, details: String?) {
    Column {
        for ((label, value) in rows) {
            if (!value.isNullOrBlank()) {
                InfoRow(label, value)
            }
        }
        if (!details.isNullOrBlank()) {
            Text(
                details,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}
