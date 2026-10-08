package com.playingwithclouds.veil.backend

import java.io.File
import java.io.IOException
import java.io.InputStream

/** A read-only folder tree of files, i.e. the APK's assets. */
interface AssetTree {

    /** Names directly below the path; empty for files and missing paths. */
    fun list(path: String): List<String>

    /** Opens the file at the path. */
    fun open(path: String): InputStream
}

/**
 * Extracts the plugin bundles shipped in the APK (assets/plugins) so the backend can load them
 * from disk. The backend copies them into its own plugin folder only when they are newer than
 * what it has (see pluginstore.Seed).
 */
class PluginSeeder(private val assets: AssetTree) {

    /** Copies the shipped plugin folder to the target; does nothing when the APK bundles no plugins. */
    @Throws(IOException::class)
    fun copyShippedPlugins(path: String, target: File) {
        if (assets.list(path).isEmpty()) {
            return
        }
        copyTree(path, target)
    }

    /** Copies the asset folder tree at the path to the target folder. */
    @Throws(IOException::class)
    fun copyTree(path: String, target: File) {
        val children = assets.list(path)
        if (children.isEmpty()) {
            copyFile(path, target)
            return
        }
        target.mkdirs()
        for (child in children) {
            copyTree("$path/$child", File(target, child))
        }
    }

    /** Copies one asset file to the target. */
    @Throws(IOException::class)
    private fun copyFile(path: String, target: File) {
        assets.open(path).use { input ->
            target.outputStream().use { output -> input.copyTo(output) }
        }
    }

    companion object {

        /** Whether the seed has to be extracted again: the first start, or after an app install or update. */
        fun needsRefresh(seedDirExists: Boolean, seededUpdateTime: Long, packageUpdateTime: Long): Boolean {
            if (!seedDirExists) {
                return true
            }
            return seededUpdateTime != packageUpdateTime
        }
    }
}
