package com.playingwithclouds.veil.scenes

/**
 * What a stretch is about when the specific heads (acts, positions, …) recognise nothing: the
 * scene's phase (`scene_phase`, read by the temporal graph only) and, where the phase is the main
 * part or unknown, how explicit the picture is (`rating`). For the player's sections only: these
 * never become ai segments, so they don't feed taste, search, compilations or card previews.
 */
object BroadLabels {

    /** The titles a broad section can have, in the order of [distribution]'s columns. */
    val titles = listOf("Intro", "Foreplay", "Climax", "Outro", "Talking", "Teasing", "Posing", "Solo", "Sex")

    /** `scene_phase` labels named on their own; main and transition fall through to the rating. */
    private val phaseTitles = mapOf("intro" to "Intro", "foreplay" to "Foreplay", "climax" to "Climax", "outro" to "Outro")

    /** `rating` labels by title. */
    private val ratingTitles = mapOf(
        "sfw" to "Talking",
        "suggestive" to "Teasing",
        "nudity" to "Posing",
        "sexual_solo" to "Solo",
        "sexual_partner" to "Sex",
    )

    private const val PHASE_HEAD = "scene_phase"
    private const val RATING_HEAD = "rating"
    private const val SEX = "Sex"

    /**
     * One frame's probabilities over [titles]: the named phases as they are, the rest of the phase
     * (main, transition; everything without a phase reading) split by the rating, or all "Sex"
     * without one. Null when the frame has neither head.
     */
    fun distribution(probabilities: Map<String, FloatArray>, manifest: SceneModelManifest): FloatArray? {
        val phase = labelled(PHASE_HEAD, probabilities, manifest)
        val rating = labelled(RATING_HEAD, probabilities, manifest)
        if (phase == null && rating == null) {
            return null
        }
        val out = FloatArray(titles.size)
        var rest = 1f
        for ((label, value) in phase.orEmpty()) {
            val title = phaseTitles[label] ?: continue
            out[titles.indexOf(title)] += value
            rest -= value
        }
        if (rating == null) {
            out[titles.indexOf(SEX)] += rest
            return out
        }
        for ((label, value) in rating) {
            val title = ratingTitles[label] ?: continue
            out[titles.indexOf(title)] += value * rest
        }
        return out
    }

    /**
     * The broad sections of frames grabbed at [times], each standing for the [stepSeconds] after it,
     * by smoothed argmax under [rules]; frames without a reading get none.
     */
    fun segments(times: List<Double>, stepSeconds: Double, frames: List<Map<String, FloatArray>>, manifest: SceneModelManifest, rules: SegmentRules): List<TimedSegment> {
        val rows = frames.map { frame -> distribution(frame, manifest) ?: FloatArray(titles.size) }
        val runs = SceneSegments.exclusive(rows, titles, rules)
        val found = runs.map { run -> TimedSegment(run.label, times[run.start], times[run.end] + stepSeconds) }
        return SceneSegments.mergeTouching(found)
    }

    /** One head's probabilities by label name; null when the model or the frame lacks the head. */
    private fun labelled(headName: String, probabilities: Map<String, FloatArray>, manifest: SceneModelManifest): Map<String, Float>? {
        val head = manifest.head(headName) ?: return null
        val values = probabilities[headName] ?: return null
        return head.labels.withIndex().associate { (index, label) -> label.name to values.getOrElse(index) { 0f } }
    }
}
