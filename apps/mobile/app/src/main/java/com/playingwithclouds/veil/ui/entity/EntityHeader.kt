package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * The top of a performer, studio or tag page: photo, name, a muted line, then actions. Draws no
 * side padding; it sits in a grid that already pads with the gutter.
 */
@Composable
fun EntityHeader(
    name: String,
    imagePath: String?,
    subtitle: String?,
    showAvatar: Boolean = true,
    actions: @Composable () -> Unit = {},
) {
    Column(
        Modifier.fillMaxWidth().padding(vertical = VeilSpacing.large),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large), verticalAlignment = Alignment.CenterVertically) {
            if (showAvatar) {
                Avatar(imagePath, Modifier.size(88.dp))
            }
            Column(Modifier.weight(1f)) {
                Text(name, style = MaterialTheme.typography.headlineMedium)
                if (subtitle != null) {
                    Text(subtitle, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
                }
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
            actions()
        }
    }
}
