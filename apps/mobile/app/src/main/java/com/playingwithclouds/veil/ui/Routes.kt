package com.playingwithclouds.veil.ui

import android.net.Uri
import androidx.navigation.NavHostController

/** Route patterns of every destination, plus builders that fill in the arguments. */
object Routes {
    /** The swipeable tab pager: Home, Following, Galleries, Library. */
    const val TABS = "tabs"
    const val COLLECTIONS = "collections"
    const val HISTORY = "history"
    const val LIBRARY_SECTION = "library/{section}"
    const val SEARCH = "search?query={query}"
    const val SCENE = "scene/{id}"
    const val PERFORMERS = "performers"
    const val PERFORMER = "performer/{id}"
    const val STUDIOS = "studios"
    const val STUDIO = "studio/{id}"
    const val TAGS = "tags"
    const val TAG = "tag/{id}"
    const val COLLECTION = "collection/{id}"
    const val GALLERY_CATEGORY = "galleries/category/{name}"
    const val GALLERY = "gallery/{id}"
    const val PLUGINS = "plugins"
    const val SETTINGS = "settings"
    const val RANDOM = "random"
    const val DESIGN_KIT = "design-kit"
    const val SUBSCRIPTION = "subscription/{id}"

    /** Search, optionally with a query to run right away. */
    fun search(query: String = ""): String {
        return "search?query=${Uri.encode(query)}"
    }

    fun librarySection(section: String): String = "library/${Uri.encode(section)}"

    fun scene(id: String): String = "scene/${Uri.encode(id)}"

    fun performer(id: String): String = "performer/${Uri.encode(id)}"

    fun studio(id: String): String = "studio/${Uri.encode(id)}"

    fun tag(id: String): String = "tag/${Uri.encode(id)}"

    fun collection(id: String): String = "collection/${Uri.encode(id)}"

    fun galleryCategory(name: String): String = "galleries/category/${Uri.encode(name)}"

    fun gallery(id: String): String = "gallery/${Uri.encode(id)}"

    fun subscription(id: String): String = "subscription/${Uri.encode(id)}"
}

/**
 * Navigation actions that screens call, so they never touch route strings. [selectTab] turns the
 * tab pager to a page.
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

    fun openSearch(query: String = "") = controller.navigate(Routes.search(query))

    fun openScene(id: String) = controller.navigate(Routes.scene(id))

    /** Opens a scene in place of the scene page on top, so playing through a queue does not stack pages. */
    fun replaceWithScene(id: String) {
        controller.navigate(Routes.scene(id)) {
            val current = controller.currentDestination?.id
            if (current != null) {
                popUpTo(current) { inclusive = true }
            }
        }
    }

    fun openCollections() = controller.navigate(Routes.COLLECTIONS)

    fun openHistory() = controller.navigate(Routes.HISTORY)

    /** Downloads, queue or watchlist, by [com.playingwithclouds.veil.ui.library.LibrarySection] name. */
    fun openLibrarySection(section: String) = controller.navigate(Routes.librarySection(section))

    fun openPerformers() = controller.navigate(Routes.PERFORMERS)

    fun openPerformer(id: String) = controller.navigate(Routes.performer(id))

    fun openStudios() = controller.navigate(Routes.STUDIOS)

    fun openStudio(id: String) = controller.navigate(Routes.studio(id))

    fun openTags() = controller.navigate(Routes.TAGS)

    fun openTag(id: String) = controller.navigate(Routes.tag(id))

    fun openCollection(id: String) = controller.navigate(Routes.collection(id))


    fun openGalleryCategory(name: String) = controller.navigate(Routes.galleryCategory(name))

    fun openGallery(id: String) = controller.navigate(Routes.gallery(id))

    fun openPlugins() = controller.navigate(Routes.PLUGINS)

    fun openSettings() = controller.navigate(Routes.SETTINGS)

    fun openRandom() = controller.navigate(Routes.RANDOM)

    fun openDesignKit() = controller.navigate(Routes.DESIGN_KIT)

    fun openSubscription(id: String) = controller.navigate(Routes.subscription(id))
}
