import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';

/** A cam room as a card shows it. */
export interface CamRow {
	id: string;
	plugin: string;
	displayName: string;
	thumbnail: { uri: string } | null;
	// Null while the room is offline.
	live: { viewers: number; subject: string | null; thumbnail: { uri: string } | null } | null;
}

const CAM_CARD = {
	id: true,
	plugin: true,
	displayName: true,
	thumbnail: { uri: true },
	live: { viewers: true, subject: true, thumbnail: { uri: true } }
} as const;

/** A recording running now. */
export interface ActiveRecordingRow {
	camModel: CamRow | null;
	startedAt: string;
	durationSeconds: number;
	quality: string | null;
}

/** Recordings running now. */
export async function fetchActiveRecordings(): Promise<ActiveRecordingRow[]> {
	const data = await gqlClient.query({
		activeRecordings: { camModel: CAM_CARD, startedAt: true, durationSeconds: true, quality: true }
	});
	return data.activeRecordings;
}

/** Finished recordings of the stored rooms, as the videos they became. */
export async function fetchRecordedVideos(limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		camModels: { __args: { limit: 50 }, recordings: VIDEO_CARD }
	});
	return data.camModels.flatMap((model) => model.recordings).slice(0, limit);
}

/** Rooms live now on every cam site, most viewers first. */
export async function fetchLiveCams(limit = 20): Promise<CamRow[]> {
	const data = await gqlClient.query({ camRooms: { __args: { query: { limit } }, rooms: CAM_CARD } });
	return data.camRooms.rooms;
}

/** One stored room. */
export async function fetchCam(id: string): Promise<CamRow | null> {
	const data = await gqlClient.query({ camModel: { __args: { id }, ...CAM_CARD } });
	return data.camModel ?? null;
}

/** A playable URL for a room's live stream. */
export async function watchCam(id: string): Promise<{ url: string; mimeType: string }> {
	const data = await gqlClient.mutation({ watchCam: { __args: { id }, url: true, mimeType: true } });
	return { url: backendUrl(data.watchCam.url), mimeType: data.watchCam.mimeType };
}
