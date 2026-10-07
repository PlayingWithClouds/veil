// In-memory snapshot of Home's recommendation feed, so leaving for a scene and
// coming back shows the same cards at the same scroll position instead of a
// freshly re-ranked feed. Lives for the page session only (a reload re-ranks).
import type { GridItem } from '$lib/explore';

/** How long a snapshot is reused before Home asks the engine again. */
const HOME_FEED_TTL_MS = 10 * 60 * 1000;

export interface HomeFeedSnapshot {
	// Media filter and site selection the feed was built for.
	key: string;
	savedAt: number;
	gridItems: GridItem[];
	sceneOffset: number;
	galleryOffset: number;
	sceneFeedDone: boolean;
	galleryFeedDone: boolean;
	recommendationSources: Map<string, string>;
	scrollTop: number;
}

let snapshot: HomeFeedSnapshot | null = null;

/** Cache key for a feed: which media and which sites it shows. */
export function homeFeedKey(mediaFilter: string, sources: string[]): string {
	return `${mediaFilter}|${[...sources].sort().join(',')}`;
}

/** Stores the feed's current state, replacing any older snapshot. */
export function saveHomeFeed(next: HomeFeedSnapshot): void {
	snapshot = next;
}

/** The snapshot for this key, unless it is missing or older than the TTL. */
export function loadHomeFeed(key: string): HomeFeedSnapshot | null {
	if (!snapshot || snapshot.key !== key) return null;
	if (Date.now() - snapshot.savedAt > HOME_FEED_TTL_MS) return null;
	return snapshot;
}

/** Drops the snapshot, e.g. when the user asks for fresh recommendations. */
export function clearHomeFeed(): void {
	snapshot = null;
}
