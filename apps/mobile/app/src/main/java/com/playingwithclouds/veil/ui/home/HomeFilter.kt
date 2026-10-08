package com.playingwithclouds.veil.ui.home

import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.FilterPreset
import com.playingwithclouds.veil.data.RuntimeWindow
import com.playingwithclouds.veil.data.SceneSummary

/** The runtime chips of the Home filter row. */
enum class RuntimeChip(val label: String, val window: RuntimeWindow) {
    SHORT("Under 10 min", RuntimeWindow(null, 10 * 60)),
    MEDIUM("10–30 min", RuntimeWindow(10 * 60, 30 * 60)),
    LONG("Over 30 min", RuntimeWindow(30 * 60, null)),
    ;

    companion object {

        /** The chip whose window is exactly [window], or null. */
        fun matching(window: RuntimeWindow): RuntimeChip? {
            return entries.firstOrNull { chip -> chip.window == window }
        }
    }
}

/**
 * What Home is narrowed to: sites, one tag and a runtime window. Nothing set means the ranked
 * feed; a tag switches to a plain listing of that tag.
 */
data class HomeFilter(
    val sources: Set<String> = emptySet(),
    val tagId: String? = null,
    val runtime: RuntimeWindow = RuntimeWindow.ANY,
) {

    /** Whether anything narrows the feed. */
    val isActive: Boolean
        get() = sources.isNotEmpty() || tagId != null || !runtime.isOpen

    /** The filter with [name] added to the picked sites, or removed when it already is. */
    fun toggleSource(name: String): HomeFilter {
        if (sources.contains(name)) {
            return copy(sources = sources - name)
        }
        return copy(sources = sources + name)
    }

    /** The filter on [id], or off it when it is the current tag. */
    fun toggleTag(id: String): HomeFilter {
        if (tagId == id) {
            return copy(tagId = null)
        }
        return copy(tagId = id)
    }

    /** The filter on [chip]'s runtime, or off it when that is the current one. */
    fun toggleRuntime(chip: RuntimeChip): HomeFilter {
        if (runtime == chip.window) {
            return copy(runtime = RuntimeWindow.ANY)
        }
        return copy(runtime = chip.window)
    }

    /** Whether [preset] describes exactly this filter. */
    fun matches(preset: FilterPreset): Boolean {
        return sources == preset.sources.toSet() && tagId == preset.tagId && runtime == preset.runtime
    }

    companion object {

        /** The filter a preset stands for. */
        fun of(preset: FilterPreset): HomeFilter {
            return HomeFilter(preset.sources.toSet(), preset.tagId, preset.runtime)
        }
    }
}

/** The tags that occur most among [scenes], most frequent first: the content-type chips. */
fun topTags(scenes: List<SceneSummary>, limit: Int): List<EntityRef> {
    val counts = scenes.flatMap { scene -> scene.tags }.groupingBy { tag -> tag.id }.eachCount()
    val byId = scenes.flatMap { scene -> scene.tags }.associateBy { tag -> tag.id }
    return counts.entries
        .sortedWith(compareByDescending<Map.Entry<String, Int>> { entry -> entry.value }.thenBy { entry -> byId.getValue(entry.key).name })
        .take(limit)
        .map { entry -> byId.getValue(entry.key) }
}

/** A name for a preset made from the current filter, e.g. "Anal · 10–30 min · eporner". */
fun suggestedPresetName(filter: HomeFilter, tagName: String?): String {
    val parts = mutableListOf<String>()
    if (tagName != null) {
        parts.add(tagName)
    }
    val chip = RuntimeChip.matching(filter.runtime)
    if (chip != null) {
        parts.add(chip.label)
    }
    parts.addAll(filter.sources.sorted())
    if (parts.isEmpty()) {
        return "My filter"
    }
    return parts.joinToString(" · ")
}
