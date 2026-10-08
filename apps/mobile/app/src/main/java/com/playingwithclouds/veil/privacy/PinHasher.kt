package com.playingwithclouds.veil.privacy

import java.security.MessageDigest
import java.security.SecureRandom
import javax.crypto.SecretKeyFactory
import javax.crypto.spec.PBEKeySpec

/** Salted PBKDF2 hashes of the app lock PIN, stored as `iterations:salt:hash` (hex). */
object PinHasher {

    private const val ITERATIONS = 120_000
    private const val SALT_BYTES = 16
    private const val HASH_BITS = 256
    private const val ALGORITHM = "PBKDF2WithHmacSHA256"

    /** Hashes [pin] with a fresh random salt. */
    fun hash(pin: String): String {
        val salt = ByteArray(SALT_BYTES).also { bytes -> SecureRandom().nextBytes(bytes) }
        return "$ITERATIONS:${toHex(salt)}:${toHex(derive(pin, salt, ITERATIONS))}"
    }

    /** Whether [pin] matches a [stored] hash; malformed hashes never match. */
    fun verify(pin: String, stored: String): Boolean {
        val parts = stored.split(":")
        if (parts.size != 3) {
            return false
        }
        val iterations = parts[0].toIntOrNull() ?: return false
        val salt = fromHex(parts[1]) ?: return false
        val expected = fromHex(parts[2]) ?: return false
        return MessageDigest.isEqual(derive(pin, salt, iterations), expected)
    }

    /** The PBKDF2 key for [pin]. */
    private fun derive(pin: String, salt: ByteArray, iterations: Int): ByteArray {
        val spec = PBEKeySpec(pin.toCharArray(), salt, iterations, HASH_BITS)
        return SecretKeyFactory.getInstance(ALGORITHM).generateSecret(spec).encoded
    }

    /** Lower-case hex of [bytes]. */
    private fun toHex(bytes: ByteArray): String {
        return bytes.joinToString("") { byte -> "%02x".format(byte) }
    }

    /** The bytes of a hex string, or null when it is not valid hex. */
    private fun fromHex(text: String): ByteArray? {
        if (text.length % 2 != 0) {
            return null
        }
        val bytes = ByteArray(text.length / 2)
        for (index in bytes.indices) {
            val value = text.substring(index * 2, index * 2 + 2).toIntOrNull(16) ?: return null
            bytes[index] = value.toByte()
        }
        return bytes
    }
}
