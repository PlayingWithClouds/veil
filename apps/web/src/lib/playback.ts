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

export class Playback {
	readonly video: HTMLVideoElement;
	private hls: Hls | null = null;
	private loadedSrc: string | null = null;

	constructor() {
		this.video = document.createElement('video');
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
		onerror?: (message: string) => void
	) {
		if (this.loadedSrc === src) return;
		this.detachSource();
		this.loadedSrc = src;
		if (!isHlsSource(src, mimeType)) {
			this.video.src = src;
			this.video.play().catch(() => {});
			return;
		}
		this.loadHls(src, headers, onerror);
	}

	/** Attaches an HLS source through hls.js, or natively where MSE is missing (Safari/iOS). */
	private loadHls(
		src: string,
		headers: Record<string, string>,
		onerror?: (message: string) => void
	) {
		import('hls.js').then(({ default: HlsClass }) => {
			if (this.loadedSrc !== src) return;
			if (HlsClass.isSupported()) {
				const instance = new HlsClass({
					xhrSetup(xhr) {
						for (const [name, value] of Object.entries(headers)) {
							xhr.setRequestHeader(name, value);
						}
					}
				});
				instance.loadSource(src);
				instance.attachMedia(this.video);
				instance.on(HlsClass.Events.MANIFEST_PARSED, () => {
					this.video.play().catch(() => {});
				});
				instance.on(HlsClass.Events.ERROR, (_, data) => {
					if (data.fatal) onerror?.(data.details ?? 'HLS playback error');
				});
				this.hls = instance;
				return;
			}
			if (!window.MediaSource && this.video.canPlayType('application/vnd.apple.mpegurl')) {
				this.video.src = src;
				this.video.play().catch(() => {});
				return;
			}
			onerror?.('HLS not supported in this browser (missing H.264 codec?)');
		});
	}

	/** Drops the current source and its hls.js instance. */
	private detachSource() {
		this.hls?.destroy();
		this.hls = null;
		this.loadedSrc = null;
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
