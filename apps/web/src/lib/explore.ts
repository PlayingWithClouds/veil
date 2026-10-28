// Data layer for Home: the recommendation feed (scenes, galleries, titled
// rows, Continue watching), the card shapes, and the name searches beside the
// live search.
import { gqlClient } from '$lib/veil';
import { cacheUrl } from '$lib/img';
import {
	pluginSearchStream,
	type PluginSearchResult,
	type SearchStream
} from '$lib/search';

export type ExploreScope = 'all' | 'scenes' | 'galleries' | 'performers' | 'studios';

export interface PerformerHit {
	id: string;
	name: string;
	imagePath: string | null;
	birthdate: string | null;
}

export interface StudioHit {
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
}

export interface GalleryHit {
	id: string;
	title: string;
	coverUrl: string | null;
	imageCount: number;
	tags: string[];
	// Page the gallery came from; its host names the site.
	sourceUrl: string;
}

export interface SceneHit {
	id: string;
	title: string;
	bannerUrl: string | null;
	previewImages: string[];
	previewVideo: string | null;
	durationSeconds: number | null;
	tags: string[];
	// Page the scene came from; its host names the site.
	sourceUrl: string;
	// Channel/studio and credited performers; empty for stubs never visited.
	studioName: string | null;
	studioImageUrl: string | null;
	performerNames: string[];
	performerImageUrl: string | null;
	// Release date, else when it was stored.
	date: string | null;
	createdAt: string | null;
	// The site's own view count and rating (0..10), when it gives them.
	viewCount?: number | null;
	rating?: number | null;
	// Why the recommendation engine picked it, e.g. "More from <performer>".
	reason?: string | null;
}

// A single item of Home's feed: either a gallery or a scene, tagged by kind.
export type GridItem =
	| { kind: 'scene'; scene: SceneHit }
	| { kind: 'gallery'; gallery: GalleryHit };

export async function searchPerformers(query: string, limit = 20): Promise<PerformerHit[]> {
	if (!query.trim()) return [];
	try {
		const data = await gqlClient.query({
			performers: {
				__args: { search: query, limit },
				id: true,
				name: true,
				imagePath: true,
				birthdate: true
			}
		});
		return (data.performers ?? []) as unknown as PerformerHit[];
	} catch {
		return [];
	}
}

export async function searchStudios(query: string, limit = 20): Promise<StudioHit[]> {
	if (!query.trim()) return [];
	try {
		const data = await gqlClient.query({
			studios: {
				__args: { search: query, limit },
				id: true,
				name: true,
				imagePath: true,
				sceneCount: true
			}
		});
		return (data.studios ?? []) as unknown as StudioHit[];
	} catch {
		return [];
	}
}

interface GalleryRow {
	id: string;
	title: string;
	sourceUrl: string;
	coverPath: string | null;
	imageCount: number;
	tags: { name: string }[] | null;
}

function galleryRowToHit(row: GalleryRow): GalleryHit {
	return {
		id: row.id,
		title: row.title,
		coverUrl: cacheUrl(row.coverPath),
		imageCount: row.imageCount,
		tags: (row.tags ?? []).map((tag) => tag.name),
		sourceUrl: row.sourceUrl
	};
}

/** The sources argument for a DB listing: omitted when every site is selected. */
function sourcesArgument(sources: string[]): string[] | undefined {
	if (sources.length === 0) return undefined;
	return sources;
}

/** DB gallery search, restricted to galleries found through the given sites (empty = all). */
export async function searchGalleries(
	query: string,
	sources: string[] = [],
	limit = 30,
	offset = 0
): Promise<GalleryHit[]> {
	try {
		const data = await gqlClient.query({
			galleries: {
				__args: { search: query || undefined, sources: sourcesArgument(sources), limit, offset },
				id: true,
				title: true,
				sourceUrl: true,
				coverPath: true,
				imageCount: true,
				tags: { name: true }
			}
		});
		const rows = (data.galleries ?? []) as unknown as GalleryRow[];
		return rows.map(galleryRowToHit);
	} catch {
		return [];
	}
}

