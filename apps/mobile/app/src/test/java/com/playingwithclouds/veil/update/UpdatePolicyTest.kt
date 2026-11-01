package com.playingwithclouds.veil.update

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class UpdatePolicyTest {

    @Test
    fun versionCodeMatchesTheGradleFormula() {
        assertEquals(1_002_003, UpdatePolicy.versionCodeOf("1.2.3"))
        assertEquals(3, UpdatePolicy.versionCodeOf("0.0.3"))
        assertEquals(12_034_005, UpdatePolicy.versionCodeOf("12.34.5"))
    }

    @Test
    fun versionCodeIgnoresTheVPrefixAndASuffix() {
        assertEquals(1_002_003, UpdatePolicy.versionCodeOf("v1.2.3"))
        assertEquals(1_002_003, UpdatePolicy.versionCodeOf("v1.2.3-beta"))
    }

    @Test
    fun malformedTagsHaveNoVersionCode() {
        assertNull(UpdatePolicy.versionCodeOf("plugins"))
        assertNull(UpdatePolicy.versionCodeOf("v1.2"))
        assertNull(UpdatePolicy.versionCodeOf("v1.2.3.4"))
        assertNull(UpdatePolicy.versionCodeOf("v1.x.3"))
        assertNull(UpdatePolicy.versionCodeOf(""))
    }

    @Test
    fun newerReleaseIsAnUpdate() {
        assertTrue(UpdatePolicy.isNewer("v0.0.3", 2))
        assertTrue(UpdatePolicy.isNewer("v1.0.0", 999_999))
    }

    @Test
    fun equalOrOlderReleaseIsNoUpdate() {
        assertFalse(UpdatePolicy.isNewer("v0.0.3", 3))
        assertFalse(UpdatePolicy.isNewer("v0.0.2", 3))
    }

    @Test
    fun malformedTagIsNoUpdate() {
        assertFalse(UpdatePolicy.isNewer("plugins", 0))
        assertFalse(UpdatePolicy.isNewer("latest", 0))
    }

    @Test
    fun checksAtMostOnceADay() {
        val day = UpdatePolicy.CHECK_INTERVAL_MILLISECONDS
        assertTrue(UpdatePolicy.isCheckDue(0, 1_000))
        assertFalse(UpdatePolicy.isCheckDue(10_000, 10_000 + day - 1))
        assertTrue(UpdatePolicy.isCheckDue(10_000, 10_000 + day))
    }

    @Test
    fun clockSetBackMakesACheckDue() {
        assertTrue(UpdatePolicy.isCheckDue(10_000, 5_000))
    }
}
