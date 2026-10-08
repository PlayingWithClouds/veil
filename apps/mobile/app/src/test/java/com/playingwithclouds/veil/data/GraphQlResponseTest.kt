package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Operation
import com.apollographql.apollo.api.json.jsonReader
import com.apollographql.apollo.api.parseResponse
import com.playingwithclouds.veil.graphql.HomeFeedQuery
import com.playingwithclouds.veil.graphql.JobsQuery
import com.playingwithclouds.veil.graphql.SceneDetailQuery
import okio.Buffer
import org.json.JSONArray
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/** Parses backend-shaped JSON through the generated Apollo adapters, then through the app's mappers. */
class GraphQlResponseTest {

    /** The GraphQL type behind each field name used in the sample payloads; Apollo needs `__typename` to apply fragments. */
    private val typeNames = mapOf(
        "recommendations" to "RecommendedScene",
        "reason" to "RecommendationReason",
        "scene" to "Scene",
        "studio" to "Studio",
        "performers" to "Performer",
        "tags" to "Tag",
        "mediaStreams" to "Stream",
        "jobs" to "Job",
    )

    /** Adds the `__typename` that a real server returns for every object. */
    private fun withTypeNames(value: Any?, fieldName: String?): Any? {
        if (value is JSONArray) {
            for (index in 0 until value.length()) {
                withTypeNames(value.get(index), fieldName)
            }
            return value
        }
        if (value !is JSONObject) {
            return value
        }
        for (key in value.keys().asSequence().toList()) {
            withTypeNames(value.get(key), key)
        }
        val typeName = typeNames[fieldName]
        if (typeName != null) {
            value.put("__typename", typeName)
        }
        return value
    }

    private fun <D : Operation.Data> Operation<D>.parse(json: String): D {
        val typed = withTypeNames(JSONObject(json), null).toString()
        val response = parseResponse(Buffer().writeUtf8(typed).jsonReader())
        return requireNotNull(response.data) { "no data in response: ${response.errors} ${response.exception}" }
    }

    private val sceneJson = """
        {"id":"scene:1","externalId":"e1","sourceUrl":"https://x/1","title":"A title","date":"2024-05-06",
         "durationSeconds":754,"rating":9.2,"viewCount":4200,"posterPath":"https://cdn/p.jpg","previewVideo":null,"previewImages":["a","b"],
         "studio":{"id":"studio:1","name":"Studio","imagePath":null},
         "performers":[{"id":"performer:1","name":"Jane","imagePath":"/api/blob/j.jpg"}],
         "tags":[{"id":"tag:1","name":"Tag"}]}
    """.trimIndent()

    @Test
    fun homeFeedMapsToRecommendedScenes() {
        val json = """{"data":{"recommendations":[{"source":"related:site","reason":{"kind":"related","text":"Because you watched X"},"scene":$sceneJson}]}}"""
        val recommended = HomeFeedQuery().parse(json).toRecommendedScenes().single()
        assertEquals("related:site", recommended.source)
        assertEquals("Because you watched X", recommended.reason)
        val scene = recommended.scene
        assertEquals("scene:1", scene.id)
        assertEquals("A title", scene.title)
        assertEquals(754, scene.durationSeconds)
        assertEquals(9.2, scene.rating!!, 0.001)
        assertEquals(4200, scene.viewCount)
        assertEquals("Studio", scene.studio?.name)
        assertEquals(listOf("Jane"), scene.performers.map { performer -> performer.name })
        assertEquals(listOf("a", "b"), scene.previewImages)
        assertNull(scene.previewVideo)
    }

    @Test
    fun sceneDetailMapsSceneAndStreams() {
        val json = """
            {"data":{"scene":{"id":"scene:1","title":"T","sourceUrl":"https://x/1","details":"About","date":null,
              "durationSeconds":null,"rating":null,"viewCount":12,"posterPath":null,"previewVideo":null,"previewImages":[],
              "studio":{"id":"studio:1","name":"S","imagePath":null,"sceneCount":40},
              "performers":[],"tags":[{"id":"tag:1","name":"X"}]},
             "mediaStreams":[{"id":"stream:1","url":"https://x/v.m3u8","kind":"stream","label":null,"provider":null,
               "resolution":"1080p","width":1920,"height":1080,"language":null,"format":null,"mimeType":null,
               "expectedSpeedBps":1000000.5,"fileSizeBytes":null,"verified":true,"pluginName":"eporner"}]}}
        """.trimIndent()
        val (detail, streams) = SceneDetailQuery("scene:1").parse(json).toSceneAndStreams()
        assertEquals("T", detail?.title)
        assertEquals(12, detail?.viewCount)
        assertEquals(40, detail?.studio?.sceneCount)
        assertEquals("1080p", streams.single().qualityLabel())
        assertTrue(streams.single().verified)
        assertEquals(1000000.5, streams.single().expectedSpeedBps!!, 0.001)
    }

    @Test
    fun missingSceneMapsToNullDetail() {
        val json = """{"data":{"scene":null,"mediaStreams":[]}}"""
        val (detail, streams) = SceneDetailQuery("scene:9").parse(json).toSceneAndStreams()
        assertNull(detail)
        assertTrue(streams.isEmpty())
    }

    @Test
    fun jobsSplitIntoDownloadsAndBackgroundJobs() {
        val json = """
            {"data":{"jobs":[
              {"id":"job:1","kind":"download","status":"running","pluginName":"eporner","error":null,"updatedAt":"2026-01-01T00:00:00Z",
               "downloadTitle":"Fallback","downloadProgress":0.25,"downloadBytesReceived":1048576.0,"downloadBytesTotal":4194304.0,"scene":$sceneJson},
              {"id":"job:2","kind":"enrich","status":"failed","pluginName":"x","error":"boom","updatedAt":"2026-01-01T00:00:00Z",
               "downloadTitle":null,"downloadProgress":null,"downloadBytesReceived":null,"downloadBytesTotal":null,"scene":null}
            ]}}
        """.trimIndent()
        val queue = JobsQuery().parse(json).toJobQueue()
        val download = queue.downloads.single()
        assertEquals("A title", download.title)
        assertEquals(0.25, download.progress!!, 0.0001)
        assertTrue(download.isActive)
        assertEquals("boom", queue.background.single().error)
    }
}
