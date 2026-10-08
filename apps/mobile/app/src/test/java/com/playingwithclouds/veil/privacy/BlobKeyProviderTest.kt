package com.playingwithclouds.veil.privacy

import java.security.GeneralSecurityException
import java.util.Base64
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertNotNull
import org.junit.Test

class BlobKeyProviderTest {

    /** A wrapper that XORs with a pad, standing in for the Keystore; [broken] mimics a lost key. */
    private class FakeWrapper(var broken: Boolean = false) : KeyWrapper {
        override fun wrap(plain: ByteArray): ByteArray = plain.map { byte -> (byte.toInt() xor PAD).toByte() }.toByteArray()

        override fun unwrap(wrapped: ByteArray): ByteArray {
            if (broken) {
                throw GeneralSecurityException("key gone")
            }
            return wrap(wrapped)
        }
    }

    private class MemorySlot : StringSlot {
        var value: String? = null

        override fun read(): String? = value

        override fun write(value: String) {
            this.value = value
        }
    }

    @Test
    fun generatesA32ByteKeyAndKeepsItAcrossCalls() {
        val slot = MemorySlot()
        val provider = BlobKeyProvider(FakeWrapper(), slot)
        val first = provider.encodedKey()
        assertEquals(32, Base64.getDecoder().decode(first).size)
        assertEquals(first, BlobKeyProvider(FakeWrapper(), slot).encodedKey())
    }

    @Test
    fun theStoredValueIsNotTheKeyItself() {
        val slot = MemorySlot()
        val key = BlobKeyProvider(FakeWrapper(), slot).encodedKey()
        assertNotNull(slot.value)
        assertNotEquals(key, slot.value)
    }

    @Test
    fun aKeyThatCannotBeOpenedIsReplaced() {
        val slot = MemorySlot()
        val wrapper = FakeWrapper()
        val original = BlobKeyProvider(wrapper, slot).encodedKey()
        wrapper.broken = true
        val replacement = BlobKeyProvider(wrapper, slot).encodedKey()
        assertNotEquals(original, replacement)
        wrapper.broken = false
        assertEquals(replacement, BlobKeyProvider(wrapper, slot).encodedKey())
    }

    @Test
    fun garbageInTheSlotIsReplaced() {
        val slot = MemorySlot().also { memory -> memory.value = "not base64!!" }
        assertEquals(32, Base64.getDecoder().decode(BlobKeyProvider(FakeWrapper(), slot).encodedKey()).size)
    }

    private companion object {
        const val PAD = 0x5a
    }
}
