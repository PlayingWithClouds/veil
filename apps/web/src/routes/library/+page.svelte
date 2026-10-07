<script lang="ts">
	import { goto } from '$app/navigation';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import ArrowLineUpIcon from 'phosphor-svelte/lib/ArrowLineUpIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ActionSheet, { type SheetAction } from '$lib/components/ActionSheet.svelte';
	import { cacheUrl } from '$lib/img';
	import { sceneUrl } from '$lib/routes';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import DownloadCard from '$lib/components/library/DownloadCard.svelte';
	import SpeedSparkline from '$lib/components/library/SpeedSparkline.svelte';
	import { createEventStream } from '$lib/events';
	import { notifications } from '$lib/stores/notifications';
	import { fetchDownloadedScenes, type PluginSearchResult } from '$lib/search';
	import {
		deleteJob,
		deleteJobsByKind,
		fetchJobs,
		formatBytes,
		formatRate,
		isActiveDownload,
		isFailedDownload,
		reorderDownloads,
		retryJob,
		type BackgroundJob,
		type DownloadJob
	} from '$lib/downloads';

	// One speed sample per second, a minute of them in the sparkline.
	const SAMPLE_INTERVAL_MS = 1000;
	const SAMPLE_CAPACITY = 60;
	// Weight of the newest second in a download's smoothed speed.
	const RATE_SMOOTHING = 0.4;

	let downloads = $state<DownloadJob[]>([]);
	let backgroundJobs = $state<BackgroundJob[]>([]);
	let scenes = $state<PluginSearchResult[]>([]);
	let loading = $state(true);
	// Job id (or job kind) with a mutation in flight.
	let busyId = $state<string | null>(null);

	// Smoothed bytes per second per running download, and the total per second.
	let rates = $state<Record<string, number>>({});
	let speedSamples = $state<number[]>([]);
	const lastBytes = new Map<string, number>();

	let queue = $derived(downloads.filter((job) => isActiveDownload(job) || isFailedDownload(job)));
	let activeCount = $derived(downloads.filter(isActiveDownload).length);
	let currentSpeed = $derived(Object.values(rates).reduce((sum, rate) => sum + rate, 0));
	let librarySize = $derived(
		downloads
			.filter((job) => job.status === 'completed')
			.reduce((sum, job) => sum + (job.bytesTotal ?? 0), 0)
	);
	let jobKinds = $derived(countByKind(backgroundJobs));

	$effect(() => {
		reload();
	});

	// Live progress from job events; a finished download reloads the grid.
	$effect(() => {
		const stream = createEventStream(['job:updated']);
		const stopListening = stream.on('job:updated', (event) => {
			if (event.kind !== 'download') return;
			const known = downloads.find((job) => job.id === event.id);
			if (!known || event.status === 'completed' || event.status !== known.status) {
				reload();
				return;
			}
			applyProgress(known, event);
		});
		const timer = setInterval(sampleSpeed, SAMPLE_INTERVAL_MS);
		return () => {
			stopListening();
			stream.close();
			clearInterval(timer);
		};
	});

	/** Loads the jobs and the downloaded scenes. */
	async function reload() {
		const [jobs, downloaded] = await Promise.all([fetchJobs(), fetchDownloadedScenes()]);
		downloads = jobs.downloads;
		backgroundJobs = jobs.background;
		scenes = downloaded;
		loading = false;
	}

	/** Copies an event's progress onto its download. */
	function applyProgress(job: DownloadJob, event: { progress?: number; bytesReceived?: number; bytesTotal?: number }) {
		if (event.progress !== undefined) job.progress = event.progress;
		if (event.bytesReceived !== undefined) job.bytesReceived = event.bytesReceived;
		if (event.bytesTotal !== undefined) job.bytesTotal = event.bytesTotal;
	}

	/** Measures the bytes each running download got since the last second. */
	function sampleSpeed() {
		const nextRates: Record<string, number> = {};
		for (const job of downloads) {
			if (job.status !== 'running') continue;
			const received = job.bytesReceived ?? 0;
			const before = lastBytes.get(job.id);
			lastBytes.set(job.id, received);
			if (before === undefined) continue;
			nextRates[job.id] = smoothRate(rates[job.id], Math.max(0, received - before));
		}
		rates = nextRates;
		const total = Object.values(nextRates).reduce((sum, rate) => sum + rate, 0);
		speedSamples = [...speedSamples, total].slice(-SAMPLE_CAPACITY);
	}

	/** Blends the newest second into the previous speed so the numbers don't jump. */
	function smoothRate(previous: number | undefined, latest: number): number {
		if (previous === undefined) return latest;
		return previous * (1 - RATE_SMOOTHING) + latest * RATE_SMOOTHING;
	}

	/** Runs a job mutation with its row marked busy, then reloads. */
	async function withBusy(id: string, action: () => Promise<void>, failure: string) {
		busyId = id;
		try {
			await action();
			await reload();
		} catch {
			notifications.push(failure, 'error');
		} finally {
			busyId = null;
		}
	}

	/** Cancels a queued or running download, or dismisses a failed one. */
	function cancel(job: DownloadJob) {
		withBusy(job.id, () => deleteJob(job.id), 'Could not cancel the download');
	}

	/** Queues a failed download again. */
	function retry(job: DownloadJob) {
		withBusy(job.id, () => retryJob(job.id), 'Could not retry the download');
	}

	/** Removes every background job of one kind. */
	function clearKind(kind: string) {
		withBusy(kind, () => deleteJobsByKind(kind), 'Could not clear the jobs');
	}

	// The download whose action sheet is open (phone long press).
	let sheetJob = $state<DownloadJob | null>(null);

	/** Open, retry, download next, and cancel or remove, as fits the download. */
	function sheetActions(job: DownloadJob): SheetAction[] {
		const actions: SheetAction[] = [];
		const sceneId = job.scene?.sceneId;
		if (sceneId) actions.push({ label: 'Open video', icon: PlayIcon, run: () => goto(sceneUrl(sceneId)) });
		if (isFailedDownload(job)) {
			actions.push({ label: 'Retry download', icon: ArrowClockwiseIcon, run: () => retry(job) });
			actions.push({ label: 'Remove', icon: TrashIcon, run: () => cancel(job), destructive: true });
			return actions;
		}
		if (queue.indexOf(job) > 0) {
			actions.push({ label: 'Download next', icon: ArrowLineUpIcon, run: () => moveDownload(queue.indexOf(job), 0) });
		}
		actions.push({ label: 'Cancel download', icon: XIcon, run: () => cancel(job), destructive: true });
		return actions;
	}

	/** What the sheet header says under the title. */
	function sheetSubtitle(job: DownloadJob): string {
		if (isFailedDownload(job)) return 'Download failed';
		if (job.status !== 'running') return 'Queued';
		return `Downloading · ${Math.floor(job.progress ?? 0)}%`;
	}

	// Desktop drag and drop reorders the queue.
	let dragIndex = $state<number | null>(null);

	/** Moves the dragged download to targetIndex. */
	function dropAt(targetIndex: number) {
		const from = dragIndex;
		dragIndex = null;
		if (from === null) return;
		moveDownload(from, targetIndex);
	}

	/** Moves a queued download from one position to another and saves the order. */
	function moveDownload(from: number, targetIndex: number) {
		if (from === targetIndex) return;
		const reordered = [...queue];
		const [moved] = reordered.splice(from, 1);
		reordered.splice(targetIndex, 0, moved);
		const finished = downloads.filter((job) => !reordered.includes(job));
		downloads = [...reordered, ...finished];
		reorderDownloads(reordered.map((job) => job.id)).catch(() => {
			notifications.push('Could not reorder the queue', 'error');
		});
	}

	/** Background jobs per kind: total, and how many failed. */
	function countByKind(jobs: BackgroundJob[]): { kind: string; total: number; failed: number }[] {
		const counts = new Map<string, { kind: string; total: number; failed: number }>();
		for (const job of jobs) {
			let entry = counts.get(job.kind);
			if (!entry) {
				entry = { kind: job.kind, total: 0, failed: 0 };
				counts.set(job.kind, entry);
			}
			entry.total++;
			if (job.status === 'failed') entry.failed++;
		}
		return [...counts.values()];
	}

	/** "12 videos · 3.4 GB" once there is something. */
	function librarySummary(): string | null {
		if (scenes.length === 0) return null;
		let summary = `${scenes.length} videos`;
		if (scenes.length === 1) summary = '1 video';
		if (librarySize === 0) return summary;
		return `${summary} · ${formatBytes(librarySize)}`;
	}
