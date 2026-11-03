import type {
	ActiveRecording,
	CamModel,
	FieldsSelection
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';
import type { SitePlugin } from '$lib/plugins';

/** A cam room as a card shows it; live is null while the room is offline. */
const CAM_CARD = {
	id: true,
	plugin: true,
	displayName: true,
	tags: true,
	autoRecord: true,
	recordQuality: true,
	thumbnail: { uri: true },
	live: { viewers: true, subject: true, thumbnail: { uri: true } }
} as const;

/** A recording running now. */
const ACTIVE_RECORDING = {
	camModel: CAM_CARD,
	startedAt: true,
	durationSeconds: true,
	sizeBytes: true,
	quality: true
} as const;

export type CamRow = FieldsSelection<CamModel, typeof CAM_CARD>;
export type ActiveRecordingRow = FieldsSelection<ActiveRecording, typeof ACTIVE_RECORDING>;

/** What narrows the live rooms; empty values don't narrow. */
export interface CamFilter {
	plugins: string[];
	gender: string;
	tag: string;
	text: string;
}

/** The genders the cam sites sort rooms by, as filter choices. */
export const CAM_GENDERS = [
	{ value: '', label: 'All' },
	{ value: 'female', label: 'Female' },
	{ value: 'male', label: 'Male' },
	{ value: 'couple', label: 'Couples' },
	{ value: 'trans', label: 'Trans' }
];

/** The qualities a recording can run at; the backend reads "source" or a height. */
export const RECORD_QUALITIES = [
	{ value: 'source', label: 'Source' },
	{ value: '1080', label: '1080p' },
	{ value: '720', label: '720p' },
	{ value: '480', label: '480p' }
];

/** How far back a recording can start; the live buffer keeps three minutes. */
export const RECORD_REWINDS = [
	{ seconds: 0, label: 'Now' },
	{ seconds: 60, label: '1 min ago' },
	{ seconds: 180, label: '3 min ago' }
];

/** The enabled sites that list cam rooms. */
export function camSites(plugins: SitePlugin[]): SitePlugin[] {
	return plugins.filter(
		(plugin) => plugin.enabled && plugin.available && plugin.capabilities.includes('cam:list')
	);
}

/** Recordings running now. */
export async function fetchActiveRecordings(): Promise<ActiveRecordingRow[]> {
	const data = await gqlClient.query({ activeRecordings: ACTIVE_RECORDING });
	return data.activeRecordings;
}

/** The recording of one room running now, if there is one. */
export async function fetchRoomRecording(id: string): Promise<ActiveRecordingRow | null> {
	const recordings = await fetchActiveRecordings();
	return recordings.find((recording) => recording.camModel?.id === id) ?? null;
}

/** Finished recordings of the stored rooms, as the videos they became. */
export async function fetchRecordedVideos(limit = 20): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		camModels: { __args: { limit: 50 }, recordings: VIDEO_CARD }
	});
	return data.camModels.flatMap((model) => model.recordings).slice(0, limit);
}

/** One room's finished recordings, newest first. */
export async function fetchCamRecordings(id: string): Promise<VideoCardRow[]> {
	const data = await gqlClient.query({
		camModel: { __args: { id }, recordings: VIDEO_CARD }
	});
	return data.camModel?.recordings ?? [];
}

/** Rooms live now on every cam site, most viewers first. */
export async function fetchLiveCams(limit = 20): Promise<CamRow[]> {
	const data = await gqlClient.query({
		camRooms: { __args: { query: { limit } }, rooms: CAM_CARD }
	});
	return data.camRooms.rooms;
}

/** A page of the rooms live now that match the filter, most viewers first. */
export async function fetchCamRooms(
	filter: CamFilter,
	limit: number,
	offset: number
): Promise<{ rooms: CamRow[]; total: number }> {
	const query = {
		plugins: filter.plugins.length > 0 ? filter.plugins : undefined,
		gender: filter.gender || undefined,
		tag: filter.tag || undefined,
		text: filter.text || undefined,
		limit,
		offset
	};
	const data = await gqlClient.query({
		camRooms: { __args: { query }, rooms: CAM_CARD, total: true }
	});
	return data.camRooms;
}

