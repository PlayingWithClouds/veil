// The library grid's filter as it lives in the URL on home, so a filtered view
// can be shared, reloaded and gone back to. Each control reads its value from
// the search params and builds the URL that changes it.
import type {
	FieldsSelection,
	FilterInput,
	SortField,
	TagCount
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** Search params the grid filter uses; anything else on the URL is left alone. */
export const FILTER_PARAMS = [
	'tag',
	'not',
	'performer',
	'studio',
	'site',
	'len',
	'res',
	'rating',
	'unwatched',
	'fav',
	'sort'
] as const;

/** A choice in one of the panel's single-pick groups. */
export interface FilterOption {
	value: string;
	label: string;
}

export const LENGTH_OPTIONS: FilterOption[] = [
	{ value: 'short', label: 'Under 10 min' },
	{ value: 'medium', label: '10–30 min' },
	{ value: 'long', label: '30–60 min' },
	{ value: 'feature', label: 'Over 1 h' }
];

// Seconds each length option spans; null leaves that end open.
const LENGTH_BOUNDS: Record<string, [number | null, number | null]> = {
	short: [null, 600],
	medium: [600, 1800],
	long: [1800, 3600],
	feature: [3600, null]
};

export const RESOLUTION_OPTIONS: FilterOption[] = [
	{ value: '720', label: '720p+' },
	{ value: '1080', label: '1080p+' },
	{ value: '2160', label: '4K' }
];

export const RATING_OPTIONS: FilterOption[] = [
	{ value: '6', label: '6+' },
	{ value: '8', label: '8+' },
	{ value: '9', label: '9+' }
];

export const SORT_OPTIONS: FilterOption[] = [
	{ value: 'date', label: 'Release date' },
	{ value: 'duration', label: 'Longest' },
	{ value: 'rating', label: 'Your rating' },
	{ value: 'viewCount', label: 'Most watched' },
	{ value: 'random', label: 'Shuffle' }
];

// Sort values that may come from the URL, as the API's fields.
const SORT_FIELDS: Record<string, SortField> = {
	date: 'date',
	duration: 'duration',
	rating: 'rating',
	viewCount: 'viewCount',
	random: 'random'
};

/** Whether the URL narrows the grid at all (sorting alone doesn't). */
export function isFiltered(params: URLSearchParams): boolean {
	return FILTER_PARAMS.some((name) => name !== 'sort' && params.has(name));
}

/** The API filter the URL describes. */
export function filterFromParams(params: URLSearchParams): FilterInput {
	const input: FilterInput = {};
	addList(input, 'includeTagIds', params.getAll('tag'));
	addList(input, 'excludeTagIds', params.getAll('not'));
	addList(input, 'performerIds', params.getAll('performer'));
	addList(input, 'studioIds', params.getAll('studio'));
	addList(input, 'sources', params.getAll('site'));
	addLength(input, params.get('len'));
	addNumber(input, 'minHeight', params.get('res'));
	addNumber(input, 'minRating', params.get('rating'));
	if (params.has('unwatched')) {
		input.unwatched = true;
	}
	if (params.has('fav')) {
		input.favorite = true;
	}
	const sortField = SORT_FIELDS[params.get('sort') ?? ''];
	if (sortField) {
		input.sort = { field: sortField, descending: true };
	}
	return input;
}

type ListField = 'includeTagIds' | 'excludeTagIds' | 'performerIds' | 'studioIds' | 'sources';

/** Sets a list field when the URL gave any values. */
function addList(input: FilterInput, field: ListField, values: string[]): void {
	if (values.length > 0) {
		input[field] = values;
	}
}

/** Sets a number field when the URL's value is a number. */
function addNumber(
	input: FilterInput,
	field: 'minHeight' | 'minRating',
	value: string | null
): void {
	const number = Number(value);
	if (value && Number.isFinite(number)) {
		input[field] = number;
	}
}

/** Sets the duration bounds of a length option. */
function addLength(input: FilterInput, length: string | null): void {
	const bounds = LENGTH_BOUNDS[length ?? ''];
	if (!bounds) return;
	const [minimum, maximum] = bounds;
	if (minimum !== null) {
		input.minDurationSeconds = minimum;
	}
	if (maximum !== null) {
		input.maxDurationSeconds = maximum;
	}
}

/**
 * The URL with one single-pick param set, or removed when value is null or
 * already set. A search (?q=) stays: filters narrow its results too.
 */
export function toggledParam(url: URL, name: string, value: string | null): string {
	const next = new URL(url);
	if (value === null || url.searchParams.get(name) === value) {
		next.searchParams.delete(name);
	} else {
		next.searchParams.set(name, value);
	}
	return pathAndQuery(next);
}

/** The URL with value added to a multi-value param, or taken out when it's there. */
export function toggledListParam(url: URL, name: string, value: string): string {
	const next = new URL(url);
	const values = next.searchParams.getAll(name);
	next.searchParams.delete(name);
	const kept = values.filter((existing) => existing !== value);
	if (kept.length === values.length) {
		kept.push(value);
	}
	for (const keptValue of kept) {
		next.searchParams.append(name, keptValue);
	}
	return pathAndQuery(next);
}

/**
 * The URL with every value of a group in a multi-value param, or with all of
 * them taken out when every one is there already.
 */
export function toggledListGroup(url: URL, name: string, group: string[]): string {
	const next = new URL(url);
	const values = next.searchParams.getAll(name);
	const allThere = group.every((value) => values.includes(value));
	next.searchParams.delete(name);
	let kept = values.filter((value) => !group.includes(value));
	if (!allThere) {
		kept = [...kept, ...group];
	}
	for (const keptValue of kept) {
		next.searchParams.append(name, keptValue);
	}
	return pathAndQuery(next);
}

/** The URL with a tag moved between required and excluded (or added excluded). */
export function excludedTagUrl(url: URL, tagId: string): string {
	const next = new URL(url);
	const required = next.searchParams.getAll('tag').filter((id) => id !== tagId);
	next.searchParams.delete('tag');
	for (const id of required) {
		next.searchParams.append('tag', id);
	}
	if (!next.searchParams.getAll('not').includes(tagId)) {
		next.searchParams.append('not', tagId);
	}
	return pathAndQuery(next);
}

/** The URL without any grid filter (the sort stays). */
export function clearedFilterUrl(url: URL): string {
	const next = new URL(url);
	for (const name of FILTER_PARAMS) {
		if (name !== 'sort') next.searchParams.delete(name);
	}
	return pathAndQuery(next);
}

/** The path and query of a URL, for goto and links. */
function pathAndQuery(url: URL): string {
	return url.pathname + url.search;
}

/** A tag and how many grid videos carry it. */
const TAG_COUNT = { tag: { id: true, name: true }, count: true } as const;

export type TagCountRow = FieldsSelection<TagCount, typeof TAG_COUNT>;

/** The tags most common among the videos matching filter, most common first. */
export async function fetchTagCounts(filter: FilterInput, limit = 24): Promise<TagCountRow[]> {
	const data = await gqlClient.query({
		videoTagCounts: { __args: { filter, limit }, ...TAG_COUNT }
	});
	return data.videoTagCounts;
}

/** Names of the tags, performers and studios the URL filters on, by id. */
export async function fetchFilterNames(params: URLSearchParams): Promise<Map<string, string>> {
	const lookups = [
		...[...params.getAll('tag'), ...params.getAll('not')].map(tagName),
		...params.getAll('performer').map(performerName),
		...params.getAll('studio').map(studioName)
	];
	const found = await Promise.all(lookups);
	const names = new Map<string, string>();
	for (const named of found) {
		if (named) names.set(named.id, named.name);
	}
	return names;
}

/** One tag's id and name. */
async function tagName(id: string): Promise<{ id: string; name: string } | null> {
	const data = await gqlClient.query({ tag: { __args: { id }, id: true, name: true } });
	return data.tag ?? null;
}

/** One performer's id and name. */
async function performerName(id: string): Promise<{ id: string; name: string } | null> {
	const data = await gqlClient.query({ performer: { __args: { id }, id: true, name: true } });
	return data.performer ?? null;
}

/** One studio's id and name. */
async function studioName(id: string): Promise<{ id: string; name: string } | null> {
	const data = await gqlClient.query({ studio: { __args: { id }, id: true, name: true } });
	return data.studio ?? null;
}

/** The user's latest searches, newest first. */
export async function fetchRecentSearches(limit = 6): Promise<string[]> {
	const data = await gqlClient.query({ recentSearches: { __args: { limit } } });
	return data.recentSearches;
}
