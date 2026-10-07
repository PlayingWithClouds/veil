<script lang="ts">
	// Floating player for the scene the user left while it played. It shows the
	// main playback's own element, so the video carries on uninterrupted and
	// returns to the scene page the same way. Drag it anywhere and it springs to
	// the nearest corner; pinch it or drag its grip to resize; fling it off the
	// side to close. Corner and size are remembered.
	import { untrack } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import ArrowsOutSimpleIcon from 'phosphor-svelte/lib/ArrowsOutSimpleIcon';
	import CompactPlayer, { type CompactPlayerState } from './CompactPlayer.svelte';
	import { miniPlayer, updateMiniPlayback } from '$lib/stores/miniPlayer';
	import { mainPlayback, type Playback } from '$lib/playback';
	import { recordProgress } from '$lib/discovery';
	import { isCompact } from '$lib/stores/viewport';
	import { sceneUrl } from '$lib/routes';

	type Corner = 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';

	// Progress is saved at most this often while playing, plus on pause/close.
	const SAVE_INTERVAL_SECONDS = 15;
	// The stored position (what a reload resumes from) is refreshed this often.
	const PERSIST_INTERVAL_SECONDS = 1;
	const COMPLETE_FRACTION = 0.9;

	const FRAME_STORAGE_KEY = 'veil:mini-player-frame';
	const PHONE_WIDTH = 176;
	const DESKTOP_WIDTH = 384;
	const MIN_WIDTH = 128;
	const MAX_WIDTH = 640;
	// Movement (px) before a press counts as a drag rather than a tap.
	const DRAG_SLOP = 6;
	// Sideways release speed (px/ms) that throws the player away.
	const FLING_VELOCITY = 1.1;
	const CONTROLS_HIDE_MS = 3000;
	const SNAP_EASING = 'cubic-bezier(0.22, 1.2, 0.36, 1)';

	let playback = $state<Playback | null>(null);
	let lastSavedAt = 0;
	let lastPersisted: CompactPlayerState | null = null;
	let playing = $state(false);
	let position = $state(0);
	let duration = $state(0);

	let frameElement = $state<HTMLElement | null>(null);
	let corner = $state<Corner>('bottom-right');
	let width = $state(0);
	let dragOffsetX = $state(0);
	let dragOffsetY = $state(0);
	let dragging = $state(false);
	let controlsShown = $state(false);
	let controlsTimer: ReturnType<typeof setTimeout> | undefined;

	// Only a new source (re)attaches, not the position updates in the store.
	let src = $derived($miniPlayer?.src);

	$effect(() => {
		if (!src) {
			playback = null;
			return;
		}
		playback = untrack(resumePlayback);
	});

	// The remembered frame, else a size that suits the screen.
	$effect(() => {
		const compact = $isCompact;
		untrack(() => restoreFrame(compact));
	});

	/**
	 * The main playback, already running when handed over from the scene page.
	 * After a reload it is empty, so load the stored source and resume it.
	 */
	function resumePlayback(): Playback | null {
		const stored = $miniPlayer;
		if (!stored) return null;
		const main = mainPlayback();
		lastSavedAt = stored.position;
		if (main.src === stored.src) return main;
		main.video.muted = stored.muted;
		main.video.addEventListener(
			'loadedmetadata',
			() => {
				main.video.currentTime = stored.position;
				// The browser may also refuse unmuted autoplay after a reload; the
				// video then waits paused for the play button.
				if (!stored.playing) main.video.pause();
			},
			{ once: true }
		);
		main.load(stored.src, stored.mimeType);
		return main;
	}

	/** Persists position/play/mute for reloads and records watch progress. */
	function onStateChange(state: CompactPlayerState) {
		playing = state.playing;
		position = state.position;
		duration = state.duration;
		persist(state);
		saveProgress(state, !state.playing);
	}

	function persist(state: CompactPlayerState) {
		const unchanged =
			lastPersisted !== null &&
			lastPersisted.playing === state.playing &&
			lastPersisted.muted === state.muted &&
			Math.abs(lastPersisted.position - state.position) < PERSIST_INTERVAL_SECONDS;
		if (unchanged) return;
		lastPersisted = state;
		updateMiniPlayback(state.position, state.playing, state.muted);
	}

	/** Records the watch position for resume and recommendations. */
	function saveProgress(state: CompactPlayerState, force: boolean) {
		const stored = $miniPlayer;
		if (!stored) return;
		if (!force && Math.abs(state.position - lastSavedAt) < SAVE_INTERVAL_SECONDS) return;
		if (force && state.position === lastSavedAt) return;
		lastSavedAt = state.position;
		recordProgress({
			mediaId: stored.sceneId,
			progressSeconds: state.position,
			durationSeconds: state.duration > 0 ? state.duration : undefined,
			completed: state.duration > 0 && state.position >= state.duration * COMPLETE_FRACTION
		});
	}

	/** Goes back to the scene page, which takes the playback over. */
	function expand() {
		const stored = $miniPlayer;
		if (stored) goto(sceneUrl(stored.sceneId));
	}

	function close() {
		if (playback && lastPersisted) saveProgress(lastPersisted, true);
		mainPlayback().stop();
		miniPlayer.set(null);
	}

	function togglePlay() {
		const video = mainPlayback().video;
		if (video.paused) {
			video.play().catch(() => {});
		} else {
			video.pause();
		}
		revealControls();
	}

	// --- Frame: corner and size --------------------------------------------

	/** Loads the saved corner and width, or the defaults for this screen size. */
	function restoreFrame(compact: boolean) {
		corner = 'bottom-right';
		width = DESKTOP_WIDTH;
		if (compact) width = PHONE_WIDTH;
		const stored = localStorage.getItem(FRAME_STORAGE_KEY);
		if (!stored) return;
		try {
			const frame = JSON.parse(stored) as { corner: Corner; width: number; compact: boolean };
			if (frame.compact !== compact) return;
			corner = frame.corner;
			width = clampWidth(frame.width);
		} catch {
			// A broken entry just means the defaults.
		}
	}

	/** Remembers the corner and width for next time. */
	function saveFrame() {
		localStorage.setItem(FRAME_STORAGE_KEY, JSON.stringify({ corner, width, compact: $isCompact }));
	}

	/** Keeps a width between the minimum and what fits the screen. */
	function clampWidth(value: number): number {
		const fits = Math.min(MAX_WIDTH, window.innerWidth - 24);
		return Math.min(fits, Math.max(MIN_WIDTH, value));
	}

	/** The corner nearest to where the player's centre was let go. */
	function nearestCorner(rect: DOMRect): Corner {
		const isTop = rect.top + rect.height / 2 < window.innerHeight / 2;
		const isLeft = rect.left + rect.width / 2 < window.innerWidth / 2;
		if (isTop && isLeft) return 'top-left';
		if (isTop) return 'top-right';
		if (isLeft) return 'bottom-left';
		return 'bottom-right';
	}

	/** Moves to a corner, springing from where the finger left the player. */
	async function snapTo(next: Corner) {
		const element = frameElement;
		if (!element) return;
		const before = element.getBoundingClientRect();
		corner = next;
		dragOffsetX = 0;
		dragOffsetY = 0;
		await Promise.resolve();
		const after = element.getBoundingClientRect();
		element.animate(
			[{ translate: `${before.left - after.left}px ${before.top - after.top}px` }, { translate: '0 0' }],
			{ duration: 420, easing: SNAP_EASING }
		);
		saveFrame();
	}

	/** Throws the player off the side it was flung towards, then closes it. */
	function flingAway(direction: number) {
		const element = frameElement;
		if (!element) {
			close();
			return;
		}
		const animation = element.animate(
			[
				{ translate: `${dragOffsetX}px ${dragOffsetY}px`, opacity: 1 },
				{ translate: `${dragOffsetX + direction * window.innerWidth}px ${dragOffsetY}px`, opacity: 0 }
			],
			{ duration: 260, easing: 'ease-in', fill: 'forwards' }
		);
		animation.onfinish = close;
	}

	// --- Gestures: drag, pinch, grip resize --------------------------------

	type Gesture = 'none' | 'press' | 'drag' | 'pinch' | 'resize';

	let gesture: Gesture = 'none';
	const pointers = new Map<number, { x: number; y: number }>();
	let startX = 0;
	let startY = 0;
	let startWidth = 0;
	let pinchStartDistance = 0;
	let lastMoveX = 0;
	let lastMoveTime = 0;
	let velocityX = 0;
	// The click that ends a drag or resize must not act as a tap.
	let swallowClick = false;

	/** Distance between the first two pointers on the player. */
	function pinchDistance(): number {
		const [first, second] = [...pointers.values()];
		return Math.hypot(first.x - second.x, first.y - second.y);
	}

	/** Starts a press, or a pinch when a second finger lands. */
	function onPointerDown(event: PointerEvent) {
		if ((event.target as HTMLElement).closest('button')) return;
		pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
		// Captured only once it's a drag or pinch: capturing a tap would send its
		// click to the frame instead of the video.
		if (pointers.size === 2) {
			for (const pointerId of pointers.keys()) frameElement?.setPointerCapture(pointerId);
			gesture = 'pinch';
			pinchStartDistance = pinchDistance();
			startWidth = width;
			dragOffsetX = 0;
			dragOffsetY = 0;
			return;
		}
		gesture = 'press';
		startX = event.clientX;
		startY = event.clientY;
		lastMoveX = event.clientX;
		lastMoveTime = event.timeStamp;
		velocityX = 0;
	}

	/** Starts resizing from the grip. */
	function onGripDown(event: PointerEvent) {
		event.stopPropagation();
		gesture = 'resize';
		startX = event.clientX;
		startWidth = width;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	/** Follows the finger: moves, pinches or resizes. */
	function onPointerMove(event: PointerEvent) {
		if (gesture === 'resize') {
			resizeFromGrip(event);
			return;
		}
		if (!pointers.has(event.pointerId)) return;
		pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
		if (gesture === 'pinch') {
			width = clampWidth(startWidth * (pinchDistance() / pinchStartDistance));
			return;
		}
		const deltaX = event.clientX - startX;
		const deltaY = event.clientY - startY;
		if (gesture === 'press' && Math.hypot(deltaX, deltaY) > DRAG_SLOP) {
			gesture = 'drag';
			dragging = true;
			frameElement?.setPointerCapture(event.pointerId);
		}
		if (gesture !== 'drag') return;
		const elapsed = event.timeStamp - lastMoveTime;
		if (elapsed > 0) velocityX = (event.clientX - lastMoveX) / elapsed;
		lastMoveX = event.clientX;
		lastMoveTime = event.timeStamp;
		dragOffsetX = deltaX;
		dragOffsetY = deltaY;
	}

	/** Grows the player as the grip is pulled away from its anchored corner. */
	function resizeFromGrip(event: PointerEvent) {
		let direction = 1;
		if (corner.endsWith('right')) direction = -1;
		width = clampWidth(startWidth + (event.clientX - startX) * direction);
	}

	/** Ends the gesture: snaps, flings away, or saves the new size. */
	function onPointerUp(event: PointerEvent) {
		pointers.delete(event.pointerId);
		const ended = gesture;
		if (ended === 'pinch' && pointers.size > 0) return;
		gesture = 'none';
		dragging = false;
		if (ended === 'resize' || ended === 'pinch') {
			swallowClick = true;
			saveFrame();
			return;
		}
		if (ended !== 'drag') return;
		swallowClick = true;
		if (Math.abs(velocityX) > FLING_VELOCITY) {
			flingAway(Math.sign(velocityX));
			return;
		}
		if (frameElement) snapTo(nearestCorner(frameElement.getBoundingClientRect()));
	}

	/** A tap: phones show the controls, desktop goes back to the scene. */
	function onVideoTap() {
		if (swallowClick) {
			swallowClick = false;
			return;
		}
		if (!$isCompact) {
			expand();
			return;
		}
		if (controlsShown) {
			controlsShown = false;
			return;
		}
		revealControls();
	}

	/** Shows the phone controls for a few seconds. */
	function revealControls() {
		controlsShown = true;
		clearTimeout(controlsTimer);
		controlsTimer = setTimeout(() => (controlsShown = false), CONTROLS_HIDE_MS);
	}
</script>

{#if $miniPlayer && playback}
	<div
		bind:this={frameElement}
		class="mini-player z-50 aspect-video overflow-hidden rounded-2xl shadow-2xl ring-1 ring-white/10"
		class:mini-top={corner.startsWith('top')}
		class:mini-bottom={corner.startsWith('bottom')}
		class:mini-left={corner.endsWith('left')}
		class:mini-right={corner.endsWith('right')}
		class:mini-compact={$isCompact}
		class:mini-dragging={dragging}
		style:width="{width}px"
		style:translate="{dragOffsetX}px {dragOffsetY}px"
		data-swipe-ignore
		role="presentation"
		onpointerdown={onPointerDown}
		onpointermove={onPointerMove}
		onpointerup={onPointerUp}
		onpointercancel={onPointerUp}
		transition:scale={{ start: 0.85, duration: 220, easing: cubicOut }}
	>
		<CompactPlayer {playback} chrome={!$isCompact} onvideoclick={onVideoTap} onstatechange={onStateChange}>
			{#snippet header()}
				<div class="flex items-start gap-2">
					<p class="line-clamp-2 min-w-0 flex-1 text-sm font-medium text-white">
						{$miniPlayer?.title}
					</p>
					<button
						type="button"
						class="btn btn-square btn-ghost btn-xs text-white"
						aria-label="Close player"
						onclick={close}
					>
						<XIcon size={14} />
					</button>
				</div>
			{/snippet}
			{#snippet actions()}
				<button
					type="button"
					class="btn btn-square btn-ghost btn-sm"
					aria-label="Back to scene"
					onclick={expand}
				>
					<ArrowsOutSimpleIcon size={18} />
				</button>
			{/snippet}
		</CompactPlayer>

		{#if $isCompact && controlsShown}
			<div class="pointer-events-none absolute inset-0 bg-black/40" transition:fade={{ duration: 150 }}>
				<button
					type="button"
					class="pointer-events-auto absolute top-1 left-1 flex h-8 w-8 items-center justify-center text-white"
					aria-label="Back to scene"
					onclick={expand}
				>
					<ArrowsOutSimpleIcon size={16} weight="bold" />
				</button>
				<button
					type="button"
					class="pointer-events-auto absolute top-1 right-1 flex h-8 w-8 items-center justify-center text-white"
					aria-label="Close player"
					onclick={close}
				>
					<XIcon size={16} weight="bold" />
				</button>
				<button
					type="button"
					class="glass-round pointer-events-auto absolute top-1/2 left-1/2 flex h-10 w-10 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full text-white"
					aria-label={playing ? 'Pause' : 'Play'}
					onclick={togglePlay}
				>
					{#if playing}
						<PauseIcon size={18} weight="fill" />
					{:else}
						<PlayIcon size={18} weight="fill" />
					{/if}
				</button>
			</div>
		{/if}

		<!-- Resize grip in the corner facing the middle of the screen. -->
		<span
			class="mini-grip absolute z-10 h-6 w-6"
			class:top-0={corner.startsWith('bottom')}
			class:bottom-0={corner.startsWith('top')}
			class:left-0={corner.endsWith('right')}
			class:right-0={corner.endsWith('left')}
			role="presentation"
			aria-hidden="true"
			onpointerdown={onGripDown}
		>
			<span class="mini-grip-mark"></span>
		</span>

		{#if $isCompact && duration > 0}
			<div class="pointer-events-none absolute inset-x-0 bottom-0 h-[3px] bg-white/20">
				<div class="h-full bg-[#ff1f4b]" style:width="{(position / duration) * 100}%"></div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.mini-player {
		position: fixed;
		touch-action: none;
		background-color: black;
		transition: width 200ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	.mini-dragging {
		transition: none;
		cursor: grabbing;
	}

	.mini-top {
		top: 1rem;
	}

	.mini-bottom {
		bottom: 1rem;
	}

	.mini-left {
		left: 1rem;
	}

	.mini-right {
		right: 1rem;
	}

	/* Phones: clear of the top bar and the floating tab bar. */
	.mini-compact.mini-top {
		top: calc(var(--chrome-offset) + 0.5rem);
	}

	.mini-compact.mini-bottom {
		bottom: calc(5.25rem + var(--safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
	}

	.mini-compact.mini-left {
		left: 0.75rem;
	}

	.mini-compact.mini-right {
		right: 0.75rem;
	}

	.mini-grip {
		cursor: nwse-resize;
		touch-action: none;
	}

	/* A small quarter-circle notch in the grip's corner. */
	.mini-grip-mark {
		position: absolute;
		inset: 0.3rem;
		border-radius: 9999px;
		background: rgba(255, 255, 255, 0.35);
		scale: 0.5;
	}
</style>
