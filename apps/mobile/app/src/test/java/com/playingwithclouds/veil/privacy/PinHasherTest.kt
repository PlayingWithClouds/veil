package com.playingwithclouds.veil.privacy

import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class PinHasherTest {

    @Test
    fun theRightPinVerifies() {
        assertTrue(PinHasher.verify("1234", PinHasher.hash("1234")))
    }

    @Test
    fun aWrongPinDoesNot() {
        assertFalse(PinHasher.verify("1235", PinHasher.hash("1234")))
    }

    @Test
    fun eachHashUsesItsOwnSalt() {
        assertNotEquals(PinHasher.hash("1234"), PinHasher.hash("1234"))
    }

    @Test
    fun malformedHashesNeverMatch() {
        assertFalse(PinHasher.verify("1234", ""))
        assertFalse(PinHasher.verify("1234", "x:y:z"))
        assertFalse(PinHasher.verify("1234", "1:abc:def"))
    }
}
