import { gqlClient } from '$lib/veil';
import { cacheUrl } from '$lib/img';
import type { SceneCardData } from '$lib/components/SceneCard.svelte';

// A browse/discovery filter. All fields optional; empty means unconstrained.
export interface SceneFilter {
	search?: string | null;
	studioId?: string | null;
	performerId?: string | null;
	tagId?: string | null;
	minRating?: number | null;
	minDuration?: number | null;
	maxDuration?: number | null;
	dateFrom?: string | null;
	dateTo?: string | null;
	sort?: string | null;
}

export interface SavedFilter {
	id: string;
	name: string;
	filter: SceneFilter;
	createdAt: string;
}

// True when at least one dimension of the filter is set.
export function isFilterActive(filter: SceneFilter): boolean {
	return Boolean(
		(filter.search && filter.search.trim()) ||
			filter.studioId ||
			filter.performerId ||
			filter.tagId ||
			filter.minRating ||
			filter.minDuration ||
			filter.maxDuration ||
			filter.dateFrom ||
			filter.dateTo ||
			(filter.sort && filter.sort !== 'created_at')
	);
}

export function emptyFilter(): SceneFilter {
	return {};
}

export interface SceneRow {
	id: string;
	title: string;
	posterPath: string | null;
	previewVideo: string | null;
	previewImages: string[];
	durationSeconds: number | null;
	sourceUrl: string;
	date: string | null;
	studio: { name: string; imagePath: string | null } | null;
	// "direct" | "inherited" — set when the query filtered by tag; inherited
	// means the tag came from the scene's studio or a credited performer.
	tagMatch?: string | null;
}

// Scene fields every grid card needs, for the SceneRow queries.
export const SCENE_ROW_FIELDS = {
	id: true,
	title: true,
	posterPath: true,
	previewVideo: true,
	previewImages: true,
	durationSeconds: true,
	sourceUrl: true,
	date: true,
	studio: { name: true, imagePath: true },
	tagMatch: true
} as const;

// True when a structured (non-text) filter dimension is set — the signal to
// switch a page from plugin/live results to DB-backed filtered results.
export function hasStructuredFilter(filter: SceneFilter): boolean {
	return Boolean(
		filter.studioId ||
			filter.performerId ||
			filter.tagId ||
			filter.minRating ||
			filter.minDuration ||
			filter.maxDuration ||
			filter.dateFrom ||
			filter.dateTo ||
			(filter.sort && filter.sort !== 'created_at')
	);
}

export function sceneToCard(scene: SceneRow): SceneCardData {
	const previewImages = (scene.previewImages ?? [])
		.map((url) => cacheUrl(url))
		.filter((url): url is string => url !== null);
	return {
		id: scene.id,
		title: scene.title,
		bannerUrl: cacheUrl(scene.posterPath),
		previewImages,
		previewVideo: cacheUrl(scene.previewVideo),
		durationSeconds: scene.durationSeconds,
		channelName: scene.studio?.name ?? null,
		channelImageUrl: cacheUrl(scene.studio?.imagePath),
		sourceUrl: scene.sourceUrl,
		date: scene.date,
		inheritedTag: scene.tagMatch === 'inherited'
	};
}

// Fetches DB scenes matching the filter. The backend applies the blocklist.
export async function fetchFilteredScenes(
	filter: SceneFilter,
	limit: number,
	offset: number
): Promise<SceneRow[]> {
	const result = await gqlClient.query({
		scenes: {
			__args: {
				limit,
				offset,
				search: filter.search || undefined,
				studioId: filter.studioId || undefined,
				performerId: filter.performerId || undefined,
				tagId: filter.tagId || undefined,
				minRating: filter.minRating || undefined,
				minDuration: filter.minDuration || undefined,
				maxDuration: filter.maxDuration || undefined,
				dateFrom: filter.dateFrom || undefined,
				dateTo: filter.dateTo || undefined,
				sort: filter.sort || undefined
			},
			...SCENE_ROW_FIELDS
		}
	});
	return (result.scenes ?? []) as SceneRow[];
}

// Fetches a random selection of scenes, respecting the blocklist.
export async function fetchRandomScenes(limit: number): Promise<SceneRow[]> {
	const result = await gqlClient.query({
		randomScenes: {
			__args: { limit },
			...SCENE_ROW_FIELDS
		}
	});
	return (result.randomScenes ?? []) as SceneRow[];
}

export interface FilterOption {
	id: string;
	name: string;
}

// Loads selectable tag/studio/performer options for the filter dropdowns.
export async function fetchFilterOptions(): Promise<{
	tags: FilterOption[];
	studios: FilterOption[];
	performers: FilterOption[];
}> {
	const result = await gqlClient.query({
		tags: { __args: { limit: 300 }, id: true, name: true },
		studios: { __args: { limit: 300 }, id: true, name: true },
		performers: { __args: { limit: 300 }, id: true, name: true }
	});
	return {
		tags: (result.tags ?? []) as FilterOption[],
		studios: (result.studios ?? []) as FilterOption[],
		performers: (result.performers ?? []) as FilterOption[]
	};
}

export async function fetchSavedFilters(): Promise<SavedFilter[]> {
	const result = await gqlClient.query({
		savedFilters: {
			id: true,
			name: true,
			createdAt: true,
			filter: {
				search: true,
				studioId: true,
				performerId: true,
				tagId: true,
				minRating: true,
				minDuration: true,
				maxDuration: true,
				dateFrom: true,
				dateTo: true,
				sort: true
			}
		}
	});
	return (result.savedFilters ?? []) as SavedFilter[];
}

export async function createSavedFilter(
	name: string,
	filter: SceneFilter
): Promise<SavedFilter> {
	const result = await gqlClient.mutation({
		createSavedFilter: {
			__args: {
				name,
				filter: {
					search: filter.search || null,
					studioId: filter.studioId || null,
					performerId: filter.performerId || null,
					tagId: filter.tagId || null,
					minRating: filter.minRating || null,
					minDuration: filter.minDuration || null,
					maxDuration: filter.maxDuration || null,
					dateFrom: filter.dateFrom || null,
					dateTo: filter.dateTo || null,
					sort: filter.sort || null
				}
			},
			id: true,
			name: true,
			createdAt: true,
			filter: {
				search: true,
				studioId: true,
				performerId: true,
				tagId: true,
				minRating: true,
				minDuration: true,
				maxDuration: true,
				dateFrom: true,
				dateTo: true,
				sort: true
			}
		}
	});
	return result.createSavedFilter as SavedFilter;
}

export async function deleteSavedFilter(filterId: string): Promise<void> {
	await gqlClient.mutation({ deleteSavedFilter: { __args: { filterId } } });
}