interface SceneRow {
	id: string;
	title: string;
	sourceUrl: string;
	posterPath: string | null;
	previewVideo: string | null;
	previewImages: string[] | null;
	durationSeconds: number | null;
	date: string | null;
	createdAt: string | null;
	tags: { name: string }[] | null;
	studio: { name: string; imagePath: string | null } | null;
	performers: { name: string; imagePath: string | null }[] | null;
	rating: number | null;
	viewCount: number | null;
}

function sceneRowToHit(row: SceneRow): SceneHit {
	const performers = row.performers ?? [];
	return {
		id: row.id,
		title: row.title,
		bannerUrl: cacheUrl(row.posterPath),
		previewImages: row.previewImages ?? [],
		previewVideo: row.previewVideo,
		durationSeconds: row.durationSeconds,
		tags: (row.tags ?? []).map((tag) => tag.name),
		sourceUrl: row.sourceUrl,
		studioName: row.studio?.name ?? null,
		studioImageUrl: cacheUrl(row.studio?.imagePath),
		performerNames: performers.map((performer) => performer.name),
		performerImageUrl: firstPerformerImage(performers),
		date: row.date,
		createdAt: row.createdAt,
		rating: row.rating,
		viewCount: row.viewCount
	};
}

/** The first credited performer's image, for the card avatar. */
function firstPerformerImage(performers: { imagePath: string | null }[]): string | null {
	for (const performer of performers) {
		if (performer.imagePath) return cacheUrl(performer.imagePath);
	}
	return null;
}

/** Caption byline: the channel/studio, else up to two performers. */
export function sceneByline(scene: SceneHit): string | null {
	if (scene.studioName) return scene.studioName;
	if (scene.performerNames.length === 0) return null;
	return scene.performerNames.slice(0, 2).join(', ');
}

// Scene fields every Explore card needs.
const SCENE_HIT_FIELDS = {
	id: true,
	title: true,
	sourceUrl: true,
	posterPath: true,
	previewVideo: true,
	previewImages: true,
	durationSeconds: true,
	date: true,
	createdAt: true,
	tags: { name: true },
	studio: { name: true, imagePath: true },
	performers: { name: true, imagePath: true },
	rating: true,
	viewCount: true
} as const;

/** A recommended scene with the candidate source the engine served it from. */
export interface RecommendedSceneHit {
	scene: SceneHit;
	source: string;
}

/** Why the engine recommends something: its kind, display text and what it points at. */
export interface RecommendationReasonRef {
	kind: string;
	text: string;
	entityId: string | null;
}

/** One titled shelf of recommendations, e.g. "More from <performer>". */
export interface RecommendationShelf {
	key: string;
	title: string;
	reason: RecommendationReasonRef;
	items: RecommendedSceneHit[];
}

/** The engine's titled rows ("Because you watched …", "More from …"), strongest first. */
export async function fetchRecommendationShelves(
	rowLimit: number,
	perRow: number
): Promise<RecommendationShelf[]> {
	try {
		const data = await gqlClient.query({
			recommendedRows: {
				__args: { rowLimit, perRow },
				key: true,
				title: true,
				reason: { kind: true, text: true, entityId: true },
				items: { scene: SCENE_HIT_FIELDS, source: true }
			}
		});
		const rows = (data.recommendedRows ?? []) as unknown as {
			key: string;
			title: string;
			reason: RecommendationReasonRef;
			items: { scene: SceneRow; source: string }[];
		}[];
		return rows.map((row) => ({
			key: row.key,
			title: row.title,
			reason: row.reason,
			items: row.items.map((item) => ({ scene: sceneRowToHit(item.scene), source: item.source }))
		}));
	} catch {
		return [];
	}
}

/**
 * Scenes started and not finished, most recent first: at least
 * CONTINUE_MIN_SECONDS played and more than CONTINUE_END_SECONDS left.
 */
