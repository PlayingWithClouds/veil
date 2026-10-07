import { gqlClient, getWsClient } from '$lib/veil';
import { apiBase, pluginIconUrl } from '$lib/server';

// Guards against duplicate externalIds (multiple plugins, or a source listing
// the same item twice) which would otherwise break keyed {#each} rendering.
function dedupeByExternalId(items: PluginSearchResult[]): PluginSearchResult[] {
	const seen = new Set<string>();
	const unique: PluginSearchResult[] = [];
	for (const item of items) {
		if (seen.has(item.externalId)) continue;
		seen.add(item.externalId);
		unique.push(item);
	}
	return unique;
}

export interface PluginSearchResult {
	externalId: string;
	title: string;
	mediaType: string;
	plugin: string;
	sourceUrl: string;
	date: string | null;
	posterUrl: string | null;
	previewImages: string[];
	previewVideo: string | null;
	// Set for locally-stored scenes (DB browse feed): opening navigates straight
	// to the detail page by id instead of scraping the source first.
	sceneId?: string;
	// Canonical DB record id for the item, present for both DB and freshly-ingested
	// plugin results (galleries navigate by this).
	dbId?: string;
	// Runtime in seconds, known only for locally-stored scenes (null for fresh
	// plugin results that haven't been scraped yet).
	durationSeconds?: number | null;
	// Marks a feed entry that arrived since the user last looked (search
	// subscription feeds); rendered as a "New" badge on the card.
	isNew?: boolean;
	// Channel/studio, known only for locally-stored scenes whose page was visited.
	studioName?: string | null;
	studioImagePath?: string | null;
}

export interface SearchPlugin {
	name: string;
	displayName: string | null;
	iconUrl: string | null;
	capabilities: string[];
	enabled: boolean;
	// Hostnames the plugin serves, used to label cards with their site.
	domains: string[];
}

/** The plugin whose domains cover url's host, or null when none does. */
export function pluginForUrl(url: string, plugins: SearchPlugin[]): SearchPlugin | null {
	let host: string;
	try {
		host = new URL(url).hostname.replace(/^www\./, '');
	} catch {
		return null;
	}
	for (const plugin of plugins) {
		const matches = plugin.domains.some((domain) => host === domain || host.endsWith(`.${domain}`));
		if (matches) return plugin;
	}
	return null;
}

// Maps plugin name to its icon URL, for showing the source provider on result
// banners.
export async function fetchPluginIcons(): Promise<Record<string, string | null>> {
	try {
		const data = await gqlClient.query({
			plugins: {
				name: true,
				iconUrl: true
			}
		});
		const all = (data.plugins ?? []) as unknown as { name: string; iconUrl: string | null }[];
		const icons: Record<string, string | null> = {};
		for (const plugin of all) {
			icons[plugin.name] = pluginIconUrl(plugin.iconUrl);
		}
		return icons;
	} catch {
		return {};
	}
}

// Installed plugins that can keyword-search scenes (the `scene:list` capability
// serves keyword search when given a query), for the search page's source badges.
/** Whether the plugin lists scenes or galleries (i.e. is a searchable site, not a resolver). */
function listsContent(plugin: SearchPlugin): boolean {
	return plugin.capabilities.includes('scene:list') || plugin.capabilities.includes('gallery:list');
}

export async function fetchSearchPlugins(): Promise<SearchPlugin[]> {
	try {
		const data = await gqlClient.query({
			plugins: {
				name: true,
				displayName: true,
				iconUrl: true,
				capabilities: true,
				enabled: true,
				available: true,
				domains: true
			}
		});
		const all = (data.plugins ?? []) as unknown as (SearchPlugin & { available: boolean })[];
		for (const plugin of all) {
			plugin.iconUrl = pluginIconUrl(plugin.iconUrl);
		}
		// Plugins missing a requirement (FlareSolverr) can't search in this deployment.
		return all.filter((plugin) => plugin.available && listsContent(plugin));
	} catch {
		return [];
	}
}

