package com.playingwithclouds.veil.data

import kotlin.random.Random

/** What happens at the end of the queue, or of the video: stop, start the queue over, play the video again. */
enum class RepeatMode(val label: String) {
    OFF("Repeat off"),
    ALL("Repeat all"),
    ONE("Repeat one");

    /** The next mode of the repeat button: off → all → one → off. */
    fun next(): RepeatMode {
        return entries[(ordinal + 1) % entries.size]
    }
}

/**
 * The play order of the queue, as a music app keeps it: the queue's order, or a shuffled one that
 * stays put while the queue changes (new videos go to its end, removed ones drop out). The page
 * playing need not be in the queue: then the queue's first video is next.
 */
object QueueOrder {

    /** A new shuffled order of [ids]; [currentId] goes first when it's in the queue, so what plays stays. */
    fun shuffled(ids: List<String>, currentId: String?, random: Random): List<String> {
        val rest = ids.filter { id -> id != currentId }.shuffled(random)
        if (currentId != null && currentId in ids) {
            return listOf(currentId) + rest
        }
        return rest
    }

    /** [shuffledIds] kept in step with the queue [ids]: gone ones dropped, new ones at the end. */
    fun reconcile(shuffledIds: List<String>, ids: List<String>): List<String> {
        val present = ids.toSet()
        val kept = shuffledIds.filter { id -> id in present }
        val keptSet = kept.toSet()
        return kept + ids.filter { id -> id !in keptSet }
    }

    /** The order videos play in: [ids] as queued, or [shuffledIds] reconciled with them when shuffling. */
    fun playOrder(ids: List<String>, shuffledIds: List<String>?): List<String> {
        if (shuffledIds == null) {
            return ids
        }
        return reconcile(shuffledIds, ids)
    }

    /**
     * What plays after [currentId]: the same again on repeat one; the next in [order], or the
     * first when the playing one isn't queued; past the end, the first again on repeat all (the
     * playing one when it's the only one), else nothing.
     */
    fun next(order: List<String>, currentId: String, repeat: RepeatMode): String? {
        if (repeat == RepeatMode.ONE) {
            return currentId
        }
        if (order.isEmpty()) {
            return null
        }
        val index = order.indexOf(currentId)
        if (index < 0) {
            return order.first()
        }
        if (index < order.size - 1) {
            return order[index + 1]
        }
        if (repeat == RepeatMode.ALL) {
            return order.first()
        }
        return null
    }

    /** What the next button skips to: like [next], but repeat one moves on as repeat all does. */
    fun skip(order: List<String>, currentId: String, repeat: RepeatMode): String? {
        var mode = repeat
        if (mode == RepeatMode.ONE) {
            mode = RepeatMode.ALL
        }
        val target = next(order, currentId, mode)
        if (target == currentId) {
            return null
        }
        return target
    }

    /** [ids] with the one at [from] moved to [to]. */
    fun moved(ids: List<String>, from: Int, to: Int): List<String> {
        if (from !in ids.indices || to !in ids.indices) {
            return ids
        }
        val out = ids.toMutableList()
        out.add(to, out.removeAt(from))
        return out
    }
}
