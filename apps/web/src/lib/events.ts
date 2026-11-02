// The event log the recommendation engine learns from (see
// internal/recommend/events.go for the data each kind carries). Events are
// sent in batches: after a short quiet period, at once for a choice, and when
// the page is hidden.
import type { EventKind, RecordKind } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** Where a card was shown: a surface name and its 0-based place there. */
export interface Shown {
	surface: string;
	position: number;
}

/** A record an event is about. */
export interface Subject {
	kind: RecordKind;
	id: string;
}

/** One event to log, about a record unless it's a search. */
export interface LoggedEvent {
	kind: EventKind;
	subject?: Subject;
	data?: Record<string, unknown>;
	sessionId?: string;
}

const FLUSH_DELAY_MS = 3000;
const BATCH_LIMIT = 200;
const MAX_QUEUED = 1000;

let pending: LoggedEvent[] = [];
let flushTimer: ReturnType<typeof setTimeout> | null = null;
// surface|kind|id of cards already logged as shown, so a re-render or
// scrolling back doesn't count them twice in one page load.
const shownKeys = new Set<string>();

if (typeof window !== 'undefined') {
	window.addEventListener('pagehide', () => void flushEvents());
}

/** Queues an event; a full batch is sent right away. */
export function logEvent(event: LoggedEvent): void {
	pending.push(event);
	if (pending.length >= BATCH_LIMIT) {
		void flushEvents();
		return;
	}
	if (flushTimer === null) {
		flushTimer = setTimeout(() => void flushEvents(), FLUSH_DELAY_MS);
	}
}

/** Logs that a card was shown on a surface, once per page load. */
export function logShown(subject: Subject, shown: Shown): void {
	const key = `${shown.surface}|${subject.kind}|${subject.id}`;
	if (shownKeys.has(key)) return;
	shownKeys.add(key);
	logEvent({ kind: 'impression', subject, data: { ...shown } });
}

/** Logs that a card was picked from a surface, and sends at once. */
export function logChosen(subject: Subject, shown: Shown): void {
	logEvent({ kind: 'choice', subject, data: { ...shown } });
	void flushEvents();
}

/** Forgets what was logged as shown on a surface, when its list is replaced. */
export function resetShown(surface: string): void {
	for (const key of shownKeys) {
		if (key.startsWith(`${surface}|`)) shownKeys.delete(key);
	}
}

/**
 * Sends the queued events. A failed batch goes back in the queue for the next
 * send, up to MAX_QUEUED events so a backend that's down doesn't pile them up.
 */
export async function flushEvents(): Promise<void> {
	if (flushTimer !== null) {
		clearTimeout(flushTimer);
		flushTimer = null;
	}
	if (pending.length === 0) return;
	const batch = pending;
	pending = [];
	try {
		await gqlClient.mutation({ recordEvents: { __args: { events: batch } } });
	} catch (failure) {
		console.warn('events were not logged, retrying with the next batch', failure);
		pending = [...batch, ...pending].slice(-MAX_QUEUED);
	}
}
