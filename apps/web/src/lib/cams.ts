// Live cam rooms: the grid across cam plugins, followed models, and a room's
// live stream. A cam model is a room that comes and goes, not a scene.
import { browser } from '$app/environment';
import { gqlClient } from '$lib/veil';
import { SCENE_ROW_FIELDS, sceneToCard, type SceneRow } from '$lib/filters';
import type { SceneCardData } from '$lib/components/SceneCard.svelte';
import { backendUrl } from '$lib/server';

export interface CamModel {
	id: string;
	pluginName: string;
	username: string;
	displayName: string;
	sourceUrl: string;
	thumbnailUrl: string | null;
	tags: string[];
	gender: string | null;
	subject: string | null;
	viewers: number;
	online: boolean;
	lastSeenOnlineAt: string | null;
	followed: boolean;
	// Recorded automatically whenever the model is live.
	autoRecord: boolean;
	// What auto-record captures: "source" (the best rendition) or a height like "480".
	recordQuality: string;
}

export interface CamSource {
	name: string;
	displayName: string;
	iconUrl: string | null;
}

export interface CamRoomFilter {
	sources: string[];
	tag: string;
	query: string;
	gender: string;
}

export interface CamRoomPage {
	rooms: CamModel[];
	totalCount: number;
}

export interface CamStream {
	url: string;
	mimeType: string;
}

const CAM_MODEL_FIELDS = {
	id: true,
	pluginName: true,
	username: true,
	displayName: true,
	sourceUrl: true,
	thumbnailUrl: true,
	tags: true,
	gender: true,
	subject: true,
	viewers: true,
	online: true,
	lastSeenOnlineAt: true,
	followed: true,
	autoRecord: true,
	recordQuality: true
} as const;

/** One page of the rooms online now, in the order the sites rank them. */
export async function fetchCamRooms(
	filter: CamRoomFilter,
	limit: number,
	offset: number
): Promise<CamRoomPage> {
	const data = await gqlClient.query({
		camRooms: {
			__args: {
				sources: filter.sources,
				tag: filter.tag || undefined,
				query: filter.query || undefined,
				gender: filter.gender || undefined,
				limit,
				offset
			},
			totalCount: true,
			rooms: CAM_MODEL_FIELDS
		}
	});
	return data.camRooms as unknown as CamRoomPage;
}

/** The cam sites installed, for the site filter. */
export async function fetchCamSources(): Promise<CamSource[]> {
	try {
		const data = await gqlClient.query({
			camSources: { name: true, displayName: true, iconUrl: true }
		});
		return data.camSources as unknown as CamSource[];
	} catch {
		return [];
	}
}

/** Followed models, online ones first. */
export async function fetchFollowedCamModels(): Promise<CamModel[]> {
	try {
		const data = await gqlClient.query({ followedCamModels: CAM_MODEL_FIELDS });
		return data.followedCamModels as unknown as CamModel[];
	} catch {
		return [];
	}
}

/** One model with its online status checked against the site; null when unknown. */
export async function fetchCamModel(id: string): Promise<CamModel | null> {
	try {
		const data = await gqlClient.query({ camModel: { __args: { id }, ...CAM_MODEL_FIELDS } });
		return (data.camModel ?? null) as unknown as CamModel | null;
	} catch {
		return null;
	}
}

/** The room's live HLS stream, through the backend proxy; throws when the room is offline. */
export async function fetchCamStream(id: string): Promise<CamStream> {
	const data = await gqlClient.query({ camStream: { __args: { id }, url: true, mimeType: true } });
	const stream = data.camStream as unknown as CamStream;
	return { url: backendUrl(stream.url), mimeType: stream.mimeType };
}

/** Follows or unfollows a model. */
export async function setCamModelFollowed(id: string, followed: boolean): Promise<CamModel> {
	if (followed) {
		const data = await gqlClient.mutation({
			followCamModel: { __args: { id }, ...CAM_MODEL_FIELDS }
		});
		return data.followCamModel as unknown as CamModel;
	}
	const data = await gqlClient.mutation({
		unfollowCamModel: { __args: { id }, ...CAM_MODEL_FIELDS }
	});
	return data.unfollowCamModel as unknown as CamModel;
}

/** A recording of a live room: in progress, or finished with its scene. */
export interface CamRecording {
	id: string;
	camModelId: string;
	displayName: string;
	thumbnailUrl: string | null;
	// The rendition captured, like "720p"; null before the stream is read.
	quality: string | null;
	status: 'recording' | 'finished' | 'failed';
	stopReason: string | null;
	startedAt: string;
	endedAt: string | null;
	durationSeconds: number;
	sizeBytes: number;
	sceneId: string | null;
}

const CAM_RECORDING_FIELDS = {
	id: true,
	camModelId: true,
	displayName: true,
	thumbnailUrl: true,
	quality: true,
	status: true,
	stopReason: true,
	startedAt: true,
	endedAt: true,
	durationSeconds: true,
	sizeBytes: true,
	sceneId: true
} as const;

