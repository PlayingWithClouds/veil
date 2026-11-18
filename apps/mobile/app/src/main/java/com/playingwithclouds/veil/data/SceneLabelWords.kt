package com.playingwithclouds.veil.data

/** What typed search words name: the scene labels found, and the other words, kept as text. */
data class LabelWords(val labels: List<TagSummary>, val text: String)

/**
 * Finds the scene labels a search names ("pov blowjob" → POV and Blowjob, nothing left to search
 * the sites for). Words are compared without case, spaces or dashes ("doggy style" and
 * "doggy-style" name Doggystyle), a plural "s" may be dropped ("blowjobs"), up to three words name
 * one label ("reverse cowgirl"), and a few common short forms are known ("bj", "doggy").
 */
object SceneLabelWords {

    /** The most words one label's name spans. */
    private const val LONGEST_NAME = 3

    /** Short forms people type, by the compact name they stand for. */
    private val SHORT_FORMS = mapOf(
        "bj" to "blowjob",
        "bjs" to "blowjob",
        "doggy" to "doggystyle",
    )

    /** The labels [text] names, in the order typed, and the words left over. */
    fun match(text: String, labels: List<TagSummary>): LabelWords {
        val byKey = labels.associateBy { label -> compact(label.name) }
        val words = text.split(Regex("\\s+")).filter { word -> word.isNotBlank() }
        val found = mutableListOf<TagSummary>()
        val rest = mutableListOf<String>()
        var index = 0
        while (index < words.size) {
            val (label, used) = labelAt(words, index, byKey)
            if (label == null) {
                rest.add(words[index])
                index++
                continue
            }
            if (label !in found) {
                found.add(label)
            }
            index += used
        }
        return LabelWords(found, rest.joinToString(" "))
    }

    /** The label the words from [start] name, longest name first, with how many words it took. */
    private fun labelAt(words: List<String>, start: Int, byKey: Map<String, TagSummary>): Pair<TagSummary?, Int> {
        for (size in minOf(LONGEST_NAME, words.size - start) downTo 1) {
            val key = compact(words.subList(start, start + size).joinToString(""))
            val label = lookUp(key, byKey)
            if (label != null) {
                return label to size
            }
        }
        return null to 0
    }

    /** The label a compact key names: as typed, through a short form, or without a plural "s". */
    private fun lookUp(key: String, byKey: Map<String, TagSummary>): TagSummary? {
        if (key.isEmpty()) {
            return null
        }
        val direct = byKey[key] ?: byKey[SHORT_FORMS[key]]
        if (direct != null || !key.endsWith("s")) {
            return direct
        }
        return byKey[key.dropLast(1)]
    }

    /** Lower case letters and digits only. */
    private fun compact(text: String): String {
        return text.lowercase().filter { character -> character.isLetterOrDigit() }
    }
}
