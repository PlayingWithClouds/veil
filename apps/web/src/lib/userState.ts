// The user's state on records from a card's ⋮ menu: blocking, favourites and
// "not interested". Hidden records also land in a local store so every card
// on the page hides at once, before any list reloads.
import { writable } from 'svelte/store';
import type { RecordKind } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { logEvent } from '$lib/events';

/** "kind:id" of every record blocked or dismissed during this page load. */
export const hiddenRecords = writable<Set<string>>(new Set());

/** The key a record has in hiddenRecords. */
export function hiddenKey(kind: RecordKind, id: string): string {
	return `${kind}:${id}`;
}

/** Hides a record's cards, or shows them again. */
export function setHidden(kind: RecordKind, id: string, hidden: boolean): void {
	hiddenRecords.update((current) => {
		const next = new Set(current);
		if (hidden) {
			next.add(hiddenKey(kind, id));
		} else {
			next.delete(hiddenKey(kind, id));
		}
		return next;
	});
}

/** Blocks or unblocks a record: it and media crediting it leave every list. */
export async function setBlocked(kind: RecordKind, id: string, blocked: boolean): Promise<void> {
	setHidden(kind, id, blocked);
	await gqlClient.mutation({
		setBlocked: { __args: { subject: { kind, id }, blocked }, blocked: true }
	});
}

/** Marks a record favourite or not. */
export async function setFavorite(kind: RecordKind, id: string, favorite: boolean): Promise<void> {
	await gqlClient.mutation({
		setFavorite: { __args: { subject: { kind, id }, favorite }, favorite: true }
	});
}

/**
 * Logs a thumbs down: the engine never recommends the record again. The log
 * has no way to take one event back, so callers send it once undo has passed.
 */
export function logNotInterested(kind: RecordKind, id: string): void {
	logEvent({ kind: 'like', subject: { kind, id }, data: { value: -1 } });
}