</script>

<div class="flex flex-col gap-6">
	<PageHeader title="Library" description={librarySummary()} />

	{#if queue.length > 0}
		<section class="bg-base-200 flex flex-col gap-4 rounded-2xl p-4" aria-label="Downloads">
			<div class="flex items-end justify-between gap-4">
				<div class="flex flex-col">
					<h2 class="text-base font-semibold">
						{#if activeCount > 0}Downloading{:else}Downloads{/if}
					</h2>
					<span class="text-base-content/50 text-xs">
						{activeCount} active{#if queue.length > activeCount}
							· {queue.length - activeCount} failed{/if}
					</span>
				</div>
				<span class="text-success text-xl font-semibold tabular-nums">{formatRate(currentSpeed)}</span>
			</div>
			<div class="h-10">
				<SpeedSparkline samples={speedSamples} capacity={SAMPLE_CAPACITY} />
			</div>
			<ul class="flex flex-col gap-3">
				{#each queue as job, index (job.id)}
					<li
						draggable="true"
						ondragstart={() => (dragIndex = index)}
						ondragover={(event) => event.preventDefault()}
						ondrop={() => dropAt(index)}
						class:opacity-50={dragIndex === index}
					>
						<DownloadCard
							{job}
							rate={rates[job.id] ?? null}
							busy={busyId === job.id}
							onretry={() => retry(job)}
							oncancel={() => cancel(job)}
							onlongpress={() => (sheetJob = job)}
						/>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	{#if !loading && scenes.length === 0 && queue.length === 0}
		<div class="flex flex-col items-center gap-3 py-24 text-center">
			<DownloadSimpleIcon size={48} class="text-base-content/20" />
			<p class="text-base-content/50">Nothing downloaded yet.</p>
			<p class="text-base-content/30 max-w-xs text-sm">
				Tap Download on a scene to keep it here, playable without the site.
			</p>
		</div>
	{:else if scenes.length > 0 || loading}
		<section class="flex flex-col gap-3" aria-label="Downloaded">
			{#if queue.length > 0}
				<h2 class="text-base font-semibold">Downloaded</h2>
			{/if}
			<NsfwGrid items={scenes} loading={loading && scenes.length === 0} ariaLabel="Downloaded videos" />
		</section>
	{/if}

	{#if jobKinds.length > 0}
		<details class="bg-base-200 rounded-2xl px-4 py-3 text-sm">
			<summary class="text-base-content/60 cursor-pointer font-medium">
				Background jobs · {backgroundJobs.length}
			</summary>
			<ul class="mt-3 flex flex-col gap-2">
				{#each jobKinds as entry (entry.kind)}
					<li class="flex items-center gap-3">
						<span class="font-mono text-xs">{entry.kind}</span>
						<span class="text-base-content/50 flex-1 text-xs tabular-nums">
							{entry.total}{#if entry.failed > 0}<span class="text-error"> · {entry.failed} failed</span>{/if}
						</span>
						<button
							type="button"
							class="btn btn-xs btn-ghost"
							disabled={busyId !== null}
							onclick={() => clearKind(entry.kind)}
						>
							Clear
						</button>
					</li>
				{/each}
			</ul>
		</details>
	{/if}
</div>

{#if sheetJob}
	{@const job = sheetJob}
	<ActionSheet
		title={job.title}
		subtitle={sheetSubtitle(job)}
		imageUrl={cacheUrl(job.scene?.posterUrl)}
		actions={sheetActions(job)}
		onclose={() => (sheetJob = null)}
	/>
{/if}
