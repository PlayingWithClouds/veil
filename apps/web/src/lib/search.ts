// Searching the sites: one page at a time over the searchResults
// subscription, each site's results as it answers.
import {
	generateSubscriptionOp,
	type FieldsSelection,
	type FilterInput,
	type ListKind,
	type SearchBatch
} from '@playingwithclouds/veil-sdk/generated';
import { getWsClient, gqlClient } from '$lib/veil';
import { GALLERY_CARD, VIDEO_CARD } from '$lib/library';
import { AUDIO_CARD, STORY_CARD } from '$lib/media';

// How long the sites' preference scores are reused before they're asked again.
const PREFERENCE_TTL_MS = 10 * 60 * 1000;

let preferenceCache: { at: number; scores: Promise<Map<string, number>> } | null = null;

/**
 * Each site's preference score (its pick rate over the overall one), by
 * plugin. Empty when the backend can't say; the search then treats every
 * site alike.
 */
export function sitePreferences(): Promise<Map<string, number>> {
	if (preferenceCache && Date.now() - preferenceCache.at < PREFERENCE_TTL_MS)
		return preferenceCache.scores;
	const scores = gqlClient
		.query({ sitePreferences: { plugin: true, score: true } })
		.then((result) => new Map(result.sitePreferences.map((site) => [site.plugin, site.score])))
		.catch(() => new Map<string, number>());
	preferenceCache = { at: Date.now(), scores };
	return scores;
}

/** What the search page reads of one batch. */
const SEARCH_BATCH = {
	plugin: true,
	sites: true,
	stored: true,
	final: true,
	exhausted: true,
	returned: true,
	listed: { videos: VIDEO_CARD, galleries: GALLERY_CARD, audios: AUDIO_CARD, stories: STORY_CARD }
} as const;

export type SearchBatchRow = FieldsSelection<SearchBatch, typeof SEARCH_BATCH>;

/** A running search page; cancel stops listening, done settles when it ends. */
export interface SearchPage {
	done: Promise<void>;
	cancel: () => void;
}

/**
 * Asks the sites for one page of a search (only plugins, when given, and
 * only the sites carrying content, such as "cam-shows"), calling onBatch with
 * each part of a site's answer as it arrives; filter narrows the results to
 * those it matches.
 */
export function searchPage(
	kind: ListKind,
	query: string,
	page: number,
	plugins: string[] | null,
	filter: FilterInput | null,
	content: string | null,
	onBatch: (batch: SearchBatchRow) => void
): SearchPage {
	const client = getWsClient();
	if (!client) return { done: Promise.resolve(), cancel: () => {} };
	const operation = generateSubscriptionOp({
		searchResults: { __args: { kind, query, page, plugins, filter, content }, ...SEARCH_BATCH }
	});
	let cancel = () => {};
	// A cancelled page settles too, so nothing waits on it forever.
	const done = new Promise<void>((resolve, reject) => {
		const unsubscribe = client.subscribe<{ searchResults: SearchBatchRow }>(operation, {
			next: (message) => {
				if (message.data) onBatch(message.data.searchResults);
			},
			error: reject,
			complete: resolve
		});
		cancel = () => {
			unsubscribe();
			resolve();
		};
	});
	return { done, cancel };
}

/** One page of a video search, every site's answer together once the last one is in. */
export async function collectSearchPage(
	query: string,
	page: number
): Promise<SearchBatchRow['listed']['videos']> {
	const videos: SearchBatchRow['listed']['videos'] = [];
	const running = searchPage('VIDEO', query, page, null, null, null, (batch) => {
		videos.push(...batch.listed.videos);
	});
	await running.done;
	return videos;
}
