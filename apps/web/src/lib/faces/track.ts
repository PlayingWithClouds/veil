/**
 * Background face tracking. Opening a video scans it per face model: the
 * credited performers' unscanned pictures first, then the video's stills
 * (cover and the sites' preview frames, plain image fetches), then 25
 * evenly spaced frames. Each frame is taken from the page's player as its
 * playback passes that time, so watching costs no extra download; while
 * nothing plays, a hidden player seeks to the frames still missing, unless
 * the stills already showed enough faces. The scan is saved as it grows,
 * so the guesses refine while it runs, and a later visit picks up a scan
 * that didn't get all its frames. All 25 frames also make the video's
 * perceptual hash and, when it has none, its trickplay sheet. One record
 * at a time (a newer request cancels the older), models in a worker.
 */
import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { cacheUrl } from '$lib/img';
import { play } from '$lib/library';
import { uploadPicture } from '$lib/performers';
import { analyzeFrame } from './analyzer';
import * as api from './api';
import { FACE_MODEL, type Frame } from './engine';
import { grayscale, perceptualHash } from './phash';
import {
	SAMPLE_COUNT,
	compactVector,
	frameSize,
	pickStream,
	regionOf,
	sampleTimes,
	sheetLayout,
	type ScannedFace,
	type SheetLayout
} from './plan';
import { VideoSampler, readFrame } from './sampler';

/** How long tracking waits after a video opens, so its player gets the bandwidth first. */
const START_DELAY_MS = 4000;

/** At most this many pictures of one performer are scanned. */
const MAX_PICTURES = 4;

/** Stills showing a face that make seeking through the stream unnecessary. */
const ENOUGH_STILLS_WITH_FACES = 3;

/** How often the page's playback is checked for a frame at a sample time. */
const PLAYBACK_POLL_MS = 250;

/** How far from a sample time a played frame may be to stand in for it. */
const PLAYED_TOLERANCE_SECONDS = 1;

/** How far the player's length may be from the video's for it to count as playing this video. */
const LENGTH_TOLERANCE_SECONDS = 5;

/** New frames between saves of the growing scan. */
const SAVE_EVERY_FRAMES = 5;

/** Seeks failing in a row after which the hidden player gives up (a stream that can't seek). */
const MAX_SEEK_FAILURES = 3;

/** How a tracking request ended. */
export type TrackOutcome = 'scanned' | 'skipped' | 'cancelled' | 'failed';

/**
 * What tracking is doing now, for a quiet progress hint. `playback` takes
 * frames from the playing video; `waiting` means seeking waits for another
 * video on the page to stop.
 */
export interface FaceTracking {
	subject: { kind: 'video' | 'performer'; id: string };
	phase: 'pictures' | 'stills' | 'playback' | 'waiting' | 'frames' | 'saving';
	// 0..1 through the phase.
	progress: number;
}

/** What scanning a video can use from its page. */
export interface VideoTrackOptions {
	// The page's player element, null while it has none.
	player?: () => HTMLVideoElement | null;
	// Called each time the growing scan is saved, e.g. to reload the guesses.
	onsaved?: () => void;
}

/** The running tracking, null when idle. */
export const faceTracking = writable<FaceTracking | null>(null);

let current: { key: string; controller: AbortController; done: Promise<TrackOutcome> } | null =
	null;

/**
 * Scans a video for faces in the background, unless it's fully scanned with
 * this model already (then its credited performers' faces are only learned
 * again). `fetchVideoFaces` has the guesses after each save and when done.
 */
export function trackVideo(
	videoId: string,
	options: VideoTrackOptions = {}
): Promise<TrackOutcome> {
	return runExclusive(`video:${videoId}`, (signal) => scanVideo(videoId, options, signal));
}

