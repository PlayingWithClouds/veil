package com.playingwithclouds.veil.scenes

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ExportManifestTest {

    /** The shape of export.py's manifest.json (runs/v1_phone), cut to three heads. */
    private val export = """
        {
         "frame_model": "frame.fp16.onnx",
         "temporal_model": "temporal.int8.onnx",
         "input": {"name": "images", "shape": ["n", 3, 224, 224], "mean": [0.485, 0.456, 0.406], "std": [0.229, 0.224, 0.225]},
         "temporal": {"input": "embeddings", "shape": ["streams", 16, 256]},
         "sample_fps": 2.0,
         "heads": {
          "rating": {"kind": "softmax", "labels": ["sfw", "nudity"], "in": ["frame", "temporal"]},
          "act": {"kind": "sigmoid", "labels": ["blowjob", "anal"], "in": ["frame", "temporal"]},
          "scene_phase": {"kind": "softmax", "labels": ["intro", "main"], "in": ["temporal"]}
         }
        }
    """.trimIndent()

    @Test
    fun readsTheExportersManifest() {
        val manifest = SceneModelManifest.parse(export)
        assertEquals("frame.fp16.onnx", manifest.frame.file)
        // Outputs are read by name, so only membership matters (JSON key order isn't kept everywhere).
        assertEquals(setOf("rating", "act"), manifest.frame.heads.toSet())
        assertEquals(setOf("rating", "act", "scene_phase"), manifest.temporal?.heads?.toSet())
        assertEquals(224, manifest.inputSize)
        assertEquals(16, manifest.window)
        assertEquals(0.5, manifest.frameSeconds, 0.0)
        assertTrue(manifest.head("rating")!!.exclusive)
        assertTrue(manifest.head("rating")!!.context)
        assertEquals(listOf("blowjob", "anal"), manifest.head("act")!!.labels.map { label -> label.name })
        assertTrue(manifest.version.startsWith("export-"))
    }
}
