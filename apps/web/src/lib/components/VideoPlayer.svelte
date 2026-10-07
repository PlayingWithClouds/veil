<script lang="ts">
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import SpeakerXIcon from 'phosphor-svelte/lib/SpeakerXIcon';
	import ArrowsOutIcon from 'phosphor-svelte/lib/ArrowsOutIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
	import FastForwardIcon from 'phosphor-svelte/lib/FastForwardIcon';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import ActionSheet from './ActionSheet.svelte';
	import SiteIcon from './scene/SiteIcon.svelte';
	import { formatClock } from './SceneCard.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import { portal } from '$lib/portal';
	import { recordProgress } from '$lib/discovery';
	import { sites, siteDisplayName } from '$lib/stores/sites';
	import { miniPlayer } from '$lib/stores/miniPlayer';
	import { isHlsSource, mainPlayback, Playback } from '$lib/playback';
	import { matchOrientationToVideo } from '$lib/native';
	import { untrack } from 'svelte';
	import { get } from 'svelte/store';

	// Lightweight recommendation card shown in the fullscreen up-next sheet.
	// Kept source-agnostic so the player stays decoupled from the nsfw feed.
	export interface PlayerRec {
		id: string;
		title: string;
		image: string | null;
		durationSeconds?: number | null;
		siteName?: string | null;
		siteIconUrl?: string | null;
	}

	// Timeline marker shown as a tick (or a band, when it spans a range) on the
	// seek bar. personal distinguishes the user's own markers from global
	// plugin-created annotations.
	export interface PlayerMarker {
		id: string;
		label: string;
		seconds: number;
		endSeconds?: number | null;
		personal?: boolean;
	}

	// One selectable playback source, grouped by provider in the quality menu.
	export interface PlayerSource {
		id: string;
		provider: string;
		quality: string;
		verified?: boolean;
		speedBps?: number | null;
	}

	interface Props {
		src: string;
		mimeType: string;
		title: string;
		headers?: Record<string, string>;
		// Selectable sources (all providers/qualities) + which one is playing.
		// Switching from the in-player menu calls onselectsource.
		sources?: PlayerSource[];
		activeSourceId?: string;
		onselectsource?: (source: PlayerSource) => void;
		// Resolves a source id to a playable URL, used to prebuffer a higher quality
		// before auto-upgrading. Without it, only auto-downgrade is available.
		onresolvesource?: (sourceId: string) => Promise<{ url: string; mimeType: string } | null>;
		// Media record id (movie:… / episode:…). When set, playback progress is
		// recorded for resume + recommendations.
		mediaId?: string;
		// Seconds to resume playback from on load (saved watch position). 0 = start.
		startPosition?: number;
		// Inline players embed in the page (aspect-video box) instead of taking
		// over the viewport as a fullscreen modal.
		inline?: boolean;
		// Square corners for an inline player that spans the full screen width.
		square?: boolean;
		// Frame shown (dimmed) while the video loads, before its first play.
		poster?: string | null;
		// Swiping the inline player down (phones): e.g. leave to the mini player.
		onswipedown?: () => void;
		// A back button over the player, for pages that hide the top bar.
		onback?: () => void;
		// Recommendations revealed by swiping up (or scrolling) while fullscreen. Empty = off.
		recommendations?: PlayerRec[];
		onselectrec?: (rec: PlayerRec) => void;
		// Called when the fullscreen grid nears its bottom, to page in more recs.
		onloadmore?: () => void;
		// Timeline markers to draw on the seek bar. Empty = none.
		markers?: PlayerMarker[];
		// Called with the current time when the user adds a marker. Absent = hide button.
		onaddmarker?: (seconds: number) => void;
		// Play through the main playback, which the mini player takes over when
		// the user leaves the page.
		shared?: boolean;
		onclose: () => void;
	}

	let {
		src,
		mimeType,
		title,
		headers = {},
		sources = [],
		activeSourceId,
		onselectsource,
		onresolvesource,
		mediaId,
		startPosition = 0,
		inline = false,
		square = false,
		poster = null,
		onswipedown,
		onback,
		recommendations = [],
		onselectrec,
		onloadmore,
		markers = [],
		onaddmarker,
		shared = false,
		onclose
	}: Props = $props();

	// Sources grouped by provider for the in-player quality menu.
	let sourceGroups = $derived.by(() => {
		const groups = new Map<string, PlayerSource[]>();
		for (const source of sources) {
			const list = groups.get(source.provider) ?? [];
			list.push(source);
			groups.set(source.provider, list);
		}
		return [...groups.entries()].map(([provider, list]) => ({ provider, sources: list }));
	});
	let activeSource = $derived(sources.find((source) => source.id === activeSourceId));
	// The ⚙ menu: a popover on desktop, a sheet on phones.
	let settingsOpen = $state(false);

	// Playback position to restore after a source switch reloads the <video>.
	let switchResume = $state(0);

	// Auto quality: adapt up/down with network health until the user picks a
	// source by hand.
	let autoAdapt = $state(true);

	function applySwitch(source: PlayerSource, resumeAt?: number) {
		switchResume = resumeAt ?? videoEl?.currentTime ?? 0;
		onselectsource?.(source);
	}

	// Manual pick from the menu — pins the quality and disables auto-adaptation.
	function selectSource(source: PlayerSource) {
		settingsOpen = false;
		if (source.id === activeSourceId) return;
		autoAdapt = false;
		abortPrebuffer();
		applySwitch(source);
	}

	// Adaptable quality ladder for the active provider, best-first, excluding the
	// adaptive HLS entry (hls.js already bitrate-adapts on its own).
	let qualityLadder = $derived.by(() => {
		if (!activeSource) return [] as PlayerSource[];
		const group = sourceGroups.find((entry) => entry.provider === activeSource!.provider);
		if (!group) return [] as PlayerSource[];
		return group.sources.filter((source) => !/auto/i.test(source.quality));
	});
	let ladderIndex = $derived(qualityLadder.findIndex((source) => source.id === activeSourceId));

	// --- Adaptive quality ------------------------------------------------------
	// Buffering longer than this drops one quality step.
	const SLOW_LOAD_MS = 6000;
	// Buffer-ahead (seconds) considered healthy enough to consider an upgrade.
	const HEALTHY_BUFFER_SECONDS = 18;
	// A higher quality must prebuffer at least this many seconds before switching.
	const PREBUFFER_SECONDS = 8;
	// Give up prebuffering (too slow → stay put) after this long.
	const PREBUFFER_TIMEOUT_MS = 20000;
	// Backoff between adaptive switches so we don't oscillate.
	const ADAPT_COOLDOWN_MS = 15000;

	let prebufferEl = $state<HTMLVideoElement | undefined>();
	let prebufferingId = $state<string | null>(null);
	let lastAdaptAt = 0;

	// Seconds of media buffered ahead of the current playhead.
	function bufferAhead(): number {
		if (!videoEl) return 0;
		const time = videoEl.currentTime;
		for (let i = 0; i < videoEl.buffered.length; i++) {
			if (videoEl.buffered.start(i) <= time && time <= videoEl.buffered.end(i)) {
				return videoEl.buffered.end(i) - time;
			}
		}
		return 0;
	}

	function nowMs(): number {
		return performance.now();
	}

	function abortPrebuffer() {
		prebufferingId = null;
		if (prebufferEl) {
			prebufferEl.pause();
			prebufferEl.removeAttribute('src');
			prebufferEl.load();
		}
	}

	// Downgrade watchdog: sustained buffering steps down one quality.
	$effect(() => {
		if (!autoAdapt || !buffering) return;
		const timer = setTimeout(() => {
			if (ladderIndex < 0 || ladderIndex >= qualityLadder.length - 1) return;
			lastAdaptAt = nowMs();
			abortPrebuffer();
			applySwitch(qualityLadder[ladderIndex + 1]);
		}, SLOW_LOAD_MS);
		return () => clearTimeout(timer);
	});

	// Upgrade loop: while playback is healthy, prebuffer the next higher quality
	// and switch to it once enough of it is buffered.
	$effect(() => {
		if (!autoAdapt || !onresolvesource) return;
		const interval = setInterval(evaluateUpgrade, 3000);
		return () => clearInterval(interval);
	});

	function evaluateUpgrade() {
		if (!playing || buffering || prebufferingId) return;
		if (nowMs() - lastAdaptAt < ADAPT_COOLDOWN_MS) return;
		if (ladderIndex <= 0) return; // already at the best quality
		if (bufferAhead() < HEALTHY_BUFFER_SECONDS) return;
		startPrebuffer(qualityLadder[ladderIndex - 1]);
	}

	async function startPrebuffer(target: PlayerSource) {
		if (!onresolvesource || !prebufferEl) return;
		prebufferingId = target.id;
		const resolved = await onresolvesource(target.id);
		// Bail if resolution failed, the user took over, or the target changed.
		if (!resolved || !autoAdapt || prebufferingId !== target.id || !prebufferEl) {
			if (prebufferingId === target.id) abortPrebuffer();
			return;
		}
		// HLS targets adapt on their own; prebuffer only progressive sources.
		if (resolved.mimeType.includes('mpegURL') || resolved.url.includes('.m3u8')) {
			abortPrebuffer();
			return;
		}

		const probe = prebufferEl;
		probe.src = resolved.url;
		probe.currentTime = videoEl?.currentTime ?? 0;
		// Play muted off-screen so it actually buffers ahead (a paused element
		// stops loading early). Position is re-synced at the swap.
		probe.muted = true;
		probe.play().catch(() => {});
		const deadline = nowMs() + PREBUFFER_TIMEOUT_MS;

		const poll = setInterval(() => {
			const stillWanted = autoAdapt && prebufferingId === target.id;
			if (!stillWanted) {
				clearInterval(poll);
				return;
			}
			if (bufferedSeconds(probe) >= PREBUFFER_SECONDS) {
				clearInterval(poll);
				beginSeamlessSwap(target, probe);
			} else if (nowMs() > deadline) {
				// Too slow to prebuffer a higher quality — stay where we are.
				clearInterval(poll);
				abortPrebuffer();
			}
		}, 1000);
	}

	// Crossfade to the prebuffered higher quality: the probe (already buffered at
	// the current position) is shown as an overlay and plays the audio, hiding the
	// main element while it reloads the higher source underneath. Once the main
	// element is playing the new source, the overlay fades back out. This avoids
	// the black flash of reloading the visible <video> directly.
	let overlayActive = $state(false);
	let pendingSeamless = $state(false);
	// Set once the main element is told to load the upgraded source, so the
	// canplay that follows is for the new source (not the one under the overlay).
	// Plain (non-reactive) so reading it in the load effect doesn't retrigger it.
	let seamlessArmed = false;
	let priorMuted = false;

	function beginSeamlessSwap(target: PlayerSource, probe: HTMLVideoElement) {
		if (!videoEl) {
			abortPrebuffer();
			return;
		}
		lastAdaptAt = nowMs();
		prebufferingId = null;

		const position = videoEl.currentTime;
		probe.currentTime = position;
		probe.volume = videoEl.volume;
		// The overlay carries the audio through the swap; silence the reloading main.
		priorMuted = videoEl.muted;
		probe.muted = priorMuted;
		videoEl.muted = true;

		probe
			.play()
			.then(() => {
				overlayActive = true;
				pendingSeamless = true;
				seamlessArmed = true;
				// Only reload the main element once the opaque overlay has actually
				// composited, so its black reload is never visible through it.
				requestAnimationFrame(() =>
					requestAnimationFrame(() => {
						applySwitch(target, probe.currentTime);
						// Safety net: never leave the overlay stuck if main never recovers.
						setTimeout(hardFinalizeSeamless, 10000);
					})
				);
			})
			.catch(() => {
				// Overlay playback blocked — fall back to a plain (flashing) switch.
				if (videoEl) videoEl.muted = priorMuted;
				applySwitch(target, position);
				abortPrebuffer();
			});
	}

	// Swap once the main element can play the upgraded source: freeze the overlay
	// on its current frame, seek the main element to that exact frame, and uncover
	// only after the main element has actually painted it — the same content at the
	// same timestamp is on screen across the swap, so there is no flash and no jump.
	function swapToMainWhenReady() {
		if (!pendingSeamless || !videoEl || !prebufferEl) return;
		const main = videoEl;
		const overlay = prebufferEl;

		const frame = overlay.currentTime;
		overlay.pause(); // freeze the overlay on this frame
		main.currentTime = frame;
		main.muted = priorMuted;
		main.play().catch(() => {});

		waitForPaintedFrame(main, frame, () => {
			pendingSeamless = false;
			// Hide the overlay instantly on the matched frame, then blank the probe
			// only after that hide has painted (so its cleared src never shows).
			overlayActive = false;
			requestAnimationFrame(() => abortPrebuffer());
		});
	}

	// Resolves once the video has painted a frame at (or past) target seconds.
	function waitForPaintedFrame(video: HTMLVideoElement, target: number, done: () => void) {
		type FrameVideo = HTMLVideoElement & {
			requestVideoFrameCallback?: (
				cb: (now: number, meta: { mediaTime: number }) => void
			) => number;
		};
		const withFrame = video as FrameVideo;
		if (typeof withFrame.requestVideoFrameCallback === 'function') {
			const onFrame = (_now: number, meta: { mediaTime: number }) => {
				if (meta.mediaTime >= target - 0.05) {
					done();
					return;
				}
				withFrame.requestVideoFrameCallback?.(onFrame);
			};
			withFrame.requestVideoFrameCallback(onFrame);
			return;
		}
		// Fallback for browsers without rVFC: uncover shortly after the seek settles.
		const onSeeked = () => {
			video.removeEventListener('seeked', onSeeked);
			setTimeout(done, 80);
		};
		video.addEventListener('seeked', onSeeked);
	}

	// Force the overlay down if the seamless path stalls, so it never sticks.
	function hardFinalizeSeamless() {
		if (!pendingSeamless) return;
		pendingSeamless = false;
		seamlessArmed = false;
		if (videoEl) videoEl.muted = priorMuted;
		overlayActive = false;
		abortPrebuffer();
	}

	function bufferedSeconds(element: HTMLVideoElement): number {
		let total = 0;
		for (let i = 0; i < element.buffered.length; i++) {
			total += element.buffered.end(i) - element.buffered.start(i);
		}
		return total;
	}

	function jumpToMarker(seconds: number) {
		if (!videoEl) return;
		videoEl.currentTime = seconds;
		showControls();
	}

	// Exposed for the parent (markers list) to seek the active player directly.
	export function seekTo(seconds: number) {
		if (!videoEl) return;
		videoEl.currentTime = seconds;
		videoEl.play().catch(() => {});
		showControls();
	}

	/** Current position and whether playback is running, for handing over to the mini player. */
	export function playback(): { position: number; playing: boolean; muted: boolean } {
		if (!videoEl) return { position: 0, playing: false, muted: false };
		return {
			position: videoEl.currentTime,
			playing: !videoEl.paused && !videoEl.ended,
			muted: videoEl.muted
		};
	}

	function addMarkerHere() {
		if (!onaddmarker || !videoEl) return;
		onaddmarker(videoEl.currentTime);
	}

	// Seconds skipped per left/right arrow press.
	const SEEK_STEP_SECONDS = 5;
	// Progress is saved at most every 15s while playing, plus on pause/close.
	const SAVE_INTERVAL_SECONDS = 15;
	const COMPLETE_FRACTION = 0.9;
	let lastSavedAt = 0;

	// Resume from the saved position once per source, when metadata is ready. Skip
	// when near the very end so a finished video restarts instead of resuming to
	// the credits.
	let seekedForSrc = $state('');
	function seekToStart() {
		if (!videoEl) return;
		// A source switch reloads the <video>; restore the prior position instead
		// of resuming from the saved watch point.
		if (switchResume > 0) {
			videoEl.currentTime = switchResume;
			switchResume = 0;
			seekedForSrc = src;
			return;
		}
		if (seekedForSrc === src) return;
		seekedForSrc = src;
		if (startPosition <= 0) return;
		const total = videoEl.duration;
		if (total > 0 && startPosition >= total * COMPLETE_FRACTION) return;
		videoEl.currentTime = startPosition;
	}

	function saveProgress(force = false) {
		if (!mediaId || !videoEl) return;
		const position = videoEl.currentTime;
		if (!force && position - lastSavedAt < SAVE_INTERVAL_SECONDS) return;
		lastSavedAt = position;
		const total = videoEl.duration;
		const completed = total > 0 && position >= total * COMPLETE_FRACTION;
		recordProgress({
			mediaId,
			progressSeconds: position,
			durationSeconds: isFinite(total) ? total : undefined,
			completed
		});
	}

	function close() {
		saveProgress(true);
		onclose();
	}

	// The playing element comes from a Playback mounted into videoSlot; with
	// `shared` it is the main playback that carries on in the mini player.
	let activePlayback = $state<Playback | undefined>();
	let videoEl = $state<HTMLVideoElement | undefined>();
	let videoSlot = $state<HTMLDivElement | undefined>();
	let containerEl = $state<HTMLDivElement | undefined>();

	let playing = $state(false);
	let currentTime = $state(0);
	let duration = $state(0);
	let volume = $state(1);
	let muted = $state(false);
	let buffering = $state(true);
	let error = $state<string | null>(null);
	let controlsVisible = $state(true);
	let hideTimer: ReturnType<typeof setTimeout>;
	// Until the first play only a big play button shows over the frame.
	let hasPlayed = $state(false);
	// Chosen speed; a held finger plays at HOLD_SPEED until it lets go.
	let playbackRate = $state(1);
	let holdingFast = $state(false);
	// True while a finger drags the seek bar: it grows and shows the preview.
	let scrubbing = $state(false);

	// Played fraction, used to paint the accent fill on the native range tracks
	// (cross-browser, no ::-webkit-progress support needed).
	let seekPercent = $derived(duration > 0 ? (currentTime / duration) * 100 : 0);
	let volumePercent = $derived((muted ? 0 : volume) * 100);

	// Fullscreen "up next": swiping up (or the wheel) pulls a glass sheet with a
	// row of recs over the video, which keeps playing dimmed. revealProgress
	// 0 = hidden, 1 = fully up.
	let isFullscreen = $state(false);
	let revealProgress = $state(0);
	let recoRowEl = $state<HTMLDivElement | undefined>();
	let recoSheetHeight = $state(240);
	// The finger drives the sheet directly; transitions only on release.
	let revealDragging = $state(false);
	let hasRecs = $derived(recommendations.length > 0);

	function clamp(value: number, min: number, max: number): number {
		return Math.min(max, Math.max(min, value));
	}

	// Seek-bar hover scrub: a muted mirror <video> is seeked to the hovered time
	// and shown as a thumbnail above the cursor.
	let previewVideoEl = $state<HTMLVideoElement | undefined>();
	let seekWrapEl = $state<HTMLDivElement | undefined>();
	let hoverActive = $state(false);
	let hoverTime = $state(0);
	let hoverX = $state(0);

	function onSeekHover(event: MouseEvent) {
		if (!seekWrapEl || duration <= 0) return;
		const rect = seekWrapEl.getBoundingClientRect();
		const ratio = clamp((event.clientX - rect.left) / rect.width, 0, 1);
		hoverX = ratio * rect.width;
		hoverTime = ratio * duration;
		hoverActive = true;
		// Skip while a prior seek is in flight so a fast drag doesn't thrash.
		if (previewVideoEl && !previewVideoEl.seeking && isFinite(hoverTime)) {
			previewVideoEl.currentTime = hoverTime;
		}
	}

	function onSeekLeave() {
		hoverActive = false;
	}

	/** Pages in more recs as the up-next row nears its end. */
	function onRecoScroll() {
		if (!recoRowEl || !onloadmore) return;
		const remaining = recoRowEl.scrollWidth - recoRowEl.scrollLeft - recoRowEl.clientWidth;
		if (remaining < 600) onloadmore();
	}

	// Controls fade out this long after the last interaction while playing.
	const HIDE_DELAY_MS = 2500;

	/** Shows the controls and restarts the timer that hides them again while playing. */
	function showControls() {
		controlsVisible = true;
		clearTimeout(hideTimer);
		hideTimer = setTimeout(() => {
			if (playing && !scrubbing && !settingsOpen) controlsVisible = false;
		}, HIDE_DELAY_MS);
	}

	/** Hides the controls right away. */
	function hideControls() {
		clearTimeout(hideTimer);
		controlsVisible = false;
	}

	// Mouse events a touch emulates arrive within this window; they must not
	// reveal the controls the tap just hid.
	const EMULATED_MOUSE_MS = 800;
	let lastTouchAt = 0;

	/** Reveals the controls on real mouse movement. */
	function onMouseMove() {
		if (performance.now() - lastTouchAt < EMULATED_MOUSE_MS) return;
		showControls();
	}

	// --- Touch gestures on the video -----------------------------------------
	// Two taps on the same side within this window skip; more taps keep skipping.
	const DOUBLE_TAP_MS = 280;
	const DOUBLE_TAP_SKIP_SECONDS = 10;
	// Holding a finger this long plays at HOLD_SPEED until it lets go.
	const HOLD_MS = 450;
	const HOLD_SPEED = 2;
	// Movement (px) that makes a touch a drag instead of a tap or hold.
	const TOUCH_SLOP = 10;

	type TapZone = 'back' | 'middle' | 'forward';

	let lastPointerType = 'mouse';
	let lastTapAt = 0;
	let lastTapZone: TapZone = 'middle';
	let singleTapTimer: ReturnType<typeof setTimeout> | undefined;
	let holdTimer: ReturnType<typeof setTimeout> | undefined;
	let touchStartX = 0;
	let touchStartY = 0;
	// The click after a hold must not count as a tap.
	let swallowNextClick = false;

	// The side that was double-tapped and how far it skipped, for the ripple.
	let skipFeedback = $state<{ zone: TapZone; seconds: number; key: number } | null>(null);
	let skipFeedbackTimer: ReturnType<typeof setTimeout> | undefined;

	/** Starts the hold-for-2× timer on a touch. */
	function onVideoPointerDown(event: PointerEvent) {
		lastPointerType = event.pointerType;
		if (event.pointerType !== 'touch') return;
		lastTouchAt = performance.now();
		touchStartX = event.clientX;
		touchStartY = event.clientY;
		clearTimeout(holdTimer);
		if (!playing) return;
		holdTimer = setTimeout(startHoldSpeed, HOLD_MS);
	}

	/** A finger that moves is dragging, not holding. */
	function onVideoPointerMove(event: PointerEvent) {
		if (event.pointerType !== 'touch' || holdingFast) return;
		const moved = Math.hypot(event.clientX - touchStartX, event.clientY - touchStartY);
		if (moved > TOUCH_SLOP) clearTimeout(holdTimer);
	}

	/** Ends a hold: back to the chosen speed. */
	function onVideoPointerUp(event: PointerEvent) {
		if (event.pointerType !== 'touch') return;
		lastTouchAt = performance.now();
		clearTimeout(holdTimer);
		if (!holdingFast) return;
		holdingFast = false;
		swallowNextClick = true;
		if (videoEl) videoEl.playbackRate = playbackRate;
	}

	/** Plays at HOLD_SPEED while the finger stays down. */
	function startHoldSpeed() {
		if (!videoEl) return;
		holdingFast = true;
		hideControls();
		videoEl.playbackRate = HOLD_SPEED;
		navigator.vibrate?.(10);
	}

	/** Which third of the video a tap landed in. */
	function tapZone(event: MouseEvent): TapZone {
		const target = event.currentTarget as HTMLElement;
		const rect = target.getBoundingClientRect();
		const fraction = (event.clientX - rect.left) / rect.width;
		if (fraction < 1 / 3) return 'back';
		if (fraction > 2 / 3) return 'forward';
		return 'middle';
	}

	/**
	 * Mouse clicks toggle playback. Taps show or hide the controls; a quick second
	 * tap on the left or right third skips instead.
	 */
	function onVideoClick(event: MouseEvent) {
		// A tap on the dimmed video above the up-next sheet closes it.
		if (revealProgress > 0) {
			revealProgress = 0;
			return;
		}
		if (lastPointerType !== 'touch') {
			togglePlay();
			return;
		}
		if (swallowNextClick) {
			swallowNextClick = false;
			return;
		}
		const zone = tapZone(event);
		const now = performance.now();
		const isRepeat = now - lastTapAt < DOUBLE_TAP_MS && zone === lastTapZone;
		lastTapAt = now;
		lastTapZone = zone;
		if (isRepeat && zone !== 'middle') {
			clearTimeout(singleTapTimer);
			skipFromTap(zone);
			return;
		}
		// A skip streak stays on while taps keep coming.
		if (skipFeedback && skipFeedback.zone === zone) {
			skipFromTap(zone);
			return;
		}
		clearTimeout(singleTapTimer);
		if (zone === 'middle') {
			toggleControls();
			return;
		}
		singleTapTimer = setTimeout(toggleControls, DOUBLE_TAP_MS);
	}

	/** Shows hidden controls, hides shown ones. */
	function toggleControls() {
		if (controlsVisible) {
			hideControls();
			return;
		}
		showControls();
	}

	/** Skips 10 s for a double tap and shows the ripple on that side. */
	function skipFromTap(zone: TapZone) {
		if (!videoEl) return;
		let step = DOUBLE_TAP_SKIP_SECONDS;
		if (zone === 'back') step = -DOUBLE_TAP_SKIP_SECONDS;
		const total = duration || videoEl.duration || 0;
		videoEl.currentTime = clamp(videoEl.currentTime + step, 0, total);
		let seconds = DOUBLE_TAP_SKIP_SECONDS;
		if (skipFeedback && skipFeedback.zone === zone) seconds = skipFeedback.seconds + DOUBLE_TAP_SKIP_SECONDS;
		skipFeedback = { zone, seconds, key: performance.now() };
		clearTimeout(skipFeedbackTimer);
		skipFeedbackTimer = setTimeout(() => (skipFeedback = null), 650);
	}

	/** Double clicks with a mouse toggle fullscreen; double taps skip instead. */
	function onVideoDoubleClick() {
		if (lastPointerType === 'touch') return;
		toggleFullscreen();
	}

	const SPEEDS = [0.5, 0.75, 1, 1.25, 1.5, 2];

	/** Where the ⚙ sheet goes: the player while fullscreen, else <body> (null). */
	function fullscreenHost(fullscreen: boolean): HTMLElement | null {
		if (fullscreen && containerEl) return containerEl;
		return null;
	}

	/** Back to automatic quality from the ⚙ menu. */
	function chooseAutoQuality() {
		autoAdapt = true;
		settingsOpen = false;
	}

	/** Adds a marker at the current time from the ⚙ menu. */
	function addMarkerFromMenu() {
		settingsOpen = false;
		addMarkerHere();
	}

	/** The ⚙ sheet's subtitle: what is playing and how. */
	function qualitySummary(): string {
		const parts: string[] = [];
		if (autoAdapt) parts.push('Auto');
		if (activeSource) parts.push(activeSource.quality);
		if (playbackRate !== 1) parts.push(`${playbackRate}×`);
		return parts.join(' · ');
	}

	/** Sets the playback speed from the ⚙ menu. */
	function setSpeed(rate: number) {
		playbackRate = rate;
		if (!videoEl) return;
		// The default carries the speed over a source switch's reload.
		videoEl.defaultPlaybackRate = rate;
		videoEl.playbackRate = rate;
	}

	// --- Seek bar dragging -----------------------------------------------------
	/** A finger on the seek bar keeps the controls up and shows the preview. */
	function onSeekPointerDown(event: PointerEvent) {
		if (event.pointerType !== 'touch') return;
		scrubbing = true;
		clearTimeout(hideTimer);
	}

	/** Lifting the finger off the seek bar hides the preview again. */
	function onSeekPointerUp() {
		if (!scrubbing) return;
		scrubbing = false;
		hoverActive = false;
		showControls();
	}

	/** Moves the preview thumbnail along with a dragged seek bar. */
	function followScrub(value: number) {
		if (!scrubbing || !seekWrapEl || duration <= 0) return;
		const ratio = clamp(value / duration, 0, 1);
		hoverX = ratio * seekWrapEl.getBoundingClientRect().width;
		hoverTime = value;
		hoverActive = true;
		if (previewVideoEl && !previewVideoEl.seeking && isFinite(value)) {
			previewVideoEl.currentTime = value;
		}
	}

	function fmtTime(s: number): string {
		if (!isFinite(s)) return '0:00';
		const h = Math.floor(s / 3600);
		const m = Math.floor((s % 3600) / 60);
		const sec = Math.floor(s % 60);
		if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
		return `${m}:${sec.toString().padStart(2, '0')}`;
	}

	function togglePlay() {
		if (!videoEl) return;
		if (videoEl.paused) videoEl.play();
		else videoEl.pause();
	}

	function seek(e: Event) {
		if (!videoEl) return;
		const value = Number((e.target as HTMLInputElement).value);
		videoEl.currentTime = value;
		followScrub(value);
	}

	// When the seek bar has focus, arrow keys would otherwise nudge it by the
	// range step (0.5s). Override them to the standard 5s skip instead.
	function onSeekKeydown(e: KeyboardEvent) {
		if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;
		e.preventDefault();
		if (!videoEl) return;
		const delta = e.key === 'ArrowLeft' ? -SEEK_STEP_SECONDS : SEEK_STEP_SECONDS;
		const total = duration || videoEl.duration || 0;
		videoEl.currentTime = clamp(videoEl.currentTime + delta, 0, total);
		showControls();
	}

	function setVolume(e: Event) {
		if (!videoEl) return;
		const v = Number((e.target as HTMLInputElement).value);
		videoEl.volume = v;
		videoEl.muted = false;
	}

	function toggleMute() {
		if (!videoEl) return;
		videoEl.muted = !videoEl.muted;
	}

	async function toggleFullscreen() {
		if (!containerEl) return;
		if (document.fullscreenElement) await document.exitFullscreen();
		else await containerEl.requestFullscreen();
	}

	function handleKeydown(e: KeyboardEvent) {
		// only intercept when no input is focused
		if (e.target instanceof HTMLInputElement) return;
		switch (e.key) {
			case ' ':
			case 'k':
				e.preventDefault();
				togglePlay();
				break;
			case 'f':
				e.preventDefault();
				toggleFullscreen();
				break;
			case 'm':
				e.preventDefault();
				toggleMute();
				break;
			case 'Escape':
				close();
				return;
			case 'ArrowLeft':
				e.preventDefault();
				if (videoEl) videoEl.currentTime = Math.max(0, videoEl.currentTime - SEEK_STEP_SECONDS);
				break;
			case 'ArrowRight':
				e.preventDefault();
				if (videoEl)
					videoEl.currentTime = Math.min(duration, videoEl.currentTime + SEEK_STEP_SECONDS);
				break;
		}
		showControls();
	}

	// Mount the playback's element. A shared playback that is already running
	// (reopened from the mini player) is adopted as it is, without a reload.
	// Only the slot is tracked: a rerun releases (stops) the playback, so state
	// read while syncing (e.g. `playing` flipping on play) must not retrigger it.
	$effect(() => {
		const slot = videoSlot;
		if (!slot) return;
		const current = untrack(() => adoptPlayback(slot));
		const stopListening = listenToVideo(current.video);
		return () => {
			stopListening();
			releasePlayback(current, slot);
		};
	});

	/** Mounts the shared (or a fresh) playback into slot and takes over its state. */
	function adoptPlayback(slot: HTMLElement): Playback {
		const current = shared ? mainPlayback() : new Playback();
		current.video.className = 'h-full w-full';
		current.mount(slot);
		syncFromElement(current.video);
		activePlayback = current;
		videoEl = current.video;
		return current;
	}

	/** Picks up the state of an element that may already be playing. */
	function syncFromElement(video: HTMLVideoElement) {
		playing = !video.paused && !video.ended;
		// Adopted mid-playback (from the mini player): no first-play screen.
		if (playing || video.currentTime > 0) hasPlayed = true;
		if (playing) showControls();
		currentTime = video.currentTime;
		duration = isFinite(video.duration) ? video.duration : 0;
		volume = video.volume;
		muted = video.muted;
		buffering = video.readyState < HTMLMediaElement.HAVE_FUTURE_DATA;
	}

	/** Mirrors the element's events into player state. Returns the unsubscribe. */
	function listenToVideo(video: HTMLVideoElement): () => void {
		const handlers: Record<string, () => void> = {
			play: () => {
				playing = true;
				buffering = false;
				hasPlayed = true;
				showControls();
			},
			pause: () => {
				playing = false;
				saveProgress(true);
				// Paused video keeps its controls up.
				clearTimeout(hideTimer);
				controlsVisible = true;
			},
			waiting: () => {
				buffering = true;
			},
			playing: () => {
				buffering = false;
			},
			// Ready but not started (autoplay refused or pending): stop showing
			// the loading state so the play button appears.
			canplay: () => {
				buffering = false;
			},
			timeupdate: () => {
				currentTime = video.currentTime;
				saveProgress();
			},
			durationchange: () => {
				duration = video.duration;
			},
			loadedmetadata: seekToStart,
			volumechange: () => {
				volume = video.volume;
				muted = video.muted;
			}
		};
		for (const [name, handler] of Object.entries(handlers)) {
			video.addEventListener(name, handler);
		}
		return () => {
			for (const [name, handler] of Object.entries(handlers)) {
				video.removeEventListener(name, handler);
			}
		};
	}

	/**
	 * Lets go of the playback on unmount: a shared one the mini player is taking
	 * over is parked for it (unless it already moved there); anything else stops.
	 */
	function releasePlayback(current: Playback, slot: HTMLElement) {
		if (!current.isMountedIn(slot)) return;
		const handoff = get(miniPlayer);
		if (shared && handoff && handoff.src === current.src) {
			current.park();
			return;
		}
		current.stop();
	}

	$effect(() => {
		if (!activePlayback) return;
		const alreadyLoaded = activePlayback.src === src;
		activePlayback.load(src, mimeType, headers, (message) => (error = message));
		if (alreadyLoaded || !seamlessArmed || isHlsSource(src, mimeType)) return;
		// Seamless upgrade: once this new source can play, crossfade off the
		// overlay that is masking this reload.
		seamlessArmed = false;
		const element = activePlayback.video;
		const onReady = () => {
			element.removeEventListener('canplay', onReady);
			swapToMainWhenReady();
		};
		element.addEventListener('canplay', onReady);
	});

	// Lowest-quality copy of the active provider (ladder is best-first), used for
	// the cheap hover-scrub thumbnails regardless of the playing quality.
	let previewSource = $derived(
		qualityLadder.length > 0 ? qualityLadder[qualityLadder.length - 1] : undefined
	);
	// Guards against re-resolving the preview source on every play-quality change.
	let previewResolvedFor = '';

	// Preview scrubber source: pinned to the lowest quality so hover-seeking stays
	// cheap — the lowest HLS level, or the provider's lowest progressive copy
	// (never the high-res source that is actually playing).
	$effect(() => {
		if (!previewVideoEl) return;
		const isHLS = mimeType.includes('mpegURL') || src.includes('.m3u8');
		if (!isHLS) {
			const lowest = previewSource;
			if (lowest && onresolvesource) {
				if (lowest.id === previewResolvedFor) return;
				previewResolvedFor = lowest.id;
				let cancelled = false;
				onresolvesource(lowest.id).then((resolved) => {
					if (cancelled || !previewVideoEl || !resolved) return;
					previewVideoEl.src = resolved.url;
				});
				return () => {
					cancelled = true;
				};
			}
			previewResolvedFor = '';
			previewVideoEl.src = src;
			return;
		}
		previewResolvedFor = '';
		let previewHls: { destroy(): void } | null = null;
		import('hls.js').then(({ default: Hls }) => {
			if (!previewVideoEl) return;
			if (!Hls.isSupported()) {
				if (previewVideoEl.canPlayType('application/vnd.apple.mpegurl')) previewVideoEl.src = src;
				return;
			}
			const instance = new Hls({
				startLevel: 0,
				xhrSetup(xhr) {
					for (const [name, value] of Object.entries(headers)) {
						xhr.setRequestHeader(name, value);
					}
				}
			});
			instance.loadSource(src);
			instance.attachMedia(previewVideoEl);
			instance.on(Hls.Events.MANIFEST_PARSED, () => {
				instance.currentLevel = 0;
			});
			previewHls = instance;
		});
		return () => previewHls?.destroy();
	});

	// Track native fullscreen so the rec overlay only engages there; collapse it
	// again whenever we leave fullscreen.
	$effect(() => {
		function onChange() {
			isFullscreen = document.fullscreenElement === containerEl;
			if (!isFullscreen) revealProgress = 0;
			matchOrientationToVideo(isFullscreen, videoEl);
		}
		document.addEventListener('fullscreenchange', onChange);
		return () => document.removeEventListener('fullscreenchange', onChange);
	});

	// Scroll-to-reveal: while fullscreen, the wheel drives the up-next sheet and
	// it snaps open or shut once the wheel settles.
	let wheelSettleTimer: ReturnType<typeof setTimeout>;
	function handleWheel(event: WheelEvent) {
		if (!isFullscreen || !hasRecs) return;
		// Sideways scrolling belongs to the rec row.
		if (Math.abs(event.deltaX) > Math.abs(event.deltaY)) return;
		event.preventDefault();
		revealProgress = clamp(revealProgress + event.deltaY / 400, 0, 1);
		clearTimeout(wheelSettleTimer);
		wheelSettleTimer = setTimeout(() => {
			revealProgress = Math.round(revealProgress);
		}, 160);
	}

	// --- Fullscreen vertical swipes ------------------------------------------
	// Swipe up pulls the up-next sheet, swipe down closes it, and with the
	// sheet shut, swipe down leaves fullscreen.
	const EXIT_SWIPE_PX = 90;
	const SHEET_FLING_VELOCITY = 0.4;

	let swipeStartX = 0;
	let swipeStartY = 0;
	let swipeStartProgress = 0;
	let swipeAxis: 'vertical' | 'other' | null = null;
	let swipeLastY = 0;
	let swipeLastTime = 0;
	let swipeVelocity = 0;
	// How far a downward swipe with the sheet shut has pulled the video.
	let exitPull = $state(0);

	/** Notes where a fullscreen touch began. */
	function onSwipeStart(event: TouchEvent) {
		swipeAxis = 'other';
		if (!swipesEnabled() || event.touches.length !== 1) return;
		swipeAxis = null;
		const target = event.target as HTMLElement;
		// The seek bar and the ⚙ sheet handle their own drags.
		if (target.closest('input, [role="dialog"]')) {
			swipeAxis = 'other';
			return;
		}
		swipeStartX = event.touches[0].clientX;
		swipeStartY = event.touches[0].clientY;
		swipeLastY = swipeStartY;
		swipeLastTime = event.timeStamp;
		swipeVelocity = 0;
		swipeStartProgress = revealProgress;
	}

	/**
	 * Vertical swipes: in fullscreen they drive the up-next sheet and leave
	 * fullscreen; inline, a downward swipe hands off (onswipedown).
	 */
	function swipesEnabled(): boolean {
		if (isFullscreen) return true;
		return inline && onswipedown !== undefined;
	}

	/** Follows a vertical swipe with the sheet, or with the video to leave. */
	function onSwipeMove(event: TouchEvent) {
		if (swipeAxis === 'other') return;
		const touch = event.touches[0];
		const deltaX = touch.clientX - swipeStartX;
		const deltaY = touch.clientY - swipeStartY;
		if (swipeAxis === null) {
			if (Math.max(Math.abs(deltaX), Math.abs(deltaY)) < TOUCH_SLOP) return;
			swipeAxis = 'other';
			if (Math.abs(deltaY) > Math.abs(deltaX) * 1.2) swipeAxis = 'vertical';
			// Inline, only a pull down is ours; anything else scrolls the page.
			if (!isFullscreen && deltaY < 0) swipeAxis = 'other';
			if (swipeAxis !== 'vertical') return;
			clearTimeout(holdTimer);
			revealDragging = true;
		}
		if (swipeAxis !== 'vertical') return;
		event.preventDefault();
		const elapsed = event.timeStamp - swipeLastTime;
		if (elapsed > 0) swipeVelocity = (touch.clientY - swipeLastY) / elapsed;
		swipeLastY = touch.clientY;
		swipeLastTime = event.timeStamp;
		if (swipeStartProgress === 0 && deltaY > 0) {
			exitPull = deltaY;
			return;
		}
		if (!isFullscreen) {
			exitPull = 0;
			return;
		}
		exitPull = 0;
		if (!hasRecs) return;
		revealProgress = clamp(swipeStartProgress - deltaY / recoSheetHeight, 0, 1);
	}

	/** Settles the sheet open or shut, or leaves fullscreen after a long pull down. */
	function onSwipeEnd() {
		if (swipeAxis !== 'vertical') return;
		swipeAxis = null;
		revealDragging = false;
		swallowNextClick = true;
		setTimeout(() => (swallowNextClick = false), 350);
		if (exitPull > 0) {
			const leave = exitPull > EXIT_SWIPE_PX || swipeVelocity > SHEET_FLING_VELOCITY * 2;
			exitPull = 0;
			if (!leave) return;
			if (isFullscreen) {
				document.exitFullscreen().catch(() => {});
				return;
			}
			onswipedown?.();
			return;
		}
		revealProgress = settledReveal();
	}

	/** Where the sheet rests after a swipe: open or shut, by distance or fling. */
	function settledReveal(): number {
		if (swipeVelocity < -SHEET_FLING_VELOCITY) return 1;
		if (swipeVelocity > SHEET_FLING_VELOCITY) return 0;
		if (swipeStartProgress === 0) return Number(revealProgress > 0.25);
		return Number(revealProgress > 0.75);
	}

	/** Opens the up-next sheet from its grab bar. */
	function openUpNext() {
		revealProgress = 1;
		hideControls();
	}

	// wheel and touchmove must be non-passive listeners to allow preventDefault.
	$effect(() => {
		const container = containerEl;
		if (!container) return;
		container.addEventListener('wheel', handleWheel, { passive: false });
		container.addEventListener('touchstart', onSwipeStart, { passive: true });
		container.addEventListener('touchmove', onSwipeMove, { passive: false });
		container.addEventListener('touchend', onSwipeEnd);
		container.addEventListener('touchcancel', onSwipeEnd);
		return () => {
			container.removeEventListener('wheel', handleWheel);
			container.removeEventListener('touchstart', onSwipeStart);
			container.removeEventListener('touchmove', onSwipeMove);
			container.removeEventListener('touchend', onSwipeEnd);
			container.removeEventListener('touchcancel', onSwipeEnd);
		};
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Fullscreen: z-[60] renders above MediaOverlay (z-50). Inline: embeds in-page. -->
<div
	bind:this={containerEl}
	data-swipe-ignore
	class="flex flex-col bg-black {inline
		? 'relative aspect-video w-full overflow-hidden'
		: 'fixed inset-0 z-[60]'}"
	class:rounded-xl={inline && !square}
	onmousemove={onMouseMove}
	role="presentation"
	class:cursor-none={!controlsVisible}
>
	<!-- Top bar: title + close, for the standalone (non-inline) player only; the
	     inline one sits under the page's own title. -->
	{#if !inline}
		<div
			class="pointer-events-none absolute inset-x-0 top-0 z-10 flex items-center justify-between bg-gradient-to-b from-black/70 to-transparent p-4 transition-opacity duration-300"
			class:opacity-0={!controlsVisible || revealProgress > 0.05}
			class:opacity-100={controlsVisible && revealProgress <= 0.05}
		>
			<span class="max-w-[70%] truncate text-sm font-medium text-white/80">{title}</span>
			<button
				type="button"
				class="pointer-events-auto flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
				onclick={close}
				aria-label="Close player"
			>
				<XIcon size={18} weight="bold" />
			</button>
		</div>
	{/if}

	<!-- Video: the playback's element is mounted in here. -->
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div
		bind:this={videoSlot}
		class="parallax-video h-full w-full select-none"
		class:parallax-dragging={revealDragging}
		style="transform: translateY({exitPull * 0.5}px) scale({1 - revealProgress * 0.04}); filter: brightness({1 -
			revealProgress * 0.5});"
		onpointerdown={onVideoPointerDown}
		onpointermove={onVideoPointerMove}
		onpointerup={onVideoPointerUp}
		onpointercancel={onVideoPointerUp}
		oncontextmenu={(event) => event.preventDefault()}
		onclick={onVideoClick}
		ondblclick={onVideoDoubleClick}
	></div>

	<!-- Double-tap skip ripple on the tapped side. -->
	{#if skipFeedback}
		{#key skipFeedback.key}
			<div
				class="skip-ripple pointer-events-none absolute inset-y-0 z-10 flex w-2/5 items-center justify-center"
				class:left-0={skipFeedback.zone === 'back'}
				class:skip-ripple-left={skipFeedback.zone === 'back'}
				class:right-0={skipFeedback.zone === 'forward'}
				class:skip-ripple-right={skipFeedback.zone === 'forward'}
			>
				<span class="text-sm font-semibold text-white drop-shadow">
					{#if skipFeedback.zone === 'back'}−{:else}+{/if}{skipFeedback.seconds}s
				</span>
			</div>
		{/key}
	{/if}

	<!-- Held finger: 2× speed pill. -->
	{#if holdingFast}
		<div
			class="glass-round pointer-events-none absolute top-3 left-1/2 z-10 flex -translate-x-1/2 items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-semibold text-white"
			transition:fade={{ duration: 150 }}
		>
			{HOLD_SPEED}× <FastForwardIcon size={14} weight="fill" />
		</div>
	{/if}

	<!-- Back, for pages that hide their top bar over the player. -->
	{#if onback && !isFullscreen && (controlsVisible || !hasPlayed) && revealProgress <= 0.05}
		<button
			type="button"
			class="glass-round absolute top-2 left-2 z-20 flex h-9 w-9 items-center justify-center rounded-full text-white"
			onclick={onback}
			aria-label="Back"
			transition:fade={{ duration: 150 }}
		>
			<ArrowLeftIcon size={18} weight="bold" />
		</button>
	{/if}

	<!-- Big play/pause: the only control before the first play, and the phone's
	     centre control after it. -->
	{#if !error && revealProgress <= 0.05 && ((!hasPlayed && !playing) || (!buffering && $isCompact && controlsVisible))}
		<button
			type="button"
			class="glass-round absolute top-1/2 left-1/2 z-10 flex h-16 w-16 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full text-white"
			onclick={togglePlay}
			aria-label={playing ? 'Pause' : 'Play'}
			transition:scale={{ start: 0.8, duration: 180, easing: cubicOut }}
		>
			{#if playing}
				<PauseIcon size={28} weight="fill" />
			{:else}
				<PlayIcon size={28} weight="fill" />
			{/if}
		</button>
	{/if}

	<!-- Thin progress line while the controls are hidden. -->
	{#if hasPlayed && duration > 0}
		<div
			class="pointer-events-none absolute inset-x-0 bottom-0 z-10 h-[3px] bg-white/20 transition-opacity duration-300"
			class:opacity-0={controlsVisible || revealProgress > 0.05}
		>
			<div class="h-full bg-[#ff1f4b]" style="width: {seekPercent}%"></div>
		</div>
	{/if}

	<!-- Prebuffer/crossfade overlay: buffers a higher quality off-screen, then
	     covers the main element during the swap so its reload never flashes. -->
	<!-- svelte-ignore a11y_media_has_caption -->
	<video
		bind:this={prebufferEl}
		class="pointer-events-none absolute inset-0 z-[5] h-full w-full"
		class:opacity-0={!overlayActive}
		class:opacity-100={overlayActive}
		preload="auto"
		playsinline
	></video>

	<!-- Loading: the poster (dimmed) until the first frame, and a shimmer
	     running along the bottom edge while buffering. -->
	{#if buffering && !error && !hasPlayed && poster}
		<img
			src={poster}
			alt=""
			class="pointer-events-none absolute inset-0 h-full w-full object-cover brightness-75"
			transition:fade={{ duration: 200 }}
		/>
	{/if}
	{#if buffering && !error}
		<div
			class="pointer-events-none absolute inset-x-0 bottom-0 z-10 h-[3px] overflow-hidden bg-white/10"
			transition:fade={{ duration: 200 }}
		>
			<div class="buffer-shimmer h-full w-1/3"></div>
		</div>
	{/if}

	<!-- Error -->
	{#if error}
		<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
			<p class="text-error rounded-xl bg-black/70 px-6 py-4 text-sm">{error}</p>
		</div>
	{/if}

	<!-- Controls: time and buttons, then the seek bar along the bottom edge. -->
	<div
		class="pointer-events-none absolute inset-x-0 bottom-0 z-10 flex flex-col gap-1 bg-gradient-to-t from-black/75 to-transparent px-3 pt-10 pb-1 transition-opacity duration-300"
		class:opacity-0={!controlsVisible || !hasPlayed || revealProgress > 0.05}
		class:opacity-100={controlsVisible && hasPlayed && revealProgress <= 0.05}
		class:invisible={!hasPlayed}
		class:controls-off={!controlsVisible || revealProgress > 0.05}
		class:pb-4={isFullscreen}
	>
		{@render controlRow()}

		<!-- Seek bar + hover-scrub thumbnail -->
		<div
			bind:this={seekWrapEl}
			class="pointer-events-auto relative"
			class:seek-scrubbing={scrubbing}
			role="presentation"
			onmousemove={onSeekHover}
			onmouseleave={onSeekLeave}
		>
			<input
				type="range"
				min="0"
				max={duration || 0}
				value={currentTime}
				step="0.5"
				class="player-range w-full"
				style="--fill:{seekPercent}%"
				oninput={seek}
				onkeydown={onSeekKeydown}
				onpointerdown={onSeekPointerDown}
				onpointerup={onSeekPointerUp}
				onpointercancel={onSeekPointerUp}
				onchange={onSeekPointerUp}
				aria-label="Seek"
			/>
			<!-- Markers: point markers are ticks, span markers translucent bands.
			     Personal markers are amber, global (plugin) annotations white. -->
			{#if duration > 0}
				{#each markers as marker (marker.id)}
					{@const spanEnd = marker.endSeconds ?? null}
					{#if spanEnd !== null && spanEnd > marker.seconds}
						<button
							type="button"
							class="group absolute top-1/2 z-10 h-2 -translate-y-1/2 rounded-sm {marker.personal
								? 'bg-amber-400/40 hover:bg-amber-400/60'
								: 'bg-white/25 hover:bg-white/40'}"
							style="left:{(marker.seconds / duration) * 100}%; width:{((spanEnd - marker.seconds) /
								duration) *
								100}%"
							onclick={() => jumpToMarker(marker.seconds)}
							aria-label="Jump to {marker.label}"
						>
							<span
								class="pointer-events-none absolute bottom-full left-1/2 mb-2 hidden -translate-x-1/2 rounded bg-black/80 px-1.5 py-0.5 text-[10px] font-medium whitespace-nowrap text-white group-hover:block"
							>
								{marker.label}
							</span>
						</button>
					{:else}
						<button
							type="button"
							class="group absolute top-1/2 z-10 h-3 w-1 -translate-x-1/2 -translate-y-1/2 rounded-full ring-1 ring-black/30 transition-transform hover:scale-y-150 {marker.personal
								? 'bg-amber-400'
								: 'bg-white/70'}"
							style="left:{(marker.seconds / duration) * 100}%"
							onclick={() => jumpToMarker(marker.seconds)}
							aria-label="Jump to {marker.label}"
						>
							<span
								class="pointer-events-none absolute bottom-full left-1/2 mb-2 hidden -translate-x-1/2 rounded bg-black/80 px-1.5 py-0.5 text-[10px] font-medium whitespace-nowrap text-white group-hover:block"
							>
								{marker.label}
							</span>
						</button>
					{/if}
				{/each}
			{/if}
			<div
				class="pointer-events-none absolute bottom-full mb-3 flex -translate-x-1/2 flex-col items-center gap-1 {hoverActive &&
				duration > 0
					? ''
					: 'hidden'}"
				style="left:{hoverX}px"
			>
				<div
					class="shadow-overlay overflow-hidden rounded-md ring-1 ring-white/15"
					style="width:160px;height:90px;"
				>
					<!-- svelte-ignore a11y_media_has_caption -->
					<video
						bind:this={previewVideoEl}
						class="h-full w-full object-cover"
						muted
						playsinline
						preload="auto"
					></video>
				</div>
				<span
					class="text-2xs rounded bg-black/70 px-1.5 py-0.5 font-mono text-white/90 tabular-nums"
				>
					{fmtTime(hoverTime)}
				</span>
			</div>
		</div>

		<!-- Fullscreen: the grab bar that pulls up the up-next sheet. -->
		{#if isFullscreen && hasRecs}
			<button
				type="button"
				class="pointer-events-auto mx-auto flex h-5 w-16 items-center justify-center"
				onclick={openUpNext}
				aria-label="Up next"
			>
				<span class="h-1 w-10 rounded-full bg-white/60"></span>
			</button>
		{/if}
	</div>

	<!-- Desktop ⚙ popover; phones get a sheet (below). -->
	{#if settingsOpen && !$isCompact}
		<div
			class="player-menu absolute right-3 bottom-16 z-30 max-h-[70%] w-60 overflow-y-auto rounded-2xl p-1.5"
			transition:scale={{ start: 0.95, duration: 150, easing: cubicOut }}
		>
			{@render settingsMenu()}
		</div>
	{/if}
	<!-- Fullscreen up-next: a glass sheet with one row of recs, pulled up over
	     the still-playing video. -->
	{#if hasRecs && isFullscreen}
		<div
			bind:offsetHeight={recoSheetHeight}
			class="reco-sheet absolute inset-x-3 bottom-3 z-20 flex flex-col gap-3 rounded-[1.75rem] pt-2 pb-4"
			class:reco-sheet-dragging={revealDragging}
			style="translate: 0 calc({1 - revealProgress} * (100% + 1rem));"
			class:pointer-events-none={revealProgress < 0.5}
			aria-hidden={revealProgress <= 0.02}
		>
			<span class="mx-auto h-1 w-9 shrink-0 rounded-full bg-white/30"></span>
			<h2 class="px-4 text-base font-semibold text-white">Up next</h2>
			<div
				bind:this={recoRowEl}
				onscroll={onRecoScroll}
				class="no-scrollbar flex snap-x snap-mandatory gap-3 overflow-x-auto scroll-px-4 px-4"
			>
				{#each recommendations as rec (rec.id)}
					{@render recCard(rec)}
				{/each}
			</div>
		</div>
	{/if}

	<!-- Moved to <body> so no wrapper's stacking keeps it under the tab bar;
	     into the player while fullscreen, where only its subtree is drawn. -->
	{#if settingsOpen && $isCompact}
		<div use:portal={fullscreenHost(isFullscreen)}>
			<ActionSheet title="Playback" subtitle={qualitySummary()} onclose={() => (settingsOpen = false)}>
				{@render settingsMenu()}
			</ActionSheet>
		</div>
	{/if}
</div>

{#snippet controlRow()}
	<div class="pointer-events-auto flex items-center gap-1">
		<!-- Phones use the big centre button. -->
		{#if !$isCompact}
			<button
				type="button"
				class="flex h-9 w-9 flex-none items-center justify-center rounded-full text-white hover:bg-white/10"
				onclick={togglePlay}
				aria-label={playing ? 'Pause' : 'Play'}
			>
				{#if playing}
					<PauseIcon size={20} weight="fill" />
				{:else}
					<PlayIcon size={20} weight="fill" />
				{/if}
			</button>
		{/if}

		<span class="px-1 font-mono text-xs whitespace-nowrap text-white/75 tabular-nums">
			{fmtTime(currentTime)} / {fmtTime(duration)}
		</span>

		<div class="flex-1"></div>

		{#if onaddmarker && !$isCompact}
			<button
				type="button"
				class="flex h-9 w-9 flex-none items-center justify-center text-white/70 hover:text-amber-400"
				onclick={addMarkerHere}
				aria-label="Add marker at current time"
				title="Add marker here"
			>
				<BookmarkSimpleIcon size={16} weight="fill" />
			</button>
		{/if}

		<!-- Phones set the volume with their hardware keys. -->
		{#if !$isCompact}
			<div class="flex items-center gap-2">
				<button
					type="button"
					class="flex h-9 w-9 flex-none items-center justify-center text-white/70 hover:text-white"
					onclick={toggleMute}
					aria-label={muted ? 'Unmute' : 'Mute'}
				>
					{#if muted || volume === 0}
						<SpeakerXIcon size={16} weight="fill" />
					{:else}
						<SpeakerHighIcon size={16} weight="fill" />
					{/if}
				</button>
				<input
					type="range"
					min="0"
					max="1"
					step="0.05"
					value={muted ? 0 : volume}
					class="player-range w-20"
					style="--fill:{volumePercent}%"
					oninput={setVolume}
					aria-label="Volume"
				/>
			</div>
		{/if}

		<button
			type="button"
			class="flex h-9 w-9 flex-none items-center justify-center text-white/80 transition-transform hover:text-white"
			class:rotate-45={settingsOpen}
			onclick={() => (settingsOpen = !settingsOpen)}
			aria-label="Playback settings"
			title="Quality, speed and source"
		>
			<GearSixIcon size={18} weight="bold" />
		</button>

		<button
			type="button"
			class="flex h-9 w-9 flex-none items-center justify-center text-white/80 hover:text-white"
			onclick={toggleFullscreen}
			aria-label="Toggle fullscreen"
		>
			<ArrowsOutIcon size={18} weight="bold" />
		</button>
	</div>
{/snippet}

{#snippet settingsMenu()}
	<div class="flex flex-col gap-1 pb-1">
		{#if sources.length > 0}
			<p class="menu-heading">Quality</p>
			<button type="button" class="menu-row" onclick={chooseAutoQuality}>
				<span>Auto</span>
				{#if autoAdapt}
					<CheckIcon size={16} weight="bold" class="text-primary" />
				{/if}
			</button>
			{#each sourceGroups as group (group.provider)}
				<p class="menu-heading">{siteDisplayName(group.provider, $sites)}</p>
				{#each group.sources as source (source.id)}
					<button type="button" class="menu-row" onclick={() => selectSource(source)}>
						<span>{source.quality}</span>
						{#if source.id === activeSourceId}
							<CheckIcon size={16} weight="bold" class="text-primary" />
						{/if}
					</button>
				{/each}
			{/each}
		{/if}

		<p class="menu-heading">Speed</p>
		<div class="flex flex-wrap gap-1.5 px-2 pb-1">
			{#each SPEEDS as speed (speed)}
				<button
					type="button"
					class="speed-chip rounded-full px-3 py-1.5 text-sm font-medium tabular-nums"
					class:speed-chip-active={playbackRate === speed}
					onclick={() => setSpeed(speed)}
				>
					{speed}×
				</button>
			{/each}
		</div>

		{#if onaddmarker && $isCompact}
			<button type="button" class="menu-row" onclick={addMarkerFromMenu}>
				<span class="flex items-center gap-3">
					<BookmarkSimpleIcon size={18} weight="fill" class="text-amber-400" />
					Add marker at {fmtTime(currentTime)}
				</span>
			</button>
		{/if}
	</div>
{/snippet}

{#snippet recCard(rec: PlayerRec)}
	<button
		type="button"
		class="flex w-56 shrink-0 snap-start flex-col gap-2 text-left"
		onclick={() => onselectrec?.(rec)}
	>
		<span class="relative block aspect-video w-full overflow-hidden rounded-xl bg-white/5">
			{#if rec.image}
				<img src={rec.image} alt="" loading="lazy" class="h-full w-full object-cover" />
			{/if}
			{#if formatClock(rec.durationSeconds)}
				<span
					class="absolute right-1.5 bottom-1.5 rounded bg-black/75 px-1.5 py-0.5 text-[11px] font-medium text-white tabular-nums"
				>
					{formatClock(rec.durationSeconds)}
				</span>
			{/if}
		</span>
		<span class="line-clamp-2 text-[13px] leading-[18px] font-medium text-white">{rec.title}</span>
		{#if rec.siteName}
			<span class="flex items-center gap-1.5 text-xs text-white/60">
				<SiteIcon iconUrl={rec.siteIconUrl} class="h-3.5 w-3.5" />
				{rec.siteName}
			</span>
		{/if}
	</button>
{/snippet}

<style>
	/* Smooth the wheel- and swipe-driven dimming between steps. */
	.parallax-video {
		transition:
			transform 300ms cubic-bezier(0.22, 1, 0.36, 1),
			filter 300ms cubic-bezier(0.22, 1, 0.36, 1);
		transform-origin: center top;
	}

	.parallax-dragging {
		transition: none;
	}

	/* Hidden controls must not catch taps meant for the video. */
	.controls-off :global(*) {
		pointer-events: none !important;
	}

	.reco-sheet {
		background-color: color-mix(in oklab, var(--color-base-200) 55%, transparent);
		backdrop-filter: blur(28px) saturate(1.8);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow: 0 -10px 40px rgba(0, 0, 0, 0.5);
		transition: translate 340ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	.reco-sheet-dragging {
		transition: none;
	}

	.player-menu {
		background-color: color-mix(in oklab, var(--color-base-200) 70%, transparent);
		backdrop-filter: blur(24px) saturate(1.6);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
	}

	.menu-heading {
		padding: 0.5rem 0.75rem 0.25rem;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: color-mix(in oklab, var(--color-base-content) 50%, transparent);
	}

	.menu-row {
		display: flex;
		min-height: 2.75rem;
		flex-shrink: 0;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		border-radius: 1rem;
		padding-inline: 0.75rem;
		text-align: left;
		font-size: 15px;
	}

	.menu-row:hover {
		background-color: color-mix(in oklab, var(--color-base-content) 8%, transparent);
	}

	.speed-chip {
		background-color: color-mix(in oklab, var(--color-base-content) 10%, transparent);
		transition: background-color 150ms;
	}

	.speed-chip-active {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}

	/* The seek bar thickens under a dragging finger. */
	.seek-scrubbing :global(.player-range) {
		background-size: 100% 6px;
	}

	.seek-scrubbing :global(.player-range::-webkit-slider-thumb) {
		transform: scale(1.3);
	}

	.buffer-shimmer {
		background: linear-gradient(to right, transparent, rgba(255, 255, 255, 0.85), transparent);
		animation: buffer-shimmer 1.1s ease-in-out infinite;
	}

	@keyframes buffer-shimmer {
		from {
			translate: -100% 0;
		}
		to {
			translate: 300% 0;
		}
	}

	/* Double-tap skip: a soft light that swells from the tapped edge. */
	.skip-ripple {
		animation: skip-ripple 650ms ease-out forwards;
	}

	.skip-ripple-left {
		background: radial-gradient(circle at 0% 50%, rgba(255, 255, 255, 0.22), transparent 70%);
		border-radius: 0 50% 50% 0;
	}

	.skip-ripple-right {
		background: radial-gradient(circle at 100% 50%, rgba(255, 255, 255, 0.22), transparent 70%);
		border-radius: 50% 0 0 50%;
	}

	@keyframes skip-ripple {
		0% {
			opacity: 0;
		}
		15% {
			opacity: 1;
		}
		100% {
			opacity: 0;
		}
	}
</style>