/** Scans a performer's unscanned pictures for their face in the background. */
export function trackPerformer(performerId: string): Promise<TrackOutcome> {
	return runExclusive(`performer:${performerId}`, async (signal) => {
		await delay(START_DELAY_MS, signal);
		const scanned = await scanPictures(performerId, signal);
		if (scanned === 0) return 'skipped';
		return 'scanned';
	});
}

/** Stops the running tracking; what was recorded stays. */
export function cancelTracking(): void {
	current?.controller.abort();
	current = null;
	faceTracking.set(null);
}

/**
 * Runs one tracking job, cancelling any other; asking again for the
 * running one joins it.
 */
function runExclusive(
	key: string,
	job: (signal: AbortSignal) => Promise<TrackOutcome>
): Promise<TrackOutcome> {
	if (!browser) return Promise.resolve('skipped');
	if (current?.key === key) return current.done;
	cancelTracking();
	const controller = new AbortController();
	const done = settle(job(controller.signal)).finally(() => {
		if (current?.controller !== controller) return;
		current = null;
		faceTracking.set(null);
	});
	current = { key, controller, done };
	return done;
}

/** Turns a cancelled or failed job into its outcome. */
async function settle(job: Promise<TrackOutcome>): Promise<TrackOutcome> {
	try {
		return await job;
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') return 'cancelled';
		console.warn('Face tracking failed:', error);
		return 'failed';
	}
}

/**
 * Scans a video's credited performers' pictures, then (unless fully done
 * before) its stills and its frames, saving the scan as it grows.
 */
async function scanVideo(
	videoId: string,
	options: VideoTrackOptions,
	signal: AbortSignal
): Promise<TrackOutcome> {
	await delay(START_DELAY_MS, signal);
	const video = await api.fetchScanVideo(videoId);
	if (!video) return 'failed';
	for (const performer of video.performers) {
		await scanPictures(performer.id, signal);
	}
	const earlierFrames = await api.scannedFrames('video', videoId, FACE_MODEL);
	if (earlierFrames !== null && earlierFrames >= SAMPLE_COUNT) {
		await api.learnFaces(videoId, FACE_MODEL);
		return 'skipped';
	}
	const stills = await scanStills(video, signal);
	const seeker = new Seeker(video, signal);
	try {
		const duration = await videoLength(video, options.player, seeker);
		if (duration <= 0) throw new Error('The video has no length to sample');
		const scan = new FrameScan(video, stills, duration, earlierFrames ?? 0, options.onsaved);
		await scan.save();
		// Enough faces in the stills: frames only refine it while it plays, without seeking.
		const seeking = stills.withFaces < ENOUGH_STILLS_WITH_FACES;
		await scanFrames(scan, options.player, seeking, seeker, signal);
	} finally {
		seeker.close();
	}
	return 'scanned';
}

/**
 * The video's length, as its stream says: the page's player's when it
 * plays a video this long, else the hidden player's, else the site's.
 */
async function videoLength(
	video: ScanVideo,
	player: (() => HTMLVideoElement | null) | undefined,
	seeker: Seeker
): Promise<number> {
	const element = player?.();
	if (element && Number.isFinite(element.duration) && element.duration > 0) {
		if (sameLength(element.duration, video.durationSeconds)) return element.duration;
	}
	const sought = await seeker.duration();
	if (sought > 0) return sought;
	return video.durationSeconds ?? 0;
}

/** Whether a player's length fits the video's (unknown fits anything). */
function sameLength(seconds: number, videoSeconds: number | null | undefined): boolean {
	if (!videoSeconds) return true;
	return Math.abs(seconds - videoSeconds) <= LENGTH_TOLERANCE_SECONDS;
}

/** A video as scanning reads it. */
type ScanVideo = NonNullable<Awaited<ReturnType<typeof api.fetchScanVideo>>>;

/** What the stills gave: how many could be read, how many show a face, their faces. */
interface StillsScan {
	read: number;
	withFaces: number;
	faces: ScannedFace[];
}

/** A still to scan and roughly when in the video it is (null: unknown). */
interface Still {
	uri: string;
	atSeconds: number | null;
}

