package com.playingwithclouds.veil.ui

import android.net.Uri
import androidx.navigation.NavHostController

/** Route patterns of every destination, plus builders that fill in the arguments. */
object Routes {
    /** The tab screens under the bottom bar: Home, Cams, Library. */
    const val TABS = "tabs"
    const val SETTINGS = "settings"
    const val DESIGN_KIT = "design-kit"
    const val VIDEO = "video/{id}"

    /** A screen not rebuilt for the new backend yet, by its title. */
    const val PENDING = "pending/{title}"

    fun pending(title: String): String = "pending/${Uri.encode(title)}"

    fun video(id: String): String = "video/${Uri.encode(id)}"
}

/**
 * Navigation actions that screens call, so they never touch route strings. [selectTab] turns the
 * tab pager to a page. Destinations not rebuilt yet open the placeholder screen; porting one only
 * changes its function here.
 */
class AppNavigator(private val controller: NavHostController, private val selectTab: (Tab) -> Unit) {

    /** Goes one screen back. */
    fun back() {
        controller.popBackStack()
    }

    /** Closes every screen above the tabs and turns to [tab]. */
    fun openTab(tab: Tab) {
        controller.popBackStack(Routes.TABS, inclusive = false)
        selectTab(tab)
    }

    /** Opens a drawer entry: a tab when it has one, else its screen. */
    fun openDestination(destination: Destination) {
        val tab = destination.tab
        if (tab != null) {
            openTab(tab)
            return
        }
        openPending(destination.label)
    }

    fun openSettings() = controller.navigate(Routes.SETTINGS)

    fun openDesignKit() = controller.navigate(Routes.DESIGN_KIT)

    fun openSearch() = openPending("Search")

    fun openVideo(id: String) = controller.navigate(Routes.video(id))

    fun openStudio(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Studio")

    fun openTag(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Tag")

    fun openCam(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Cam room")

    fun openStory(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Story")

    fun openAudio(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Audio")

    fun openPerformer(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Performer")

    /** The placeholder of a screen not rebuilt yet. */
    private fun openPending(title: String) = controller.navigate(Routes.pending(title))
}
