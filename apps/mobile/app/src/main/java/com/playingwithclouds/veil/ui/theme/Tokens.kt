package com.playingwithclouds.veil.ui.theme

import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.unit.dp

/**
 * The spacing scale. Every padding and gap in the app is one of these steps, so screens line up
 * with each other; sizes of things (icons, buttons, cells) are not spacing and stay literal.
 */
object VeilSpacing {
    val hairline = 2.dp
    val extraSmall = 4.dp
    val small = 8.dp
    val medium = 12.dp
    val large = 16.dp
    val extraLarge = 24.dp
    val huge = 32.dp

    /** Side margin of every screen: content, headings and the first card of a shelf start here. */
    val gutter = large

    /** Gap between cards in grids and shelves. */
    val cardGap = medium
}

/** The corner radius scale, from badges up to sheets. */
object VeilShapes {
    val extraSmall = RoundedCornerShape(6.dp)
    val small = RoundedCornerShape(10.dp)
    val medium = RoundedCornerShape(16.dp)
    val large = RoundedCornerShape(20.dp)
    val extraLarge = RoundedCornerShape(28.dp)
    val capsule = CircleShape

    /** Labels drawn over an image (runtime, NEW). */
    val badge = extraSmall

    /** Media cards and their artwork, text fields, list thumbnails. */
    val card = medium

    /** Flat panels ([com.playingwithclouds.veil.ui.design.VeilCard]) and popup menus. */
    val panel = large

    /** Dialogs and the top of bottom sheets. */
    val sheet = extraLarge
}
