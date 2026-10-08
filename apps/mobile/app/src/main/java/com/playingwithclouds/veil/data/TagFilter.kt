package com.playingwithclouds.veil.data

/**
 * Tags a list must carry ([include]) and tags it must not ([exclude]). A scene carries a tag when
 * the scene, its studio or a credited performer is tagged with it, as on the backend.
 */
data class TagFilter(val include: List<EntityRef> = emptyList(), val exclude: List<EntityRef> = emptyList()) {

    /** Whether no tag is required or forbidden. */
    val isEmpty: Boolean
        get() = include.isEmpty() && exclude.isEmpty()

    /** Ids of the required tags. */
    fun includeIds(): List<String> {
        return include.map { tag -> tag.id }
    }

    /** Ids of the forbidden tags. */
    fun excludeIds(): List<String> {
        return exclude.map { tag -> tag.id }
    }

    /** The state of [tag] in this filter. */
    fun stateOf(tag: EntityRef): TagFilterState {
        if (include.any { entry -> entry.id == tag.id }) {
            return TagFilterState.INCLUDED
        }
        if (exclude.any { entry -> entry.id == tag.id }) {
            return TagFilterState.EXCLUDED
        }
        return TagFilterState.NEUTRAL
    }

    /** The filter with [tag] moved to [state], leaving the other tags as they are. */
    fun with(tag: EntityRef, state: TagFilterState): TagFilter {
        val remainingInclude = include.filter { entry -> entry.id != tag.id }
        val remainingExclude = exclude.filter { entry -> entry.id != tag.id }
        return when (state) {
            TagFilterState.INCLUDED -> TagFilter(remainingInclude + tag, remainingExclude)
            TagFilterState.EXCLUDED -> TagFilter(remainingInclude, remainingExclude + tag)
            TagFilterState.NEUTRAL -> TagFilter(remainingInclude, remainingExclude)
        }
    }

    /**
     * Whether a scene passes, judged by the tags on its card. Live results are checked this way
     * because the backend cannot filter what it has not stored yet.
     */
    fun accepts(scene: SceneSummary): Boolean {
        val sceneTagIds = scene.tags.map { tag -> tag.id }.toSet()
        if (excludeIds().any { tagId -> tagId in sceneTagIds }) {
            return false
        }
        return includeIds().all { tagId -> tagId in sceneTagIds }
    }
}

/** How a tag takes part in a [TagFilter]. */
enum class TagFilterState {
    NEUTRAL,
    INCLUDED,
    EXCLUDED,
}
