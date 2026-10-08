package com.playingwithclouds.veil.privacy

/** How long the app may stay in the background before it asks for the PIN again. */
enum class LockDelay(val label: String, val milliseconds: Long) {
    IMMEDIATELY("Immediately", 0),
    ONE_MINUTE("After 1 minute", 60_000),
    FIVE_MINUTES("After 5 minutes", 300_000);

    companion object {
        /** The stored delay by name; unknown or missing names mean immediately. */
        fun fromName(name: String?): LockDelay {
            return entries.firstOrNull { delay -> delay.name == name } ?: IMMEDIATELY
        }
    }
}

/** The pure rules of when the app locks and how wrong PINs are throttled. */
object LockPolicy {

    private const val FREE_ATTEMPTS = 5
    private const val LOCKOUT_STEP_MILLISECONDS = 30_000L
    private const val LOCKOUT_MAX_MILLISECONDS = 300_000L

    /** Whether returning to the app at [nowMilliseconds] needs unlocking, given when it left. */
    fun shouldLockOnReturn(enabled: Boolean, delay: LockDelay, backgroundedAtMilliseconds: Long?, nowMilliseconds: Long): Boolean {
        if (!enabled || backgroundedAtMilliseconds == null) {
            return false
        }
        return nowMilliseconds - backgroundedAtMilliseconds >= delay.milliseconds
    }

    /** How long to refuse PIN entry after [failedAttempts] wrong tries in a row. */
    fun lockoutMilliseconds(failedAttempts: Int): Long {
        if (failedAttempts < FREE_ATTEMPTS) {
            return 0
        }
        val steps = failedAttempts - FREE_ATTEMPTS + 1
        return minOf(steps * LOCKOUT_STEP_MILLISECONDS, LOCKOUT_MAX_MILLISECONDS)
    }
}
