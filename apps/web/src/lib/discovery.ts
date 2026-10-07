// Data helpers for scene browsing, the watchlist rail, and recording playback
// progress. These talk to the GraphQL API via the raw genql client.
import { gqlClient } from '$lib/veil';
import { cacheUrl } from '$lib/img';
import { sceneUrl } from '$lib/routes';
import type { CardItem } from '$lib/components/MediaCard.svelte';

export interface RailItem extends CardItem {
	mediaType: string;
	backdropUrl: string | null;
	// Watch fraction (0..1) for the continue-watching rail; undefined elsewhere.
	progress?: number;
}

function detailPath(item: { id: string }): string {
	return sceneUrl(item.id);
}

interface SceneCardData {
	id: string;
	title: string;
	posterPath: string | null;
	date: string | null;
}

function toRailItem(scene: SceneCardData): RailItem {
	return {
		id: scene.id,
		mediaType: 'scene',
		title: scene.title,
		posterUrl: cacheUrl(scene.posterPath),
		backdropUrl: null,
		subtitle: scene.date ?? undefined
	};
}

export interface SceneFilters {
	search?: string;
	studioId?: string;
	performerId?: string;
	tagId?: string;
	sort?: string;
}

export const SCENE_PAGE_SIZE = 40;

// One page of the scene browse grid.
export async function fetchScenes(filters: SceneFilters, page: number): Promise<RailItem[]> {
	const data = await gqlClient
		.query({
			scenes: {
				__args: {
					limit: SCENE_PAGE_SIZE,
					offset: page * SCENE_PAGE_SIZE,
					search: filters.search || undefined,
					studioId: filters.studioId || undefined,
					performerId: filters.performerId || undefined,
					tagId: filters.tagId || undefined,
					sort: filters.sort || undefined
				},
				id: true,
				title: true,
				posterPath: true,
				date: true
			}
		})
		.catch(() => ({ scenes: [] as SceneCardData[] }));
	return (data.scenes ?? []).map((scene) => toRailItem(scene as SceneCardData));
}

// Normalized cards for a set of scene ids (watch-history / watchlist rows carry
// only ids), resolved in one round trip.
export async function fetchMediaCards(ids: string[]): Promise<RailItem[]> {
	if (ids.length === 0) return [];
	const data = await gqlClient
		.query({
			mediaCards: {
				__args: { ids },
				mediaId: true,
				mediaType: true,
				title: true,
				posterPath: true,
				backdropPath: true
			}
		})
		.catch(() => ({ mediaCards: [] as never[] }));

	return (data.mediaCards ?? []).map((card) => ({
		id: card.mediaId,
		mediaType: card.mediaType,
		title: card.title,
		posterUrl: cacheUrl(card.posterPath),
		backdropUrl: cacheUrl(card.backdropPath)
	}));
}

export async function fetchWatchlist(): Promise<RailItem[]> {
	const items = await gqlClient
		.query({ watchlist: { media: true } })
		.then((d) => d.watchlist ?? [])
		.catch(() => []);
	return fetchMediaCards(items.map((i) => i.media));
}

export async function addToWatchlist(mediaId: string): Promise<void> {
	await gqlClient.mutation({
		addToWatchlist: { __args: { mediaId }, id: true }
	});
}

export async function removeFromWatchlist(mediaId: string): Promise<void> {
	await gqlClient.mutation({
		removeFromWatchlist: { __args: { mediaId } }
	});
}

// Records playback progress for resume. No-op without a media id.
export async function recordProgress(input: {
	mediaId: string;
	progressSeconds: number;
	durationSeconds?: number;
	completed?: boolean;
}): Promise<void> {
	await gqlClient
		.mutation({
			upsertWatchHistory: {
				__args: {
					input: {
						media: input.mediaId,
						progressSeconds: Math.floor(input.progressSeconds),
						durationSeconds: input.durationSeconds ? Math.floor(input.durationSeconds) : undefined,
						completed: input.completed
					}
				},
				id: true
			}
		})
		.catch(() => {
			// Progress tracking is best-effort; never interrupt playback.
		});
}

export { detailPath };
