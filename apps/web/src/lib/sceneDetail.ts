import { gqlClient, getWsClient } from '$lib/veil';
import { backendUrl, pluginIconUrl } from '$lib/server';

const STREAM_SELECT = {
	id: true,
	url: true,
	kind: true,
	label: true,
	provider: true,
	resolution: true,
	width: true,
	height: true,
	language: true,
	format: true,
	mimeType: true,
	expectedSpeedBps: true,
	fileSizeBytes: true,
	verified: true,
	pluginName: true
} as const;

export interface SceneStream {
	id: string;
	url: string;
	kind: string;
	label: string | null;
	provider: string | null;
	resolution: string | null;
	width: number | null;
	height: number | null;
	language: string | null;
	format: string | null;
	mimeType: string | null;
	expectedSpeedBps: number | null;
	fileSizeBytes: number | null;
	verified: boolean;
	pluginName: string | null;
}

export interface PerformerRef {
	id: string;
	name: string;
	imagePath: string | null;
}

export interface TagRef {
	id: string;
	name: string;
}

export interface StudioRef {
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
}

export interface SceneDetail {
	id: string;
	title: string;
	sourceUrl: string;
	details: string | null;
	date: string | null;
	durationSeconds: number | null;
	rating: number | null;
	viewCount: number;
	posterPath: string | null;
	previewVideo: string | null;
	previewImages: string[];
	studio: StudioRef | null;
	performers: PerformerRef[];
	tags: TagRef[];
}

export interface FullSceneDetail {
	scene: SceneDetail | null;
	streams: SceneStream[];
}

const EMPTY_DETAIL: FullSceneDetail = { scene: null, streams: [] };

// Loads a scene's detail plus its stored playback sources in one call.
export async function loadSceneDetail(id: string): Promise<FullSceneDetail> {
	try {
		const [sceneData, streams] = await Promise.all([
			gqlClient.query({
				scene: {
					__args: { id },
					id: true,
					title: true,
					sourceUrl: true,
					details: true,
					date: true,
					durationSeconds: true,
					rating: true,
					viewCount: true,
					posterPath: true,
					previewVideo: true,
					previewImages: true,
					studio: { id: true, name: true, imagePath: true, sceneCount: true },
					performers: { id: true, name: true, imagePath: true },
					tags: { id: true, name: true }
				}
			}),
			loadStreams(id)
		]);
		if (!sceneData.scene) return EMPTY_DETAIL;
		return { scene: sceneData.scene as unknown as SceneDetail, streams };
	} catch {
		return EMPTY_DETAIL;
	}
}

// Stored streams for a scene, typed for StreamList.
export async function loadStreams(mediaId: string): Promise<SceneStream[]> {
	try {
		const data = await gqlClient.query({
			mediaStreams: { __args: { mediaId }, ...STREAM_SELECT }
		});
		return (data.mediaStreams ?? []) as unknown as SceneStream[];
	} catch {
		return [];
	}
}

// Populates a stub scene's sources on demand (discovered-but-never-scraped
// scenes have no streams). Returns the freshly ingested, ranked streams.
export async function ensureSceneStreams(sceneId: string): Promise<SceneStream[]> {
	try {
		const data = await gqlClient.mutation({
			ensureSceneStreams: { __args: { sceneId }, ...STREAM_SELECT }
		});
		return (data.ensureSceneStreams ?? []) as unknown as SceneStream[];
	} catch (err) {
		console.error('ensureSceneStreams failed', sceneId, err);
		return [];
	}
}

const STREAMS_CHANGED_SUBSCRIPTION = `
	subscription StreamsChanged($mediaId: ID!) {
		streamsChanged(mediaId: $mediaId) {
			id
			url
			kind
			label
			provider
			resolution
			width
			height
			language
			format
			mimeType
			expectedSpeedBps
			fileSizeBytes
			verified
			pluginName
		}
	}
`;

// Live-update a scene's sources over WebSocket. Returns an unsubscribe function;
// a no-op during SSR.
export function subscribeStreams(
	mediaId: string,
	onStreams: (streams: SceneStream[]) => void
): () => void {
	const client = getWsClient();
	if (!client) return () => {};

	return client.subscribe(
		{ query: STREAMS_CHANGED_SUBSCRIPTION, variables: { mediaId } },
		{
			next(message) {
				const streams = message.data?.streamsChanged as SceneStream[] | undefined;
				if (streams) onStreams(streams);
			},
			error(err) {
				console.error('streamsChanged subscription error', err);
			},
			complete() {}
		}
	);
}

