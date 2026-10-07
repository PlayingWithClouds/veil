// Watch page up-next feed: the scene's related scenes (linked when it is
// visited) first, then the personalized feed and live plugin fetches (mirrors
// Home), paged in as the rail scrolls.
import {
	fetchRecommended,
	fetchRecommendedBrowse,
	pluginBrowse,
	type PluginSearchResult
} from '$lib/search';

const RECS_PAGE_SIZE = 12;
const PLUGIN_PAGE_SIZE = 18;

export class RelatedFeed {
	related = $state<PluginSearchResult[]>([]);
	recommendations = $state<PluginSearchResult[]>([]);
	loading = $state(false);
	done = $state(false);

	// Related scenes lead; feed items already shown as related are dropped.
	items = $derived.by(() => {
		const relatedIds = new Set(this.related.map((item) => item.externalId));
		const rest = this.recommendations.filter((item) => !relatedIds.has(item.externalId));
		return [...this.related, ...rest];
	});

	// Two-phase pagination: local recommendedFeed, then live plugin fetches.
	#source: 'recs' | 'plugin' = 'recs';
	#recsOffset = 0;
	#pluginOffset = 0;
	#usedRecommendedBrowse = false;
	#seen = new Set<string>();
	#isCurrentScene: (item: PluginSearchResult) => boolean;

	/** Creates a feed that never lists the scene the predicate identifies as on screen. */
	constructor(isCurrentScene: (item: PluginSearchResult) => boolean) {
		this.#isCurrentScene = isCurrentScene;
	}

	/** Clears everything, for a newly viewed scene. */
	reset() {
		this.related = [];
		this.recommendations = [];
		this.loading = false;
		this.done = false;
		this.#source = 'recs';
		this.#recsOffset = 0;
		this.#pluginOffset = 0;
		this.#usedRecommendedBrowse = false;
		this.#seen.clear();
	}

	/** Replaces the related-scenes head of the feed (from the live subscription). */
	setRelated(items: PluginSearchResult[]) {
		this.related = items.filter((item) => !this.#isCurrentScene(item));
	}

	/** Loads the next page, from the local feed first and plugins once it runs dry. */
	async loadMore() {
		if (this.done || this.loading) return;
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
		return fresh.length;
	}
}
