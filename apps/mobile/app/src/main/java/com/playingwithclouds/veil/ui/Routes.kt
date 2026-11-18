package com.playingwithclouds.veil.ui

import android.net.Uri
import androidx.navigation.NavHostController
import com.playingwithclouds.veil.data.TagSummary

/** Route patterns of every destination, plus builders that fill in the arguments. */
object Routes {
    /** The tab screens under the bottom bar: Home, Cams, Library. */
    const val TABS = "tabs"
    const val SETTINGS = "settings"
    const val DESIGN_KIT = "design-kit"
    /** A video's page; [at] (seconds, optional) is where playback starts instead of the resume point. */
    const val VIDEO = "video/{id}?at={at}"
    const val SHORTS = "shorts"

    /** The compilation of the liked scenes; [video] (optional) plays first. */
    const val COMPILATION = "compilation?video={video}"
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
    const val CATEGORIES = "categories"
    const val TASTE = "taste"
    const val PLUGINS = "plugins"

    /** A category's videos; the name is the title. */
    const val CATEGORY = "category/{id}?name={name}"

    /** A tag's videos; the name is the title. */
    const val TAG = "tag/{id}?name={name}"

    /** The live rooms with one tag; the Cams tab lists them all. */
    const val CAMS_TAGGED = "cams/tag/{tag}"

    /** An uploader's videos on one site: the plugin, their channel page and their name (the title). */
    const val UPLOADER = "uploader?plugin={plugin}&url={url}&name={name}"

    fun video(id: String): String = "video/${Uri.encode(id)}"

    /** A video's page starting at [seconds]. */
    fun videoAt(id: String, seconds: Double): String = "video/${Uri.encode(id)}?at=$seconds"

    fun compilation(videoId: String): String = "compilation?video=${Uri.encode(videoId)}"

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

    fun category(id: String, name: String): String = "category/${Uri.encode(id)}?name=${Uri.encode(name)}"

    fun tag(id: String, name: String): String = "tag/${Uri.encode(id)}?name=${Uri.encode(name)}"

    fun uploader(plugin: String, url: String, name: String): String {
        return "uploader?plugin=${Uri.encode(plugin)}&url=${Uri.encode(url)}&name=${Uri.encode(name)}"
    }
}

/** Navigation actions that screens call, so they never touch route strings. [selectTab] turns the tab pager to a page. */
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
        controller.navigate(routeOf(destination))
    }

    fun openSettings() = controller.navigate(Routes.SETTINGS)

    fun openDesignKit() = controller.navigate(Routes.DESIGN_KIT)

    fun openSearch(query: String = "") = controller.navigate(Routes.search(query))

    fun openVideo(id: String) = controller.navigate(Routes.video(id))

    /** A video's page, playing from [seconds] (a short's moment). */
    fun openVideoAt(id: String, seconds: Double) = controller.navigate(Routes.videoAt(id, seconds))

    /** The compilation of the liked scenes, the queue's matches first. */
    fun openCompilation() = controller.navigate(Routes.compilation(""))

    fun openStudio(id: String) = controller.navigate(Routes.studio(id))

    fun openCollection(id: String) = controller.navigate(Routes.collection(id))

    fun openGallery(id: String) = controller.navigate(Routes.gallery(id))

    /** A tag's videos. */
    fun openTag(tag: TagSummary) = controller.navigate(Routes.tag(tag.id, tag.name))

    /** A category's videos. */
    fun openCategory(id: String, name: String) = controller.navigate(Routes.category(id, name))

    fun openCam(id: String) = controller.navigate(Routes.cam(id))

    /** The live rooms with [tag]. */
    fun openCamsTagged(tag: String) = controller.navigate(Routes.camsTagged(tag))

    fun openStory(id: String) = controller.navigate(Routes.story(id))

    /** The reader at the chapter at [position]. */
    fun openStoryReader(id: String, position: Int) = controller.navigate(Routes.storyReader(id, position))

    fun openAudio(id: String) = controller.navigate(Routes.audio(id))

    fun openPerformer(id: String) = controller.navigate(Routes.performer(id))

    /** An uploader's videos, listed from their channel page [url] on [plugin]'s site. */
    fun openUploader(plugin: String, url: String, name: String) = controller.navigate(Routes.uploader(plugin, url, name))

    /** The route of a drawer entry without a tab. */
    private fun routeOf(destination: Destination): String {
        return when (destination) {
            Destination.SHORTS -> Routes.SHORTS
            Destination.GALLERIES -> Routes.GALLERIES
            Destination.AUDIO -> Routes.AUDIOS
            Destination.STORIES -> Routes.STORIES
            Destination.CATEGORIES -> Routes.CATEGORIES
            Destination.PERFORMERS -> Routes.PERFORMERS
            Destination.STUDIOS -> Routes.STUDIOS
            Destination.TASTE -> Routes.TASTE
            Destination.PLUGINS -> Routes.PLUGINS
            Destination.HOME, Destination.CAMS, Destination.LIBRARY -> Routes.TABS
        }
    }
}
