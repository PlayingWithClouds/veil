import { gqlClient } from '$lib/veil';
import type { SceneStream } from '$lib/sceneDetail';

// One ranked alternate source for a scene, gathered live across every search
// plugin. Matching is performer/duration/poster driven, so a copy with a totally
// different title still surfaces.
export interface AlikeCandidate {
	externalId: string;
	title: string;
	plugin: string;
	sourceUrl: string;
	posterUrl: string | null;
	date: string | null;
	previewImages: string[];
	previewVideo: string | null;
	durationSeconds: number | null;
	// 0..1 match confidence; higher is more likely the same scene.
	matchScore: number;
}

// Stream fields requested when attaching a source, matching the detail page's
// stream shape.
const STREAM_FIELDS = {
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

// Ranked alternate sources for a scene. Fans out live across every search
// plugin, so this can take a few seconds.
export async function findAlikeSources(sceneId: string, limit = 12): Promise<AlikeCandidate[]> {
	try {
		const data = await gqlClient.query({
			findAlikeSources: {
				__args: { sceneId, limit },
				externalId: true,
				title: true,
				plugin: true,
				sourceUrl: true,
				posterUrl: true,
				date: true,
				previewImages: true,
				previewVideo: true,
				durationSeconds: true,
				matchScore: true
			}
		});
		return (data.findAlikeSources ?? []) as unknown as AlikeCandidate[];
	} catch (err) {
		console.error('findAlikeSources failed', sceneId, err);
		return [];
	}
}

// Scrape a chosen match and attach its playback source(s) to this scene. Returns
// the scene's ranked streams, including the newly attached ones.
export async function attachAlikeSource(
	sceneId: string,
	pluginName: string,
	url: string
): Promise<SceneStream[]> {
	const data = await gqlClient.mutation({
		attachAlikeSource: { __args: { sceneId, pluginName, url }, ...STREAM_FIELDS }
	});
	return (data.attachAlikeSource ?? []) as unknown as SceneStream[];
}
