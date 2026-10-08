package com.playingwithclouds.veil.privacy

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import java.security.GeneralSecurityException
import java.security.KeyStore
import java.security.SecureRandom
import java.util.Base64
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

/** Seals and opens small byte strings with a key that never leaves secure hardware. */
interface KeyWrapper {

    /** The sealed form of [plain]. */
    fun wrap(plain: ByteArray): ByteArray

    /** The bytes sealed by [wrap]; throws when the wrapping key is gone or the data was altered. */
    fun unwrap(wrapped: ByteArray): ByteArray
}

/** A place for one string that survives restarts. */
interface StringSlot {
    fun read(): String?
    fun write(value: String)
}

/**
 * The AES-256 key the backend encrypts downloads and cached streams with. A random data key is
 * generated once and stored only in sealed form (the sealing key lives in the Android Keystore),
 * so the stored bytes are useless off this device. If the sealed key cannot be opened any more
 * (Keystore reset, app data restored from a backup onto another phone) a new key replaces it;
 * blobs encrypted under the old one are then unreadable and get fetched again.
 */
class BlobKeyProvider(private val wrapper: KeyWrapper, private val slot: StringSlot) {

    /** The data key as base64, ready for the backend's `BLOB_ENCRYPTION_KEY`. */
    fun encodedKey(): String {
        val stored = slot.read()
        if (stored != null) {
            val opened = openStored(stored)
            if (opened != null) {
                return Base64.getEncoder().encodeToString(opened)
            }
        }
        val fresh = ByteArray(KEY_BYTES).also { bytes -> SecureRandom().nextBytes(bytes) }
        slot.write(Base64.getEncoder().encodeToString(wrapper.wrap(fresh)))
        return Base64.getEncoder().encodeToString(fresh)
    }

    /** The key inside a stored sealed value, or null when it cannot be opened. */
    private fun openStored(stored: String): ByteArray? {
        try {
            val key = wrapper.unwrap(Base64.getDecoder().decode(stored))
            if (key.size == KEY_BYTES) {
                return key
            }
        } catch (error: GeneralSecurityException) {
            return null
        } catch (error: IllegalArgumentException) {
            return null
        }
        return null
    }

    private companion object {
        const val KEY_BYTES = 32
    }
}

/** [KeyWrapper] backed by an AES-GCM key in the Android Keystore; the IV is stored in front of the sealed bytes. */
class KeystoreKeyWrapper : KeyWrapper {

    /** The Keystore key, created on first use. */
    private fun wrappingKey(): SecretKey {
        val keyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }
        val existing = keyStore.getKey(ALIAS, null)
        if (existing is SecretKey) {
            return existing
        }
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
        generator.init(
            KeyGenParameterSpec.Builder(ALIAS, KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setKeySize(256)
                .build(),
        )
        return generator.generateKey()
    }

    /** Seals [plain] as IV followed by ciphertext. */
    override fun wrap(plain: ByteArray): ByteArray {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, wrappingKey())
        return cipher.iv + cipher.doFinal(plain)
    }

    /** Opens what [wrap] produced. */
    override fun unwrap(wrapped: ByteArray): ByteArray {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        val spec = GCMParameterSpec(TAG_BITS, wrapped, 0, IV_BYTES)
        cipher.init(Cipher.DECRYPT_MODE, wrappingKey(), spec)
        return cipher.doFinal(wrapped, IV_BYTES, wrapped.size - IV_BYTES)
    }

    private companion object {
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val ALIAS = "veil_blob_key_wrap"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
        const val IV_BYTES = 12
        const val TAG_BITS = 128
    }
}

/** The production key provider: Keystore-wrapped, kept in a private preference file. */
object BlobKeys {

    private const val PREFERENCES = "blob-key"
    private const val KEY_SEALED = "sealed"

    /** The base64 data key for this install. */
    fun encodedKey(context: Context): String {
        val preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        val slot = object : StringSlot {
            override fun read(): String? = preferences.getString(KEY_SEALED, null)

            override fun write(value: String) {
                preferences.edit().putString(KEY_SEALED, value).apply()
            }
        }
        return BlobKeyProvider(KeystoreKeyWrapper(), slot).encodedKey()
    }
}
