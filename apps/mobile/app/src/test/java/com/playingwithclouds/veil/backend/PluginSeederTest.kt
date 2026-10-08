package com.playingwithclouds.veil.backend

import java.io.ByteArrayInputStream
import java.io.File
import java.io.InputStream
import kotlin.io.path.createTempDirectory
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PluginSeederTest {

    /** An in-memory asset folder: file path to content. */
    private class FakeAssets(private val files: Map<String, String>) : AssetTree {

        override fun list(path: String): List<String> {
            val prefix = "$path/"
            return files.keys
                .filter { name -> name.startsWith(prefix) }
                .map { name -> name.removePrefix(prefix).substringBefore("/") }
                .distinct()
        }

        override fun open(path: String): InputStream {
            return ByteArrayInputStream(files.getValue(path).toByteArray())
        }
    }

    @Test
    fun copyTreeReproducesNestedFolders() {
        val assets = FakeAssets(
            mapOf(
                "plugins/eporner/plugin.js" to "module.exports = 1",
                "plugins/eporner/package.json" to "{}",
                "plugins/xhamster/plugin.js" to "module.exports = 2",
            ),
        )
        val target = createTempDirectory("seed").toFile()
        PluginSeeder(assets).copyTree("plugins", target)
        assertEquals("module.exports = 1", File(target, "eporner/plugin.js").readText())
        assertEquals("{}", File(target, "eporner/package.json").readText())
        assertEquals("module.exports = 2", File(target, "xhamster/plugin.js").readText())
        target.deleteRecursively()
    }

    @Test
    fun anApkWithoutPluginsLeavesTheSeedFolderUntouched() {
        val target = createTempDirectory("seed").toFile()
        val missing = File(target, "plugin-seed")
        PluginSeeder(FakeAssets(emptyMap())).copyShippedPlugins("plugins", missing)
        assertFalse(missing.exists())
        target.deleteRecursively()
    }

    @Test
    fun seedIsRefreshedOnFirstStart() {
        assertTrue(PluginSeeder.needsRefresh(seedDirExists = false, seededUpdateTime = 0, packageUpdateTime = 5))
    }

    @Test
    fun seedIsRefreshedAfterAnAppUpdate() {
        assertTrue(PluginSeeder.needsRefresh(seedDirExists = true, seededUpdateTime = 5, packageUpdateTime = 9))
    }

    @Test
    fun seedIsKeptWhileTheAppIsUnchanged() {
        assertFalse(PluginSeeder.needsRefresh(seedDirExists = true, seededUpdateTime = 5, packageUpdateTime = 5))
    }
}
