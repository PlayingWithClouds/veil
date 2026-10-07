<script lang="ts" module>
	interface HoverSource {
		// The scene stream it was resolved from.
		streamId: string;
		url: string;
		mimeType: string;
		headers: Record<string, string>;
	}

	// Per scene, where the last hover left off and the stream it played, so
	// hovering the card again resumes at once. Lives for the page session.
	const resumeAt = new Map<string, { position: number; muted: boolean }>();
	const resolvedSources = new Map<string, HoverSource>();
</script>

<script lang="ts">
	// Plays the scene itself inside a card once the hover has outlasted the
	// preview. Opening the scene hands this very playback to the scene page, so
	// the video continues there without reloading.
	import { untrack } from 'svelte';
	import { get } from 'svelte/store';
	import CompactPlayer from './CompactPlayer.svelte';
	import { ensureStreamsCached } from '$lib/cardActions';
	import { resolveStream } from '$lib/sceneDetail';
	import { handOffToScene, Playback, promoteToMain } from '$lib/playback';
	import { miniPlayer } from '$lib/stores/miniPlayer';
	import { watchProgress } from '$lib/stores/watchProgress';

	interface Props {
		sceneId: string;
		title: string;
		// Preview length: frames of the strip, or the preview clip's seconds (0 = none).
		frameCount: number;
		clipSeconds: number;
		// Called with true once the video is playing (the card then stops its
		// own preview) and false when it goes.
		onshow: (shown: boolean) => void;
		onopen: () => void;
	}

	let { sceneId, title, frameCount, clipSeconds, onshow, onopen }: Props = $props();

	// One frame of the preview strip is shown this long (matches the cards).
	const FRAME_MS = 700;
	// Wait at least this long, e.g. for a card without any preview.
	const MIN_DELAY_MS = 2500;
	// Same threshold as the cards' stream prefetch.
	const RESOLVE_AFTER_MS = 1400;

	const hoveredAt = performance.now();
	// A card hovered before skips the preview and the wait for its stream.
	// (Each hover mounts a fresh player, so the scene is fixed for its lifetime.)
	const resuming = untrack(() => resumeAt.has(sceneId));
	let delayPassed = $state(resuming);
	let source = $state<HoverSource | null>(untrack(() => resolvedSources.get(sceneId)) ?? null);
	let playback = $state<Playback | null>(null);
	let visible = $state(false);
	let opened = false;

	// Resolve the stream in the background once the hover is deliberate: it
	// visits the scene (a detail fetch for stubs), which a pass-over shouldn't.
	$effect(() => {
		if (source) return;
		let cancelled = false;
		const timer = setTimeout(() => {
			resolveBestSource(sceneId).then((resolved) => {
				if (cancelled || !resolved) return;
				resolvedSources.set(sceneId, resolved);
				source = resolved;
			});
		}, RESOLVE_AFTER_MS);
		return () => {
			cancelled = true;
			clearTimeout(timer);
		};
	});

	// Start once the preview has had its turn; the clip's length may only be
	// known after the timer started, so count from the hover's start.
	$effect(() => {
		if (delayPassed) return;
		const remaining = hoveredAt + previewMs(frameCount, clipSeconds) - performance.now();
		const timer = setTimeout(() => (delayPassed = true), Math.max(0, remaining));
		return () => clearTimeout(timer);
	});

	$effect(() => {
		if (!delayPassed || !source || playback) return;
		playback = startPlayback(source);
	});

	// Leaving the card remembers the spot and stops the video, unless it went on
	// to the scene page.
	$effect(() => {
		return () => {
			onshow(false);
			if (!playback || opened) return;
			if (visible) {
				resumeAt.set(sceneId, {
					position: playback.video.currentTime,
					muted: playback.video.muted
				});
			}
			playback.stop();
		};
	});

	/** How long the card's own preview runs before real playback takes over. */
	function previewMs(frames: number, clip: number): number {
		if (clip > 0) return Math.max(MIN_DELAY_MS, clip * 1000);
		return Math.max(MIN_DELAY_MS, frames * FRAME_MS);
	}

	/** The scene's best stream as a playable URL, or null. */
	async function resolveBestSource(id: string): Promise<HoverSource | null> {
		try {
			const streams = await ensureStreamsCached(id);
			if (streams.length === 0) return null;
			const resolved = await resolveStream(streams[0].url);
			if (!resolved) return null;
			const headers: Record<string, string> = {};
			for (const header of resolved.headers ?? []) {
				headers[header.name] = header.value;
			}
			return { streamId: streams[0].id, url: resolved.url, mimeType: resolved.mimeType, headers };
		} catch {
			return null;
		}
	}

	/**
	 * A playback of the source from where the last hover stopped, else the saved
	 * watch position; muted unless the last hover unmuted it. Shown once it plays.
	 */
	function startPlayback(from: HoverSource): Playback {
		const started = new Playback();
		const previous = resumeAt.get(sceneId);
		started.video.muted = previous ? previous.muted : true;
		const position = startPosition(previous);
		if (position > 0) {
			started.video.addEventListener(
				'loadedmetadata',
				() => (started.video.currentTime = position),
				{ once: true }
			);
		}
		started.video.addEventListener(
			'playing',
			() => {
				visible = true;
				onshow(true);
			},
			{ once: true }
		);
		started.load(from.url, from.mimeType, from.headers);
		return started;
	}

	/** Seconds to start from: the last hover's spot, else the saved watch position. */
	function startPosition(previous: { position: number } | undefined): number {
		if (previous) return previous.position;
		const progress = get(watchProgress).get(sceneId);
		if (progress) return progress.progressSeconds;
		return 0;
	}

	/**
	 * Opens the scene, handing it this playback and its exact stream so the page
	 * continues it in place (unmuted: the click allows sound).
	 */
	function open() {
		if (playback && source) {
			opened = true;
			resumeAt.delete(sceneId);
			miniPlayer.set(null);
			promoteToMain(playback);
			playback.video.muted = false;
			handOffToScene({
				sceneId,
				title,
				src: source.url,
				mimeType: source.mimeType,
				sourceId: source.streamId,
				position: playback.video.currentTime
			});
		}
		onopen();
	}
</script>

{#if playback}
	<!-- Takes the place of the card's banner (same box, rounding and hover scale),
	     hidden until frames play. Not .tv-card: its position: relative would
	     push the player below the banner. -->
	<div
		class="absolute inset-x-0 top-0 z-10 aspect-video w-full overflow-hidden rounded-[var(--tv-radius)] transition-[transform,opacity] duration-200 group-hover:scale-[1.02]"
		class:opacity-0={!visible}
		class:pointer-events-none={!visible}
	>
		<CompactPlayer {playback} fit="cover" onvideoclick={open} />
	</div>
{/if}
