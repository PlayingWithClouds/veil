package com.playingwithclouds.veil.scenes

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Test

/** A small manifest: an exclusive head, a multi head, a conditional head and an image-skipped one. */
internal val testManifestJson = """
{
  "format": 1,
  "version": "test-1",
  "taxonomyVersion": 2,
  "input": {"size": 224, "mean": [0.485, 0.456, 0.406], "std": [0.229, 0.224, 0.225]},
  "frame": {"file": "frame.onnx", "input": "images", "embedding": "embedding", "heads": ["rating", "position", "oral_intensity"]},
  "temporal": {"file": "temporal.onnx", "input": "embeddings", "window": 16, "frameSeconds": 0.5, "heads": ["rating", "position", "oral_intensity"]},
  "heads": [
    {"name": "rating", "kind": "exclusive", "labels": [{"name": "sfw"}, {"name": "sexual_partner"}]},
    {"name": "position", "kind": "multi", "context": true, "labels": [{"name": "cowgirl"}, {"name": "kneeling"}, {"name": "new_pose", "title": "lotus"}]},
    {"name": "oral_intensity", "kind": "multi", "condition": {"head": "rating", "label": "sexual_partner"}, "labels": [{"name": "regular", "title": "blowjob"}]}
  ],
  "closeupShots": ["close_up"]
}
""".trimIndent()

class SceneModelTest {

    @Test
    fun manifestIsRead() {
        val manifest = SceneModelManifest.parse(testManifestJson)
        assertEquals("test-1", manifest.version)
        assertEquals(224, manifest.inputSize)
        assertEquals(16, manifest.window)
        assertEquals(listOf("rating", "position", "oral_intensity"), manifest.frame.heads)
        assertTrue(manifest.head("rating")!!.exclusive)
        assertEquals(HeadCondition("rating", "sexual_partner"), manifest.head("oral_intensity")!!.condition)
        assertEquals(setOf("close_up"), manifest.closeupShots)
    }

    @Test
    fun unknownFormatsAndHeadsAreRefused() {
        assertThrows(IllegalArgumentException::class.java) { SceneModelManifest.parse(testManifestJson.replace("\"format\": 1", "\"format\": 2")) }
        assertThrows(IllegalArgumentException::class.java) {
            SceneModelManifest.parse(testManifestJson.replace("\"heads\": [\"rating\", \"position\", \"oral_intensity\"]}", "\"heads\": [\"missing\"]}"))
        }
    }

    @Test
    fun labelsBecomeSiteTagNames() {
        assertEquals("Blowjob", SceneLabels.tagName("oral_intensity", ModelLabel("regular", "blowjob", false)))
        assertEquals("Doggystyle", SceneLabels.tagName("position", ModelLabel("doggy", "", false)))
        assertEquals("Lotus", SceneLabels.tagName("position", ModelLabel("new_pose", "lotus", false)))
        assertEquals("Hair Pulling", SceneLabels.tagName("fetish", ModelLabel("hair_pulling", "", false)))
        assertNull(SceneLabels.tagName("position", ModelLabel("kneeling", "", false)))
        assertNull(SceneLabels.tagName("rating", ModelLabel("sfw", "", false)))
    }

    @Test
    fun pickerGroupsFollowTheModel() {
        val groups = SceneLabels.pickerGroups(SceneModelManifest.parse(testManifestJson))
        assertEquals(listOf("Oral", "Positions"), groups.map { group -> group.title })
        assertEquals(listOf("Cowgirl", "Lotus"), groups[1].tags)
        assertTrue(SceneLabels.pickerGroups(null).first { group -> group.title == "Positions" }.tags.contains("Reverse Cowgirl"))
    }

    @Test
    fun conditionalHeadsAreGated() {
        val manifest = SceneModelManifest.parse(testManifestJson)
        val logits = mapOf("rating" to floatArrayOf(0f, 0f), "oral_intensity" to floatArrayOf(10f))
        val probabilities = SceneSegments.probabilities(logits, manifest)
        assertEquals(0.5f, probabilities.getValue("rating")[1], 1e-4f)
        assertEquals(0.5f, probabilities.getValue("oral_intensity")[0], 1e-3f)
    }
}