export async function fetchContinueWatching(limit: number): Promise<SceneHit[]> {
	const CONTINUE_MIN_SECONDS = 30;
	const CONTINUE_END_SECONDS = 60;
	try {
		const data = await gqlClient.query({
			watchHistory: {
				__args: { limit: limit * 3 },
				progressSeconds: true,
				durationSeconds: true,
				completed: true,
				scene: SCENE_HIT_FIELDS
			}
		});
		const rows = (data.watchHistory ?? []) as unknown as {
			progressSeconds: number;
			durationSeconds: number | null;
			completed: boolean;
			scene: SceneRow | null;
		}[];
		const unfinished = rows.filter((row) => {
			if (row.completed || !row.scene || row.progressSeconds < CONTINUE_MIN_SECONDS) return false;
			const duration = row.durationSeconds || row.scene.durationSeconds;
			if (!duration) return true;
			return duration - row.progressSeconds > CONTINUE_END_SECONDS;
		});
		return unfinished.slice(0, limit).map((row) => sceneRowToHit(row.scene as SceneRow));
	} catch {
		return [];
	}
}

/**
 * One page of the ranked recommendation feed, for Home. Offset 0 keeps the
 * ranking for 10 minutes (so a reload shows the same feed) unless refresh is
 * set; later offsets continue that ranking so pages don't overlap. Non-empty
 * sources page through those sites' share of the ranking; when its first
 * page comes up short the backend pulls fresh scenes from them first.
 */
export async function fetchRecommendedSceneHits(
	limit: number,
	offset: number,
	refresh = false,
	sources: string[] = []
): Promise<RecommendedSceneHit[]> {
	try {
		const data = await gqlClient.query({
			recommendations: {
				__args: { limit, offset, refresh, sources },
				scene: SCENE_HIT_FIELDS,
				source: true,
				reason: { text: true }
			}
		});
		const rows = (data.recommendations ?? []) as unknown as {
			scene: SceneRow;
			source: string;
			reason: { text: string };
		}[];
		return rows.map((row) => ({
			scene: { ...sceneRowToHit(row.scene), reason: row.reason.text },
			source: row.source
		}));
	} catch {
		return [];
	}
}

/** A random pick of stored scenes (blocklist applied), for random mode. */
export async function fetchRandomSceneHits(count: number): Promise<SceneHit[]> {
	try {
		const data = await gqlClient.query({
			randomScenes: { __args: { limit: count }, ...SCENE_HIT_FIELDS }
		});
		const rows = (data.randomScenes ?? []) as unknown as SceneRow[];
		return rows.map(sceneRowToHit);
	} catch {
		return [];
	}
}

/** Which entity a performer/studio page lists content for. */
export type EntityFilter = { performerId: string } | { studioId: string };

/** The listing filter for a performer or studio. */
export function entityFilter(kind: 'performer' | 'studio', id: string): EntityFilter {
	if (kind === 'performer') return { performerId: id };
	return { studioId: id };
}

// Library scenes credited to a performer or studio, newest release first.
export async function fetchEntitySceneHits(
	filter: EntityFilter,
	limit = 40,
	offset = 0
): Promise<SceneHit[]> {
	try {
		const data = await gqlClient.query({
			scenes: {
				__args: { ...filter, sort: 'date', limit, offset },
				...SCENE_HIT_FIELDS
			}
		});
		const rows = (data.scenes ?? []) as unknown as SceneRow[];
		return rows.map(sceneRowToHit);
	} catch {
		return [];
	}
}

// Galleries credited to a performer or studio.
export async function fetchEntityGalleries(
	filter: EntityFilter,
	limit = 30,
	offset = 0
): Promise<GalleryHit[]> {
	try {
		const data = await gqlClient.query({
			galleries: {
				__args: { ...filter, limit, offset },
				id: true,
				title: true,
				sourceUrl: true,
				coverPath: true,
				imageCount: true,
				tags: { name: true }
			}
		});
		const rows = (data.galleries ?? []) as unknown as GalleryRow[];
		return rows.map(galleryRowToHit);
	} catch {
		return [];
	}
}

// The preferred tag names, most-watched first. Derived from the recommendation
// categories, which are ordered by how much has been watched from each tag — so
// the head of this list is the strongest taste.
export async function fetchPreferredTags(limit = 40): Promise<string[]> {
	try {
		const data = await gqlClient.query({
			recommendedCategories: {
				__args: { categoryLimit: limit, perCategory: 1 },
				tag: { name: true }
			}
		});
		const categories = (data.recommendedCategories ?? []) as unknown as {
			tag: { name: string };
		}[];
		return categories.map((category) => category.tag.name);
	} catch {
		return [];
	}
}

