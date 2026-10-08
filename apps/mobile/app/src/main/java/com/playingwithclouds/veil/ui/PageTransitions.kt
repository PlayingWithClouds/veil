package com.playingwithclouds.veil.ui

import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.navigation.NavBackStackEntry

/**
 * Navigation animations: opening a page slides it in from the right over the parent drifting left,
 * and going back reverses that. Tabs need none; they live side by side in the tab pager.
 */
object PageTransitions {
    private const val SLIDE_MILLISECONDS = 320

    fun enter(from: NavBackStackEntry, to: NavBackStackEntry): EnterTransition {
        return slideInHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> width / 3 } +
            fadeIn(tween(SLIDE_MILLISECONDS / 2, easing = LinearOutSlowInEasing))
    }

    fun exit(from: NavBackStackEntry, to: NavBackStackEntry): ExitTransition {
        return slideOutHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> -width / 10 } +
            fadeOut(tween(SLIDE_MILLISECONDS))
    }

    fun popEnter(from: NavBackStackEntry, to: NavBackStackEntry): EnterTransition {
        return slideInHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> -width / 10 } +
            fadeIn(tween(SLIDE_MILLISECONDS))
    }

    fun popExit(from: NavBackStackEntry, to: NavBackStackEntry): ExitTransition {
        return slideOutHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> width / 3 } +
            fadeOut(tween(SLIDE_MILLISECONDS / 2))
    }
}