// Plugin-direct search. Results come straight from source plugins (banner image
// + preview frames). Selection resolves to a scene via ingestPluginResult. An
// optional pluginNames list restricts the search to those sources (empty = all).
export async function pluginSearch(
	query: string,
	limit = 20,
	pluginNames: string[] = []
): Promise<PluginSearchResult[]> {
	if (!query.trim()) return [];
	try {
		const data = await gqlClient.query({
			pluginSearch: {
				__args: { query, limit, pluginNames },
				externalId: true,
				title: true,
				mediaType: true,
				plugin: true,
				sourceUrl: true,
				date: true,
				posterUrl: true,
				previewImages: true,
				previewVideo: true
			}
		});
		return dedupeByExternalId((data.pluginSearch ?? []) as unknown as PluginSearchResult[]);
	} catch {
		return [];
	}
}

export interface SearchStream {
	close(): void;
}

// Raw shape of a `result` SSE event's `item` (mirrors plugins.DiscoveredItem).
interface StreamItem {
	title: string;
	media_type: string;
	source_url: string;
	external_id: string;
	date?: string;
	poster_path?: string;
	preview_images?: string[];
	preview_video?: string;
}

interface StreamResultEvent {
	source: string;
	id?: string;
	item: StreamItem;
}

function streamEventToResult(event: StreamResultEvent): PluginSearchResult {
	const isDbHit = event.source === 'db';
	return {
		externalId: event.item.external_id,
		title: event.item.title,
		mediaType: event.item.media_type,
		plugin: isDbHit ? '' : event.source,
		sourceUrl: event.item.source_url,
		date: event.item.date || null,
		posterUrl: event.item.poster_path || null,
		previewImages: event.item.preview_images ?? [],
		previewVideo: event.item.preview_video || null,
		// DB hits are already-ingested scenes: open by id, no scrape needed.
		sceneId: isDbHit ? event.id : undefined,
		dbId: event.id
	};
}

// Streaming plugin search over the /api/search SSE endpoint. Local DB matches
// arrive first, then each source plugin's results as it finishes — so the grid
// fills incrementally instead of waiting for the slowest provider. An empty
// sources list searches every enabled plugin.
//
// onResult fires per item (caller dedupes); onDone fires when all sources are
// exhausted. Returns a handle whose close() aborts the stream.
export function pluginSearchStream(
	query: string,
	sources: string[],
	onResult: (result: PluginSearchResult) => void,
	onDone: () => void,
	options?: { offset?: number; limit?: number }
): SearchStream {
	const url = new URL(`${apiBase()}/api/search`);
	url.searchParams.set('q', query);
	if (sources.length > 0) {
		url.searchParams.set('sources', sources.join(','));
	}
	// An explicit offset makes the backend paginate the plugin listings — used
	// by infinite-scroll feeds (e.g. galleries).
	if (options?.offset !== undefined) {
		url.searchParams.set('offset', String(options.offset));
	}
	if (options?.limit !== undefined) {
		url.searchParams.set('limit', String(options.limit));
	}

	const source = new EventSource(url.toString());

	source.addEventListener('result', (event) => {
		try {
			const parsed = JSON.parse((event as MessageEvent).data) as StreamResultEvent;
			onResult(streamEventToResult(parsed));
		} catch {
			// malformed event — ignore
		}
	});

	source.addEventListener('done', () => {
		source.close();
		onDone();
	});

	// The server closes the stream after `done`; a raw error on an open stream
	// (network drop) also ends the search rather than hanging the spinner.
	source.addEventListener('error', () => {
		source.close();
		onDone();
	});

	return {
		close(): void {
			source.close();
		}
	};
}

// Latest items from source plugins (empty-query browse), for the landing feed.
export async function pluginBrowse(limit = 40, offset = 0): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			pluginBrowse: {
				__args: { limit, offset },
				externalId: true,
				title: true,
				mediaType: true,
				plugin: true,
				sourceUrl: true,
				date: true,
				posterUrl: true,
				previewImages: true,
				previewVideo: true
			}
		});
		return dedupeByExternalId((data.pluginBrowse ?? []) as unknown as PluginSearchResult[]);
	} catch {
		return [];
	}
}

