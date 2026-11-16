package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.Dp
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * A headed row of cards that scrolls sideways, for a list without side padding: the heading takes
 * the gutter and the first card lines up with it. Every card is [itemWidth] wide, so the next one
 * peeks in at the edge. Draws nothing while [items] is empty, so empty rows vanish. [card] draws
 * one item, given its 0-based position in the row.
 */
@Composable
fun <T> Shelf(
    title: String,
    items: List<T>,
    key: (T) -> Any,
    itemWidth: Dp,
    modifier: Modifier = Modifier,
    state: LazyListState = rememberLazyListState(),
    card: @Composable (item: T, position: Int) -> Unit,
) {
    if (items.isEmpty()) {
        return
    }
    Column(modifier.fillMaxWidth()) {
        SectionHeading(title, modifier = Modifier.gutterPadding())
        LazyRow(
            Modifier.fillMaxWidth(),
            state = state,
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        ) {
            itemsIndexed(items, key = { _, item -> key(item) }) { position, item ->
                Column(Modifier.width(itemWidth)) { card(item, position) }
            }
        }
    }
}
