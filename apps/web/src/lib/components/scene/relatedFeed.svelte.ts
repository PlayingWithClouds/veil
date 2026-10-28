// Watch page up-next feed: the scene's related scenes (linked when it is
// visited) first, then the personalized feed and live plugin fetches (mirrors
// Home), paged in as the rail scrolls.
//
// The list stays hidden until the related scenes are known: a scene's first
// visit links them only once its page is scraped, and showing the feed before
// would put the related scenes on top of cards the reader is already looking
// at. Once shown, cards never move: later arrivals are appended.
import {
	fetchRecommended,
	fetchRecommendedBrowse,
	pluginBrowse,
	type PluginSearchResult
} from '$lib/search';
import { resetShownImpressions } from '$lib/recommendations';

const RECS_PAGE_SIZE = 12;
const PLUGIN_PAGE_SIZE = 18;
// After the visit ends, how long a related list it linked may take to arrive.
const VISIT_GRACE_MS = 1000;
// The longest the rail waits for related scenes before showing the feed anyway.
const MAX_WAIT_MS = 8000;

export class RelatedFeed {
	related = $state.raw<PluginSearchResult[]>([]);
	recommendations = $state.raw<PluginSearchResult[]>([]);
	// What the rail shows, in a fixed order; empty until ready.
	items = $state.raw<PluginSearchResult[]>([]);
	// Whether the related scenes are known and the list is shown.
	ready = $state(false);
	loading = $state(false);
	done = $state(false);

	// Two-phase pagination: local recommendedFeed, then live plugin fetches.
	#source: 'recs' | 'plugin' = 'recs';
	#recsOffset = 0;
	#pluginOffset = 0;
	#usedRecommendedBrowse = false;
	#seen = new Set<string>();
	#isCurrentScene: (item: PluginSearchResult) => boolean;
	// Whether the subscription sent the stored related list yet.
	#heardRelated = false;
	// Whether the visit ended and its grace passed.
	#visitSettled = false;
	#visitTimer: ReturnType<typeof setTimeout> | null = null;
	#waitTimer: ReturnType<typeof setTimeout> | null = null;

	/** Creates a feed that never lists the scene the predicate identifies as on screen. */
	constructor(isCurrentScene: (item: PluginSearchResult) => boolean) {
		this.#isCurrentScene = isCurrentScene;
	}

	/** Clears everything, for a newly viewed scene. */
	reset() {
		resetShownImpressions('related');
		this.related = [];
		this.recommendations = [];
		this.items = [];
		this.ready = false;
		this.loading = false;
		this.done = false;
		this.#source = 'recs';
		this.#recsOffset = 0;
		this.#pluginOffset = 0;
		this.#usedRecommendedBrowse = false;
		this.#seen.clear();
		this.#heardRelated = false;
		this.#visitSettled = false;
		this.#clearTimers();
		this.#waitTimer = setTimeout(() => this.#show(), MAX_WAIT_MS);
	}

	/** Stops the pending timers, for a page being left. */
	dispose() {
		this.#clearTimers();
	}

	/**
	 * Tells the feed the scene's visit is over (or was never needed), so the
	 * related list the subscription has is all there is for now.
	 */
	markVisited() {
		if (this.ready || this.#visitTimer) return;
		this.#visitTimer = setTimeout(() => {
			this.#visitSettled = true;
			if (this.#heardRelated) this.#show();
		}, VISIT_GRACE_MS);
	}

	/** Replaces the related-scenes head of the feed (from the live subscription). */
	setRelated(items: PluginSearchResult[]) {
		this.#heardRelated = true;
		this.related = items.filter((item) => !this.#isCurrentScene(item));
		if (this.related.length > 0 || this.#visitSettled) {
			this.#show();
			return;
		}
		this.#publish();
	}

	/** Loads the next page, from the local feed first and plugins once it runs dry. */
	async loadMore() {
		if (this.done || this.loading) return;
		// While hidden, the first page is all the rail needs at hand.
		if (!this.ready && this.recommendations.length > 0) return;
		this.loading = true;
		try {
			if (this.#source === 'recs') {
				await this.#loadRecommendedPage();
			} else {
				await this.#loadPluginPage();
			}
		} finally {
			this.loading = false;
		}
	}

	/** Shows the list, related scenes first. */
	#show() {
		if (this.ready) return;
		this.ready = true;
		this.#clearTimers();
		this.#publish();
	}

	/** Appends to the shown list whatever it lacks, related scenes before the feed. */
	#publish() {
		if (!this.ready) return;
		const shownIds = new Set(this.items.map((item) => item.externalId));
		const added = [...this.related, ...this.recommendations].filter((item) => {
			if (shownIds.has(item.externalId)) return false;
			shownIds.add(item.externalId);
			return true;
		});
		if (added.length > 0) this.items = [...this.items, ...added];
	}

	/** Next page of the local recommended feed; switches to plugins when exhausted. */
	async #loadRecommendedPage() {
		const page = await fetchRecommended(RECS_PAGE_SIZE, this.#recsOffset);
		this.#recsOffset += page.length;
		this.#appendFresh(page);
		if (page.length < RECS_PAGE_SIZE) this.#source = 'plugin';
	}

	/** First plugin page is taste-biased; deeper pages fall back to generic browse. */
	async #loadPluginPage() {
		if (!this.#usedRecommendedBrowse) {
			this.#usedRecommendedBrowse = true;
			this.#appendFresh(await fetchRecommendedBrowse(PLUGIN_PAGE_SIZE));
			return;
		}
		const page = await pluginBrowse(PLUGIN_PAGE_SIZE, this.#pluginOffset);
		this.#pluginOffset += page.length;
		const added = this.#appendFresh(page);
		if (page.length < PLUGIN_PAGE_SIZE || added === 0) this.done = true;
	}

	/** Appends unseen items that aren't the current scene; returns how many were added. */
	#appendFresh(page: PluginSearchResult[]): number {
		const fresh: PluginSearchResult[] = [];
		for (const item of page) {
			if (this.#seen.has(item.externalId) || this.#isCurrentScene(item)) continue;
			this.#seen.add(item.externalId);
			fresh.push(item);
		}
		this.recommendations = [...this.recommendations, ...fresh];
		this.#publish();
		return fresh.length;
	}

	/** Drops the visit and wait timers. */
	#clearTimers() {
		if (this.#visitTimer) clearTimeout(this.#visitTimer);
		this.#visitTimer = null;
		if (this.#waitTimer) clearTimeout(this.#waitTimer);
		this.#waitTimer = null;
	}
}