/**
 * Scans the video's cover and preview frames, all fetched at once. Preview
 * frames get their rough time (sites space them evenly), so faces of one
 * still are never taken for one person.
 */
async function scanStills(video: ScanVideo, signal: AbortSignal): Promise<StillsScan> {
	const stills = stillsOf(video);
	const loads = stills.map((still) => loadPicture(still.uri, signal));
	const scan: StillsScan = { read: 0, withFaces: 0, faces: [] };
	for (const [index, still] of stills.entries()) {
		faceTracking.set({
			subject: { kind: 'video', id: video.id },
			phase: 'stills',
			progress: index / stills.length
		});
		const frame = await loads[index];
		if (!frame) continue;
		scan.read++;
		const faces = await facesOf(frame, still.atSeconds);
		if (faces.length > 0) scan.withFaces++;
		scan.faces.push(...faces);
	}
	return scan;
}

/** The video's cover, then its preview frames placed evenly over its length. */
function stillsOf(video: ScanVideo): Still[] {
	const stills: Still[] = [];
	if (video.cover) stills.push({ uri: video.cover.uri, atSeconds: null });
	const previews = video.previewFrames ?? [];
	const duration = video.durationSeconds ?? 0;
	for (const [index, preview] of previews.entries()) {
		let atSeconds: number | null = null;
		if (duration > 0) atSeconds = Math.round((duration * (index + 1)) / (previews.length + 1));
		stills.push({ uri: preview.uri, atSeconds });
	}
	return stills;
}

/**
 * Fills the scan's frames until it has them all: from the page's player as
 * its playback passes a missing time, and, with seeking, by having the
 * hidden player seek to one while nothing on the page plays. Then saves
 * the hash, the sheet and the scan.
 */
async function scanFrames(
	scan: FrameScan,
	player: (() => HTMLVideoElement | null) | undefined,
	seeking: boolean,
	seeker: Seeker,
	signal: AbortSignal
) {
	let readablePlayer = true;
	while (!scan.complete) {
		const playing = readablePlayer ? playingElement(player, scan.duration) : null;
		if (playing) {
			readablePlayer = await takePlayed(scan, playing, signal);
			continue;
		}
		if (!seeking || !seeker.usable || isPlaying()) {
			await waitForPlayback(scan, seeking && seeker.usable, signal);
			continue;
		}
		await takeSought(scan, seeker, signal);
	}
	scan.show('saving', 0);
	await scan.finish();
}

/**
 * Takes the playing frame when it's at a missing time, else waits a moment.
 * Returns false when the player's pixels can't be read, so it's not asked again.
 */
async function takePlayed(
	scan: FrameScan,
	element: HTMLVideoElement,
	signal: AbortSignal
): Promise<boolean> {
	scan.show('playback', scan.progress);
	const seconds = element.currentTime;
	const index = scan.missingNear(seconds);
	if (index === null) {
		await delay(PLAYBACK_POLL_MS, signal);
		return true;
	}
	let frame: Frame;
	try {
		frame = readFrame(element, scan.canvas);
	} catch {
		return false;
	}
	scan.tile(index, element);
	await scan.add(index, frame, seconds);
	return true;
}

/** Has the hidden player seek to the first missing time and takes that frame. */
async function takeSought(scan: FrameScan, seeker: Seeker, signal: AbortSignal) {
	scan.show('frames', scan.progress);
	await idle(signal);
	const index = scan.firstMissing();
	if (index === null) return;
	const seconds = scan.times[index];
	const frame = await seeker.frameAt(seconds, (source) => scan.tile(index, source));
	if (!frame) return;
	await scan.add(index, frame, seconds);
}

/**
 * Waits for the page's playback to change. Seeking waits for the page's
 * other video to stop; without seeking there's nothing to show meanwhile.
 */
