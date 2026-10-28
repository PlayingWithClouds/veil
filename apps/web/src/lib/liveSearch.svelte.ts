// One live search over /api/search, ordered by the backend's relevance score
// (or the picked sort): library matches, site videos and galleries in separate
// lists, the library entities the query names, and how far each site got.
//
// The first second of arrivals is held back and sorted together, so the grid
// opens in order. From then on a late result is sorted in below the first
// FROZEN_VISIBLE_COUNT results (and below everything shown when a further page
// loads), so the screen never reshuffles under the reader's eyes. A result
// that `replaces` an earlier copy of the same video takes its place, unless
// that copy is in the library, which always wins.
import {
	pluginSearchStream,
	type PluginSearchResult,
	type SearchMatchedEntity,
	type SearchStream
} from '$lib/search';
import { logSearchClick, logSearchShown, SHOWN_DWELL_MS } from '$lib/searchLog';
import {
	FIRST_RESULTS_WINDOW_MS,
	FROZEN_VISIBLE_COUNT,
	insertArrival,
	isGalleryResult,
	sortResults,
	type SortMode
} from '$lib/searchOrder';

/** Each site's share of one page of live results. */
const PER_PLUGIN = 20;

/** How far one site's search got: still running, finished, or failed. */
export type SiteState = 'searching' | 'done' | 'failed';

/** One site's search: its state and how many results it added. */
export interface SiteStatus {
	state: SiteState;
	count: number;
}

export class LiveSearch {
	// Stored scenes the query matches, best first.
	library = $state.raw<PluginSearchResult[]>([]);
	// Site videos, best first.
	videos = $state.raw<PluginSearchResult[]>([]);
	// Galleries (stored or from sites), best first, ranked apart from the videos.
	galleries = $state.raw<PluginSearchResult[]>([]);
	// Performers, studios and tags of the library the query names.
	matched = $state.raw<SearchMatchedEntity[]>([]);
	// Per searched site (plugin name).
	sites = $state.raw<Map<string, SiteStatus>>(new Map());
	sortMode = $state<SortMode>('relevance');
	loading = $state(false);
	// True once a further page added nothing.
	done = $state(false);
	query = $state('');

	#sources: string[] = [];
	#stream: SearchStream | null = null;
	#token = 0;
	#offset = 0;
	#seen = new Set<string>();
	#held: PluginSearchResult[] = [];
	#firstShown = false;
	#firstTimer: ReturnType<typeof setTimeout> | null = null;
	#shownTimer: ReturnType<typeof setTimeout> | null = null;
	#frozenVideos = 0;
	#frozenGalleries = 0;

	/**
	 * Starts a search for query on sources (plugin names, empty = all);
	 * siteNames are the sites it will ask, for their status.
	 */
	run(query: string, sources: string[], siteNames: string[]) {
		this.stop();
		this.#token++;
		this.query = query;
		this.#sources = sources;
		this.library = [];
		this.videos = [];
		this.galleries = [];
		this.matched = [];
		this.sites = new Map(siteNames.map((name) => [name, { state: 'searching', count: 0 }]));
		this.done = false;
		this.#offset = 0;
		this.#seen = new Set();
		this.#held = [];
		this.#firstShown = false;
		this.#frozenVideos = 0;
		this.#frozenGalleries = 0;
		if (!query) return;
		this.#streamPage(this.#token);
	}

	/** Loads the sites' next page, unless one is loading or the last added nothing. */
	loadMore() {
		if (this.loading || this.done || !this.query) return;
		this.#streamPage(this.#token);
	}

	/** Closes the stream and drops pending timers. */
	stop() {
		if (this.#stream) {
			this.#stream.close();
			this.#stream = null;
		}
		this.#clearTimer(this.#firstTimer);
		this.#firstTimer = null;
		this.#clearTimer(this.#shownTimer);
		this.#shownTimer = null;
		this.loading = false;
	}

	/** Re-sorts everything shown when the reader picks another order. */
	changeSort(mode: SortMode) {
		this.sortMode = mode;
		this.library = sortResults(this.library, mode);
		this.videos = sortResults(this.videos, mode);
		this.galleries = sortResults(this.galleries, mode);
		this.#held = sortResults(this.#held, mode);
		this.#frozenVideos = Math.min(this.videos.length, FROZEN_VISIBLE_COUNT);
		this.#frozenGalleries = Math.min(this.galleries.length, FROZEN_VISIBLE_COUNT);
	}

	/** Logs that the user opened a video or gallery of this search (at its place on screen). */
	logOpened(result: PluginSearchResult) {
		logSearchClick(this.query, result, this.#positionOf(result));
	}

	/** Place of a result in the order shown: library first, then site videos; galleries apart. */
	#positionOf(result: PluginSearchResult): number {
		if (isGalleryResult(result)) return this.galleries.indexOf(result);
		const libraryIndex = this.library.indexOf(result);
		if (libraryIndex >= 0) return libraryIndex;
		return this.library.length + this.videos.indexOf(result);
	}

