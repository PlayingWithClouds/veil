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

// Fields of a ranked match.
const CANDIDATE_FIELDS = {
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
} as const;

// Ranked alternate sources for a scene. Fans out live across every search
// plugin, so this can take a few seconds.
export async function findAlikeSources(sceneId: string, limit = 12): Promise<AlikeCandidate[]> {
	try {
		const data = await gqlClient.query({
			findAlikeSources: { __args: { sceneId, limit }, ...CANDIDATE_FIELDS }
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

/** Videos a gallery's photo set may come from, ranked live across every site. */
export async function findGalleryVideos(galleryId: string, limit = 12): Promise<AlikeCandidate[]> {
	try {
		const data = await gqlClient.query({
			findGalleryVideos: { __args: { galleryId, limit }, ...CANDIDATE_FIELDS }
		});
		return (data.findGalleryVideos ?? []) as unknown as AlikeCandidate[];
	} catch (err) {
		console.error('findGalleryVideos failed', galleryId, err);
		return [];
	}
}

/** Stores a picked match as the gallery's video and returns that scene's id. */
export async function linkGalleryVideo(
	galleryId: string,
	pluginName: string,
	url: string
): Promise<string> {
	const data = await gqlClient.mutation({
		linkGalleryVideo: { __args: { galleryId, pluginName, url }, id: true }
	});
	return data.linkGalleryVideo.id;
}