async function waitForPlayback(scan: FrameScan, seeking: boolean, signal: AbortSignal) {
	if (seeking) {
		scan.show('waiting', scan.progress);
	} else {
		faceTracking.set(null);
	}
	await delay(PLAYBACK_POLL_MS, signal);
}

/** The page's player while it plays a video of the scan's length, else null. */
function playingElement(
	player: (() => HTMLVideoElement | null) | undefined,
	duration: number
): HTMLVideoElement | null {
	const element = player?.();
	if (!element || element.paused || element.ended || element.seeking) return null;
	if (!Number.isFinite(element.duration)) return null;
	if (Math.abs(element.duration - duration) > LENGTH_TOLERANCE_SECONDS) return null;
	return element;
}

/** Whether any video in the page is playing. */
function isPlaying(): boolean {
	return [...document.querySelectorAll('video')].some(
		(element) => !element.paused && !element.ended
	);
}

/**
 * A video's frame scan in progress: which sample times have their frame,
 * the faces found so far (after the stills'), and the trickplay sheet.
 */
class FrameScan {
	readonly times: number[];
	readonly canvas = document.createElement('canvas');
	private readonly filled: boolean[];
	private readonly faces: ScannedFace[] = [];
	private sheet: TrickplayCanvas | null = null;
	private frames = 0;
	private unsaved = 0;

	constructor(
		private readonly video: ScanVideo,
		private readonly stills: StillsScan,
		readonly duration: number,
		// Frames of the scan saved before; a save never replaces it with fewer.
		private savedFrames: number,
		private readonly onsaved: (() => void) | undefined
	) {
		this.times = sampleTimes(duration);
		this.filled = this.times.map(() => false);
	}

	/** Whether every sample time has its frame. */
	get complete(): boolean {
		return this.frames === this.times.length;
	}

	/** How far through the frames, 0..1. */
	get progress(): number {
		return this.frames / this.times.length;
	}

	/** The missing sample time a played frame at `seconds` can stand in for, else null. */
	missingNear(seconds: number): number | null {
		const index = this.times.findIndex(
			(time, position) =>
				!this.filled[position] && Math.abs(time - seconds) <= PLAYED_TOLERANCE_SECONDS
		);
		if (index === -1) return null;
		return index;
	}

	/** The first sample time without its frame, else null. */
	firstMissing(): number | null {
		const index = this.filled.indexOf(false);
		if (index === -1) return null;
		return index;
	}

	/** Shows what the scan is doing. */
	show(phase: FaceTracking['phase'], progress: number) {
		faceTracking.set({ subject: { kind: 'video', id: this.video.id }, phase, progress });
	}

	/** Draws a sample time's frame into its trickplay tile; the sheet is sized by the first. */
	tile(index: number, source: HTMLVideoElement) {
		if (!this.sheet) {
			this.sheet = new TrickplayCanvas(
				sheetLayout(this.times.length, source.videoWidth, source.videoHeight)
			);
		}
		this.sheet.draw(index, source);
	}

	/** Adds a sample time's frame and its faces; saves every few frames. */
	async add(index: number, frame: Frame, seconds: number) {
		this.filled[index] = true;
		this.frames++;
		this.faces.push(...(await facesOf(frame, Math.round(seconds * 100) / 100)));
		this.unsaved++;
		if (this.unsaved >= SAVE_EVERY_FRAMES) await this.save();
	}

	/** Saves the scan so far, unless the saved one looked at as many frames. */
	async save() {
		const frames = this.frames + this.stills.read;
		if (frames <= this.savedFrames) return;
		await api.recordFaceScan({
			kind: 'video',
			id: this.video.id,
			model: FACE_MODEL,
			frames,
			faces: [...this.stills.faces, ...this.faces]
		});
		this.savedFrames = frames;
		this.unsaved = 0;
		this.onsaved?.();
	}

	/** Saves the hash and sheet of the complete frames, then the scan. */
	async finish() {
		if (this.sheet) await saveSheet(this.video, this.sheet, this.duration, this.times.length);
		await this.save();
	}
}

