/**
 * Reads frames out of a video without showing it: a detached, muted video
 * element seeks to each time and the frame is drawn to a canvas. Streams come
 * through the backend's proxy, which allows reading their pixels
 * (crossorigin); HLS goes through hls.js where the browser can't play it.
 */
import type Hls from 'hls.js';
import { isHlsSource } from '$lib/playback';
import { frameSize, samplingHeightIndex } from './plan';
import type { Frame } from './engine';

/** How long opening the stream may take. */
const OPEN_TIMEOUT_MS = 30_000;

/** How long one seek may take before the frame is skipped. */
const SEEK_TIMEOUT_MS = 20_000;

/** How far a seek may land from the asked time. */
const SEEK_TOLERANCE_SECONDS = 2;

/** A video opened for sampling. */
export class VideoSampler {
	private hls: Hls | null = null;
	private readonly canvas = document.createElement('canvas');

	private constructor(private readonly element: HTMLVideoElement) {}

	/** Opens a stream for sampling; fails when it doesn't load or the signal aborts. */
	static async open(url: string, mimeType: string, signal: AbortSignal): Promise<VideoSampler> {
		const element = document.createElement('video');
		element.muted = true;
		element.playsInline = true;
		element.crossOrigin = 'anonymous';
		// Only what each seek needs: reading ahead would download stretches
		// between the sampled frames.
		element.preload = 'metadata';
		const sampler = new VideoSampler(element);
		try {
			await sampler.attach(url, mimeType);
			await once(element, 'loadedmetadata', OPEN_TIMEOUT_MS, signal);
		} catch (error) {
			sampler.close();
			throw error;
		}
		return sampler;
	}

	/** The video's length in seconds, 0 when the stream doesn't say (live). */
	get duration(): number {
		if (!Number.isFinite(this.element.duration)) return 0;
		return this.element.duration;
	}

	/** The video's pixel size. */
	get size(): { width: number; height: number } {
		return { width: this.element.videoWidth, height: this.element.videoHeight };
	}

	/** Plays src in the element, through hls.js where the browser can't play HLS itself. */
	private async attach(url: string, mimeType: string) {
		const nativeHls = this.element.canPlayType('application/vnd.apple.mpegurl') !== '';
		if (!isHlsSource(url, mimeType) || nativeHls) {
			this.element.src = url;
			return;
		}
		const { default: HlsClass } = await import('hls.js');
		// One segment around each seek is all a frame needs.
		const hls = new HlsClass({ maxBufferLength: 1, maxMaxBufferLength: 2, autoStartLoad: true });
		hls.on(HlsClass.Events.MANIFEST_PARSED, () => {
			hls.currentLevel = samplingHeightIndex(hls.levels.map((level) => level.height));
		});
		hls.loadSource(url);
		hls.attachMedia(this.element);
		this.hls = hls;
	}

	/**
	 * Seeks to a time and returns the frame there, scaled to the detector's
	 * size; `draw` may also paint the full frame elsewhere (a trickplay tile).
	 * Throws when the stream doesn't allow reading its pixels.
	 */
	async frameAt(
		seconds: number,
		signal: AbortSignal,
		draw?: (source: HTMLVideoElement) => void
	): Promise<Frame> {
		const seeked = once(this.element, 'seeked', SEEK_TIMEOUT_MS, signal);
		this.element.currentTime = seconds;
		await seeked;
		// A stream served without range requests can't seek and lands back at the start.
		if (Math.abs(this.element.currentTime - seconds) > SEEK_TOLERANCE_SECONDS) {
			throw new Error(`The video can't seek to ${seconds} s`);
		}
		const { width, height } = frameSize(this.element.videoWidth, this.element.videoHeight);
		this.canvas.width = width;
		this.canvas.height = height;
		const context = this.canvas.getContext('2d', { willReadFrequently: true });
		if (!context) throw new Error('No 2D canvas to draw frames on');
		context.drawImage(this.element, 0, 0, width, height);
		draw?.(this.element);
		const pixels = context.getImageData(0, 0, width, height);
		return { rgba: pixels.data, width, height };
	}

	/** Stops loading and lets go of the stream. */
	close() {
		this.hls?.destroy();
		this.hls = null;
		this.element.removeAttribute('src');
		this.element.load();
	}
}

/** Waits for an element's event; rejects on its error, the timeout or the signal. */
function once(
	element: HTMLVideoElement,
	name: string,
	timeoutMs: number,
	signal: AbortSignal
): Promise<void> {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(
			() => finish(new Error(`The video took too long (${name})`)),
			timeoutMs
		);
		const onEvent = () => finish(null);
		const onError = () => finish(new Error('The video could not be read'));
		const onAbort = () => finish(new DOMException('Cancelled', 'AbortError'));
		/** Removes the listeners and settles the promise. */
		function finish(error: Error | null) {
			clearTimeout(timer);
			element.removeEventListener(name, onEvent);
			element.removeEventListener('error', onError);
			signal.removeEventListener('abort', onAbort);
			if (error) {
				reject(error);
				return;
			}
			resolve();
		}
		if (signal.aborted) {
			onAbort();
			return;
		}
		element.addEventListener(name, onEvent);
		element.addEventListener('error', onError);
		signal.addEventListener('abort', onAbort);
	});
}
