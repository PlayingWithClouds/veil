package com.playingwithclouds.veil.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class StoryReadingTest {

    private fun chapter(id: String, position: Int, parent: String? = null): Chapter {
        return Chapter(id, position, null, 1000, parent, null)
    }

    private val linear = listOf(chapter("a", 0), chapter("b", 1), chapter("c", 2))

    // a → (b, c), b → d
    private val branching = listOf(chapter("a", 0), chapter("b", 1, "a"), chapter("c", 2, "a"), chapter("d", 3, "b"))

    @Test
    fun aLinearStoryGoesOnToTheNextChapter() {
        assertEquals(listOf("b"), StoryReading.followingChapters(linear, linear[0]).map { chapter -> chapter.id })
        assertEquals(emptyList<Chapter>(), StoryReading.followingChapters(linear, linear[2]))
        assertEquals("a", StoryReading.previousChapter(linear, linear[1])?.id)
        assertNull(StoryReading.previousChapter(linear, linear[0]))
    }

    @Test
    fun aBranchingStoryOffersTheChoicesAndGoesBackToTheParent() {
        assertEquals(listOf("b", "c"), StoryReading.followingChapters(branching, branching[0]).map { chapter -> chapter.id })
        assertEquals("b", StoryReading.previousChapter(branching, branching[3])?.id)
        assertNull(StoryReading.previousChapter(branching, branching[0]))
    }

    @Test
    fun theContentsOfABranchingStoryGoDepthFirst() {
        val contents = StoryReading.contentsOf(branching)
        assertEquals(listOf("a" to 0, "b" to 1, "d" to 2, "c" to 1), contents.map { entry -> entry.chapter.id to entry.depth })
    }

    @Test
    fun progressKeepsTheChapterAndHowFarIntoIt() {
        assertEquals(3.4, StoryReading.progressOf(ReadingPoint(3, 0.4)), 1e-9)
        assertEquals(3.999, StoryReading.progressOf(ReadingPoint(3, 1.0)), 1e-9)
        assertEquals(ReadingPoint(3, 0.5), StoryReading.readingPointOf(3.5))
        assertNull(StoryReading.readingPointOf(null))
    }

    @Test
    fun paragraphsSplitOnBlankLines() {
        assertEquals(listOf("One\ntwo", "Three"), StoryReading.paragraphsOf("One\ntwo\n\n  \nThree\n"))
    }

    @Test
    fun minutesLeftIsAtLeastOne() {
        assertEquals(2, StoryReading.minutesLeft(920, 0.5))
        assertEquals(1, StoryReading.minutesLeft(10, 0.9))
    }
}
