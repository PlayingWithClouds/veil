package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.DeleteJobMutation
import com.playingwithclouds.veil.graphql.DeleteJobsByKindMutation
import com.playingwithclouds.veil.graphql.JobsQuery
import com.playingwithclouds.veil.graphql.RetryJobMutation
import com.playingwithclouds.veil.graphql.VpnStatusQuery

/** A download job as the library shows it. */
data class DownloadJob(
    val id: String,
    val status: String,
    val title: String,
    val scene: SceneSummary?,
    val progress: Double?,
    val bytesReceived: Double?,
    val bytesTotal: Double?,
    val error: String?,
) {

    /** Whether the download is still queued or running. */
    val isActive: Boolean
        get() = status !in FINISHED_STATUSES

    /** Whether the download ended in failure (or was cancelled) and can be retried. */
    val isFailed: Boolean
        get() = status != "completed" && status in FINISHED_STATUSES

    companion object {
        /** Download states that are over, one way or the other. */
        val FINISHED_STATUSES = listOf("completed", "failed", "cancelled", "canceled")
    }
}

/** A non-download job (scrape, enrich, speed check), listed for maintenance. */
data class BackgroundJob(val id: String, val kind: String, val status: String, val pluginName: String, val error: String?)

/** The job queue split into downloads (queue order) and background jobs. */
data class JobQueue(val downloads: List<DownloadJob>, val background: List<BackgroundJob>)

/** Downloads and background jobs of the backend. */
object JobRepository {

    /** Loads every job. */
    suspend fun jobs(): JobQueue {
        return VeilApi.client.query(JobsQuery()).execute().dataOrThrow().toJobQueue()
    }


    /** The scene's title, else the job's own, else a placeholder. */
    fun downloadTitle(jobTitle: String?, scene: SceneSummary?): String {
        if (scene != null) {
            return scene.title
        }
        if (!jobTitle.isNullOrEmpty()) {
            return jobTitle
        }
        return "Download"
    }

    /** Queues a failed download again. */
    suspend fun retry(id: String) {
        VeilApi.client.mutation(RetryJobMutation(id)).execute().dataOrThrow()
    }

    /** Removes a job; a download's file stays where it is. */
    suspend fun delete(id: String) {
        VeilApi.client.mutation(DeleteJobMutation(id)).execute().dataOrThrow()
    }

    /** Removes every job of one kind. */
    suspend fun deleteKind(kind: String) {
        VeilApi.client.mutation(DeleteJobsByKindMutation(kind)).execute().dataOrThrow()
    }

    /** Whether the backend's VPN is connected, which downloads can require. */
    suspend fun vpnConnected(): Boolean {
        return VeilApi.client.query(VpnStatusQuery()).execute().dataOrThrow().vpnStatus.connected
    }
}

/** Splits the jobs response into downloads (queue order) and background jobs. */
fun JobsQuery.Data.toJobQueue(): JobQueue {
    val downloads = mutableListOf<DownloadJob>()
    val background = mutableListOf<BackgroundJob>()
    for (job in jobs) {
        if (job.kind != "download") {
            background.add(BackgroundJob(job.id, job.kind, job.status, job.pluginName, job.error))
            continue
        }
        val scene = job.scene?.sceneCardFields?.toSummary()
        downloads.add(
            DownloadJob(
                id = job.id,
                status = job.status,
                title = JobRepository.downloadTitle(job.downloadTitle, scene),
                scene = scene,
                progress = job.downloadProgress,
                bytesReceived = job.downloadBytesReceived,
                bytesTotal = job.downloadBytesTotal,
                error = job.error,
            ),
        )
    }
    return JobQueue(downloads, background)
}
