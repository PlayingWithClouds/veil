package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.tagLabel

/** Tags as a wrapping row of tag pills, inset by the gutter. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun TagChips(tags: List<EntityRef>, onClick: (EntityRef) -> Unit, modifier: Modifier = Modifier) {
    if (tags.isEmpty()) {
        return
    }
    FlowRow(
        modifier = modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        for (tag in tags) {
            TagPill(tagLabel(tag.name), onClick = { onClick(tag) })
        }
    }
}
