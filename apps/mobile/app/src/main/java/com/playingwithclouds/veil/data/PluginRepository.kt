package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.InstallPluginMutation
import com.playingwithclouds.veil.graphql.PluginPackagesQuery
import com.playingwithclouds.veil.graphql.PluginsQuery
import com.playingwithclouds.veil.graphql.TogglePluginMutation
import com.playingwithclouds.veil.graphql.UninstallPluginMutation
import com.playingwithclouds.veil.graphql.UpdatePluginSettingsMutation
import com.playingwithclouds.veil.graphql.UpdatePluginsMutation
import com.playingwithclouds.veil.graphql.type.PluginSettingValueInput

/** One setting a plugin declares. */
data class PluginSettingField(val key: String, val label: String, val description: String?, val type: String, val defaultValue: String?)

/** An installed plugin. */
data class PluginInfo(
    val name: String,
    val displayName: String?,
    val iconUrl: String?,
    val description: String?,
    val version: String,
    val capabilities: List<String>,
    val domains: List<String>,
    val enabled: Boolean,
    val available: Boolean,
    val requiresSolver: Boolean,
    val localBuild: Boolean,
    val settings: List<PluginSettingField>,
    val settingValues: Map<String, String>,
) {

    /** The name to show: the display name, else the identifier. */
    val label: String
        get() = displayName ?: name
}

/** A plugin package published in the GitHub plugin index. */
data class PluginPackage(val name: String, val version: String, val description: String?, val installedVersion: String?)

/** Installed plugins and the published plugin catalog. */
object PluginRepository {

    /** Every installed plugin. */
    suspend fun plugins(): List<PluginInfo> {
        val data = VeilApi.client.query(PluginsQuery()).execute().dataOrThrow()
        return data.plugins.map { plugin ->
            PluginInfo(
                name = plugin.name,
                displayName = plugin.displayName,
                iconUrl = plugin.iconUrl,
                description = plugin.description,
                version = plugin.version,
                capabilities = plugin.capabilities,
                domains = plugin.domains,
                enabled = plugin.enabled,
                available = plugin.available,
                requiresSolver = plugin.requiresSolver,
                localBuild = plugin.localBuild,
                settings = plugin.settings.map { field ->
                    PluginSettingField(field.key, field.label, field.description, field.type, field.default)
                },
                settingValues = plugin.settingValues.associate { value -> value.key to value.value },
            )
        }
    }

    /** Enables or disables a plugin. */
    suspend fun toggle(name: String, enabled: Boolean) {
        VeilApi.client.mutation(TogglePluginMutation(name, enabled)).execute().dataOrThrow()
    }

    /** Saves a plugin's setting values. */
    suspend fun saveSettings(pluginName: String, values: Map<String, String>) {
        val inputs = values.map { (key, value) -> PluginSettingValueInput(key, value) }
        VeilApi.client.mutation(UpdatePluginSettingsMutation(pluginName, inputs)).execute().dataOrThrow()
    }

    /** Published plugin packages matching the query; all when empty. */
    suspend fun packages(query: String): List<PluginPackage> {
        val data = VeilApi.client.query(PluginPackagesQuery(Optional.present(query))).execute().dataOrThrow()
        return data.pluginPackages.map { item -> PluginPackage(item.name, item.version, item.description, item.installedVersion) }
    }

    /** Installs or updates a published plugin package; the backend loads it within a second. */
    suspend fun install(packageName: String) {
        VeilApi.client.mutation(InstallPluginMutation(packageName)).execute().dataOrThrow()
    }

    /** Removes an installed plugin. */
    suspend fun uninstall(name: String) {
        VeilApi.client.mutation(UninstallPluginMutation(name)).execute().dataOrThrow()
    }

    /** Installs newer published versions of the installed plugins; returns the updated package names. */
    suspend fun updateAll(): List<String> {
        return VeilApi.client.mutation(UpdatePluginsMutation()).execute().dataOrThrow().updatePlugins
    }
}
