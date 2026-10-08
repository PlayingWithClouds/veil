package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.tagLabel

/**
 * Tags as a wrapping row of tag pills, inset by the gutter. With a [limit], only that many show
 * at first, followed by a "… more" pill that reveals the rest.
 */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun TagChips(tags: List<EntityRef>, onClick: (EntityRef) -> Unit, modifier: Modifier = Modifier, limit: Int? = null) {
    if (tags.isEmpty()) {
        return
    }
    var expanded by rememberSaveable { mutableStateOf(false) }
    val collapsed = limit != null && !expanded && tags.size > limit
    var shown = tags
    if (collapsed) {
        shown = tags.take(limit)
    }
    FlowRow(
        modifier = modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        for (tag in shown) {
            TagPill(tagLabel(tag.name), onClick = { onClick(tag) })
        }
        if (collapsed) {
            TagPill("… ${tags.size - shown.size} more", onClick = { expanded = true })
        }
    }
}
