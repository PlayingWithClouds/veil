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
    const val CAM = "cam/{id}"
    const val COLLECTION = "collection/{id}"
    const val SEARCH = "search?query={query}"
    const val PERFORMERS = "performers"
    const val PERFORMER = "performer/{id}"
    const val STUDIOS = "studios"
    const val STUDIO = "studio/{id}"
    const val GALLERIES = "galleries"
    const val GALLERY = "gallery/{id}"
    const val STORIES = "stories"
    const val STORY = "story/{id}"
    const val STORY_READER = "story/{id}/read/{position}"
    const val AUDIOS = "audio"
    const val AUDIO = "audio/{id}"

    /** The live rooms with one tag; the Cams tab lists them all. */
    const val CAMS_TAGGED = "cams/tag/{tag}"

    /** A screen not rebuilt for the new backend yet, by its title. */
    const val PENDING = "pending/{title}"

    fun pending(title: String): String = "pending/${Uri.encode(title)}"

    fun video(id: String): String = "video/${Uri.encode(id)}"

    fun cam(id: String): String = "cam/${Uri.encode(id)}"

    fun collection(id: String): String = "collection/${Uri.encode(id)}"

    /** Search, with a query to run right away when it isn't empty. */
    fun search(query: String): String = "search?query=${Uri.encode(query)}"

    fun camsTagged(tag: String): String = "cams/tag/${Uri.encode(tag)}"

    fun performer(id: String): String = "performer/${Uri.encode(id)}"

    fun studio(id: String): String = "studio/${Uri.encode(id)}"

    fun gallery(id: String): String = "gallery/${Uri.encode(id)}"

    fun story(id: String): String = "story/${Uri.encode(id)}"

    fun storyReader(id: String, position: Int): String = "story/${Uri.encode(id)}/read/$position"

    fun audio(id: String): String = "audio/${Uri.encode(id)}"
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
        when (destination) {
            Destination.PERFORMERS -> controller.navigate(Routes.PERFORMERS)
            Destination.STUDIOS -> controller.navigate(Routes.STUDIOS)
            Destination.GALLERIES -> controller.navigate(Routes.GALLERIES)
            Destination.STORIES -> controller.navigate(Routes.STORIES)
            Destination.AUDIO -> controller.navigate(Routes.AUDIOS)
            else -> openPending(destination.label)
        }
    }

    fun openSettings() = controller.navigate(Routes.SETTINGS)

    fun openDesignKit() = controller.navigate(Routes.DESIGN_KIT)

    fun openSearch(query: String = "") = controller.navigate(Routes.search(query))

    fun openVideo(id: String) = controller.navigate(Routes.video(id))

    fun openStudio(id: String) = controller.navigate(Routes.studio(id))

    fun openCollection(id: String) = controller.navigate(Routes.collection(id))

    fun openGallery(id: String) = controller.navigate(Routes.gallery(id))

    fun openTag(@Suppress("UNUSED_PARAMETER") id: String) = openPending("Tag")

    fun openCam(id: String) = controller.navigate(Routes.cam(id))

    /** The live rooms with [tag]. */
    fun openCamsTagged(tag: String) = controller.navigate(Routes.camsTagged(tag))

    fun openStory(id: String) = controller.navigate(Routes.story(id))

    /** The reader at the chapter at [position]. */
    fun openStoryReader(id: String, position: Int) = controller.navigate(Routes.storyReader(id, position))

    fun openAudio(id: String) = controller.navigate(Routes.audio(id))

    fun openPerformer(id: String) = controller.navigate(Routes.performer(id))

    /** The placeholder of a screen not rebuilt yet. */
    private fun openPending(title: String) = controller.navigate(Routes.pending(title))
}