/**
 * The hidden player that seeks to sample times, opened on first use. Gives
 * up after a few failed seeks in a row, or when the video has no stream.
 */
class Seeker {
	private sampler: VideoSampler | null = null;
	private failures = 0;

	constructor(
		private readonly video: ScanVideo,
		private readonly signal: AbortSignal
	) {}

	/** Whether seeking still works. */
	get usable(): boolean {
		return this.failures < MAX_SEEK_FAILURES;
	}

	/** The stream's length in seconds; 0 when it can't be opened or doesn't say. */
	async duration(): Promise<number> {
		const sampler = await this.open();
		if (!sampler) return 0;
		return sampler.duration;
	}

	/**
	 * The frame at a time, null when the seek fails; a cancel or a stream
	 * that forbids reading its pixels ends the scan.
	 */
	async frameAt(seconds: number, draw: (source: HTMLVideoElement) => void): Promise<Frame | null> {
		const sampler = await this.open();
		if (!sampler) return null;
		try {
			const frame = await sampler.frameAt(seconds, this.signal, draw);
			this.failures = 0;
			return frame;
		} catch (error) {
			if (error instanceof DOMException) throw error;
			this.failures++;
			return null;
		}
	}

	/** Lets go of the stream. */
	close() {
		this.sampler?.close();
		this.sampler = null;
	}

	/** The open sampler, opening it first; null (and unusable) when it can't. */
	private async open(): Promise<VideoSampler | null> {
		if (this.sampler) return this.sampler;
		if (!this.usable) return null;
		try {
			this.sampler = await this.openStream();
		} catch (error) {
			if (error instanceof DOMException && error.name === 'AbortError') throw error;
			this.sampler = null;
		}
		if (!this.sampler) this.failures = MAX_SEEK_FAILURES;
		return this.sampler;
	}

	/** Opens the video's stream best suited to sampling, null when it has none. */
	private async openStream(): Promise<VideoSampler | null> {
		const stream = pickStream(this.video.streams);
		if (!stream) return null;
		const playable = await play(stream.url);
		if (!playable) return null;
		return VideoSampler.open(playable.url, playable.mimeType, this.signal);
	}
}

/** The faces of a frame as the backend stores them. */
async function facesOf(frame: Frame, atSeconds: number | null): Promise<ScannedFace[]> {
	const { width, height } = frame;
	const found = await analyzeFrame(frame);
	return found.map((face) => ({
		vector: compactVector(face.vector),
		atSeconds,
		region: regionOf(face.box, width, height)
	}));
}

/** Stores the sheet's perceptual hash and, when the video has none, the sheet as its trickplay. */
async function saveSheet(
	video: ScanVideo,
	sheet: TrickplayCanvas,
	duration: number,
	frames: number
) {
	if (!video.fingerprints.some((fingerprint) => fingerprint.algorithm === 'phash')) {
		await api.addPhash(video.id, sheet.phash(), duration);
	}
	if (video.trickplay) return;
	const uri = await uploadPicture(await sheet.file());
	await api.setTrickplay(video.id, { uri, ...sheet.layout, intervalSeconds: duration / frames });
}

/** Scans a performer's unscanned pictures for their face; returns how many were scanned. */
async function scanPictures(performerId: string, signal: AbortSignal): Promise<number> {
	const pictures = await api.fetchFacePictures(performerId, FACE_MODEL);
	const waiting = pictures.filter((picture) => !picture.scanned).slice(0, MAX_PICTURES);
	for (const [index, picture] of waiting.entries()) {
		await idle(signal);
		faceTracking.set({
			subject: { kind: 'performer', id: performerId },
			phase: 'pictures',
			progress: index / waiting.length
		});
		const frame = await loadPicture(picture.uri, signal);
		if (!frame) continue;
		// A picture of a performer shows them biggest; other faces in it are someone else.
		const faces = (await facesOf(frame, null)).slice(0, 1);
		await api.recordFaceScan({
			kind: 'performer',
			id: performerId,
			model: FACE_MODEL,
			picture: picture.uri,
			frames: 1,
			faces
		});
	}
	return waiting.length;
}

