package com.playingwithclouds.veil.privacy

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LockPolicyTest {

    @Test
    fun locksAfterTheDelayHasPassed() {
        assertTrue(LockPolicy.shouldLockOnReturn(true, LockDelay.ONE_MINUTE, 1_000, 61_000))
        assertFalse(LockPolicy.shouldLockOnReturn(true, LockDelay.ONE_MINUTE, 1_000, 60_999))
    }

    @Test
    fun immediateLocksOnEveryReturn() {
        assertTrue(LockPolicy.shouldLockOnReturn(true, LockDelay.IMMEDIATELY, 5, 5))
    }

    @Test
    fun neverLocksWhenOffOrNeverLeft() {
        assertFalse(LockPolicy.shouldLockOnReturn(false, LockDelay.IMMEDIATELY, 0, 10_000))
        assertFalse(LockPolicy.shouldLockOnReturn(true, LockDelay.IMMEDIATELY, null, 10_000))
    }

    @Test
    fun wrongPinsAreThrottledFromTheFifthTry() {
        assertEquals(0L, LockPolicy.lockoutMilliseconds(4))
        assertEquals(30_000L, LockPolicy.lockoutMilliseconds(5))
        assertEquals(60_000L, LockPolicy.lockoutMilliseconds(6))
        assertEquals(300_000L, LockPolicy.lockoutMilliseconds(100))
    }

    @Test
    fun unknownStoredNamesFallBack() {
        assertEquals(LockDelay.IMMEDIATELY, LockDelay.fromName("nope"))
        assertEquals(Disguise.NONE, Disguise.fromName(null))
        assertEquals(Disguise.NOTES, Disguise.fromName("NOTES"))
    }
}
