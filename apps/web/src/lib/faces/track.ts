/**
 * Background face tracking. Opening a video scans it once per face model:
 * the credited performers' unscanned pictures first, then the video's
 * stills (cover and the sites' preview frames, plain image fetches). Only
 * when those show too few faces does it seek through 25 evenly spaced
 * frames of the stream, and only while nothing on the page plays. Those
 * frames also make the video's perceptual hash and, when it has none, its
 * trickplay sheet. One record at a time (a newer request cancels the
 * older), models in a worker.
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
	compactVector,
	frameSize,
	pickStream,
	regionOf,
	sampleTimes,
	sheetLayout,
	type ScannedFace,
	type SheetLayout
} from './plan';
import { VideoSampler } from './sampler';

/** How long tracking waits after a video opens, so its player gets the bandwidth first. */
const START_DELAY_MS = 4000;

/** At most this many pictures of one performer are scanned. */
const MAX_PICTURES = 4;

/** Stills showing a face that make seeking through the stream unnecessary. */
const ENOUGH_STILLS_WITH_FACES = 3;

/** How often seeking checks whether the page's playback has stopped. */
const PLAYBACK_POLL_MS = 2000;

/** How a tracking request ended. */
export type TrackOutcome = 'scanned' | 'skipped' | 'cancelled' | 'failed';

/**
 * What tracking is doing now, for a quiet progress hint. `waiting` means the
 * stream's frames wait for playback to pause.
 */
export interface FaceTracking {
	subject: { kind: 'video' | 'performer'; id: string };
	phase: 'pictures' | 'stills' | 'waiting' | 'frames' | 'saving';
	// 0..1 through the phase.
	progress: number;
}

/** The running tracking, null when idle. */
export const faceTracking = writable<FaceTracking | null>(null);

let current: { key: string; controller: AbortController; done: Promise<TrackOutcome> } | null =
	null;

/**
 * Scans a video for faces in the background, unless it's scanned with this
 * model already (then its credited performers' faces are only learned
 * again). Resolves when done; `fetchVideoFaces` then has the guesses.
 */
export function trackVideo(videoId: string): Promise<TrackOutcome> {
	return runExclusive(`video:${videoId}`, (signal) => scanVideo(videoId, signal));
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
 * Scans a video's credited performers' pictures, then (unless done before)
 * its stills, and its stream's frames when the stills show too few faces.
 */
async function scanVideo(videoId: string, signal: AbortSignal): Promise<TrackOutcome> {
	await delay(START_DELAY_MS, signal);
	const video = await api.fetchScanVideo(videoId);
	if (!video) return 'failed';
	for (const performer of video.performers) {
		await scanPictures(performer.id, signal);
	}
	if (await api.hasFaceScan('video', videoId, FACE_MODEL)) {
		await api.learnFaces(videoId, FACE_MODEL);
		return 'skipped';
	}
	const stills = await scanStills(video, signal);
	if (stills.withFaces >= ENOUGH_STILLS_WITH_FACES) {
		await api.recordFaceScan({
			kind: 'video',
			id: video.id,
			model: FACE_MODEL,
			frames: stills.read,
			faces: stills.faces
		});
		return 'scanned';
	}
	const stream = pickStream(video.streams);
	if (!stream) return 'failed';
	const playable = await play(stream.url);
	if (!playable) return 'failed';
	const sampler = await VideoSampler.open(playable.url, playable.mimeType, signal);
	try {
		await scanFrames(video, sampler, stills, signal);
	} finally {
		sampler.close();
	}
	return 'scanned';
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
 * Samples the stream's frames while nothing plays, finds their faces, and
 * records them with the stills' as one scan, plus the hash and sheet.
 */
async function scanFrames(
	video: ScanVideo,
	sampler: VideoSampler,
	stills: StillsScan,
	signal: AbortSignal
) {
	let duration = sampler.duration;
	if (duration <= 0) duration = video.durationSeconds ?? 0;
	if (duration <= 0) throw new Error('The video has no length to sample');
	const times = sampleTimes(duration);
	const sheet = new TrickplayCanvas(
		sheetLayout(times.length, sampler.size.width, sampler.size.height)
	);
	const faces: ScannedFace[] = [...stills.faces];
	let frames = 0;
	for (const [index, seconds] of times.entries()) {
		await quietPlayback(video.id, index / times.length, signal);
		await idle(signal);
		const frame = await sampleFrame(sampler, seconds, signal, (source) =>
			sheet.draw(index, source)
		);
		faceTracking.set({
			subject: { kind: 'video', id: video.id },
			phase: 'frames',
			progress: (index + 1) / times.length
		});
		if (!frame) continue;
		frames++;
		faces.push(...(await facesOf(frame, Math.round(seconds * 100) / 100)));
	}
	if (frames < times.length / 2)
		throw new Error(`Only ${frames} of ${times.length} frames could be read`);
	faceTracking.set({ subject: { kind: 'video', id: video.id }, phase: 'saving', progress: 0 });
	if (frames === times.length) await saveSheet(video, sheet, duration, times.length);
	await api.recordFaceScan({
		kind: 'video',
		id: video.id,
		model: FACE_MODEL,
		frames: frames + stills.read,
		faces
	});
}

/**
 * Waits until no video on the page plays (paused or ended), so seeking
 * never takes bandwidth from what's being watched. The sampler's own
 * element isn't in the page.
 */
async function quietPlayback(videoId: string, progress: number, signal: AbortSignal) {
	while (isPlaying()) {
		faceTracking.set({ subject: { kind: 'video', id: videoId }, phase: 'waiting', progress });
		await delay(PLAYBACK_POLL_MS, signal);
	}
}

/** Whether any video in the page is playing. */
function isPlaying(): boolean {
	return [...document.querySelectorAll('video')].some(
		(element) => !element.paused && !element.ended
	);
}

/**
 * The frame at a time, null when the seek times out; a cancel or a stream
 * that forbids reading its pixels ends the scan.
 */
async function sampleFrame(
	sampler: VideoSampler,
	seconds: number,
	signal: AbortSignal,
	draw: (source: HTMLVideoElement) => void
): Promise<Frame | null> {
	try {
		return await sampler.frameAt(seconds, signal, draw);
	} catch (error) {
		if (error instanceof DOMException) throw error;
		return null;
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
