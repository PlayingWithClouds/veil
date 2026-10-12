// The suggestions behind a search field: the library's (fast) and the plugins'
// own search-box completions (slower, fetched live), kept apart so the quick
// ones show first and the others join the list when they arrive.
import {
	fetchSearchSuggestions,
	fetchSourceSuggestions,
	withSourceSuggestions,
	type SearchSuggestion
} from '$lib/searchSuggestions';

// Wait this long after the last keystroke before asking the backend.
const LIBRARY_DEBOUNCE_MS = 120;
// The sites are asked later and only for real input: each question is a live request to them.
const SOURCE_DEBOUNCE_MS = 250;
const SOURCE_MINIMUM_LENGTH = 2;

/**
 * Starts fetching suggestions for whatever query() returns while active() is
 * true. Call it while a component initialises; it owns the effects.
 */
export function createSuggestionFeed(options: { query: () => string; active: () => boolean }) {
	let library = $state<SearchSuggestion[]>([]);
	let fromSources = $state<SearchSuggestion[]>([]);
	const items = $derived(withSourceSuggestions(library, fromSources));
	// Answers can arrive out of order; only the newest request of each kind may land.
	let libraryToken = 0;
	let sourceToken = 0;

	$effect(() => {
		if (!options.active()) return;
		const query = options.query().trim();
		const token = ++libraryToken;
		const timer = setTimeout(async () => {
			try {
				const next = await fetchSearchSuggestions(query);
				if (token === libraryToken) library = next;
			} catch {
				if (token === libraryToken) library = [];
			}
		}, LIBRARY_DEBOUNCE_MS);
		return () => clearTimeout(timer);
	});

	$effect(() => {
		if (!options.active()) return;
		const query = options.query().trim();
		const token = ++sourceToken;
		if (query.length < SOURCE_MINIMUM_LENGTH) {
			fromSources = [];
			return;
		}
		const timer = setTimeout(async () => {
			try {
				const next = await fetchSourceSuggestions(query);
				if (token === sourceToken) fromSources = next;
			} catch {
				if (token === sourceToken) fromSources = [];
			}
		}, SOURCE_DEBOUNCE_MS);
		return () => clearTimeout(timer);
	});

	return {
		/** The library's suggestions alone: recent searches and taste picks when nothing is typed. */
		get library() {
			return library;
		},
		/** The library's suggestions followed by the sites' completions. */
		get items() {
			return items;
		},
		/** Drops a suggestion from the list, e.g. a recent search being forgotten. */
		remove(suggestion: SearchSuggestion) {
			library = library.filter((entry) => entry !== suggestion);
		}
	};
}
