import { get, writable } from 'svelte/store';
import type {
	ActiveRecording,
	CamModel,
	CamShow,
	CamSort,
	FieldsSelection
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { backendUrl } from '$lib/server';
import { VIDEO_CARD, fetchVideosMatching, type VideoCardRow } from '$lib/library';
import type { SitePlugin } from '$lib/plugins';
import { searchPage, type SearchPage } from '$lib/search';
import { flushEvents, logEvent, type Subject } from '$lib/events';

/** A cam room as a card shows it; live is null while the room is offline. */
const CAM_CARD = {
	id: true,
	plugin: true,
	username: true,
	displayName: true,
	tags: true,
	autoRecord: true,
	recordQuality: true,
	lastSeenOnlineAt: true,
	performer: { id: true, name: true },
	thumbnail: { uri: true },
	live: { viewers: true, subject: true, startedAt: true, thumbnail: { uri: true } }
} as const;

/** A recording running now. */
const ACTIVE_RECORDING = {
	camModel: CAM_CARD,
	startedAt: true,
	durationSeconds: true,
	sizeBytes: true,
	quality: true
} as const;

/** A finished show of a room. */
const CAM_SHOW = { startedAt: true, endedAt: true } as const;

// reason says why a room ranked by taste is listed ("Because you like #asian").
export type CamRow = FieldsSelection<CamModel, typeof CAM_CARD> & { reason?: string };
export type ActiveRecordingRow = FieldsSelection<ActiveRecording, typeof ACTIVE_RECORDING>;
export type CamShowRow = FieldsSelection<CamShow, typeof CAM_SHOW>;

/** What narrows the live rooms and how they're ordered; empty values don't narrow. */
export interface CamFilter {
	plugins: string[];
	gender: string;
	tag: string;
	text: string;
	sort: CamSort;
}

/** The genders the cam sites sort rooms by, as filter choices. */
export const CAM_GENDERS = [
	{ value: '', label: 'All' },
	{ value: 'female', label: 'Female' },
	{ value: 'male', label: 'Male' },
	{ value: 'couple', label: 'Couples' },
	{ value: 'trans', label: 'Trans' }
];

/** The orders of the live rooms. */
export const CAM_SORTS: { value: CamSort; label: string }[] = [
	{ value: 'viewers', label: 'Most viewers' },
	{ value: 'forYou', label: 'For you' },
	{ value: 'started', label: 'Just started' },
	{ value: 'fewest', label: 'Fewest viewers' }
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

/** The rooms the user follows: room id → follow id. Filled by loadFollowedCams. */
export const followedCams = writable<Map<string, string>>(new Map());

let followedLoad: Promise<void> | null = null;

/** The enabled sites that list cam rooms. */
export function camSites(plugins: SitePlugin[]): SitePlugin[] {
	return plugins.filter(
		(plugin) => plugin.enabled && plugin.available && plugin.capabilities.includes('cam:list')
	);
}

/**
 * The room's subject without hashtags and token counters ("[840 tokens
 * left]"); null when nothing else is in it.
 */
export function roomTopic(subject: string | null | undefined): string | null {
	if (!subject) return null;
	const cleaned = subject
		.replace(/#[\p{L}\p{N}_-]+/gu, '')
		.replace(/\[[^\]]*\]/g, '')
		.replace(/\s+/g, ' ')
		.replace(/[\s/|,-]+$/u, '')
		.trim();
	if (cleaned === '') return null;
	return cleaned;
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

/**
 * The library's videos of a room: its recordings, then videos credited to
 * its performer or naming its username (cam show archives title them so).
 */
export async function fetchRoomVideos(room: CamRow, limit = 60): Promise<VideoCardRow[]> {
	const requests = [
		fetchCamRecordings(room.id),
		fetchVideosMatching({ text: room.username }, limit)
	];
	if (room.performer) {
		requests.push(fetchVideosMatching({ performerIds: [room.performer.id] }, limit));
	}
	const lists = await Promise.all(requests);
	return uniqueVideos(lists.flat());
}

/** The videos without repeats, first place kept. */
export function uniqueVideos(videos: VideoCardRow[]): VideoCardRow[] {
	const seen = new Set<string>();
	return videos.filter((video) => {
		if (seen.has(video.id)) return false;
		seen.add(video.id);
		return true;
	});
}

/**
 * Searches every site of recorded cam shows for the room's username, calling
 * onvideos with the videos each site finds as it answers (they're stored in
 * the library), and onsites with how many sites were asked and answered.
 */
export function findRoomShows(
	room: CamRow,
	onvideos: (videos: VideoCardRow[]) => void,
	onsites: (asked: number, answered: number) => void
): SearchPage {
	let answered = 0;
	return searchPage('VIDEO', room.username, 0, null, null, 'cam-shows', (batch) => {
		if (batch.listed.videos.length > 0) onvideos(batch.listed.videos);
		if (batch.final) answered += 1;
		onsites(batch.sites, answered);
	});
}

/** The cam rooms of the person in a video, live state checked now. */
export async function fetchVideoRooms(videoId: string): Promise<CamRow[]> {
	const data = await gqlClient.query({ video: { __args: { id: videoId }, camRooms: CAM_CARD } });
	return data.video?.camRooms ?? [];
}

/** One room's shows of the last eight weeks, newest first. */
export async function fetchCamShows(id: string): Promise<CamShowRow[]> {
	const data = await gqlClient.query({ camModel: { __args: { id }, shows: CAM_SHOW } });
	return data.camModel?.shows ?? [];
}

/** Rooms live now on every cam site, most viewers first. */
export async function fetchLiveCams(limit = 20): Promise<CamRow[]> {
	const data = await gqlClient.query({
		camRooms: { __args: { query: { limit } }, rooms: CAM_CARD }
	});
	return data.camRooms.rooms;
}

/** A page of the rooms live now that match the filter, in its order. */
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
		sort: filter.sort,
		limit,
		offset
	};
	const data = await gqlClient.query({
		camRooms: { __args: { query }, rooms: CAM_CARD, total: true, reasons: true }
	});
	return {
		rooms: withReasons(data.camRooms.rooms, data.camRooms.reasons),
		total: data.camRooms.total
	};
}

/** Rooms live now that match what the user likes, best first, each with its reason. */
export async function fetchRecommendedCams(limit = 20): Promise<CamRow[]> {
	const data = await gqlClient.query({
		camRecommendations: { __args: { limit }, rooms: CAM_CARD, reasons: true }
	});
	return withReasons(data.camRecommendations.rooms, data.camRecommendations.reasons);
}

/** Gives each room the reason at its position; an empty reason is none. */
function withReasons(rooms: CamRow[], reasons: string[]): CamRow[] {
	return rooms.map((room, index) => {
		const reason = reasons[index];
		if (!reason) return room;
		return { ...room, reason };
	});
}

/**
 * Rooms live now sharing the room's first tags, without it: each tag's rooms
 * ranked by the user's taste, taken in turns.
 */
export async function fetchSimilarRooms(room: CamRow, limit = 12): Promise<CamRow[]> {
	const tags = room.tags.slice(0, 2);
	const pages = await Promise.all(
		tags.map((tag) =>
			fetchCamRooms({ plugins: [], gender: '', tag, text: '', sort: 'forYou' }, limit, 0)
		)
	);
	const seen = new Set([room.id]);
	const similar: CamRow[] = [];
	for (const candidate of inTurns(pages.map((found) => found.rooms))) {
		if (seen.has(candidate.id)) continue;
		seen.add(candidate.id);
		similar.push(candidate);
	}
	return similar.slice(0, limit);
}

/** The lists' first entries, then their second ones, and so on. */
function inTurns<Entry>(lists: Entry[][]): Entry[] {
	const merged: Entry[] = [];
	const longest = Math.max(0, ...lists.map((list) => list.length));
	for (let index = 0; index < longest; index++) {
		for (const list of lists) {
			if (index < list.length) merged.push(list[index]);
		}
	}
	return merged;
}

/**
 * The rooms the user keeps: followed or auto-recorded, live ones first
 * (most viewers first), then the offline ones by name.
 */
export async function fetchYourRooms(): Promise<CamRow[]> {
	const [, autoRecorded] = await Promise.all([
		loadFollowedCams(),
		gqlClient.query({ camModels: { __args: { autoRecord: true, limit: 200 }, ...CAM_CARD } })
	]);
	const known = new Set(autoRecorded.camModels.map((room) => room.id));
	const missingIds = [...get(followedCams).keys()].filter((id) => !known.has(id));
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

/** Loads the followed rooms into followedCams, once per page load. */
export function loadFollowedCams(): Promise<void> {
	if (followedLoad === null) {
		followedLoad = fetchFollowedCams().then((followed) => followedCams.set(followed));
	}
	return followedLoad;
}

/** The rooms with an enabled follow, each with its follow's id. */
async function fetchFollowedCams(): Promise<Map<string, string>> {
	const data = await gqlClient.query({
		follows: { id: true, enabled: true, targetKind: true, targetId: true }
	});
	const followed = new Map<string, string>();
	for (const follow of data.follows) {
		if (!follow.enabled || follow.targetKind !== 'camModel' || !follow.targetId) continue;
		followed.set(follow.targetId, follow.id);
	}
	return followed;
}

/** Follows a room (going live notifies), or ends its follow. */
export async function toggleCamFollow(id: string): Promise<void> {
	await loadFollowedCams();
	const followId = get(followedCams).get(id);
	if (followId) {
		await gqlClient.mutation({ unfollow: { __args: { id: followId } } });
		followedCams.update((current) => withoutKey(current, id));
		return;
	}
	const data = await gqlClient.mutation({
		followRecord: { __args: { subject: { kind: 'camModel', id }, notify: true }, id: true }
	});
	followedCams.update((current) => new Map(current).set(id, data.followRecord.id));
}

/** A copy of the map without one key. */
function withoutKey(map: Map<string, string>, key: string): Map<string, string> {
	const copy = new Map(map);
	copy.delete(key);
	return copy;
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

/**
 * A playable URL for a room's live stream. A preview (a card's hover) starts
 * no live buffer, so it can't be rewound.
 */
export async function watchCam(
	id: string,
	preview = false
): Promise<{ url: string; mimeType: string }> {
	const data = await gqlClient.mutation({
		watchCam: { __args: { id, preview }, url: true, mimeType: true }
	});
	return { url: backendUrl(data.watchCam.url), mimeType: data.watchCam.mimeType };
}

// A room's preview stream stays playable about a minute (its token expires).
const PREVIEW_STREAM_MS = 45_000;

const previewStreams = new Map<string, { at: number; url: Promise<string> }>();

/**
 * The preview stream of a room, resolved once and reused for a while, so a
 * card can ask as the pointer enters and the preview finds it ready. fresh
 * asks again, for a stream whose token ran out.
 */
export function previewStream(id: string, fresh = false): Promise<string> {
	const known = previewStreams.get(id);
	if (!fresh && known && Date.now() - known.at < PREVIEW_STREAM_MS) return known.url;
	const url = watchCam(id, true).then((stream) => stream.url);
	previewStreams.set(id, { at: Date.now(), url });
	url.catch(() => previewStreams.delete(id));
	return url;
}

/** Turns auto-recording of a room on or off, at a quality when given. */
export async function setCamAutoRecord(id: string, on: boolean, quality?: string): Promise<CamRow> {
	const data = await gqlClient.mutation({
		setCamAutoRecord: { __args: { id, on, quality }, ...CAM_CARD }
	});
	return data.setCamAutoRecord;
}

/** Links the room to a library performer; null unlinks. */
export async function linkCamPerformer(id: string, performerId: string | null): Promise<CamRow> {
	const data = await gqlClient.mutation({
		linkCamPerformer: { __args: { id, performerId }, ...CAM_CARD }
	});
	return data.linkCamPerformer;
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

const WEEK_MS = 7 * 24 * 60 * 60 * 1000;

/**
 * When the room is usually on: per weekday (0 = Sunday) and hour, in local
 * time, the share of the weeks since its oldest show it was on then, 0..1.
 */
export function usualHours(shows: CamShowRow[]): number[][] {
	const hours = Array.from({ length: 7 }, () => new Array<number>(24).fill(0));
	if (shows.length === 0) return hours;
	for (const show of shows) {
		markShowHours(hours, new Date(show.startedAt), new Date(show.endedAt));
	}
	const oldest = Math.min(...shows.map((show) => new Date(show.startedAt).getTime()));
	const weeks = Math.max(1, Math.ceil((Date.now() - oldest) / WEEK_MS));
	return hours.map((day) => day.map((count) => Math.min(1, count / weeks)));
}

/** Counts each local hour a show was on. */
function markShowHours(hours: number[][], start: Date, end: Date) {
	const cursor = new Date(start);
	cursor.setMinutes(0, 0, 0);
	while (cursor < end) {
		hours[cursor.getDay()][cursor.getHours()] += 1;
		cursor.setHours(cursor.getHours() + 1);
	}
}

// Where cam room cards are shown, for their impressions and choices.
export const CAMS_SURFACE = 'cams';
// How often a room being watched reports how long it has played.
const WATCH_REPORT_SECONDS = 30;

/** The event subject of a cam room. */
export function camSubject(id: string): Subject {
	return { kind: 'camModel', id };
}

/**
 * Logs that the room was opened, then how long it played while isPlaying
 * says so: every WATCH_REPORT_SECONDS of playing, and once more when the
 * returned stop is called (leaving the room). The longest report of the
 * session is how long the room was watched; one under ten seconds counts as
 * leaving it at once. A room that never played (a stream that failed)
 * reports nothing.
 */
export function trackRoomWatch(id: string, isPlaying: () => boolean): () => void {
	const subject = camSubject(id);
	const sessionId = crypto.randomUUID();
	let watched = 0;
	let reported = 0;
	logEvent({ kind: 'view', subject, data: { surface: CAMS_SURFACE }, sessionId });

	/** Sends how long the room has played, unless that was sent already. */
	function report() {
		if (watched === reported) return;
		reported = watched;
		logEvent({ kind: 'progress', subject, data: { watched }, sessionId });
	}

	const timer = setInterval(() => {
		if (!isPlaying()) return;
		watched += 1;
		if (watched % WATCH_REPORT_SECONDS === 0) report();
	}, 1000);
	return () => {
		clearInterval(timer);
		report();
		void flushEvents();
	};
}
