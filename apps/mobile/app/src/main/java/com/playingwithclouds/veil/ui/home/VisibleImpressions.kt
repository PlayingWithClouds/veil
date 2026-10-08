package com.playingwithclouds.veil.ui.home

/** Where a laid-out list item sits along the scroll axis. */
data class ItemExtent(val index: Int, val offset: Int, val size: Int)

/** Which items count as shown to the user. */
object VisibleImpressions {

    /** The share of an item that has to be on screen for it to count as shown. */
    const val MINIMUM_VISIBLE_FRACTION = 0.5f

    /** Indices of the items of which at least the given fraction lies inside the viewport. */
    fun shownIndices(
        items: List<ItemExtent>,
        viewportStart: Int,
        viewportEnd: Int,
        minimumFraction: Float = MINIMUM_VISIBLE_FRACTION,
    ): List<Int> {
        return items
            .filter { item -> visibleFraction(item, viewportStart, viewportEnd) >= minimumFraction }
            .map { item -> item.index }
    }

    /** The share of the item inside the viewport, 0 when it is outside or empty. */
    private fun visibleFraction(item: ItemExtent, viewportStart: Int, viewportEnd: Int): Float {
        if (item.size <= 0) {
            return 0f
        }
        val visibleStart = maxOf(item.offset, viewportStart)
        val visibleEnd = minOf(item.offset + item.size, viewportEnd)
        val visibleSize = maxOf(0, visibleEnd - visibleStart)
        return visibleSize.toFloat() / item.size
    }
}
