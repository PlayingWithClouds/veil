<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import SealCheckIcon from 'phosphor-svelte/lib/SealCheckIcon';
	import FocusList from './FocusList.svelte';
	import { formatSpeed } from '$lib/tmdb';
	import { resolveStream, queueDownload, type SceneStream } from '$lib/sceneDetail';
	import { notifications } from '$lib/stores/notifications';
	import { downloads } from '$lib/stores/downloads';
	import VideoPlayer from './VideoPlayer.svelte';

	interface Props {
		streams: SceneStream[];
		title: string;
		// Media record id for playback progress tracking (movie:… / episode:…).
		mediaId?: string;
		// TV navigation: edge callback from the internal FocusList.
		onedge?: (direction: 'left' | 'right' | 'up' | 'down') => void;
	}

	let { streams, title, mediaId, onedge }: Props = $props();

	// One FocusList entry per interactive control, so the whole sources row is
	// arrow-key navigable: Play, Download, the more-sources toggle, then each
	// alternative source once expanded.
	type SourceEntry =
		| { kind: 'play' }
		| { kind: 'download' }
		| { kind: 'toggle' }
		| { kind: 'alt'; stream: SceneStream };

	let entries = $derived.by<SourceEntry[]>(() => {
		if (streams.length === 0) return [];
		const list: SourceEntry[] = [{ kind: 'play' }, { kind: 'download' }];
		if (alternatives.length > 0) list.push({ kind: 'toggle' });
		if (showAlternatives) {
			for (const stream of alternatives) list.push({ kind: 'alt', stream });
		}
		return list;
	});

	let focusList = $state<{ focus: () => void; focusFirst: () => void }>();

	/** Enter the sources row at the Play button (TV navigation). */
	export function focusFirst() {
		focusList?.focusFirst();
	}

	export function focus() {
		focusList?.focus();
	}

	function activate(entry: SourceEntry) {
		if (entry.kind === 'play') play(best);
		else if (entry.kind === 'download') download(best);
		else if (entry.kind === 'toggle') showAlternatives = !showAlternatives;
		else play(entry.stream);
	}

	let resolving = $state<string | null>(null);
	let queueing = $state<string | null>(null);
	let toast = $state<{ msg: string; type: 'error' | 'info' } | null>(null);
	let activeStream = $state<{ src: string; mimeType: string; title: string } | null>(null);

	// Streams arrive ranked (verified first, then measured speed). The best one
	// backs the primary Play button; the rest hide behind a toggle.
	let best = $derived(streams[0] ?? null);
	let alternatives = $derived(streams.slice(1));
	let showAlternatives = $state(false);

	function showToast(msg: string, type: 'error' | 'info' = 'info') {
		toast = { msg, type };
		setTimeout(() => (toast = null), 4000);
	}

	function fileSize(bytes: number | null | undefined): string | null {
		if (!bytes) return null;
		if (bytes >= 1_000_000_000) return `${(bytes / 1_000_000_000).toFixed(1)} GB`;
		if (bytes >= 1_000_000) return `${(bytes / 1_000_000).toFixed(0)} MB`;
		return `${Math.round(bytes / 1000)} KB`;
	}

	function streamLabel(s: SceneStream): string {
		return s.label ?? s.provider ?? s.resolution ?? 'Source';
	}

	async function play(s: SceneStream) {
		resolving = s.id;
		try {
			const resolved = await resolveStream(s.url);
			if (!resolved) throw new Error('no stream');
			activeStream = {
				src: resolved.url,
				mimeType: resolved.mimeType,
				title: `${title} – ${streamLabel(s)}`
			};
		} catch {
			showToast('Could not resolve stream', 'error');
		} finally {
			resolving = null;
		}
	}

	async function download(s: SceneStream) {
		queueing = s.id;
		try {
			const jobId = await queueDownload(s.url, `${title} – ${streamLabel(s)}`);
			if (mediaId) downloads.start(jobId, mediaId);
			notifications.push(`Download started: ${title}`, 'success');
		} catch (e) {
			notifications.push(String(e), 'error');
		} finally {
			queueing = null;
		}
	}
</script>

{#if activeStream}
	<VideoPlayer
		src={activeStream.src}
		mimeType={activeStream.mimeType}
		title={activeStream.title}
		{mediaId}
		onclose={() => (activeStream = null)}
	/>
{/if}

{#if toast}
	<div class="toast toast-top toast-end z-50">
		<div class="alert {toast.type === 'error' ? 'alert-error' : 'alert-info'} text-sm">
			{toast.msg}
		</div>
	</div>
{/if}

{#if streams.length === 0}
	<p class="text-base-content/40 text-sm">No sources available yet.</p>
{:else}
	<div class="flex flex-col gap-3">
		<FocusList
			bind:this={focusList}
			items={entries}
			orientation="horizontal"
			ariaLabel="Sources"
			class="flex-wrap"
			scroll="top"
			onselect={(entry) => activate(entry)}
			{onedge}
		>
			{#snippet item(entry)}
				{#if entry.kind === 'play'}
					<span
						class="flex items-center gap-2 rounded-lg bg-white px-5 py-2 text-sm font-semibold text-black"
					>
						<PlayIcon size={14} weight="fill" />
						{resolving === best.id ? 'Loading…' : 'Play'}
					</span>
				{:else if entry.kind === 'download'}
					<span
						class="flex items-center gap-2 rounded-lg bg-white/10 px-3 py-2 text-white transition-colors hover:bg-white/20"
						title="Download to NAS"
					>
						{#if queueing === best.id}
							<span class="loading loading-xs"></span>
						{:else}
							<DownloadSimpleIcon size={16} />
						{/if}
					</span>
				{:else if entry.kind === 'toggle'}
					<span
						class="text-base-content/50 hover:text-base-content flex h-full items-center px-1 text-xs transition-colors"
					>
						{showAlternatives ? 'Hide sources' : `More sources (${alternatives.length})`}
					</span>
				{:else}
					<span
						class="flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-1.5 text-xs text-white"
						title="Play from {streamLabel(entry.stream)}"
					>
						<PlayIcon size={11} weight="fill" />
						<span class="font-medium">
							{resolving === entry.stream.id ? 'Loading…' : streamLabel(entry.stream)}
						</span>
						{#if entry.stream.verified}
							<SealCheckIcon size={12} weight="fill" class="text-success" />
						{/if}
						{#if entry.stream.resolution}<span class="opacity-60">{entry.stream.resolution}</span>{/if}
						{#if formatSpeed(entry.stream.expectedSpeedBps)}
							<span class="text-success">~{formatSpeed(entry.stream.expectedSpeedBps)}</span>
						{/if}
					</span>
				{/if}
			{/snippet}
		</FocusList>

		<span class="text-base-content/50 flex items-center gap-1.5 text-xs">
			{streamLabel(best)}
			{#if best.verified}
				<SealCheckIcon size={13} weight="fill" class="text-success" />
			{/if}
			{#if formatSpeed(best.expectedSpeedBps)}
				<span class="text-success">~{formatSpeed(best.expectedSpeedBps)}</span>
			{/if}
		</span>
	</div>
{/if}