// Taste-biased live fetch: the backend runs a category-driven and a main-page
// discover per provider, interleaved by relevance. Plugin results (not scenes),
// so selection ingests on open.
export async function fetchRecommendedBrowse(limit = 30): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			recommendedBrowse: {
				__args: { limit },
				externalId: true,
				title: true,
				mediaType: true,
				plugin: true,
				sourceUrl: true,
				date: true,
				posterUrl: true,
				previewImages: true,
				previewVideo: true
			}
		});
		return dedupeByExternalId((data.recommendedBrowse ?? []) as unknown as PluginSearchResult[]);
	} catch {
		return [];
	}
}

export interface BrowseSceneData {
	id: string;
	externalId: string;
	sourceUrl: string;
	title: string;
	date: string | null;
	durationSeconds: number | null;
	posterPath: string | null;
	previewVideo: string | null;
	previewImages: string[];
	studio?: { name: string; imagePath: string | null } | null;
}

// Adapts a locally-stored scene to the browse feed's result shape. `sceneId`
// marks it as already-ingested so selection navigates by id.
export function browseSceneToResult(scene: BrowseSceneData): PluginSearchResult {
	return {
		studioName: scene.studio?.name ?? null,
		studioImagePath: scene.studio?.imagePath ?? null,
		externalId: scene.externalId,
		title: scene.title,
		mediaType: 'scene',
		plugin: '',
		sourceUrl: scene.sourceUrl,
		date: scene.date,
		durationSeconds: scene.durationSeconds,
		posterUrl: scene.posterPath,
		previewImages: scene.previewImages ?? [],
		previewVideo: scene.previewVideo,
		sceneId: scene.id
	};
}

const RELATED_CHANGED_SUBSCRIPTION = `
	subscription RelatedChanged($sceneId: ID!) {
		relatedChanged(sceneId: $sceneId) {
			id
			externalId
			sourceUrl
			title
			date
			durationSeconds
			posterPath
			previewVideo
			previewImages
			studio {
				name
				imagePath
			}
		}
	}
`;

/**
 * Streams a scene's related scenes: the current list on connect, then again
 * whenever a visit links more (the site's related list, then fallback searches).
 * Returns the unsubscribe function.
 */
export function subscribeRelated(
	sceneId: string,
	onRelated: (items: PluginSearchResult[]) => void
): () => void {
	const client = getWsClient();
	if (!client) return () => {};
	return client.subscribe(
		{ query: RELATED_CHANGED_SUBSCRIPTION, variables: { sceneId } },
		{
			next(message) {
				const scenes = message.data?.relatedChanged as BrowseSceneData[] | undefined;
				if (scenes) onRelated(scenes.map(browseSceneToResult));
			},
			error(err) {
				console.error('relatedChanged subscription error', err);
			},
			complete() {}
		}
	);
}

// Locally-stored scenes (newest first), for the DB-backed browse feed that fills
// in before falling back to live plugin fetches.
export async function fetchBrowseScenes(limit = 25, offset = 0): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			scenes: { __args: { limit, offset }, ...SCENE_CARD_FIELDS }
		});
		const scenes = (data.scenes ?? []) as unknown as BrowseSceneData[];
		return scenes.map(browseSceneToResult);
	} catch {
		return [];
	}
}

// Personalized feed: scenes ranked by the backend against the taste signal,
// already-watched media excluded. Same result shape as
// the DB browse feed, so results open by scene id.
export async function fetchRecommended(limit = 25, offset = 0): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			recommendedFeed: {
				__args: { limit, offset },
				id: true,
				externalId: true,
				sourceUrl: true,
				title: true,
				date: true,
				durationSeconds: true,
				posterPath: true,
				previewVideo: true,
				previewImages: true,
				studio: { name: true, imagePath: true }
			}
		});
		const scenes = (data.recommendedFeed ?? []) as unknown as BrowseSceneData[];
		return scenes.map(browseSceneToResult);
	} catch {
		return [];
	}
}

