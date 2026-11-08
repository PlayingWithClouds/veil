// Downloads kept for good: each a background job, and once it's done the
// stored file. Started from a video's page (downloadVideo in library.ts).
import type { JobStatus } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, type VideoCardRow } from '$lib/library';

/** One download as the library lists it. */
export interface DownloadRow {
	jobId: string;
	status: JobStatus;
	// 0..1
	progress: number;
	error: string | null;
	sizeBytes: number | null;
	video: VideoCardRow;
}

/** The downloads of videos, newest first; downloads of records since gone are left out. */
export async function fetchDownloads(): Promise<DownloadRow[]> {
	const data = await gqlClient.query({
		downloads: {
			job: {
				id: true,
				status: true,
				progress: true,
				error: true,
				createdAt: true,
				subject: { kind: true, id: true }
			},
			file: { sizeBytes: true }
		}
	});
	const downloads = data.downloads
		.filter((download) => download.job.subject?.kind === 'video')
		.toSorted((first, second) => second.job.createdAt.localeCompare(first.job.createdAt));
	const videos = await Promise.all(
		downloads.map((download) => fetchVideoCard(download.job.subject?.id ?? ''))
	);
	return downloads.flatMap((download, index) => {
		const video = videos[index];
		if (!video) return [];
		return [
			{
				jobId: download.job.id,
				status: download.job.status,
				progress: download.job.progress,
				error: download.job.error ?? null,
				sizeBytes: download.file?.sizeBytes ?? null,
				video
			}
		];
	});
}

/** A video as its card shows it; null when it's gone. */
async function fetchVideoCard(id: string): Promise<VideoCardRow | null> {
	const data = await gqlClient.query({ video: { __args: { id }, ...VIDEO_CARD } });
	return data.video ?? null;
}

/** Stops a download and deletes it with its file. */
export async function deleteDownload(jobId: string): Promise<void> {
	await gqlClient.mutation({ deleteDownload: { __args: { jobId } } });
}
