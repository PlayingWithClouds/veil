// Search click logging: which results of a query were shown and which one the
// user opened. The backend keeps them (surface "search", with the query) and
// ranks later searches by them: scenes opened before for the same or a similar
// query rank higher, and sites whose results get opened for the query's words
// get a larger share. Separate from recommendations.ts because every event
// carries the query and the same scene may be shown for several queries.
import { gqlClient } from '$lib/veil';
import type { PluginSearchResult } from '$lib/search';

/** How many of the top results count as shown: roughly the first screens of the grid. */
export const SHOWN_RESULT_LIMIT = 24;

/** How long a result list must stay put before it counts as shown, in milliseconds. */
export const SHOWN_DWELL_MS = 2500;

/** One search event as the recordImpressions mutation takes it. */
interface SearchEvent {
	mediaId: string;
	source: string;
	surface: 'search';
	position: number;
	query: string;
	clicked: boolean;
}

/** The event for a result, or null when it is not a stored scene (nothing to learn about). */
function toEvent(
	query: string,
	result: PluginSearchResult,
	position: number,
	clicked: boolean
): SearchEvent | null {
	if (!result.dbId || result.mediaType !== 'scene') return null;
	return {
		mediaId: result.dbId,
		source: result.plugin || 'db',
		surface: 'search',
		position,
		query,
		clicked
	};
}

/** Sends events, best-effort: they only tune the ranking, so failures are dropped. */
async function send(events: SearchEvent[]): Promise<void> {
	if (events.length === 0) return;
	try {
		await gqlClient.mutation({ recordImpressions: { __args: { impressions: events } } });
	} catch {
		// Never surface logging failures.
	}
}

/** Logs the top results of a settled search (in the order they are shown) as shown. */
export function logSearchShown(query: string, results: PluginSearchResult[]): void {
	const events: SearchEvent[] = [];
	results.slice(0, SHOWN_RESULT_LIMIT).forEach((result, position) => {
		const event = toEvent(query, result, position, false);
		if (event) events.push(event);
	});
	void send(events);
}

/** Logs that the user opened a result of the search. */
export function logSearchClick(query: string, result: PluginSearchResult, position: number): void {
	const event = toEvent(query, result, position, true);
	if (event) void send([event]);
}
