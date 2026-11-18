package com.playingwithclouds.veil.scenes

import org.json.JSONArray
import org.json.JSONObject

/** The manifest format this app reads; a newer one is refused instead of misread. */
const val SUPPORTED_MANIFEST_FORMAT = 1

/** One label of a head, as taxonomy.yaml names it; [title] is the display name when it differs. */
data class ModelLabel(val name: String, val title: String, val deprecated: Boolean)

/** A label of another head that must be on for a conditional head to count, as `nsr.model.probabilities` gates it. */
data class HeadCondition(val head: String, val label: String)

/**
 * One taxonomy head: [exclusive] heads are softmax (one label at a time), the others sigmoid
 * (any number). [context] heads need the surrounding seconds, so a single close-up frame says
 * nothing about them.
 */
data class ModelHead(
    val name: String,
    val exclusive: Boolean,
    val title: String,
    val context: Boolean,
    val condition: HeadCondition?,
    val labels: List<ModelLabel>,
)

/** One ONNX graph: its file, its input's name and its head outputs in output order. */
data class ModelGraph(val file: String, val input: String, val heads: List<String>)

/**
 * `model.json`, written by the classifier's export.py next to `frame.onnx` and `temporal.onnx`
 * (contract in `.claude/rules/mobile.md`). [version] names the export: videos are scanned again
 * when it changes. Frames are resized to [inputSize] square and normalised with [mean] and [std]
 * per RGB channel. [embedding] is the frame graph's first output; the temporal graph reads
 * [window] embeddings spaced [frameSeconds] apart, as it was trained.
 */
data class SceneModelManifest(
    val version: String,
    val taxonomyVersion: Int,
    val inputSize: Int,
    val mean: FloatArray,
    val std: FloatArray,
    val frame: ModelGraph,
    val embedding: String,
    val temporal: ModelGraph?,
    val window: Int,
    val frameSeconds: Double,
    val heads: List<ModelHead>,
    val closeupShots: Set<String>,
) {

    /** The head named [name]; null when the model has none. */
    fun head(name: String): ModelHead? {
        return heads.firstOrNull { head -> head.name == name }
    }

    companion object {

        /** Reads a manifest; throws [IllegalArgumentException] when it's not one this app can use. */
        fun parse(text: String): SceneModelManifest {
            val root = JSONObject(text)
            val format = root.optInt("format", 0)
            require(format == SUPPORTED_MANIFEST_FORMAT) { "Model format $format isn't supported (this app reads $SUPPORTED_MANIFEST_FORMAT)" }
            val input = root.getJSONObject("input")
            val frame = root.getJSONObject("frame")
            val temporal = root.optJSONObject("temporal")
            val manifest = SceneModelManifest(
                version = root.getString("version"),
                taxonomyVersion = root.optInt("taxonomyVersion", 0),
                inputSize = input.getInt("size"),
                mean = floats(input.getJSONArray("mean")),
                std = floats(input.getJSONArray("std")),
                frame = graphOf(frame),
                embedding = frame.optString("embedding", "embedding"),
                temporal = temporal?.let(::graphOf),
                window = temporal?.optInt("window", 0) ?: 0,
                frameSeconds = temporal?.optDouble("frameSeconds", DEFAULT_FRAME_SECONDS) ?: DEFAULT_FRAME_SECONDS,
                heads = headsOf(root.getJSONArray("heads")),
                closeupShots = strings(root.optJSONArray("closeupShots")).toSet(),
            )
            validate(manifest)
            return manifest
        }

        /** The frame spacing the temporal head was trained on when the manifest doesn't say: labels are cut at 2 fps. */
        private const val DEFAULT_FRAME_SECONDS = 0.5

        /** Fails when the graphs name heads the manifest doesn't describe, or the normalisation isn't RGB. */
        private fun validate(manifest: SceneModelManifest) {
            require(manifest.mean.size == 3 && manifest.std.size == 3) { "mean and std need three values (RGB)" }
            require(manifest.inputSize > 0) { "input size must be positive" }
            val known = manifest.heads.map { head -> head.name }.toSet()
            val named = manifest.frame.heads + manifest.temporal?.heads.orEmpty()
            val unknown = named.filter { name -> name !in known }
            require(unknown.isEmpty()) { "Graph outputs without a head: ${unknown.joinToString()}" }
            if (manifest.temporal != null) {
                require(manifest.window > 0) { "the temporal graph needs a window" }
            }
        }

        /** A graph entry: file, input name and head outputs. */
        private fun graphOf(json: JSONObject): ModelGraph {
            return ModelGraph(json.getString("file"), json.getString("input"), strings(json.getJSONArray("heads")))
        }

        /** The heads with their labels. */
        private fun headsOf(array: JSONArray): List<ModelHead> {
            return (0 until array.length()).map { index ->
                val json = array.getJSONObject(index)
                ModelHead(
                    name = json.getString("name"),
                    exclusive = json.getString("kind") == "exclusive",
                    title = json.optString("title", ""),
                    context = json.optBoolean("context", false),
                    condition = conditionOf(json.optJSONObject("condition")),
                    labels = labelsOf(json.getJSONArray("labels")),
                )
            }
        }

        /** A head's condition, or null when it has none. */
        private fun conditionOf(json: JSONObject?): HeadCondition? {
            if (json == null) {
                return null
            }
            return HeadCondition(json.getString("head"), json.getString("label"))
        }

        /** A head's labels in output order. */
        private fun labelsOf(array: JSONArray): List<ModelLabel> {
            return (0 until array.length()).map { index ->
                val json = array.getJSONObject(index)
                ModelLabel(json.getString("name"), json.optString("title", ""), json.optBoolean("deprecated", false))
            }
        }

        /** A JSON array of numbers as floats. */
        private fun floats(array: JSONArray): FloatArray {
            return FloatArray(array.length()) { index -> array.getDouble(index).toFloat() }
        }

        /** A JSON array of strings; empty for null. */
        private fun strings(array: JSONArray?): List<String> {
            if (array == null) {
                return emptyList()
            }
            return (0 until array.length()).map { index -> array.getString(index) }
        }
    }
}
