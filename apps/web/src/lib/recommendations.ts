// Recommendation impression logging: what was shown and what was opened, which
// feeds the engine's fatigue penalty and "shown but not clicked" signal. Home
// fetches the feed itself (explore.ts).
import { gqlClient } from '$lib/veil';

/** One shown (or opened) recommendation, as the recordImpressions mutation takes it. */
export interface ImpressionEvent {
	mediaId: string;
	source?: string;
	// Where it was shown, e.g. "feed" or "row".
	surface?: string;
	position?: number;
	clicked?: boolean;
}

// Impressions are sent in batches: after a short quiet period, or at once when
// the batch fills up or the page is hidden.
const FLUSH_DELAY_MS = 3000;
const BATCH_LIMIT = 500;

let pending: ImpressionEvent[] = [];
let flushTimer: ReturnType<typeof setTimeout> | null = null;
// surface|mediaId pairs already logged as shown, so re-renders don't count twice.
const shownKeys = new Set<string>();

if (typeof window !== 'undefined') {
	window.addEventListener('pagehide', () => void flushImpressions());
}

/** Queues a "shown" impression, once per scene and surface per page load. */
export function recordShown(event: Omit<ImpressionEvent, 'clicked'>): void {
	const key = `${event.surface}|${event.mediaId}`;
	if (shownKeys.has(key)) return;
	shownKeys.add(key);
	enqueue({ ...event, clicked: false });
}

/** Logs that the user opened a recommendation, sending the batch right away. */
export function recordClicked(event: Omit<ImpressionEvent, 'clicked'>): void {
	enqueue({ ...event, clicked: true });
	void flushImpressions();
}

/** Forgets which scenes were logged as shown, e.g. when the feed is reloaded. */
export function resetShownImpressions(): void {
	shownKeys.clear();
}

/**
 * Adds an event to the batch and schedules or triggers the send. Events
 * without a media id are dropped: one invalid id would fail the whole batch.
 */
function enqueue(event: ImpressionEvent): void {
	if (event.mediaId === '') return;
	pending.push(event);
	if (pending.length >= BATCH_LIMIT) {
		void flushImpressions();
		return;
	}
	if (flushTimer === null) {
		flushTimer = setTimeout(() => void flushImpressions(), FLUSH_DELAY_MS);
	}
}

/** Sends every queued impression. Best-effort: failures are dropped. */
export async function flushImpressions(): Promise<void> {
	if (flushTimer !== null) {
		clearTimeout(flushTimer);
		flushTimer = null;
	}
	while (pending.length > 0) {
		const batch = pending.slice(0, BATCH_LIMIT);
		pending = pending.slice(BATCH_LIMIT);
		try {
			await gqlClient.mutation({ recordImpressions: { __args: { impressions: batch } } });
		} catch {
			// Impressions are a ranking signal only; never surface failures.
		}
	}
}

/**
 * Svelte action logging a "shown" impression once at least half of the node
 * has been on screen: `use:impression={{ mediaId, source, surface, position }}`.
 */
export function impression(node: HTMLElement, event: Omit<ImpressionEvent, 'clicked'>) {
	let current = event;
	const observer = new IntersectionObserver(
		(entries) => {
			if (!entries.some((entry) => entry.isIntersecting)) return;
			recordShown(current);
			observer.disconnect();
		},
		{ threshold: 0.5 }
	);
	observer.observe(node);
	return {
		update(next: Omit<ImpressionEvent, 'clicked'>) {
			current = next;
		},
		destroy() {
			observer.disconnect();
		}
	};
}