// Resolve a provider/source URL to a playable stream via the backend.
export async function resolveStream(url: string) {
	const data = await gqlClient.query({
		stream: {
			__args: { url },
			url: true,
			mimeType: true,
			quality: true,
			headers: { name: true, value: true }
		}
	});
	if (!data.stream) return data.stream;
	return { ...data.stream, url: backendUrl(data.stream.url) };
}

// Queue a download of a source URL to the NAS.
export async function queueDownload(url: string, title: string): Promise<string> {
	const data = await gqlClient.mutation({
		queueDownload: { __args: { url, title } }
	});
	return data.queueDownload;
}

/** A scene's sources from one site, qualities sorted best-first. */
export interface StreamGroup {
	provider: string;
	streams: SceneStream[];
}

/** Human label for a single source: its label, provider or resolution. */
export function streamLabel(stream: SceneStream): string {
	if (stream.label) return stream.label;
	if (stream.provider) return stream.provider;
	if (stream.resolution) return stream.resolution;
	return 'Source';
}

/** The site a stream came from, used to group qualities under one entry. */
export function providerName(stream: SceneStream): string {
	if (stream.pluginName) return stream.pluginName;
	if (stream.provider) return stream.provider;
	return 'Source';
}

/**
 * The per-quality label shown next to a provider's name. "Auto" when the site
 * doesn't say: the player picks whatever the page or playlist serves.
 */
export function qualityLabel(stream: SceneStream): string {
	if (stream.resolution) return stream.resolution;
	if (stream.height && stream.height > 0) return `${stream.height}p`;
	// A label that just repeats the site name says nothing next to it.
	if (stream.label && stream.label.toLowerCase() !== providerName(stream).toLowerCase()) return stream.label;
	if (stream.format) return stream.format.toUpperCase();
	return 'Auto';
}

/** Approximate vertical resolution, for ranking qualities best-first. */
export function streamHeight(stream: SceneStream): number {
	if (stream.height && stream.height > 0) return stream.height;
	const fromResolution = parseInt(stream.resolution ?? '', 10);
	if (Number.isFinite(fromResolution)) return fromResolution;
	return 0;
}

/** Groups streams by source site, each group's qualities sorted best-first. */
export function groupStreams(streams: SceneStream[]): StreamGroup[] {
	const groups = new Map<string, SceneStream[]>();
	for (const stream of streams) {
		const key = providerName(stream);
		const list = groups.get(key) ?? [];
		list.push(stream);
		groups.set(key, list);
	}
	for (const list of groups.values()) {
		list.sort((first, second) => streamHeight(second) - streamHeight(first));
	}
	return [...groups.entries()].map(([provider, list]) => ({ provider, streams: list }));
}

/** The site a scene was scraped from, with the matching plugin's icon when known. */
export interface SceneSite {
	name: string;
	iconUrl: string | null;
}

interface SitePlugin {
	name: string;
	displayName: string | null;
	iconUrl: string | null;
	domains: string[];
}

/** Hostname of a URL without a leading "www.", or null when it doesn't parse. */
function hostOf(url: string): string | null {
	try {
		return new URL(url).hostname.replace(/^www\./, '');
	} catch {
		return null;
	}
}

/** Whether a plugin serves a host: by a listed domain, or by its name in the host. */
function pluginServesHost(plugin: SitePlugin, host: string): boolean {
	const byDomain = plugin.domains.some(
		(domain) => domain !== '*' && (host === domain || host.endsWith(`.${domain}`))
	);
	if (byDomain) return true;
	return host.includes(plugin.name);
}

/** Resolves the site a scene's source URL belongs to (plugin name + icon, else the host). */
export async function fetchSceneSite(sourceUrl: string): Promise<SceneSite | null> {
	const host = hostOf(sourceUrl);
	if (!host) return null;
	try {
		const data = await gqlClient.query({
			plugins: { name: true, displayName: true, iconUrl: true, domains: true }
		});
		const plugins = (data.plugins ?? []) as unknown as SitePlugin[];
		const match = plugins.find((plugin) => pluginServesHost(plugin, host));
		if (!match) return { name: host, iconUrl: null };
		return { name: match.displayName || match.name, iconUrl: pluginIconUrl(match.iconUrl) };
	} catch {
		return { name: host, iconUrl: null };
	}
}
