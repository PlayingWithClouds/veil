<!--
	One download in the library queue: the scene's poster with a progress bar
	along its bottom, the title, and a status line (size, speed and time left
	while running; queued; or the error with a retry).
-->
<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { cacheUrl } from '$lib/img';
	import { longPress } from '$lib/longPress';
	import { sceneUrl } from '$lib/routes';
	import {
		formatBytes,
		formatRate,
		formatTimeLeft,
		isFailedDownload,
		type DownloadJob
	} from '$lib/downloads';

	interface Props {
		job: DownloadJob;
		// Current speed in bytes per second; null before the first measurement.
		rate: number | null;
		busy: boolean;
		onretry: () => void;
		oncancel: () => void;
		// Touch hold: the page opens the download's action sheet.
		onlongpress: () => void;
	}

	let { job, rate, busy, onretry, oncancel, onlongpress }: Props = $props();

	let failed = $derived(isFailedDownload(job));
	let running = $derived(job.status === 'running');
	let posterUrl = $derived(cacheUrl(job.scene?.posterUrl));

	/** The scene page, when the download belongs to a library scene. */
	function sceneLink(): string | undefined {
		if (!job.scene?.sceneId) return undefined;
		return sceneUrl(job.scene.sceneId);
	}

	/** "58.0 MB of 335.3 MB", or just the bytes so far when the total is unknown. */
	function sizeLabel(): string {
		const received = formatBytes(job.bytesReceived ?? 0);
		if (!job.bytesTotal) return received;
		return `${received} of ${formatBytes(job.bytesTotal)}`;
	}

	/** Speed and time left while bytes are arriving; empty otherwise. */
	function speedLabel(): string {
		if (!rate || rate <= 0) return '';
		if (!job.bytesTotal || job.bytesReceived === null) return formatRate(rate);
		const secondsLeft = (job.bytesTotal - job.bytesReceived) / rate;
		return `${formatRate(rate)} · ${formatTimeLeft(secondsLeft)}`;
	}
</script>

<div class="flex items-center gap-3" use:longPress={onlongpress}>
	<!-- Poster and text open the video; without a scene there is no href and
	     the anchor is a plain box. -->
	<a href={sceneLink()} class="download-link flex min-w-0 flex-1 items-center gap-3">
	<span class="bg-base-200 relative aspect-video w-32 shrink-0 overflow-hidden rounded-lg">
		{#if posterUrl}
			<img src={posterUrl} alt="" loading="lazy" class="h-full w-full object-cover" class:grayscale={failed} />
		{:else}
			<span class="text-base-content/20 flex h-full w-full items-center justify-center">
				<FilmStripIcon size={28} />
			</span>
		{/if}
		{#if !failed}
			<div class="absolute inset-x-0 bottom-0 h-1 bg-black/50">
				<div
					class="bg-success h-full transition-[width] duration-700"
					class:progress-pulse={running && job.progress === null}
					style:width="{job.progress ?? 0}%"
				></div>
			</div>
		{/if}
		{#if running && job.progress !== null}
			<span
				class="absolute top-1 right-1 rounded bg-black/70 px-1 text-[11px] font-semibold text-white tabular-nums"
			>
				{Math.floor(job.progress)}%
			</span>
		{/if}
	</span>

	<span class="flex min-w-0 flex-1 flex-col gap-0.5">
		<span class="line-clamp-2 text-sm leading-snug font-medium">{job.title}</span>
		{#if failed}
			<span class="text-error line-clamp-1 text-xs" title={job.error ?? undefined}>
				{job.error ?? 'Download failed'}
			</span>
		{:else if running}
			<span class="text-base-content/60 text-xs tabular-nums">{sizeLabel()}</span>
			{#if speedLabel()}
				<span class="text-success text-xs tabular-nums">{speedLabel()}</span>
			{/if}
		{:else}
			<span class="text-base-content/50 text-xs">Queued</span>
		{/if}
	</span>
	</a>

	{#if failed}
		<button
			type="button"
			class="glass-pill flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
			aria-label="Retry {job.title}"
			disabled={busy}
			onclick={onretry}
		>
			<ArrowClockwiseIcon size={16} weight="bold" />
		</button>
	{/if}
	<button
		type="button"
		class="text-base-content/50 hover:text-error flex h-9 w-9 shrink-0 items-center justify-center rounded-full transition-colors"
		aria-label="Cancel {job.title}"
		disabled={busy}
		onclick={oncancel}
	>
		{#if busy}
			<span class="loading loading-spinner loading-xs"></span>
		{:else}
			<XIcon size={16} weight="bold" />
		{/if}
	</button>
</div>

<style>
	/* Pressed feedback on the tappable part, like the scene cards. */
	.download-link {
		transition: transform 150ms;
		-webkit-touch-callout: none;
		user-select: none;
	}

	.download-link:active {
		transform: scale(0.98);
	}

	.progress-pulse {
		width: 30% !important;
		animation: progress-slide 1.2s ease-in-out infinite;
	}

	@keyframes progress-slide {
		from {
			transform: translateX(-100%);
		}
		to {
			transform: translateX(340%);
		}
	}
</style>
