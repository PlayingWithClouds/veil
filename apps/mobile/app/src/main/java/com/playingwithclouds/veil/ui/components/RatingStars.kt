package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Row
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Stars of the rating row; each is two points of the 0..10 rating. */
private const val STAR_COUNT = 5

/** Five stars for a 0..10 [rating], filled in the accent; tapping the star that's already the rating clears it. */
@Composable
fun RatingStars(rating: Double?, onRate: (Double?) -> Unit, modifier: Modifier = Modifier) {
    val stars = ((rating ?: 0.0) / 2).toInt()
    Row(modifier, verticalAlignment = Alignment.CenterVertically) {
        for (star in 1..STAR_COUNT) {
            var icon = VeilIcons.Star
            var tint = VeilColors.contentFaint
            if (star <= stars) {
                icon = VeilIcons.StarFilled
                tint = VeilColors.accent
            }
            IconTap(icon, contentDescription = "Rate $star of $STAR_COUNT", tint = tint, onClick = {
                if (star == stars) {
                    onRate(null)
                } else {
                    onRate(star * 2.0)
                }
            })
        }
    }
}
