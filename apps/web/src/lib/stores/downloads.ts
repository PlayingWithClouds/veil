import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { createEventStream } from '$lib/events';
import { gqlClient } from '$lib/veil';

// Live download progress per media. Downloads are linked to their media via the
// job's `target` (set on the backend from the stream URL); we key by jobId and
// keep a jobId -> mediaId map so progress events (which carry no media id) still
// resolve.
export interface DownloadState {
	jobId: string;
	mediaId: string;
	progress: number;
	status: string;
}

const TERMINAL_STATUSES = ['completed', 'failed', 'canceled', 'cancelled'];

const mediaByJob = new Map<string, string>();

interface DownloadJob {
	id: string;
	target: string;
	downloadProgress: number | null;
	status: string;
}

async function fetchRunningDownloads(): Promise<DownloadJob[]> {
	const data = await gqlClient
		.query({
			jobs: { id: true, kind: true, status: true, target: true, downloadProgress: true }
		})
		.catch(() => ({ jobs: [] as never[] }));
	return (data.jobs ?? [])
		.filter(
			(job) => job.kind === 'download' && !!job.target && !TERMINAL_STATUSES.includes(job.status)
		)
		.map((job) => ({
			id: job.id,
			target: job.target as string,
			downloadProgress: job.downloadProgress ?? null,
			status: job.status
		}));
}

function createDownloads() {
	const { subscribe, update } = writable<Record<string, DownloadState>>({});

	function seed(runningJobs: DownloadJob[]) {
		update((byJob) => {
			const next = { ...byJob };
			for (const job of runningJobs) {
				mediaByJob.set(job.id, job.target);
				const previous = next[job.id];
				next[job.id] = {
					jobId: job.id,
					mediaId: job.target,
					progress: job.downloadProgress ?? previous?.progress ?? 0,
					status: job.status
				};
			}
			return next;
		});
	}

	async function refresh() {
		seed(await fetchRunningDownloads());
	}

	function start(jobId: string, mediaId: string) {
		mediaByJob.set(jobId, mediaId);
		update((byJob) => ({
			...byJob,
			[jobId]: { jobId, mediaId, progress: 0, status: 'pending' }
		}));
	}

	if (browser) {
		refresh();
		const stream = createEventStream(['job:updated']);
		stream.on('job:updated', (event) => {
			if (event.kind !== 'download') return;
			const mediaId = mediaByJob.get(event.id);
			if (!mediaId) {
				// A download we don't know yet (started elsewhere) — re-query to learn
				// its media, unless it already finished.
				if (!TERMINAL_STATUSES.includes(event.status)) refresh();
				return;
			}
			update((byJob) => {
				const next = { ...byJob };
				if (TERMINAL_STATUSES.includes(event.status)) {
					delete next[event.id];
					mediaByJob.delete(event.id);
					return next;
				}
				const previous = next[event.id];
				next[event.id] = {
					jobId: event.id,
					mediaId,
					progress: event.progress ?? previous?.progress ?? 0,
					status: event.status
				};
				return next;
			});
		});
	}

	return { subscribe, start, refresh };
}

export const downloads = createDownloads();

// Active downloads as a flat list.
export const activeDownloads = derived(downloads, ($downloads) => Object.values($downloads));

// Download progress (0..100) for a media id, or null when it isn't downloading.
export function progressForMedia(active: DownloadState[], mediaId: string): number | null {
	const entry = active.find((download) => download.mediaId === mediaId);
	if (!entry) return null;
	return entry.progress;
}

// Cancel a running download (removes its job from the queue).
export async function cancelDownload(jobId: string): Promise<void> {
	await gqlClient.mutation({ deleteJob: { __args: { id: jobId } } });
}
