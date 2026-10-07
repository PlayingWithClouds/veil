// Search suggestions for the phone search overlay (backend: internal/suggest),
// plus the search history they draw on.
import { gqlClient } from '$lib/veil';
import { cacheUrl } from '$lib/img';
import { performerUrl, studioUrl, tagUrl } from '$lib/routes';

export type SearchSuggestionKind = 'RECENT' | 'QUERY' | 'TAG' | 'PERFORMER' | 'STUDIO';

export interface SearchSuggestion {
	kind: SearchSuggestionKind;
	text: string;
	entityId: string | null;
	imageUrl: string | null;
	detail: string | null;
}

/** Suggestions for a partly typed query; an empty query gives recent searches and taste picks. */
export async function fetchSearchSuggestions(query: string, limit = 12): Promise<SearchSuggestion[]> {
	const data = await gqlClient.query({
		searchSuggestions: {
			__args: { query, limit },
			kind: true,
			text: true,
			entityId: true,
			imageUrl: true,
			detail: true
		}
	});
	return data.searchSuggestions.map((suggestion) => ({
		kind: suggestion.kind as SearchSuggestionKind,
		text: suggestion.text,
		entityId: suggestion.entityId ?? null,
		imageUrl: cacheUrl(suggestion.imageUrl),
		detail: suggestion.detail ?? null
	}));
}

/** Remembers a submitted search for the recent list. Failures are ignored. */
export async function recordSearch(query: string): Promise<void> {
	try {
		await gqlClient.mutation({ recordSearch: { __args: { query } } });
	} catch {
		// History is a convenience; a failed write must not block the search.
	}
}

/** Removes a query from the recent searches. */
export async function forgetSearch(query: string): Promise<void> {
	await gqlClient.mutation({ forgetSearch: { __args: { query } } });
}

/**
 * The page an entity suggestion opens, or null for suggestions that run a
 * search (recent searches and query completions).
 */
export function suggestionUrl(suggestion: SearchSuggestion): string | null {
	if (!suggestion.entityId) return null;
	if (suggestion.kind === 'PERFORMER') return performerUrl(suggestion.entityId);
	if (suggestion.kind === 'STUDIO') return studioUrl(suggestion.entityId);
	if (suggestion.kind === 'TAG') return tagUrl(suggestion.entityId);
	return null;
}
