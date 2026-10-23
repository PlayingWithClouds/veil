// Impression logging: what was shown and what was opened, with the surface it
// was shown on and its position there. The engine learns from the recommendation
// surfaces (home, row): their sources, positions and clicks set the source
// priors and the fatigue of shown-but-ignored scenes. The other surfaces are
// logged for analysis and position-bias estimates and never change the ranking.
//
// To log a surface, put the action on the element holding a scene's card and
// call recordClicked with the same event when it is opened:
//
//   <div use:impression={{ mediaId: scene.id, surface: 'tag', position: index }}>
//     <SceneCard onclick={() => { recordClicked({ mediaId: scene.id, surface: 'tag', position: index }); open(); }} />
//   </div>
//
// - position is the 0-based place in the order the user sees (for a paged feed
//   the offset of the page plus the index), required so no impression is logged
//   without one.
// - source is the engine's candidate source (RecommendedScene.source) when the
//   list comes from the engine; leave it out otherwise.
// - An empty mediaId (a result not stored yet) is skipped.
// - A scene is logged once per surface per page load; call
//   resetShownImpressions(surface) when the list on a surface is replaced, e.g.
//   on navigating to another performer.
//
// Home fetches the feed itself (explore.ts).
import { gqlClient } from '$lib/veil';

/** The surfaces impressions are logged for. */
export type ImpressionSurface =
	| 'home'
	| 'row'
	| 'related'
	| 'performer'
	| 'studio'
	| 'tag'
	| 'search';

/** One shown (or opened) scene, as the recordImpressions mutation takes it. */
export interface ImpressionEvent {
	mediaId: string;
	source?: string;
	surface: ImpressionSurface;
	position: number;
	clicked?: boolean;
}

/** What the logging functions take: an impression without the click flag. */
export type ShownEvent = Omit<ImpressionEvent, 'clicked'>;

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
export function recordShown(event: ShownEvent): void {
	const key = `${event.surface}|${event.mediaId}`;
	if (shownKeys.has(key)) return;
	shownKeys.add(key);
	enqueue({ ...event, clicked: false });
}

/** Logs that the user opened a recommendation, sending the batch right away. */
export function recordClicked(event: ShownEvent): void {
	enqueue({ ...event, clicked: true });
	void flushImpressions();
}

/**
 * Forgets which scenes were logged as shown on a surface (all surfaces without
 * one), e.g. when its list is reloaded or replaced.
 */
export function resetShownImpressions(surface?: ImpressionSurface): void {
	if (surface === undefined) {
		shownKeys.clear();
		return;
	}
	for (const key of shownKeys) {
		if (key.startsWith(`${surface}|`)) shownKeys.delete(key);
	}
}

/**
 * Adds an event to the batch and schedules or triggers the send. Events
 * that are not about a stored scene are dropped: one invalid id would fail the
 * whole batch.
 */
function enqueue(event: ImpressionEvent): void {
	if (!event.mediaId.startsWith('scene:')) return;
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
export function impression(node: HTMLElement, event: ShownEvent) {
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
		update(next: ShownEvent) {
			current = next;
		},
		destroy() {
			observer.disconnect();
		}
	};
}
