package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.util.tagLabel

/** Tags as a wrapping row of pills. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun TagChips(tags: List<EntityRef>, onClick: (EntityRef) -> Unit, modifier: Modifier = Modifier) {
    if (tags.isEmpty()) {
        return
    }
    FlowRow(
        modifier = modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        for (tag in tags) {
            Pill(tagLabel(tag.name), onClick = { onClick(tag) })
        }
    }
}
