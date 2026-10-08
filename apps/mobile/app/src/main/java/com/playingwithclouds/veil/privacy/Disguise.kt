package com.playingwithclouds.veil.privacy

import android.content.ComponentName
import android.content.Context
import android.content.pm.PackageManager

/**
 * The launcher entries the app can show. Each maps to an `activity-alias` in the manifest, and
 * exactly one is enabled at a time; [NONE] is the real name and icon.
 */
enum class Disguise(val label: String, val aliasName: String) {
    NONE("Veil", ".alias.DefaultAlias"),
    FILES("Files", ".alias.FilesAlias"),
    GALLERY("Gallery", ".alias.GalleryAlias"),
    NOTES("Notes", ".alias.NotesAlias"),
    CALCULATOR("Calculator", ".alias.CalculatorAlias");

    companion object {
        /** The stored disguise by name; unknown or missing names mean none. */
        fun fromName(name: String?): Disguise {
            return entries.firstOrNull { disguise -> disguise.name == name } ?: NONE
        }
    }
}

/** Switches which launcher alias is enabled. */
object Disguises {

    /** Enables the alias of [disguise] and disables the others, enabling first so the app never vanishes. */
    fun apply(context: Context, disguise: Disguise) {
        val packageManager = context.packageManager
        setState(context, packageManager, disguise, PackageManager.COMPONENT_ENABLED_STATE_ENABLED)
        for (other in Disguise.entries) {
            if (other != disguise) {
                setState(context, packageManager, other, PackageManager.COMPONENT_ENABLED_STATE_DISABLED)
            }
        }
    }

    /** Sets one alias's enabled state without killing the app. */
    private fun setState(context: Context, packageManager: PackageManager, disguise: Disguise, state: Int) {
        val component = ComponentName(context.packageName, context.packageName + disguise.aliasName)
        packageManager.setComponentEnabledSetting(component, state, PackageManager.DONT_KILL_APP)
    }
}
