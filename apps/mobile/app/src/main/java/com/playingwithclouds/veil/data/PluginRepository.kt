package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.InstallPluginMutation
import com.playingwithclouds.veil.graphql.InstalledPluginsQuery
import com.playingwithclouds.veil.graphql.PluginPackagesQuery
import com.playingwithclouds.veil.graphql.SetPluginEnabledMutation
import com.playingwithclouds.veil.graphql.UpdatePluginsMutation

/** An installed plugin; [available] is false when a requirement (e.g. FlareSolverr) is missing here. */
data class InstalledPlugin(
    val name: String,
    val displayName: String,
    val iconUrl: String?,
    val description: String?,
    val version: String,
    val enabled: Boolean,
    val available: Boolean,
)

/** A published plugin package; [installedVersion] is null when it isn't installed. */
data class PluginPackage(val name: String, val version: String, val description: String?, val installedVersion: String?)

/** The installed plugins and the published packages (`apps/web/src/lib/plugins.ts`). */
object PluginRepository {

    /** Every installed plugin, by display name. */
    suspend fun installed(): List<InstalledPlugin> {
        val data = VeilApi.client.query(InstalledPluginsQuery()).execute().dataOrThrow()
        return data.plugins.map { plugin ->
            InstalledPlugin(
                name = plugin.name,
                displayName = plugin.displayName ?: plugin.name,
                iconUrl = plugin.iconUrl,
                description = plugin.description,
                version = plugin.version,
                enabled = plugin.enabled,
                available = plugin.available,
            )
        }.sortedBy { plugin -> plugin.displayName.lowercase() }
    }

    /** The published packages. */
    suspend fun packages(): List<PluginPackage> {
        val data = VeilApi.client.query(PluginPackagesQuery()).execute().dataOrThrow()
        return data.pluginPackages.map { found -> PluginPackage(found.name, found.version, found.description, found.installedVersion) }
    }

    /** Turns a plugin on or off. */
    suspend fun setEnabled(name: String, enabled: Boolean) {
        VeilApi.client.mutation(SetPluginEnabledMutation(name, enabled)).execute().dataOrThrow()
    }

    /** Installs (or updates) a published package. */
    suspend fun install(packageName: String) {
        VeilApi.client.mutation(InstallPluginMutation(packageName)).execute().dataOrThrow()
    }

    /** Installs newer versions of the installed plugins; returns the updated package names. */
    suspend fun updateAll(): List<String> {
        return VeilApi.client.mutation(UpdatePluginsMutation()).execute().dataOrThrow().updatePlugins
    }
}
