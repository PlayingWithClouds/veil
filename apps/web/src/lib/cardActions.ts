// Card-level stream handling shared by the Explore grid: a per-scene stream
// cache (warmed on hover, reused by the detail page), plus the download flow
// that a card's download button and resolution dropdown drive.
import { ensureSceneStreams, queueDownload, type SceneStream } from '$lib/sceneDetail';
import { downloads } from '$lib/stores/downloads';

// sceneId -> resolved streams. Populated by the 1.4s hover prefetch so the
// download dropdown and the detail page open instantly.
const streamCache = new Map<string, SceneStream[]>();
// In-flight ensures, so overlapping hovers/clicks share one scrape.
const pending = new Map<string, Promise<SceneStream[]>>();

export function getCachedStreams(sceneId: string): SceneStream[] | null {
	return streamCache.get(sceneId) ?? null;
}

// Ensures a scene's streams exist, deduping concurrent calls and caching the
// result. Fresh (never-scraped) scenes get scraped here.
export async function ensureStreamsCached(sceneId: string): Promise<SceneStream[]> {
	const cached = streamCache.get(sceneId);
	if (cached) return cached;
	const inFlight = pending.get(sceneId);
	if (inFlight) return inFlight;

	const promise = ensureSceneStreams(sceneId)
		.then((streams) => {
			streamCache.set(sceneId, streams);
			return streams;
		})
		.finally(() => {
			pending.delete(sceneId);
		});
	pending.set(sceneId, promise);
	return promise;
}

// Human label for a stream's resolution, preferring an explicit height.
export function resolutionLabel(stream: SceneStream): string {
	if (stream.height) return `${stream.height}p`;
	if (stream.resolution) return stream.resolution;
	if (stream.label) return stream.label;
	if (stream.provider) return stream.provider;
	return 'Source';
}

// Queues a download for one stream and registers it with the progress store.
export async function downloadStream(
	sceneId: string,
	stream: SceneStream,
	title: string
): Promise<void> {
	const jobId = await queueDownload(stream.url, `${title} – ${resolutionLabel(stream)}`);
	downloads.start(jobId, sceneId);
}

// Downloads the default (best-ranked) resolution, scraping streams first if the
// scene has never been resolved.
export async function downloadDefault(sceneId: string, title: string): Promise<void> {
	const streams = await ensureStreamsCached(sceneId);
	const best = streams[0];
	if (!best) throw new Error('No downloadable source');
	await downloadStream(sceneId, best, title);
}
