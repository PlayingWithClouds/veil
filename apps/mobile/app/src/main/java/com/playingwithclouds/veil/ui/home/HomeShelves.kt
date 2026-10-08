package com.playingwithclouds.veil.ui.home

import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.SceneRow
import com.playingwithclouds.veil.data.SceneSummary

/** A horizontally swiping row placed between runs of feed cards. */
sealed interface HomeShelf {

    /** Stable identity of the row, for list keys. */
    val key: String

    /** A titled row of scenes ("Because you watched X"). */
    data class Scenes(override val key: String, val title: String, val scenes: List<SceneSummary>) : HomeShelf

    /** The performers credited most in the feed. */
    data class Performers(val performers: List<EntityRef>) : HomeShelf {
        override val key: String
            get() = "performers"
    }

    /** The studios credited most in the feed. */
    data class Studios(val studios: List<EntityRef>) : HomeShelf {
        override val key: String
            get() = "studios"
    }
}

/** Fewest entities worth a performer or studio row. */
private const val MINIMUM_ENTITIES = 4

/** Most entities in a performer or studio row. */
private const val MAXIMUM_ENTITIES = 14

/**
 * The shelves for Home in display order: the strongest scene row, performers, the next scene row,
 * studios, then the remaining scene rows. Performers and studios come from who is credited most
 * in [feedScenes]; rows too thin to be worth a shelf are left out.
 */
fun buildShelves(rows: List<SceneRow>, feedScenes: List<SceneSummary>): List<HomeShelf> {
    val sceneShelves = rows.map { row -> HomeShelf.Scenes(row.key, row.title, row.scenes) }
    val performers = mostCredited(feedScenes.flatMap { scene -> scene.performers }.filter { performer -> performer.imagePath != null })
    val studios = mostCredited(feedScenes.mapNotNull { scene -> scene.studio })
    val shelves = mutableListOf<HomeShelf>()
    sceneShelves.getOrNull(0)?.let { shelf -> shelves.add(shelf) }
    if (performers.size >= MINIMUM_ENTITIES) {
        shelves.add(HomeShelf.Performers(performers))
    }
    sceneShelves.getOrNull(1)?.let { shelf -> shelves.add(shelf) }
    if (studios.size >= MINIMUM_ENTITIES) {
        shelves.add(HomeShelf.Studios(studios))
    }
    shelves.addAll(sceneShelves.drop(2))
    return shelves
}

/** The entities that occur most often, most frequent first, at most [MAXIMUM_ENTITIES]. */
private fun mostCredited(entities: List<EntityRef>): List<EntityRef> {
    val counts = entities.groupingBy { entity -> entity.id }.eachCount()
    return entities
        .distinctBy { entity -> entity.id }
        .sortedByDescending { entity -> counts.getValue(entity.id) }
        .take(MAXIMUM_ENTITIES)
}