	/** Streams the sites' next page of results; a page adding nothing ends the list. */
	#streamPage(token: number) {
		this.loading = true;
		// A further page goes below what the reader has scrolled through.
		this.#frozenVideos = this.videos.length;
		this.#frozenGalleries = this.galleries.length;
		let added = 0;
		this.#stream = pluginSearchStream(
			this.query,
			this.#sources,
			(result) => {
				if (token !== this.#token) return;
				if (this.#receive(result)) added++;
			},
			() => {
				if (token !== this.#token) return;
				this.#showFirstResults();
				this.#finishSites();
				this.loading = false;
				this.#stream = null;
				if (added === 0) this.done = true;
				this.#scheduleShownLog();
			},
			{
				offset: this.#offset,
				limit: PER_PLUGIN,
				onMatched: (entities) => {
					if (token === this.#token) this.matched = entities;
				},
				onSite: (plugin, _count, failed) => {
					if (token === this.#token) this.#siteFinished(plugin, failed);
				}
			}
		);
		this.#offset += PER_PLUGIN;
	}

	/** Takes one streamed result; false when it is a repeat or can't be opened. */
	#receive(result: PluginSearchResult): boolean {
		// Results open by their stored record; one without can't be shown.
		if (!result.dbId) return false;
		const key = `${result.plugin}|${result.externalId}`;
		if (this.#seen.has(key)) return false;
		this.#seen.add(key);
		if (!result.plugin && !isGalleryResult(result)) {
			this.library = sortResults([...this.library, result], this.sortMode);
			return true;
		}
		if (this.#replacesLibraryCopy(result)) return false;
		this.#countForSite(result.plugin);
		if (!this.#firstShown) {
			this.#hold(result);
			return true;
		}
		if (isGalleryResult(result)) {
			this.galleries = insertArrival(this.galleries, result, this.sortMode, this.#frozenGalleries);
			return true;
		}
		this.videos = insertArrival(this.videos, result, this.sortMode, this.#frozenVideos);
		return true;
	}

	/** Whether the result is a site copy of a video the library already shows. */
	#replacesLibraryCopy(result: PluginSearchResult): boolean {
		if (!result.replaces) return false;
		return this.library.some((stored) => stored.sourceUrl === result.replaces);
	}

	/** Holds a result of the first second, dropping a held copy it supersedes. */
	#hold(result: PluginSearchResult) {
		if (result.replaces) {
			this.#held = this.#held.filter((held) => held.sourceUrl !== result.replaces);
		}
		this.#held.push(result);
		if (!this.#firstTimer) {
			this.#firstTimer = setTimeout(() => this.#showFirstResults(), FIRST_RESULTS_WINDOW_MS);
		}
	}

	/** Shows the held first results, sorted. */
	#showFirstResults() {
		this.#clearTimer(this.#firstTimer);
		this.#firstTimer = null;
		if (this.#firstShown) return;
		this.#firstShown = true;
		const heldVideos = this.#held.filter((result) => !isGalleryResult(result));
		const heldGalleries = this.#held.filter(isGalleryResult);
		this.videos = sortResults([...this.videos, ...heldVideos], this.sortMode);
		this.galleries = sortResults([...this.galleries, ...heldGalleries], this.sortMode);
		this.#held = [];
		this.#frozenVideos = Math.min(this.videos.length, FROZEN_VISIBLE_COUNT);
		this.#frozenGalleries = Math.min(this.galleries.length, FROZEN_VISIBLE_COUNT);
	}

	/** Adds one result to its site's count. */
	#countForSite(plugin: string) {
		const status = this.sites.get(plugin);
		if (!status) return;
		this.sites = new Map(this.sites).set(plugin, { ...status, count: status.count + 1 });
	}

	/** Marks a site's search finished, or failed. */
	#siteFinished(plugin: string, failed: boolean) {
		const status = this.sites.get(plugin);
		if (!status) return;
		let state: SiteState = 'done';
		if (failed && status.count === 0) state = 'failed';
		this.sites = new Map(this.sites).set(plugin, { ...status, state });
	}

	/** Ends every site still marked searching once the stream is over. */
	#finishSites() {
		const next = new Map(this.sites);
		for (const [plugin, status] of next) {
			if (status.state === 'searching') next.set(plugin, { ...status, state: 'done' });
		}
		this.sites = next;
	}

	/** Logs the top of the results as shown once they stayed put for SHOWN_DWELL_MS. */
	#scheduleShownLog() {
		this.#clearTimer(this.#shownTimer);
		const query = this.query;
		this.#shownTimer = setTimeout(() => {
			logSearchShown(query, [...this.library, ...this.videos]);
		}, SHOWN_DWELL_MS);
	}

	/** Clears a timer that may not be set. */
	#clearTimer(timer: ReturnType<typeof setTimeout> | null) {
		if (timer) clearTimeout(timer);
	}
}
