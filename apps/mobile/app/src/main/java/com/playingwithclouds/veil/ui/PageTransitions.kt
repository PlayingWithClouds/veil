package com.playingwithclouds.veil.ui

import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.navigation.NavBackStackEntry

/**
 * Navigation animations: switching tabs fades through (the old tab fades out, the new one fades
 * and scales up), opening a page slides it in from the right over the parent drifting left, and
 * going back reverses that.
 */
object PageTransitions {
    private const val SLIDE_MILLISECONDS = 320
    private const val FADE_OUT_MILLISECONDS = 90
    private const val FADE_IN_MILLISECONDS = 210

    /** Whether both ends of the navigation are tabs. */
    private fun betweenTabs(from: NavBackStackEntry, to: NavBackStackEntry): Boolean {
        return isTab(from) && isTab(to)
    }

    private fun isTab(entry: NavBackStackEntry): Boolean {
        return Tab.entries.any { tab -> tab.route == entry.destination.route }
    }

    fun enter(from: NavBackStackEntry, to: NavBackStackEntry): EnterTransition {
        if (betweenTabs(from, to)) {
            return fadeThroughIn()
        }
        return slideInHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> width / 3 } +
            fadeIn(tween(SLIDE_MILLISECONDS / 2, easing = LinearOutSlowInEasing))
    }

    fun exit(from: NavBackStackEntry, to: NavBackStackEntry): ExitTransition {
        if (betweenTabs(from, to)) {
            return fadeThroughOut()
        }
        return slideOutHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> -width / 10 } +
            fadeOut(tween(SLIDE_MILLISECONDS))
    }

    fun popEnter(from: NavBackStackEntry, to: NavBackStackEntry): EnterTransition {
        if (betweenTabs(from, to)) {
            return fadeThroughIn()
        }
        return slideInHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> -width / 10 } +
            fadeIn(tween(SLIDE_MILLISECONDS))
    }

    fun popExit(from: NavBackStackEntry, to: NavBackStackEntry): ExitTransition {
        if (betweenTabs(from, to)) {
            return fadeThroughOut()
        }
        return slideOutHorizontally(tween(SLIDE_MILLISECONDS, easing = FastOutSlowInEasing)) { width -> width / 3 } +
            fadeOut(tween(SLIDE_MILLISECONDS / 2))
    }

    /** The incoming tab waits for the old one to fade, then fades and grows into place. */
    private fun fadeThroughIn(): EnterTransition {
        return fadeIn(tween(FADE_IN_MILLISECONDS, delayMillis = FADE_OUT_MILLISECONDS)) +
            scaleIn(tween(FADE_IN_MILLISECONDS, delayMillis = FADE_OUT_MILLISECONDS), initialScale = 0.94f)
    }

    private fun fadeThroughOut(): ExitTransition {
        return fadeOut(tween(FADE_OUT_MILLISECONDS)) + scaleOut(tween(FADE_OUT_MILLISECONDS), targetScale = 0.98f)
    }
}