/**
 * The rooms the user keeps: followed or auto-recorded, live ones first
 * (most viewers first), then the offline ones by name.
 */
export async function fetchYourRooms(): Promise<CamRow[]> {
	const [followedIds, autoRecorded] = await Promise.all([
		fetchFollowedCamIds(),
		gqlClient.query({ camModels: { __args: { autoRecord: true, limit: 200 }, ...CAM_CARD } })
	]);
	const known = new Set(autoRecorded.camModels.map((room) => room.id));
	const missingIds = followedIds.filter((id) => !known.has(id));
	const followed = await Promise.all(missingIds.map((id) => fetchCam(id)));
	const rooms = [...autoRecorded.camModels, ...followed.filter((room) => room !== null)];
	return rooms.sort(compareRooms);
}

/** Orders rooms live first by viewers, then offline ones by name. */
function compareRooms(first: CamRow, second: CamRow): number {
	if (first.live && second.live) return second.live.viewers - first.live.viewers;
	if (first.live) return -1;
	if (second.live) return 1;
	return first.displayName.localeCompare(second.displayName);
}

/** The ids of the rooms with an enabled follow. */
async function fetchFollowedCamIds(): Promise<string[]> {
	const data = await gqlClient.query({
		follows: { enabled: true, targetKind: true, targetId: true }
	});
	const ids: string[] = [];
	for (const follow of data.follows) {
		if (!follow.enabled || follow.targetKind !== 'camModel' || !follow.targetId) continue;
		ids.push(follow.targetId);
	}
	return ids;
}

/** The tags most of the rooms carry, most common first. */
export function popularRoomTags(rooms: CamRow[], count: number): string[] {
	const counts = new Map<string, number>();
	for (const room of rooms) {
		for (const tag of room.tags) counts.set(tag, (counts.get(tag) ?? 0) + 1);
	}
	return [...counts.entries()]
		.sort((first, second) => second[1] - first[1])
		.slice(0, count)
		.map(([tag]) => tag);
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

/** The id of the room's follow, or null when it isn't followed. */
export async function fetchCamFollow(id: string): Promise<string | null> {
	const data = await gqlClient.query({
		followOf: { __args: { subject: { kind: 'camModel', id } }, id: true, enabled: true }
	});
	if (!data.followOf?.enabled) return null;
	return data.followOf.id;
}

/** Follows a room, so going live notifies; returns the follow's id. */
export async function followCam(id: string): Promise<string> {
	const data = await gqlClient.mutation({
		followRecord: { __args: { subject: { kind: 'camModel', id }, notify: true }, id: true }
	});
	return data.followRecord.id;
}

/** Ends a follow. */
export async function unfollowCam(followId: string): Promise<void> {
	await gqlClient.mutation({ unfollow: { __args: { id: followId } } });
}

/** Turns auto-recording of a room on or off, at a quality when given. */
export async function setCamAutoRecord(id: string, on: boolean, quality?: string): Promise<CamRow> {
	const data = await gqlClient.mutation({
		setCamAutoRecord: { __args: { id, on, quality }, ...CAM_CARD }
	});
	return data.setCamAutoRecord;
}

/** Starts recording a room, from some seconds back in the live buffer. */
export async function startCamRecording(
	id: string,
	quality: string,
	fromSecondsAgo: number
): Promise<void> {
	await gqlClient.mutation({
		startRecording: { __args: { id, quality, fromSecondsAgo }, id: true }
	});
}

/** Stops a room's recording; the backend files it as a video. */
export async function stopCamRecording(id: string): Promise<void> {
	await gqlClient.mutation({ stopRecording: { __args: { id } } });
}
