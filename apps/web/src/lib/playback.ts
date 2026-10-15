// A <video> element plus the hls.js instance feeding it, independent of any
// component. Players mount the element into their own box, so one playback can
// move between the scene page, the mini player and a card's hover player
// without reloading: moving a media element within the document keeps it
// playing.
import type Hls from 'hls.js';

/** Whether a source is HLS rather than a progressive file. */
export function isHlsSource(src: string, mimeType: string): boolean {
	return mimeType.includes('mpegURL') || src.includes('.m3u8');
}

/** One quality of an adaptive HLS stream. */
export interface PlaybackLevel {
	index: number;
	// Menu text: the height and bitrate, like "1080p · 5.1 Mbps".
	label: string;
	// Video height in pixels, 0 when the playlist does not say.
	height: number;
	// Bits per second.
	bitrate: number;
}

/** Extra behaviour for loading a source. */
export interface PlaybackOptions {
	// A live broadcast: sits near the live edge, retries network errors before
	// giving up, and starts with the session's audio choice (unmuted unless the
	// user muted it).
	live?: boolean;
	// Live only: the browser refused autoplay with sound, so playback runs muted
	// until the user unmutes.
	onMutedFallback?: () => void;
	// Called with the stream's qualities, best first, once its playlist is read.
	onLevels?: (levels: PlaybackLevel[]) => void;
}

// A live stream retries this many network errors in a row before it is
// reported as over.
const LIVE_NETWORK_RETRIES = 3;
// hls.js settings for a live broadcast: plain (not low-latency) HLS, three
// segments behind the edge.
const LIVE_HLS_CONFIG = { lowLatencyMode: false, liveSyncDurationCount: 3 };

// What the user last chose for the sound of a live stream in this session, so
// the next room (or a reconnect) starts the same way.
const liveAudio = { volume: 1, muted: false };

export class Playback {
	readonly video: HTMLVideoElement;
	private hls: Hls | null = null;
	private loadedSrc: string | null = null;
	// Set while playback runs muted only because autoplay with sound was refused.
	private mutedFallback = false;
	private liveAudioListener: (() => void) | null = null;

	constructor() {
		this.video = document.createElement('video');
		// Streams come through the backend (CORS *); anonymous mode keeps frames
		// readable for "Who is this?" frame captures.
		this.video.crossOrigin = 'anonymous';
		this.video.playsInline = true;
		this.video.preload = 'auto';
	}

	/** The source currently loaded, or null. */
	get src(): string | null {
		return this.loadedSrc;
	}

	/**
	 * Loads src and starts playing it. A no-op when src is already loaded, so a
	 * player adopting a running playback doesn't restart it.
	 */
	load(
		src: string,
		mimeType: string,
		headers: Record<string, string> = {},
		onerror?: (message: string) => void,
		options: PlaybackOptions = {}
	) {
		if (this.loadedSrc === src) return;
		this.detachSource();
		this.loadedSrc = src;
		if (options.live) this.useLiveAudio();
		if (!isHlsSource(src, mimeType)) {
			this.video.src = src;
			this.startPlayback(options);
			return;
		}
		this.loadHls(src, headers, onerror, options);
	}

	/** Starts playing; a live stream falls back to muted when sound autoplay is refused. */
	private startPlayback(options: PlaybackOptions) {
		this.video.play().catch(() => {
			if (!options.live) return;
			this.mutedFallback = true;
			this.video.muted = true;
			options.onMutedFallback?.();
			this.video.play().catch(() => {});
		});
	}

	/** Applies the session's live audio choice and keeps it up to date. */
	private useLiveAudio() {
		this.video.volume = liveAudio.volume;
		this.video.muted = liveAudio.muted;
		this.mutedFallback = false;
		if (this.liveAudioListener) return;
		this.liveAudioListener = () => {
			// Our own muted fallback is not a choice of the user.
			if (this.mutedFallback && this.video.muted) return;
			this.mutedFallback = false;
			liveAudio.volume = this.video.volume;
			liveAudio.muted = this.video.muted;
		};
		this.video.addEventListener('volumechange', this.liveAudioListener);
	}

	/** Seconds the playhead is behind the live edge; 0 when not live or unknown. */
	liveLag(): number {
		const edge = this.hls?.liveSyncPosition;
		if (edge === undefined || edge === null) return 0;
		return Math.max(0, edge - this.video.currentTime);
	}

	/** Jumps to the live edge and plays. */
	seekToLiveEdge() {
		const edge = this.hls?.liveSyncPosition;
		if (edge !== undefined && edge !== null) this.video.currentTime = edge;
		this.video.play().catch(() => {});
	}

	/** Picks a quality by level index, or -1 for automatic. */
	setLevel(index: number) {
		if (this.hls) this.hls.currentLevel = index;
	}

	/** Attaches an HLS source through hls.js, or natively where MSE is missing (Safari/iOS). */
	private loadHls(
		src: string,
		headers: Record<string, string>,
		onerror?: (message: string) => void,
		options: PlaybackOptions = {}
	) {
		import('hls.js').then(({ default: HlsClass }) => {
			if (this.loadedSrc !== src) return;
			if (HlsClass.isSupported()) {
				const instance = new HlsClass({
					...(options.live && LIVE_HLS_CONFIG),
					xhrSetup(xhr) {
						for (const [name, value] of Object.entries(headers)) {
							xhr.setRequestHeader(name, value);
						}
					}
				});
				instance.loadSource(src);
				instance.attachMedia(this.video);
				instance.on(HlsClass.Events.MANIFEST_PARSED, (_, data) => {
					options.onLevels?.(describeLevels(data.levels));
					this.startPlayback(options);
				});
				if (options.live) this.recoverLiveErrors(instance, HlsClass, onerror);
				else {
					instance.on(HlsClass.Events.ERROR, (_, data) => {
						if (data.fatal) onerror?.(data.details ?? 'HLS playback error');
					});
				}
				this.hls = instance;
				return;
			}
			if (!window.MediaSource && this.video.canPlayType('application/vnd.apple.mpegurl')) {
				this.video.src = src;
				this.startPlayback(options);
				return;
			}
			onerror?.('HLS not supported in this browser (missing H.264 codec?)');
		});
	}

