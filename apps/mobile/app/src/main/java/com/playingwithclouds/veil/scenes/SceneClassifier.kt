package com.playingwithclouds.veil.scenes

import ai.onnxruntime.OnnxTensor
import ai.onnxruntime.OrtEnvironment
import ai.onnxruntime.OrtSession
import android.graphics.Bitmap
import java.io.File
import java.nio.FloatBuffer
import kotlin.math.min

/** What the frame graph says about a batch: one embedding per frame and each image head's logits per frame. */
class FrameOutputs(val embeddings: List<FloatArray>, val logits: Map<String, List<FloatArray>>)

/**
 * Runs the scene model with ONNX Runtime: [frames] encodes pictures and reads the image heads,
 * [temporal] reads the video heads from windows of embeddings. The sessions open on first use
 * from the installed model and close when it's replaced or removed. Safe to call from any thread;
 * ONNX Runtime runs one call at a time per session here, which keeps the phone cool.
 */
object SceneClassifier {

    private var opened: OpenModel? = null

    /** The sessions of one installed model. */
    private class OpenModel(val directory: File, val manifest: SceneModelManifest, val frame: OrtSession, val temporal: OrtSession?)

    /** Whether a model is installed, so callers can skip work before they start it. */
    fun available(): Boolean {
        return SceneModelStore.state.value.ready != null
    }

    /** The installed model's manifest; null without one. */
    fun manifest(): SceneModelManifest? {
        return SceneModelStore.state.value.ready?.manifest
    }

    /** Encodes [bitmaps] (any size; each is centre-cropped and scaled) in one batch; null without a model. */
    @Synchronized
    fun frames(bitmaps: List<Bitmap>): FrameOutputs? {
        val model = open() ?: return null
        if (bitmaps.isEmpty()) {
            return FrameOutputs(emptyList(), emptyMap())
        }
        val size = model.manifest.inputSize
        val input = FloatBuffer.allocate(bitmaps.size * 3 * size * size)
        for (bitmap in bitmaps) {
            input.put(pixels(bitmap, model.manifest))
        }
        input.rewind()
        val environment = OrtEnvironment.getEnvironment()
        OnnxTensor.createTensor(environment, input, longArrayOf(bitmaps.size.toLong(), 3, size.toLong(), size.toLong())).use { tensor ->
            model.frame.run(mapOf(model.manifest.frame.input to tensor)).use { result ->
                val embeddings = rows(result, model.manifest.embedding)
                val logits = model.manifest.frame.heads.associateWith { head -> rows(result, head) }
                return FrameOutputs(embeddings, logits)
            }
        }
    }

    /**
     * The video heads' logits for the last frame of each window ([windows] of equal length, at most
     * the manifest's window); null without a model or without a temporal graph.
     */
    @Synchronized
    fun temporal(windows: List<List<FloatArray>>): Map<String, List<FloatArray>>? {
        val model = open() ?: return null
        val session = model.temporal ?: return null
        if (windows.isEmpty()) {
            return emptyMap()
        }
        val length = windows.first().size
        val dimension = windows.first().first().size
        val input = FloatBuffer.allocate(windows.size * length * dimension)
        for (window in windows) {
            for (embedding in window) {
                input.put(embedding)
            }
        }
        input.rewind()
        val environment = OrtEnvironment.getEnvironment()
        val shape = longArrayOf(windows.size.toLong(), length.toLong(), dimension.toLong())
        OnnxTensor.createTensor(environment, input, shape).use { tensor ->
            session.run(mapOf(model.manifest.temporal!!.input to tensor)).use { result ->
                return model.manifest.temporal.heads.associateWith { head -> rows(result, head) }
            }
        }
    }

    /** Closes the sessions, e.g. before the model files are replaced. */
    @Synchronized
    fun close() {
        val model = opened ?: return
        model.frame.close()
        model.temporal?.close()
        opened = null
    }

    /** The open sessions of the installed model, opening them when needed; null without a model. */
    private fun open(): OpenModel? {
        val ready = SceneModelStore.state.value.ready ?: return null
        val current = opened
        if (current != null && current.directory == ready.directory && current.manifest.version == ready.manifest.version) {
            return current
        }
        close()
        val environment = OrtEnvironment.getEnvironment()
        val options = OrtSession.SessionOptions()
        val frame = environment.createSession(File(ready.directory, ready.manifest.frame.file).path, options)
        val temporalFile = ready.manifest.temporal?.file
        var temporal: OrtSession? = null
        if (temporalFile != null) {
            temporal = environment.createSession(File(ready.directory, temporalFile).path, options)
        }
        val model = OpenModel(ready.directory, ready.manifest, frame, temporal)
        opened = model
        return model
    }

    /** One output as a row per batch item. */
    private fun rows(result: OrtSession.Result, name: String): List<FloatArray> {
        val tensor = result.get(name).orElseThrow { IllegalStateException("The model has no output $name") } as OnnxTensor
        val shape = tensor.info.shape
        val width = shape.last().toInt()
        val values = tensor.floatBuffer
        return List(shape.first().toInt()) { row ->
            FloatArray(width) { column -> values.get(row * width + column) }
        }
    }

    /** A picture as the model's input: centre square, scaled to the input size, CHW, normalised per channel. */
    private fun pixels(bitmap: Bitmap, manifest: SceneModelManifest): FloatArray {
        val size = manifest.inputSize
        val side = min(bitmap.width, bitmap.height)
        val square = Bitmap.createBitmap(bitmap, (bitmap.width - side) / 2, (bitmap.height - side) / 2, side, side)
        val scaled = Bitmap.createScaledBitmap(square, size, size, true)
        val colours = IntArray(size * size)
        scaled.getPixels(colours, 0, size, 0, 0, size, size)
        val plane = size * size
        val out = FloatArray(3 * plane)
        for ((index, colour) in colours.withIndex()) {
            out[index] = (((colour shr 16) and 0xFF) / 255f - manifest.mean[0]) / manifest.std[0]
            out[plane + index] = (((colour shr 8) and 0xFF) / 255f - manifest.mean[1]) / manifest.std[1]
            out[2 * plane + index] = ((colour and 0xFF) / 255f - manifest.mean[2]) / manifest.std[2]
        }
        return out
    }
}
