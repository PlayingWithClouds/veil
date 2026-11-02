// Searching the sites: one page at a time over the searchResults
// subscription, each site's results as it answers.
import {
	generateSubscriptionOp,
	type FieldsSelection,
	type ListKind,
	type SearchBatch
} from '@playingwithclouds/veil-sdk/generated';
import { getWsClient } from '$lib/veil';
import { VIDEO_CARD } from '$lib/library';

/** What the search page reads of one batch. */
const SEARCH_BATCH = {
	plugin: true,
	sites: true,
	stored: true,
	final: true,
	exhausted: true,
	listed: { videos: VIDEO_CARD }
} as const;

export type SearchBatchRow = FieldsSelection<SearchBatch, typeof SEARCH_BATCH>;

/** A running search page; cancel stops listening, done settles when it ends. */
export interface SearchPage {
	done: Promise<void>;
	cancel: () => void;
}

/**
 * Asks the sites for one page of a search (only plugins, when given), calling
 * onBatch with each part of a site's answer as it arrives.
 */
export function searchPage(
	kind: ListKind,
	query: string,
	page: number,
	plugins: string[] | null,
	onBatch: (batch: SearchBatchRow) => void
): SearchPage {
	const client = getWsClient();
	if (!client) return { done: Promise.resolve(), cancel: () => {} };
	const operation = generateSubscriptionOp({
		searchResults: { __args: { kind, query, page, plugins }, ...SEARCH_BATCH }
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
