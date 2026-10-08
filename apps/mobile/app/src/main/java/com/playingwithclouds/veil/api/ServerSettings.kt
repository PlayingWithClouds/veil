package com.playingwithclouds.veil.api

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import com.playingwithclouds.veil.backend.EmbeddedBackend

/**
 * Which backend this device talks to: the one running on the phone, or another server (e.g. a
 * NAS) saved in Settings. Mirrors `lib/server.ts`.
 */
object ServerSettings {

    private const val PREFERENCES = "server"
    private const val KEY_URL = "url"

    private lateinit var preferences: SharedPreferences

    /** Opens the stored settings; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
    }

    /** The address saved on this device, if any. */
    val storedUrl: String?
        get() = preferences.getString(KEY_URL, null)

    /** The saved address, else the backend the phone runs itself. */
    val baseUrl: String
        get() {
            val stored = storedUrl
            if (stored.isNullOrEmpty()) {
                return EmbeddedBackend.BASE_URL
            }
            return stored
        }

    /** Whether the app talks to the backend running on this phone. */
    val usesEmbeddedBackend: Boolean
        get() = storedUrl.isNullOrEmpty()

    /** Saves another server's address; the GraphQL client follows on its next request. */
    fun save(url: String) {
        preferences.edit { putString(KEY_URL, BackendUrls.normalizeServerUrl(url)) }
    }

    /** Forgets the saved address, so the app goes back to its on-device backend. */
    fun clear() {
        preferences.edit { remove(KEY_URL) }
    }

    /** Points a backend-built URL at the address in use now (see [BackendUrls.rebase]). */
    fun backendUrl(url: String): String {
        return BackendUrls.rebase(url, baseUrl)
    }

    /** A remote image routed through the backend's image cache (see [BackendUrls.imageUrl]). */
    fun imageUrl(url: String?): String? {
        return BackendUrls.imageUrl(url, baseUrl)
    }
}