	/**
	 * Keeps a live stream going through hiccups: network errors are retried a
	 * few times in a row and media errors recovered; anything else is reported.
	 */
	private recoverLiveErrors(
		instance: Hls,
		HlsClass: typeof Hls,
		onerror?: (message: string) => void
	) {
		let networkRetries = 0;
		instance.on(HlsClass.Events.FRAG_LOADED, () => {
			networkRetries = 0;
		});
		instance.on(HlsClass.Events.ERROR, (_, data) => {
			if (!data.fatal) return;
			if (data.type === HlsClass.ErrorTypes.MEDIA_ERROR) {
				instance.recoverMediaError();
				return;
			}
			if (
				data.type === HlsClass.ErrorTypes.NETWORK_ERROR &&
				networkRetries < LIVE_NETWORK_RETRIES
			) {
				networkRetries += 1;
				instance.startLoad();
				return;
			}
			onerror?.(data.details ?? 'HLS playback error');
		});
	}

	/** Drops the current source and its hls.js instance. */
	private detachSource() {
		this.hls?.destroy();
		this.hls = null;
		this.loadedSrc = null;
		if (this.liveAudioListener) {
			this.video.removeEventListener('volumechange', this.liveAudioListener);
			this.liveAudioListener = null;
		}
		this.video.removeAttribute('src');
	}

	/** Moves the element into container (keeps playing when already in the document). */
	mount(container: HTMLElement) {
		if (this.video.parentElement !== container) container.appendChild(this.video);
	}

	/** Whether the element currently sits in container. */
	isMountedIn(container: HTMLElement | undefined): boolean {
		return container !== undefined && this.video.parentElement === container;
	}

	/** Keeps the element in the document, out of sight, until a player mounts it. */
	park() {
		parkingLot().appendChild(this.video);
	}

	/** Stops playback for good and frees the source. */
	stop() {
		this.video.pause();
		this.detachSource();
		this.video.load();
	}
}

/** The levels of an HLS playlist as menu entries, highest first. */
function describeLevels(levels: Array<{ height?: number; bitrate?: number }>): PlaybackLevel[] {
	const described = levels.map((level, index) => {
		const height = level.height ?? 0;
		const bitrate = level.bitrate ?? 0;
		return { index, label: levelLabel(height, bitrate), height, bitrate };
	});
	described.sort((first, second) => second.height - first.height || second.bitrate - first.bitrate);
	return described;
}

/** "1080p · 5.1 Mbps", or just the bitrate when the height is unknown. */
function levelLabel(height: number, bitrate: number): string {
	const megabits = `${(bitrate / 1_000_000).toFixed(1)} Mbps`;
	if (height > 0) return `${height}p · ${megabits}`;
	return megabits;
}

/**
 * The level for a target height: the tallest at or below it, else the shortest
 * above it. Null when no level has a height or the target is 0 (automatic).
 */
export function pickLevel(levels: PlaybackLevel[], targetHeight: number): PlaybackLevel | null {
	if (targetHeight <= 0) return null;
	const known = levels.filter((level) => level.height > 0);
	const atOrBelow = known.filter((level) => level.height <= targetHeight);
	if (atOrBelow.length > 0) return atOrBelow[0];
	if (known.length === 0) return null;
	return known[known.length - 1];
}

let parking: HTMLElement | null = null;

/** A hidden body-level holder for playbacks between players. */
function parkingLot(): HTMLElement {
	if (!parking) {
		parking = document.createElement('div');
		parking.hidden = true;
		document.body.appendChild(parking);
	}
	return parking;
}

let main: Playback | null = null;

/** The playback that follows the user around: scene page ⇄ mini player. */
export function mainPlayback(): Playback {
	if (!main) main = new Playback();
	return main;
}

/** Makes playback the main one (a card's hover player opening its scene), stopping the old one. */
export function promoteToMain(playback: Playback) {
	if (main === playback) return;
	main?.stop();
	main = playback;
}

/**
 * What a scene page needs to take the main playback over as it runs: the
 * exact stream it plays (resolving again could yield another URL and reload).
 */
export interface PlaybackHandoff {
	sceneId: string;
	title: string;
	src: string;
	mimeType: string;
	// Stream id, so the page's source menu marks the playing source.
	sourceId: string;
	// Seconds; used only when the stream has to load again (e.g. after a reload).
	position: number;
}

let pendingHandoff: PlaybackHandoff | null = null;

/** Hands the main playback to the scene page about to open. */
export function handOffToScene(handoff: PlaybackHandoff) {
	pendingHandoff = handoff;
}

/** The handoff for sceneId, if one is waiting; consumed either way. */
export function takeSceneHandoff(sceneId: string): PlaybackHandoff | null {
	const handoff = pendingHandoff;
	pendingHandoff = null;
	if (!handoff || handoff.sceneId !== sceneId) return null;
	return handoff;
}
