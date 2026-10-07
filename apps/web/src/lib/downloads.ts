// Download and background jobs for the library page: the queue with live
// progress, retrying, cancelling and reordering, plus byte and rate formatting.
import { gqlClient } from '$lib/veil';
import { browseSceneToResult, SCENE_CARD_FIELDS, type BrowseSceneData, type PluginSearchResult } from '$lib/search';

// A download job as the library shows it.
export interface DownloadJob {
	id: string;
	status: string;
	title: string;
	// The scene being downloaded, when the job targets one.
	scene: PluginSearchResult | null;
	progress: number | null;
	bytesReceived: number | null;
	bytesTotal: number | null;
	url: string | null;
	error: string | null;
	updatedAt: string;
}

// A non-download job (scrape, enrich, speed-check …), listed for maintenance.
export interface BackgroundJob {
	id: string;
	kind: string;
	status: string;
	pluginName: string;
	error: string | null;
	updatedAt: string;
}

// Download states that are over, one way or the other.
const FINISHED_STATUSES = ['completed', 'failed', 'cancelled', 'canceled'];

/** Whether a download is still queued or running. */
export function isActiveDownload(job: DownloadJob): boolean {
	return !FINISHED_STATUSES.includes(job.status);
}

/** Whether a download ended in failure (or was cancelled) and can be retried. */
export function isFailedDownload(job: DownloadJob): boolean {
	return job.status !== 'completed' && FINISHED_STATUSES.includes(job.status);
}

/** Loads every job, split into downloads (queue order) and background jobs. */
export async function fetchJobs(): Promise<{ downloads: DownloadJob[]; background: BackgroundJob[] }> {
	try {
		const data = await gqlClient.query({
			jobs: {
				id: true,
				kind: true,
				status: true,
				pluginName: true,
				error: true,
				updatedAt: true,
				downloadTitle: true,
				downloadUrl: true,
				downloadProgress: true,
				downloadBytesReceived: true,
				downloadBytesTotal: true,
				scene: SCENE_CARD_FIELDS
			}
		});
		const downloads: DownloadJob[] = [];
		const background: BackgroundJob[] = [];
		for (const job of data.jobs) {
			if (job.kind !== 'download') {
				background.push({
					id: job.id,
					kind: job.kind,
					status: job.status,
					pluginName: job.pluginName,
					error: job.error ?? null,
					updatedAt: job.updatedAt
				});
				continue;
			}
			let scene: PluginSearchResult | null = null;
			if (job.scene) scene = browseSceneToResult(job.scene as unknown as BrowseSceneData);
			downloads.push({
				id: job.id,
				status: job.status,
				title: downloadTitle(job.downloadTitle, scene),
				scene,
				progress: job.downloadProgress ?? null,
				bytesReceived: job.downloadBytesReceived ?? null,
				bytesTotal: job.downloadBytesTotal ?? null,
				url: job.downloadUrl ?? null,
				error: job.error ?? null,
				updatedAt: job.updatedAt
			});
		}
		return { downloads, background };
	} catch {
		return { downloads: [], background: [] };
	}
}

/** The scene's title, else the job's own, else a placeholder. */
function downloadTitle(jobTitle: string | null | undefined, scene: PluginSearchResult | null): string {
	if (scene) return scene.title;
	if (jobTitle) return jobTitle;
	return 'Download';
}

/** Queues a failed download again. */
export async function retryJob(id: string): Promise<void> {
	await gqlClient.mutation({ retryJob: { __args: { id } } });
}

/** Removes a job; a download's file stays where it is. */
export async function deleteJob(id: string): Promise<void> {
	await gqlClient.mutation({ deleteJob: { __args: { id } } });
}

/** Removes every job of one kind. */
export async function deleteJobsByKind(kind: string): Promise<void> {
	await gqlClient.mutation({ deleteJobsByKind: { __args: { kind } } });
}

/** Saves the download queue order; earlier ids run first. */
export async function reorderDownloads(jobIds: string[]): Promise<void> {
	await gqlClient.mutation({ reorderDownloads: { __args: { jobIds } } });
}

const KILOBYTE = 1024;
const MEGABYTE = KILOBYTE * 1024;
const GIGABYTE = MEGABYTE * 1024;

/** "335.3 MB", "1.2 GB", "0 B". */
export function formatBytes(bytes: number): string {
	if (!bytes || bytes < 1) return '0 B';
	if (bytes >= GIGABYTE) return `${(bytes / GIGABYTE).toFixed(1)} GB`;
	if (bytes >= MEGABYTE) return `${(bytes / MEGABYTE).toFixed(1)} MB`;
	if (bytes >= KILOBYTE) return `${(bytes / KILOBYTE).toFixed(0)} KB`;
	return `${Math.round(bytes)} B`;
}

/** "2.0 MB/s". */
export function formatRate(bytesPerSecond: number): string {
	return `${formatBytes(bytesPerSecond)}/s`;
}

/** Time left as "3 min left", "1 h 20 min left" or "under a minute left". */
export function formatTimeLeft(seconds: number): string {
	if (seconds < 60) return 'under a minute left';
	const minutes = Math.round(seconds / 60);
	if (minutes < 60) return `${minutes} min left`;
	const hours = Math.floor(minutes / 60);
	return `${hours} h ${minutes % 60} min left`;
}
