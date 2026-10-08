package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.RecordSceneHeatMutation
import com.playingwithclouds.veil.graphql.SceneHeatQuery
import com.playingwithclouds.veil.graphql.SetSceneThumbnailMutation
import com.playingwithclouds.veil.graphql.type.HeatSpanInput
import com.playingwithclouds.veil.graphql.type.RecordSceneHeatInput

/** Where a scene gets replayed (0 to 1 per slice), the moment worth a thumbnail, and the one its thumbnail shows. */
data class SceneHeat(
    val buckets: List<Float>,
    val bestMomentSeconds: Double?,
    val thumbnailSeconds: Double?,
)

/** A stretch of continuous playback, in seconds. */
data class HeatSpan(val fromSeconds: Double, val toSeconds: Double)

/** What one viewing session did on a scene: the stretches it played and the positions it jumped to. */
data class HeatReport(val durationSeconds: Double, val spans: List<HeatSpan>, val scrubs: List<Double>)

/** The backend's replay data for scenes and the thumbnails taken from it. */
object SceneHeatRepository {

    /** The scene's replay graph and thumbnail moments. */
    suspend fun heat(sceneId: String): SceneHeat {
        val data = VeilApi.client.query(SceneHeatQuery(sceneId)).execute().dataOrThrow().sceneHeat
        return SceneHeat(data.buckets.map { bucket -> bucket.toFloat() }, data.bestMomentSeconds, data.thumbnailSeconds)
    }

    /** Adds a viewing session to the scene's replay data. */
    suspend fun record(sceneId: String, report: HeatReport) {
        val input = RecordSceneHeatInput(
            sceneId = sceneId,
            durationSeconds = report.durationSeconds,
            spans = report.spans.map { span -> HeatSpanInput(span.fromSeconds, span.toSeconds) },
            scrubs = report.scrubs,
        )
        VeilApi.client.mutation(RecordSceneHeatMutation(input)).execute().dataOrThrow()
    }

    /** Makes a JPEG frame (base64) taken at [atSeconds] the scene's thumbnail. */
    suspend fun setThumbnail(sceneId: String, atSeconds: Double, jpegBase64: String) {
        VeilApi.client.mutation(SetSceneThumbnailMutation(sceneId, atSeconds, jpegBase64)).execute().dataOrThrow()
    }
}