/** The recordings in progress, with their live duration and size. */
export async function fetchActiveCamRecordings(): Promise<CamRecording[]> {
	try {
		const data = await gqlClient.query({
			camRecordings: { __args: { activeOnly: true }, ...CAM_RECORDING_FIELDS }
		});
		return data.camRecordings as unknown as CamRecording[];
	} catch {
		return [];
	}
}

/** A model's recordings: the one in progress, then the finished ones (their scenes). */
export async function fetchCamModelRecordings(camModelId: string): Promise<CamRecording[]> {
	try {
		const data = await gqlClient.query({
			camRecordings: { __args: { camModelId }, ...CAM_RECORDING_FIELDS }
		});
		return data.camRecordings as unknown as CamRecording[];
	} catch {
		return [];
	}
}

/**
 * Starts recording a live room in the backend at a quality ("source" or a height
 * like "480"); throws when it is not live. With fromSecondsAgo the recording
 * begins that far in the past, from the room's live buffer (the backend caps it
 * to what is buffered and keeps the buffered rendition).
 */
export async function startCamRecording(
	id: string,
	quality: string,
	fromSecondsAgo?: number
): Promise<CamRecording> {
	const data = await gqlClient.mutation({
		startCamRecording: { __args: { id, quality, fromSecondsAgo }, ...CAM_RECORDING_FIELDS }
	});
	return data.startCamRecording as unknown as CamRecording;
}

/** Stops a room's recording and waits for its scene; null when none was running. */
export async function stopCamRecording(id: string): Promise<CamRecording | null> {
	const data = await gqlClient.mutation({
		stopCamRecording: { __args: { id }, ...CAM_RECORDING_FIELDS }
	});
	return (data.stopCamRecording ?? null) as unknown as CamRecording | null;
}

/**
 * Turns automatic recording on or off for a model (turning it on also follows)
 * and sets the quality it records at; the quality stays as it was when omitted.
 */
export async function setCamAutoRecord(
	id: string,
	enabled: boolean,
	quality?: string
): Promise<CamModel> {
	const data = await gqlClient.mutation({
		setCamAutoRecord: { __args: { id, enabled, quality }, ...CAM_MODEL_FIELDS }
	});
	return data.setCamAutoRecord as unknown as CamModel;
}

/** A finished recording with its scene as a card, ready for SceneCard. */
export interface RecordedSceneCard {
	recording: CamRecording;
	card: SceneCardData;
}

/**
 * The scenes of finished recordings as cards (poster, title, duration, date,
 * quality badge), in the order of the recordings. A recording whose scene is
 * gone is left out.
 */
export async function fetchRecordedSceneCards(
	recordings: CamRecording[]
): Promise<RecordedSceneCard[]> {
	const loaded = await Promise.all(
		recordings.map(async (recording) => {
			if (!recording.sceneId) return null;
			const data = await gqlClient
				.query({ scene: { __args: { id: recording.sceneId }, ...SCENE_ROW_FIELDS } })
				.catch(() => null);
			if (!data?.scene) return null;
			const card = sceneToCard(data.scene as unknown as SceneRow);
			if (recording.quality) card.badge = `Live ${recording.quality}`;
			return { recording, card };
		})
	);
	return loaded.filter((entry): entry is RecordedSceneCard => entry !== null);
}

const WATCH_QUALITY_KEY = 'veil.cam.quality';

/** The video height the user last picked for watching live rooms; 0 = automatic. */
export function loadWatchQuality(): number {
	if (!browser) return 0;
	const stored = Number(localStorage.getItem(WATCH_QUALITY_KEY));
	if (!Number.isFinite(stored) || stored < 0) return 0;
	return stored;
}

/** Remembers the picked watching quality (a video height, 0 = automatic). */
export function saveWatchQuality(height: number) {
	if (!browser) return;
	localStorage.setItem(WATCH_QUALITY_KEY, String(height));
}

/** The recording quality for a watched height: that height, or the source when automatic. */
export function recordQualityFor(height: number): string {
	if (height > 0) return String(height);
	return 'source';
}

/** "Source", or "480p" for a stored recording quality. */
export function qualityName(quality: string): string {
	if (quality === 'source') return 'Source';
	return `${quality}p`;
}

/** "1:02:03" / "4:05" clock for a recording's length in seconds. */
export function formatRecordingClock(seconds: number): string {
	const whole = Math.max(0, Math.floor(seconds));
	const hours = Math.floor(whole / 3600);
	const minutes = Math.floor((whole % 3600) / 60);
	const rest = String(whole % 60).padStart(2, '0');
	if (hours > 0) return `${hours}:${String(minutes).padStart(2, '0')}:${rest}`;
	return `${minutes}:${rest}`;
}

/** "1.4 GB" style size of a recording. */
export function formatRecordingSize(bytes: number): string {
	if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
	return `${Math.round(bytes / 1e6)} MB`;
}

/** "1.2K" style viewer counts for the card badge. */
export function formatViewers(viewers: number): string {
	if (viewers >= 10000) return `${Math.round(viewers / 1000)}K`;
	if (viewers >= 1000) return `${(viewers / 1000).toFixed(1)}K`;
	return String(viewers);
}