// The single tag on this card that best matches the user's taste: the card
// tag with the smallest preference rank. Null when nothing overlaps.
export function bestPreferredTag(cardTags: string[], preferredOrder: string[]): string | null {
	let bestTag: string | null = null;
    let bestRank = Number.MAX_SAFE_INTEGER;
	for (const tag of cardTags) {
		const rank = preferredOrder.indexOf(tag);
		if (rank === -1) continue;
		if (rank >= bestRank) continue;
		bestRank = rank;
		bestTag = tag;
	}
	return bestTag;
}

// Adapts a browse/recommended plugin result (no tags carried) to a scene hit.
function pluginResultToSceneHit(result: PluginSearchResult): SceneHit {
	return {
		id: result.sceneId ?? result.dbId ?? result.externalId,
		title: result.title,
		bannerUrl: cacheUrl(result.posterUrl),
		previewImages: result.previewImages,
		previewVideo: result.previewVideo,
		durationSeconds: result.durationSeconds ?? null,
		tags: [],
		sourceUrl: result.sourceUrl,
		studioName: null,
		studioImageUrl: null,
		performerNames: [],
		performerImageUrl: null,
		date: result.date,
		createdAt: null,
		viewCount: result.viewCount,
		rating: result.rating
	};
}

/** Adapts an up-next feed item to Home's scene card, keeping its channel. */
export function feedItemToSceneHit(item: PluginSearchResult): SceneHit {
	const hit = pluginResultToSceneHit(item);
	if (item.studioName) hit.studioName = item.studioName;
	hit.studioImageUrl = cacheUrl(item.studioImagePath);
	return hit;
}

/**
 * Turns a live plugin search result into a grid item. Null for results that
 * weren't stored (no DB id) or aren't scenes/galleries.
 */
export function streamResultToGridItem(result: PluginSearchResult): GridItem | null {
	if (!result.dbId) return null;
	if (result.mediaType === 'scene') {
		return { kind: 'scene', scene: { ...pluginResultToSceneHit(result), id: result.dbId } };
	}
	if (result.mediaType !== 'gallery') return null;
	return { kind: 'gallery', gallery: pluginResultToGalleryHit(result, result.dbId) };
}

/** Adapts a plugin result for a gallery, stored under the record id given, to a gallery card. */
export function pluginResultToGalleryHit(result: PluginSearchResult, id: string): GalleryHit {
	return {
		id,
		title: result.title,
		coverUrl: cacheUrl(result.posterUrl),
		imageCount: result.photoCount || 0,
		tags: [],
		sourceUrl: result.sourceUrl
	};
}

/** The site a page URL belongs to (its host without "www."), or empty for an unparsable URL. */
function siteKeyOfUrl(sourceUrl: string): string {
	try {
		return new URL(sourceUrl).hostname.replace(/^www\./, '');
	} catch {
		return '';
	}
}

/** The site a grid item came from. */
export function gridItemSiteKey(item: GridItem): string {
	if (item.kind === 'scene') return siteKeyOfUrl(item.scene.sourceUrl);
	return siteKeyOfUrl(item.gallery.sourceUrl);
}

/**
 * Reorders items round-robin across sites, so a site that answered with thirty
 * results in a row doesn't fill the page before the others show up. The order
 * within each site is kept.
 */
export function mixBySite<T>(items: T[], siteKeyOf: (item: T) => string): T[] {
	const queues = new Map<string, T[]>();
	for (const item of items) {
		const siteKey = siteKeyOf(item);
		const queue = queues.get(siteKey);
		if (queue) queue.push(item);
		else queues.set(siteKey, [item]);
	}
	const mixed: T[] = [];
	const rounds = Math.max(0, ...[...queues.values()].map((queue) => queue.length));
	for (let round = 0; round < rounds; round++) {
		for (const queue of queues.values()) {
			if (round < queue.length) mixed.push(queue[round]);
		}
	}
	return mixed;
}

/** Stable key for a grid item, unique across scenes and galleries. */
export function gridItemKey(item: GridItem): string {
	if (item.kind === 'scene') return `scene:${item.scene.id}`;
	return `gallery:${item.gallery.id}`;
}

export type { PluginSearchResult, SearchStream };
export { pluginSearchStream };
