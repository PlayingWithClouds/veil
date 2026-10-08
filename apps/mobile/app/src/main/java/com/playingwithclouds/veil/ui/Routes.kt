package com.playingwithclouds.veil.ui

import android.net.Uri
import androidx.navigation.NavHostController

/** Route patterns of every destination, plus builders that fill in the arguments. */
object Routes {
    const val HOME = "home"
    const val SUBSCRIPTIONS = "subscriptions"
    const val LIBRARY = "library"
    const val COLLECTIONS = "collections"
    const val HISTORY = "history"
    const val SEARCH = "search?query={query}"
    const val SCENE = "scene/{id}"
    const val PERFORMERS = "performers"
    const val PERFORMER = "performer/{id}"
    const val STUDIOS = "studios"
    const val STUDIO = "studio/{id}"
    const val TAGS = "tags"
    const val TAG = "tag/{id}"
    const val COLLECTION = "collection/{id}"
    const val GALLERIES = "galleries"
    const val GALLERY_CATEGORY = "galleries/category/{name}"
    const val GALLERY = "gallery/{id}"
    const val PLUGINS = "plugins"
    const val SETTINGS = "settings"
    const val RANDOM = "random"
    const val SUBSCRIPTION = "subscription/{id}"

    /** Search, optionally with a query to run right away. */
    fun search(query: String = ""): String {
        return "search?query=${Uri.encode(query)}"
    }

    fun scene(id: String): String = "scene/${Uri.encode(id)}"

    fun performer(id: String): String = "performer/${Uri.encode(id)}"

    fun studio(id: String): String = "studio/${Uri.encode(id)}"

    fun tag(id: String): String = "tag/${Uri.encode(id)}"

    fun collection(id: String): String = "collection/${Uri.encode(id)}"

    fun galleryCategory(name: String): String = "galleries/category/${Uri.encode(name)}"

    fun gallery(id: String): String = "gallery/${Uri.encode(id)}"

    fun subscription(id: String): String = "subscription/${Uri.encode(id)}"
}

/** Navigation actions that screens call, so they never touch route strings. */
class AppNavigator(private val controller: NavHostController) {

    /** Goes one screen back. */
    fun back() {
        controller.popBackStack()
    }

    /** Switches to a tab, keeping each tab's own state. */
    fun openTab(route: String) {
        controller.navigate(route) {
            popUpTo(Routes.HOME) { saveState = true }
            launchSingleTop = true
            restoreState = true
        }
    }

    fun openSearch(query: String = "") = controller.navigate(Routes.search(query))

    fun openScene(id: String) = controller.navigate(Routes.scene(id))

    fun openPerformers() = controller.navigate(Routes.PERFORMERS)

    fun openPerformer(id: String) = controller.navigate(Routes.performer(id))

    fun openStudios() = controller.navigate(Routes.STUDIOS)

    fun openStudio(id: String) = controller.navigate(Routes.studio(id))

    fun openTags() = controller.navigate(Routes.TAGS)

    fun openTag(id: String) = controller.navigate(Routes.tag(id))

    fun openCollection(id: String) = controller.navigate(Routes.collection(id))

    fun openGalleries() = controller.navigate(Routes.GALLERIES)

    fun openGalleryCategory(name: String) = controller.navigate(Routes.galleryCategory(name))

    fun openGallery(id: String) = controller.navigate(Routes.gallery(id))

    fun openPlugins() = controller.navigate(Routes.PLUGINS)

    fun openSettings() = controller.navigate(Routes.SETTINGS)

    fun openRandom() = controller.navigate(Routes.RANDOM)

    fun openSubscription(id: String) = controller.navigate(Routes.subscription(id))
}
