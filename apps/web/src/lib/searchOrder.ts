import type { PluginSearchResult } from '$lib/search';

/** How the search page orders its results. */
export type SortMode = 'relevance' | 'newest' | 'views' | 'rating' | 'duration';

export const SORT_OPTIONS: { mode: SortMode; label: string }[] = [
	{ mode: 'relevance', label: 'Relevance' },
	{ mode: 'newest', label: 'Newest' },
	{ mode: 'views', label: 'Most viewed' },
	{ mode: 'rating', label: 'Rating' },
	{ mode: 'duration', label: 'Duration' }
];

/**
 * How many of the best results stay put once the first ones are shown: roughly
 * the first screen of the grid. Later arrivals are sorted in below them, so
 * what the reader is looking at never jumps.
 */
export const FROZEN_VISIBLE_COUNT = 12;

/** How long the first results are held back to be sorted together, in milliseconds. */
export const FIRST_RESULTS_WINDOW_MS = 1000;

/** The number a result is sorted by in the given mode; missing values are -Infinity so they sort last. */
function sortKey(result: PluginSearchResult, mode: SortMode): number {
	let value: number | null | undefined;
	switch (mode) {
		case 'relevance':
			value = result.score;
			break;
		case 'newest':
			value = result.date ? Date.parse(result.date) : null;
			break;
		case 'views':
			value = result.viewCount;
			break;
		case 'rating':
			value = result.rating;
			break;
		case 'duration':
			value = result.durationSeconds;
			break;
	}
	if (value === null || value === undefined || Number.isNaN(value)) return -Infinity;
	return value;
}

/** Orders two results best first in the mode; relevance breaks ties. */
export function compareResults(
	mode: SortMode
): (first: PluginSearchResult, second: PluginSearchResult) => number {
	return (first, second) => {
		const byMode = sortKey(second, mode) - sortKey(first, mode);
		if (byMode !== 0 && !Number.isNaN(byMode)) return byMode;
		const byScore = sortKey(second, 'relevance') - sortKey(first, 'relevance');
		if (byScore !== 0 && !Number.isNaN(byScore)) return byScore;
		return 0;
	};
}

/** A copy of the results ordered best first (stable: ties keep their order). */
export function sortResults(results: PluginSearchResult[], mode: SortMode): PluginSearchResult[] {
	return results.toSorted(compareResults(mode));
}

/**
 * Adds a late arrival to results that are already on screen. A result that
 * replaces an earlier copy (the same video from another site, scored better)
 * takes its place without moving. Anything else goes to its sorted position,
 * but never above the first `frozenCount` results, so the visible part keeps
 * its order.
 */
export function insertArrival(
	results: PluginSearchResult[],
	arrival: PluginSearchResult,
	mode: SortMode,
	frozenCount: number
): PluginSearchResult[] {
	if (arrival.replaces) {
		const replacedIndex = results.findIndex((result) => result.sourceUrl === arrival.replaces);
		if (replacedIndex >= 0) {
			return results.with(replacedIndex, arrival);
		}
	}
	const compare = compareResults(mode);
	let position = results.length;
	for (let index = Math.min(frozenCount, results.length); index < results.length; index++) {
		if (compare(arrival, results[index]) < 0) {
			position = index;
			break;
		}
	}
	return results.toSpliced(position, 0, arrival);
}
