<!--
	One pane of the split view: resolves the pane's stream (a cam's live HLS, or
	a scene's best source) and plays it in its own VideoPlayer. Only the focused
	pane has sound, unless the view plays every pane's; the player is silenced otherwise.
-->
<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import VideoPlayer from './VideoPlayer.svelte';
	import { fetchCamModel, fetchCamStream, loadWatchQuality } from '$lib/cams';
	import { cacheUrl } from '$lib/img';
	import { ensureSceneStreams, loadStreams, resolveStream } from '$lib/sceneDetail';
	import type { SplitPane } from '$lib/stores/split';
	import { watchProgress } from '$lib/stores/watchProgress';

	interface Props {
		pane: SplitPane;
		/** Gets the keyboard shortcuts (the pane last clicked). */
		focused: boolean;
		/** Plays its sound. */
		audible: boolean;
	}

	let { pane, focused, audible }: Props = $props();

	// A dropped cam stream is retried with a fresh URL (the site's tokens expire
	// within about a minute) this many times before the room counts as offline.
	const MAX_RECONNECTS = 3;

	let source = $state<{ url: string; mimeType: string } | null>(null);
	let loading = $state(true);
	let failed = $state(false);
	let poster = $state<string | null>(null);
	let attempt = $state(0);
	let reconnects = 0;
	const watchHeight = loadWatchQuality();

	let startPosition = $derived.by(() => {
		if (pane.kind !== 'scene') return 0;
		const progress = $watchProgress.get(pane.mediaId);
		if (!progress) return 0;
		return progress.progressSeconds;
	});

	// Memoized, so the effect below reruns only when the stream itself changes,
	// not whenever the split store publishes (focus or tab moves).
	let paneKind = $derived(pane.kind);
	let paneMediaId = $derived(pane.mediaId);

	// Resolve the stream again whenever the pane changes or "Try again" is pressed.
	$effect(() => {
		void attempt;
		const kind = paneKind;
		const mediaId = paneMediaId;
		let cancelled = false;
		loading = true;
		failed = false;
		source = null;
		reconnects = 0;
		resolvePane(kind, mediaId, () => cancelled);
		return () => {
			cancelled = true;
		};
	});

	/** Resolves the pane's stream into a playable source, or marks the pane failed. */
	async function resolvePane(kind: string, mediaId: string, isCancelled: () => boolean) {
		try {
			let resolved: { url: string; mimeType: string } | null;
			if (kind === 'cam') {
				resolved = await resolveCam(mediaId);
			} else {
				resolved = await resolveScene(mediaId);
			}
			if (isCancelled()) return;
			source = resolved;
			failed = resolved === null;
		} catch {
			if (isCancelled()) return;
			failed = true;
		}
		loading = false;
	}

	/** The room's live stream, or null when it is offline. */
	async function resolveCam(id: string): Promise<{ url: string; mimeType: string } | null> {
		const camModel = await fetchCamModel(id);
		if (!camModel) return null;
		poster = cacheUrl(camModel.thumbnailUrl);
		if (!camModel.online) return null;
		return fetchCamStream(id);
	}

	/** The scene's best-ranked source, resolved to a playable URL. */
	async function resolveScene(id: string): Promise<{ url: string; mimeType: string } | null> {
		let streams = await loadStreams(id);
		if (streams.length === 0) streams = await ensureSceneStreams(id);
		if (streams.length === 0) return null;
		return resolveStream(streams[0].url);
	}

	/** A cam stream died: ask for a fresh URL, and give up after a few tries. */
	async function onStreamError() {
		if (reconnects >= MAX_RECONNECTS) {
			source = null;
			failed = true;
			return;
		}
		reconnects += 1;
		try {
			source = await fetchCamStream(pane.mediaId);
		} catch {
			source = null;
			failed = true;
		}
	}
</script>

<div class="relative h-full w-full bg-black">
	{#if source}
		{#key source.url}
			<VideoPlayer
				src={source.url}
				mimeType={source.mimeType}
				title={pane.title}
				{poster}
				mediaId={pane.kind === 'scene' ? pane.mediaId : undefined}
				{startPosition}
				live={pane.kind === 'cam'}
				levelHeight={watchHeight}
				onstreamerror={onStreamError}
				silenced={!audible}
				keyboard={focused}
				inline
				fill
				square
				onclose={() => {}}
			/>
		{/key}
	{:else if failed}
		<div class="absolute inset-0 flex flex-col items-center justify-center gap-3 p-4 text-center">
			<p class="text-sm font-semibold text-white/80">
				{pane.kind === 'cam' ? `${pane.title} is offline` : 'Could not load this video'}
			</p>
			<button type="button" class="btn btn-sm" onclick={() => (attempt += 1)}>
				<ArrowClockwiseIcon size={16} />
				Try again
			</button>
		</div>
	{:else if loading}
		<div class="absolute inset-0 flex items-center justify-center">
			<span class="loading loading-spinner text-white/40"></span>
		</div>
	{/if}
</div>
