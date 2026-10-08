package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.BackendSettingsQuery
import com.playingwithclouds.veil.graphql.BlocklistQuery
import com.playingwithclouds.veil.graphql.RemoveBlockMutation
import com.playingwithclouds.veil.graphql.UpdateBackendSettingsMutation
import com.playingwithclouds.veil.graphql.type.UpdateSettingsInput

/** The backend's global job and download settings. */
data class BackendSettings(
    val maxConcurrentJobs: Int,
    val maxJobRetries: Int,
    val downloadSpeedLimitKBps: Int,
    val allowDownloadsWhileStreaming: Boolean,
    val autoEnrichAfterScrape: Boolean,
    val requireVpn: Boolean,
)

/** A blocked tag, performer or studio. */
data class BlockedEntity(val kind: String, val targetId: String, val label: String?)

/** Global backend settings and the blocklist. */
object SettingsRepository {

    /** The current settings. */
    suspend fun settings(): BackendSettings {
        val settings = VeilApi.client.query(BackendSettingsQuery()).execute().dataOrThrow().settings
        return BackendSettings(
            maxConcurrentJobs = settings.maxConcurrentJobs,
            maxJobRetries = settings.maxJobRetries,
            downloadSpeedLimitKBps = settings.downloadSpeedLimitKBps,
            allowDownloadsWhileStreaming = settings.allowDownloadsWhileStreaming,
            autoEnrichAfterScrape = settings.autoEnrichAfterScrape,
            requireVpn = settings.requireVpn,
        )
    }

    /** Saves the settings. */
    suspend fun save(settings: BackendSettings) {
        val input = UpdateSettingsInput(
            maxConcurrentJobs = Optional.present(settings.maxConcurrentJobs),
            maxJobRetries = Optional.present(settings.maxJobRetries),
            downloadSpeedLimitKBps = Optional.present(settings.downloadSpeedLimitKBps),
            allowDownloadsWhileStreaming = Optional.present(settings.allowDownloadsWhileStreaming),
            autoEnrichAfterScrape = Optional.present(settings.autoEnrichAfterScrape),
            requireVpn = Optional.present(settings.requireVpn),
            kindLimits = Optional.Absent,
        )
        VeilApi.client.mutation(UpdateBackendSettingsMutation(input)).execute().dataOrThrow()
    }

    /** Everything on the blocklist. */
    suspend fun blocklist(): List<BlockedEntity> {
        val data = VeilApi.client.query(BlocklistQuery()).execute().dataOrThrow()
        return data.blocklist.map { entry -> BlockedEntity(entry.kind, entry.targetId, entry.label) }
    }

    /** Unblocks an entity. */
    suspend fun unblock(targetId: String) {
        VeilApi.client.mutation(RemoveBlockMutation(targetId)).execute().dataOrThrow()
    }
}
