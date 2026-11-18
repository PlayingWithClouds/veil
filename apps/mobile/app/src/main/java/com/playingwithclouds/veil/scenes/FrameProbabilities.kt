package com.playingwithclouds.veil.scenes

/**
 * Turns the two graphs' per-frame outputs into one probability series per head, as the scanner
 * and the cover labeller need them.
 */
object FrameProbabilities {

    /**
     * The temporal graph's input for every frame: the last [window] embeddings up to and including
     * it, left-padded by repeating the first, as `nsr.model.StreamClassifier` batches them.
     */
    fun windows(embeddings: List<FloatArray>, window: Int): List<List<FloatArray>> {
        return embeddings.indices.map { frame ->
            val first = maxOf(0, frame - window + 1)
            val recent = embeddings.subList(first, frame + 1)
            List(window - recent.size) { recent.first() } + recent
        }
    }

    /**
     * One frame's probabilities per head from the image heads and, when there are any, the video
     * heads: their mean where both read a head, else whichever does. A context head's image reading
     * is left out on a close-up frame (the shot head names one of the manifest's close-up shots),
     * since one close-up can't show it.
     */
    fun combine(image: Map<String, FloatArray>, video: Map<String, FloatArray>?, manifest: SceneModelManifest): Map<String, FloatArray> {
        val closeUp = isCloseUp(image, manifest)
        val out = mutableMapOf<String, FloatArray>()
        for ((name, values) in image) {
            if (closeUp && manifest.head(name)?.context == true) {
                continue
            }
            out[name] = values
        }
        for ((name, values) in video.orEmpty()) {
            val fromImage = out[name]
            if (fromImage == null) {
                out[name] = values
                continue
            }
            out[name] = FloatArray(values.size) { index -> (values[index] + fromImage[index]) / 2 }
        }
        return out
    }

    /**
     * The video heads worth keeping when frames are [stepSeconds] apart. The temporal graph was
     * trained on frames [SceneModelManifest.frameSeconds] apart (16 × 0.5 s = 8 s); spread wider
     * than [MAX_WINDOW_STRETCH] times that, its window spans minutes it never saw, so heads the
     * picture reads too go by the picture alone. Heads only the temporal graph has (the scene
     * phase, a slow thing that a wide window still suits) stay.
     */
    fun videoAtSpacing(video: Map<String, FloatArray>?, stepSeconds: Double, manifest: SceneModelManifest): Map<String, FloatArray>? {
        if (video == null || stepSeconds <= manifest.frameSeconds * MAX_WINDOW_STRETCH) {
            return video
        }
        return video.filterKeys { name -> name !in manifest.frame.heads }
    }

    /** How many times wider than trained the temporal graph's frames may be before its overlapping heads are dropped. */
    const val MAX_WINDOW_STRETCH = 4.0

    /** Whether the image heads read a close-up shot. */
    private fun isCloseUp(image: Map<String, FloatArray>, manifest: SceneModelManifest): Boolean {
        val shot = manifest.head("shot") ?: return false
        val values = image["shot"] ?: return false
        val best = values.indices.maxByOrNull { index -> values[index] } ?: return false
        return shot.labels[best].name in manifest.closeupShots
    }

    /** The image's tag names it is fairly sure of (probability ≥ [minimum]), for cover tags. */
    fun confidentTags(probabilities: Map<String, FloatArray>, manifest: SceneModelManifest, minimum: Float): Map<String, Double> {
        val out = mutableMapOf<String, Double>()
        for (headName in SceneLabels.segmentHeads) {
            val head = manifest.head(headName) ?: continue
            val values = probabilities[headName] ?: continue
            for ((index, label) in head.labels.withIndex()) {
                val tag = SceneLabels.tagName(headName, label) ?: continue
                val probability = values.getOrNull(index) ?: continue
                if (probability >= minimum && probability > (out[tag] ?: 0.0)) {
                    out[tag] = Math.round(probability * 100) / 100.0
                }
            }
        }
        return out
    }
}
