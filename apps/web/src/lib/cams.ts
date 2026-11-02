import type {
	ActiveRecording,
	CamModel,
	FieldsSelection
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';

/** A cam room as a card shows it; live is null while the room is offline. */
const CAM_CARD = {
	id: true,
	plugin: true,
	displayName: true,
	thumbnail: { uri: true },
	live: { viewers: true, subject: true, thumbnail: { uri: true } }
} as const;

/** A recording running now. */
const ACTIVE_RECORDING = {
	camModel: CAM_CARD,
	startedAt: true,
	durationSeconds: true,
	quality: true
} as const;

export type CamRow = FieldsSelection<CamModel, typeof CAM_CARD>;
export type ActiveRecordingRow = FieldsSelection<ActiveRecording, typeof ACTIVE_RECORDING>;

/** Recordings running now. */
export async function fetchActiveRecordings(): Promise<ActiveRecordingRow[]> {
	const data = await gqlClient.query({ activeRecordings: ACTIVE_RECORDING });
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
	const data = await gqlClient.query({
		camRooms: { __args: { query: { limit } }, rooms: CAM_CARD }
	});
	return data.camRooms.rooms;
}

/** One stored room. */
export async function fetchCam(id: string): Promise<CamRow | null> {
	const data = await gqlClient.query({ camModel: { __args: { id }, ...CAM_CARD } });
	return data.camModel ?? null;
}

/** A playable URL for a room's live stream. */
export async function watchCam(id: string): Promise<{ url: string; mimeType: string }> {
	const data = await gqlClient.mutation({
		watchCam: { __args: { id }, url: true, mimeType: true }
	});
	return { url: backendUrl(data.watchCam.url), mimeType: data.watchCam.mimeType };
}
