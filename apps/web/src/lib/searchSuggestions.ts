// Search suggestions for the phone search overlay (backend: internal/suggest),
// plus the search history they draw on.
import { goto } from '$app/navigation';
import { gqlClient } from '$lib/veil';
import { cacheUrl } from '$lib/img';
import { performerUrl, studioUrl, tagUrl } from '$lib/routes';
import { searchQuery } from '$lib/stores/searchQuery';

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

/**
 * Completions from the plugins' own search boxes, shaped like the library's
 * suggestions so one list can show both. They carry no entity, so choosing one
 * searches for its text.
 */
export async function fetchSourceSuggestions(query: string, limit = 8): Promise<SearchSuggestion[]> {
	const data = await gqlClient.query({
		sourceSuggestions: { __args: { query, limit }, kind: true, text: true, sources: true }
	});
	return data.sourceSuggestions.map((suggestion) => ({
		kind: suggestion.kind as SearchSuggestionKind,
		text: suggestion.text,
		entityId: null,
		imageUrl: null,
		detail: sourceDetail(suggestion.sources)
	}));
}

/** The grey line under a site completion: which sites offered it. */
function sourceDetail(sources: string[]): string {
	const shown = sources.slice(0, 2).join(', ');
	if (sources.length > 2) return `From ${shown} +${sources.length - 2}`;
	return `From ${shown}`;
}

/** How a suggestion's text is compared for duplicates: case and spacing ignored. */
function textKey(text: string): string {
	return text.trim().replace(/\s+/g, ' ').toLowerCase();
}

// How many of the library's suggestions lead the list before the sites' completions
// join it, so a long library list never pushes the sites' out of view.
const LIBRARY_LEAD = 5;

/**
 * The library's best suggestions first, then the sites' completions the library's
 * don't already cover, then the rest of the library's.
 */
export function withSourceSuggestions(
	local: SearchSuggestion[],
	fromSources: SearchSuggestion[]
): SearchSuggestion[] {
	const known = new Set(local.map((suggestion) => textKey(suggestion.text)));
	const extra = fromSources.filter((suggestion) => !known.has(textKey(suggestion.text)));
	return [...local.slice(0, LIBRARY_LEAD), ...extra, ...local.slice(LIBRARY_LEAD)];
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

/**
 * Runs a query on Home: makes it the shared search, remembers it and opens the
 * results. An empty query goes back to the feed.
 */
export function submitSearch(query: string): void {
	const trimmed = query.trim();
	searchQuery.set(trimmed);
	if (!trimmed) {
		goto('/');
		return;
	}
	recordSearch(trimmed);
	goto(`/?q=${encodeURIComponent(trimmed)}`);
}

/** Opens an entity suggestion's page, or searches for the suggestion's text. */
export function openSuggestion(suggestion: SearchSuggestion): void {
	const url = suggestionUrl(suggestion);
	if (!url) {
		submitSearch(suggestion.text);
		return;
	}
	recordSearch(suggestion.text);
	goto(url);
}