// One recommendation category: a tag and its recommended scenes.
export interface RecommendedCategory {
	tagId: string;
	tagName: string;
	items: PluginSearchResult[];
}

export const SCENE_CARD_FIELDS = {
	id: true,
	externalId: true,
	sourceUrl: true,
	title: true,
	date: true,
	durationSeconds: true,
	posterPath: true,
	previewVideo: true,
	previewImages: true,
	studio: { name: true, imagePath: true }
} as const;

// Recommended scenes grouped into categories, most-watched category first. Empty
// when there is no watch signal yet (caller falls back to the flat feed).
export async function fetchRecommendedCategories(
	categoryLimit = 8,
	perCategory = 12
): Promise<RecommendedCategory[]> {
	try {
		const data = await gqlClient.query({
			recommendedCategories: {
				__args: { categoryLimit, perCategory },
				tag: { id: true, name: true },
				scenes: SCENE_CARD_FIELDS
			}
		});
		const categories = (data.recommendedCategories ?? []) as unknown as {
			tag: { id: string; name: string };
			scenes: BrowseSceneData[];
		}[];
		return categories.map((category) => ({
			tagId: category.tag.id,
			tagName: category.tag.name,
			items: category.scenes.map(browseSceneToResult)
		}));
	} catch {
		return [];
	}
}

// Two taste-ranked scenes to show side by side. Returns 0 or 2 items.
export async function fetchRecommendationPair(): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			recommendationPair: SCENE_CARD_FIELDS
		});
		const scenes = (data.recommendationPair ?? []) as unknown as BrowseSceneData[];
		return scenes.map(browseSceneToResult);
	} catch {
		return [];
	}
}

// Records that the user picked `chosenId` over `rejectedId` in an A/B pair.
export async function recordRecommendationChoice(
	chosenId: string,
	rejectedId: string
): Promise<void> {
	try {
		await gqlClient.mutation({
			recordRecommendationChoice: {
				__args: { chosenMediaId: chosenId, rejectedMediaId: rejectedId }
			}
		});
	} catch {
		// Best-effort signal; ignore failures.
	}
}

// The watch history as scene cards, most recently watched first. Resume
// bars come from the shared watchProgress store, so only scene detail is fetched.
export async function fetchWatchedScenes(limit = 100): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			watchHistory: {
				__args: { limit },
				scene: SCENE_CARD_FIELDS
			}
		});
		const rows = (data.watchHistory ?? []) as unknown as { scene: BrowseSceneData | null }[];
		return rows
			.map((row) => row.scene)
			.filter((scene): scene is BrowseSceneData => scene !== null)
			.map(browseSceneToResult);
	} catch {
		return [];
	}
}

// Scenes downloaded to the NAS (completed download jobs), for the library.
export async function fetchDownloadedScenes(): Promise<PluginSearchResult[]> {
	try {
		const data = await gqlClient.query({
			downloadedScenes: SCENE_CARD_FIELDS
		});
		const scenes = (data.downloadedScenes ?? []) as unknown as BrowseSceneData[];
		return scenes.map(browseSceneToResult);
	} catch {
		return [];
	}
}

// Synchronously scrapes a selected plugin result's source URL and ingests the
// full scene, returning the canonical scene id to navigate to.
export async function ingestPluginResult(
	pluginName: string,
	sourceUrl: string,
	posterUrl?: string | null
): Promise<string> {
	const data = await gqlClient.mutation({
		resolvePluginResult: {
			__args: { pluginName, url: sourceUrl, posterUrl: posterUrl ?? undefined }
		}
	});
	const mediaId = data.resolvePluginResult;
	if (!mediaId) throw new Error('Scrape produced no media record');
	return mediaId;
}