/** Loads a picture as a frame, null when it can't be fetched or decoded. */
async function loadPicture(uri: string, signal: AbortSignal): Promise<Frame | null> {
	const url = cacheUrl(uri);
	if (!url) return null;
	try {
		const response = await fetch(url, { signal });
		if (!response.ok) return null;
		const bitmap = await createImageBitmap(await response.blob());
		const { width, height } = frameSize(bitmap.width, bitmap.height);
		const canvas = new OffscreenCanvas(width, height);
		const context = canvas.getContext('2d');
		if (!context) return null;
		context.drawImage(bitmap, 0, 0, width, height);
		bitmap.close();
		return { rgba: context.getImageData(0, 0, width, height).data, width, height };
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') throw error;
		return null;
	}
}

/** The trickplay sheet the sampled frames are tiled into. */
class TrickplayCanvas {
	private readonly canvas = document.createElement('canvas');
	private readonly context: CanvasRenderingContext2D | null;

	constructor(readonly layout: SheetLayout) {
		this.canvas.width = layout.columns * layout.tileWidth;
		this.canvas.height = layout.rows * layout.tileHeight;
		this.context = this.canvas.getContext('2d', { willReadFrequently: true });
	}

	/** Draws frame `index` into its tile. */
	draw(index: number, source: HTMLVideoElement) {
		const column = index % this.layout.columns;
		const row = Math.floor(index / this.layout.columns);
		this.context?.drawImage(
			source,
			column * this.layout.tileWidth,
			row * this.layout.tileHeight,
			this.layout.tileWidth,
			this.layout.tileHeight
		);
	}

	/** The perceptual hash of the whole sheet. */
	phash(): string {
		const pixels = this.context?.getImageData(0, 0, this.canvas.width, this.canvas.height);
		if (!pixels) throw new Error('No 2D canvas for the sheet');
		return perceptualHash(
			grayscale(pixels.data, pixels.width, pixels.height),
			pixels.width,
			pixels.height
		);
	}

	/** The sheet as a JPEG file to upload. */
	file(): Promise<File> {
		return new Promise((resolve, reject) => {
			this.canvas.toBlob(
				(blob) => {
					if (!blob) {
						reject(new Error('The sheet could not be encoded'));
						return;
					}
					resolve(new File([blob], 'trickplay.jpg', { type: 'image/jpeg' }));
				},
				'image/jpeg',
				0.8
			);
		});
	}
}

/** Waits for the browser to be idle (or 2 s at most), so scanning never competes with the page. */
function idle(signal: AbortSignal): Promise<void> {
	if (typeof requestIdleCallback !== 'function') return delay(100, signal);
	return abortable(signal, (resolve) => {
		const handle = requestIdleCallback(() => resolve(), { timeout: 2000 });
		return () => cancelIdleCallback(handle);
	});
}

/** Waits a while, or until the signal aborts. */
function delay(ms: number, signal: AbortSignal): Promise<void> {
	return abortable(signal, (resolve) => {
		const timer = setTimeout(resolve, ms);
		return () => clearTimeout(timer);
	});
}

/** A wait that rejects with an AbortError when the signal aborts; start returns its cleanup. */
function abortable(signal: AbortSignal, start: (resolve: () => void) => () => void): Promise<void> {
	return new Promise((resolve, reject) => {
		if (signal.aborted) {
			reject(new DOMException('Cancelled', 'AbortError'));
			return;
		}
		const onAbort = () => {
			cleanup();
			reject(new DOMException('Cancelled', 'AbortError'));
		};
		const cleanup = start(() => {
			signal.removeEventListener('abort', onAbort);
			resolve();
		});
		signal.addEventListener('abort', onAbort, { once: true });
	});
}
