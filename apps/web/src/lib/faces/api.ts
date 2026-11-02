/**
 * The backend side of face scans: what a scan needs to know about a video
 * or performer, recording scans, fingerprints and trickplay, and reading the
 * guesses of who is in a video.
 */
import type { FieldsSelection, VideoFaces } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import type { ScannedFace } from './plan';

/** What scanning a video needs: its length, credits, streams, and what it has already. */
const SCAN_VIDEO = {
	id: true,
	durationSeconds: true,
	performers: { id: true },
	trickplay: { intervalSeconds: true },
	streams: { url: true, mimeType: true, quality: true },
	fingerprints: { algorithm: true }
} as const;

/** A performer as a guess names them. */
const GUESSED_PERFORMER = { id: true, name: true, image: { uri: true } } as const;

/** Who is probably in a video, as the video page shows it. */
const VIDEO_FACES = {
	scan: { id: true, model: true, frames: true, faces: true, createdAt: true },
	people: {
		sightings: { embeddingId: true, atSeconds: true, region: true },
		candidates: {
			performer: GUESSED_PERFORMER,
			similarity: true,
			likely: true,
			credited: true,
			picture: true,
			video: { id: true, title: true }
		}
	},
	guesses: { performer: GUESSED_PERFORMER, similarity: true, credited: true, person: true }
} as const;

/** Who is probably in a video: people, their candidates, and the guesses. */
export type VideoFaceGuesses = FieldsSelection<VideoFaces, typeof VIDEO_FACES>;

/** A video as a scan reads it; null when it's gone. */
export async function fetchScanVideo(videoId: string) {
	const data = await gqlClient.query({ video: { __args: { id: videoId }, ...SCAN_VIDEO } });
	return data.video ?? null;
}

/** Whether the record has a face scan with the model. */
export async function hasFaceScan(
	kind: 'video' | 'performer',
	id: string,
	model: string
): Promise<boolean> {
	const data = await gqlClient.query({
		faceScans: { __args: { subject: { kind, id } }, model: true }
	});
	return data.faceScans.some((scan) => scan.model === model);
}

/** A performer's pictures and whether each is scanned with the model. */
export async function fetchFacePictures(
	performerId: string,
	model: string
): Promise<{ uri: string; scanned: boolean }[]> {
	const data = await gqlClient.query({
		facePictures: { __args: { performerId, model }, uri: true, scanned: true }
	});
	return data.facePictures;
}

/** What a client found: a video's frames, or one picture of a performer. */
export interface FaceScanRecord {
	kind: 'video' | 'performer';
	id: string;
	model: string;
	picture?: string;
	frames: number;
	faces: ScannedFace[];
}

/** Records a scan, replacing the record's earlier one with the model; the backend learns from it. */
export async function recordFaceScan(scan: FaceScanRecord): Promise<void> {
	await gqlClient.mutation({
		recordFaceScan: {
			__args: {
				input: {
					subject: { kind: scan.kind, id: scan.id },
					model: scan.model,
					picture: scan.picture,
					frames: scan.frames,
					faces: scan.faces
				}
			},
			id: true
		}
	});
}

/** Learns a scanned video's credited performers' faces again; returns how many were learned. */
export async function learnFaces(videoId: string, model: string): Promise<number> {
	const data = await gqlClient.mutation({ learnFaces: { __args: { videoId, model } } });
	return data.learnFaces;
}

/** Stores the video's perceptual hash (16 hex digits). */
export async function addPhash(
	videoId: string,
	value: string,
	durationSeconds: number
): Promise<void> {
	await gqlClient.mutation({
		addFingerprint: {
			__args: {
				subject: { kind: 'video', id: videoId },
				algorithm: 'phash',
				value,
				durationSeconds
			},
			id: true
		}
	});
}

/** A trickplay sheet already uploaded. */
export interface TrickplaySheet {
	uri: string;
	columns: number;
	rows: number;
	tileWidth: number;
	tileHeight: number;
	intervalSeconds: number;
}

/** Sets the video's scrubbing thumbnails to one sheet. */
export async function setTrickplay(videoId: string, sheet: TrickplaySheet): Promise<void> {
	const { uri, ...grid } = sheet;
	await gqlClient.mutation({
		setTrickplay: { __args: { videoId, trickplay: { sheetUris: [uri], ...grid } }, id: true }
	});
}

/**
 * Who is probably in a video, from its newest face scan (or the one with
 * `model`). `scan` is null when it isn't scanned yet.
 */
export async function fetchVideoFaces(videoId: string, model?: string): Promise<VideoFaceGuesses> {
	const data = await gqlClient.query({
		videoFaces: { __args: { videoId, model }, ...VIDEO_FACES }
	});
	return data.videoFaces;
}
