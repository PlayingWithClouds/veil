package com.playingwithclouds.veil.backend

import android.content.Context
import android.content.pm.PackageManager
import android.content.res.AssetManager
import android.net.ConnectivityManager
import android.net.LinkProperties
import android.net.Network
import android.util.Log
import com.playingwithclouds.veil.BuildConfig
import java.io.File
import java.io.IOException
import java.io.InputStream
import kotlin.concurrent.thread

/**
 * Runs the Go backend on the phone. The binary ships as a native library
 * (libveil.so, built by `bun run build:backend`) because the app's native
 * library folder is the only place Android lets an app execute files from.
 * It listens on loopback only, at [BASE_URL].
 */
class EmbeddedBackend(context: Context) {

    private val context = context.applicationContext
    private val dataDir = File(this.context.filesDir, "veil")
    private val seedDir = File(this.context.filesDir, "plugin-seed")
    private val dnsFile = File(dataDir, "dns-servers")
    private var restarts = 0

    /** Prepares the plugin seed and DNS file, then starts the backend on a background thread. */
    fun start() {
        thread(name = "veil-backend") { prepareAndRun() }
    }

    /** Runs the backend, restarting it when it crashes. */
    private fun prepareAndRun() {
        dataDir.mkdirs()
        try {
            refreshPluginSeed()
        } catch (error: IOException) {
            // The backend still runs; it just starts without the bundled plugins.
            Log.e(TAG, "extracting bundled plugins failed", error)
        }
        watchDnsServers()
        while (restarts <= MAX_RESTARTS) {
            val exitCode = runOnce()
            restarts++
            Log.e(TAG, "backend exited with $exitCode, restart $restarts/$MAX_RESTARTS")
            Thread.sleep(BackendEnvironment.restartDelayMilliseconds(restarts))
        }
    }

    /** Starts the backend process and blocks until it exits, forwarding its log to logcat. */
    private fun runOnce(): Int {
        val binary = context.applicationInfo.nativeLibraryDir + "/libveil.so"
        val builder = ProcessBuilder(binary).redirectErrorStream(true)
        builder.environment().putAll(
            BackendEnvironment.variables(
                dataDir = dataDir.absolutePath,
                pluginSeedDir = seedDir.absolutePath,
                dnsFile = dnsFile.absolutePath,
                tempDir = context.cacheDir.absolutePath,
                port = BuildConfig.BACKEND_PORT,
            ),
        )
        return try {
            val process = builder.start()
            forwardLog(process.inputStream)
            process.waitFor()
        } catch (error: IOException) {
            Log.e(TAG, "starting backend failed", error)
            -1
        }
    }

    /** Copies each backend output line to logcat. */
    private fun forwardLog(output: InputStream) {
        output.bufferedReader().forEachLine { line -> Log.i(TAG, line) }
    }

    /** Extracts the plugin bundles shipped in the APK once per install or update of the app. */
    private fun refreshPluginSeed() {
        val preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        val updateTime = packageUpdateTime()
        val seededUpdateTime = preferences.getLong(SEEDED_UPDATE_TIME, 0)
        if (!PluginSeeder.needsRefresh(seedDir.isDirectory, seededUpdateTime, updateTime)) {
            return
        }
        seedDir.deleteRecursively()
        PluginSeeder(AndroidAssetTree(context.assets)).copyShippedPlugins("plugins", seedDir)
        preferences.edit().putLong(SEEDED_UPDATE_TIME, updateTime).apply()
    }

    /** When the app was last installed or updated. */
    private fun packageUpdateTime(): Long {
        return try {
            context.packageManager.getPackageInfo(context.packageName, 0).lastUpdateTime
        } catch (error: PackageManager.NameNotFoundException) {
            0
        }
    }

    /**
     * Keeps the DNS file listing the current network's DNS servers. Go can't ask
     * Android for them (pure-Go build, no resolv.conf); the backend re-reads the
     * file on every lookup, so Wi-Fi/mobile/VPN switches take effect at once.
     */
    private fun watchDnsServers() {
        val connectivity = context.getSystemService(ConnectivityManager::class.java)
        val activeNetwork = connectivity.activeNetwork
        if (activeNetwork != null) {
            writeDnsServers(connectivity.getLinkProperties(activeNetwork))
        }
        connectivity.registerDefaultNetworkCallback(object : ConnectivityManager.NetworkCallback() {
            override fun onLinkPropertiesChanged(network: Network, linkProperties: LinkProperties) {
                writeDnsServers(linkProperties)
            }
        })
    }

    /** Writes the DNS servers of a network to the DNS file, replacing it atomically. */
    private fun writeDnsServers(properties: LinkProperties?) {
        if (properties == null) {
            return
        }
        val servers = properties.dnsServers.mapNotNull { server -> server.hostAddress }
        val temporary = File(dataDir, ".dns-servers.tmp")
        try {
            temporary.writeText(BackendEnvironment.dnsServerFileContents(servers))
        } catch (error: IOException) {
            Log.w(TAG, "writing DNS servers failed", error)
            return
        }
        if (!temporary.renameTo(dnsFile)) {
            Log.w(TAG, "replacing DNS server file failed")
        }
    }

    companion object {
        private const val TAG = "VeilBackend"

        /** Gives up restarting after this many crashes in a row. */
        private const val MAX_RESTARTS = 5
        private const val PREFERENCES = "embedded-backend"
        private const val SEEDED_UPDATE_TIME = "seededUpdateTime"

        /** The address the embedded backend answers on. */
        val BASE_URL = BackendEnvironment.loopbackUrl(BuildConfig.BACKEND_PORT)
    }
}

/** The APK's assets as an [AssetTree]. */
private class AndroidAssetTree(private val assets: AssetManager) : AssetTree {

    /** Names directly below the path. */
    override fun list(path: String): List<String> {
        return assets.list(path)?.toList().orEmpty()
    }

    /** Opens the asset file at the path. */
    override fun open(path: String): InputStream {
        return assets.open(path)
    }
}
