package com.playingwithclouds.veil.data

import kotlin.math.floor
import kotlin.math.roundToInt

/**
 * A chapter as navigation knows it. In a branching story [parent] is the chapter whose choice
 * leads here and [choice] that choice's words; both are null in a linear story.
 */
data class Chapter(
    val id: String,
    val position: Int,
    val title: String?,
    val wordCount: Int,
    val parent: String?,
    val choice: String?,
)

/** Where reading stopped: a chapter by position, and how far into it (0..1). */
data class ReadingPoint(val position: Int, val fraction: Double)

/** A chapter in the contents, with how deep its branch goes. */
data class ContentsEntry(val chapter: Chapter, val depth: Int)

/**
 * Finding the way through a story's chapters, and the reading position as `UserState.progress`
 * keeps it: the chapter's position plus how far into it (3.4 is 40% into the chapter at position
 * 3). A port of the web's `storyReading.ts`.
 */
object StoryReading {

    /** Kept below 1, so a chapter read to the end doesn't count as the next one. */
    private const val MAX_FRACTION = 0.999

    /** Words read per minute, for the time left. */
    private const val WORDS_PER_MINUTE = 230

    /** Whether the story branches: a chapter is reached by a choice. */
    fun isBranching(chapters: List<Chapter>): Boolean {
        return chapters.any { chapter -> chapter.parent != null }
    }

    /** The chapters a chapter's choices lead to, in story order. */
    fun choicesOf(chapters: List<Chapter>, chapterId: String): List<Chapter> {
        return chapters.filter { chapter -> chapter.parent == chapterId }
    }

    /** What follows a chapter: in a branching story the chapters its choices lead to, else the next one. Empty at an ending. */
    fun followingChapters(chapters: List<Chapter>, current: Chapter): List<Chapter> {
        if (isBranching(chapters)) {
            return choicesOf(chapters, current.id)
        }
        val index = chapters.indexOfFirst { chapter -> chapter.id == current.id }
        val next = chapters.getOrNull(index + 1)
        if (index == -1 || next == null) {
            return emptyList()
        }
        return listOf(next)
    }

    /** What comes before a chapter: its parent in a branching story, else the chapter before. */
    fun previousChapter(chapters: List<Chapter>, current: Chapter): Chapter? {
        val parent = current.parent
        if (parent != null) {
            return chapters.firstOrNull { chapter -> chapter.id == parent }
        }
        if (isBranching(chapters)) {
            return null
        }
        val index = chapters.indexOfFirst { chapter -> chapter.id == current.id }
        if (index <= 0) {
            return null
        }
        return chapters[index - 1]
    }

    /**
     * The chapters in reading order with their depth: a linear story flat, a branching one depth
     * first from its roots (chapters without a known parent), each choice's branch under it.
     */
    fun contentsOf(chapters: List<Chapter>): List<ContentsEntry> {
        if (!isBranching(chapters)) {
            return chapters.map { chapter -> ContentsEntry(chapter, 0) }
        }
        val known = chapters.map { chapter -> chapter.id }.toSet()
        val roots = chapters.filter { chapter -> chapter.parent == null || chapter.parent !in known }
        val entries = mutableListOf<ContentsEntry>()
        val visited = mutableSetOf<String>()
        // An explicit stack: real branching stories run hundreds of choices deep.
        val stack = ArrayDeque(roots.reversed().map { chapter -> ContentsEntry(chapter, 0) })
        while (stack.isNotEmpty()) {
            val entry = stack.removeLast()
            if (!visited.add(entry.chapter.id)) {
                continue
            }
            entries.add(entry)
            for (child in choicesOf(chapters, entry.chapter.id).reversed()) {
                stack.addLast(ContentsEntry(child, entry.depth + 1))
            }
        }
        return entries
    }

    /** A reading point as progress. */
    fun progressOf(point: ReadingPoint): Double {
        return point.position + point.fraction.coerceIn(0.0, MAX_FRACTION)
    }

    /** Progress as a reading point; null when nothing was read. */
    fun readingPointOf(progress: Double?): ReadingPoint? {
        if (progress == null || progress < 0) {
            return null
        }
        val position = floor(progress).toInt()
        return ReadingPoint(position, progress - position)
    }

    /** The minutes it takes to read what's left of [words], at least 1. */
    fun minutesLeft(words: Int, fraction: Double): Int {
        val left = words * (1 - fraction.coerceIn(0.0, 1.0))
        return maxOf(1, (left / WORDS_PER_MINUTE).roundToInt())
    }

    /** A chapter's text as paragraphs: blank lines part them, single line breaks stay. */
    fun paragraphsOf(text: String): List<String> {
        return text.split(Regex("\\n\\s*\\n")).map { paragraph -> paragraph.trim() }.filter { paragraph -> paragraph.isNotEmpty() }
    }
}
