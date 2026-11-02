/**
 * What a background face scan of a video samples and makes: evenly spaced
 * frame times, the trickplay sheet the frames tile into, which stream to
 * read, and faces as the backend stores them. Pure, for `bun test`.
 */
import type { FaceBox } from './detect';

/** Frames sampled per video: one 5×5 sheet. */
export const SAMPLE_COUNT = 25;

/** Longest side of a sampled frame, in pixels; the detector looks at 640. */
export const FRAME_SIDE = 640;

/** Width of one trickplay tile, in pixels. */
const TILE_WIDTH = 160;

/** Columns of the trickplay sheet. */
const SHEET_COLUMNS = 5;

/** Streams below this height lose small faces; above it they cost bandwidth for nothing. */
const PREFERRED_HEIGHT = 480;

/**
 * Frame times splitting the video into `count` equal spans, each frame in
 * the middle of its span, so trickplay tile i covers the time around its
 * frame.
 */
export function sampleTimes(durationSeconds: number, count = SAMPLE_COUNT): number[] {
	const interval = durationSeconds / count;
	const times: number[] = [];
	for (let index = 0; index < count; index++) times.push((index + 0.5) * interval);
	return times;
}

/** The grid of a trickplay sheet. */
export interface SheetLayout {
	columns: number;
	rows: number;
	tileWidth: number;
	tileHeight: number;
}

/** The sheet layout for `count` frames of a video of the given size. */
export function sheetLayout(count: number, videoWidth: number, videoHeight: number): SheetLayout {
	const columns = Math.min(SHEET_COLUMNS, count);
	const tileHeight = Math.max(1, Math.round((TILE_WIDTH * videoHeight) / videoWidth));
	return { columns, rows: Math.ceil(count / columns), tileWidth: TILE_WIDTH, tileHeight };
}

/** The pixel size of a frame scaled so its longest side is at most `side`. */
export function frameSize(
	videoWidth: number,
	videoHeight: number,
	side = FRAME_SIDE
): { width: number; height: number } {
	const scale = Math.min(1, side / Math.max(videoWidth, videoHeight));
	return {
		width: Math.max(1, Math.round(videoWidth * scale)),
		height: Math.max(1, Math.round(videoHeight * scale))
	};
}

/** A stream as the video page lists it. */
export interface StreamChoice {
	url: string;
	quality: string | null;
}

/**
 * The stream to sample: the smallest one of at least 480p, else the
 * biggest one below that; streams of unknown quality only when nothing
 * else is there.
 */
export function pickStream<T extends StreamChoice>(streams: T[]): T | null {
	const chosen = samplingHeightIndex(streams.map((stream) => qualityHeight(stream.quality)));
	if (chosen >= 0) return streams[chosen];
	return streams[0] ?? null;
}

/**
 * Which of several heights (0 = unknown) to sample: the smallest of at
 * least 480 lines, else the biggest; -1 when none is known.
 */
export function samplingHeightIndex(heights: number[]): number {
	let chosen = -1;
	for (let index = 0; index < heights.length; index++) {
		if (heights[index] <= 0) continue;
		if (chosen === -1 || isBetterHeight(heights[index], heights[chosen])) chosen = index;
	}
	return chosen;
}

/** Whether `height` suits sampling better than `current`. */
function isBetterHeight(height: number, current: number): boolean {
	if (current < PREFERRED_HEIGHT) return height > current;
	return height >= PREFERRED_HEIGHT && height < current;
}

/** The pixel height a quality label names ("720p", "1080", "4K"), 0 when unknown. */
export function qualityHeight(quality: string | null): number {
	if (!quality) return 0;
	if (/4k/i.test(quality)) return 2160;
	const digits = quality.match(/(\d{3,4})/);
	if (!digits) return 0;
	return Number(digits[1]);
}

/** A face as the backend stores it. */
export interface ScannedFace {
	vector: number[];
	atSeconds: number | null;
	region: FaceRegion | null;
}

/** A face box as shares of its frame, with the detector's score. */
export interface FaceRegion {
	x: number;
	y: number;
	width: number;
	height: number;
	score: number;
}

/** A face box in frame pixels as shares of the frame, clamped to it. */
export function regionOf(box: FaceBox, frameWidth: number, frameHeight: number): FaceRegion {
	const left = Math.max(0, box.x);
	const top = Math.max(0, box.y);
	const right = Math.min(frameWidth, box.x + box.width);
	const bottom = Math.min(frameHeight, box.y + box.height);
	return {
		x: round(left / frameWidth),
		y: round(top / frameHeight),
		width: round((right - left) / frameWidth),
		height: round((bottom - top) / frameHeight),
		score: round(box.score)
	};
}

/** A unit vector rounded to five decimals: no loss that moves a cosine, a third of the JSON. */
export function compactVector(vector: number[]): number[] {
	return vector.map((value) => Math.round(value * 100000) / 100000);
}

/** Rounds a share to four decimals, plenty for a box and short in JSON. */
function round(value: number): number {
	return Math.round(value * 10000) / 10000;
}
