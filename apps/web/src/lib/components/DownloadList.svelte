<!--
	The downloads, newest first: each video's card says how far it got, how
	big it is or why it failed, and its ⋮ menu deletes it with its file.
	Reloads every few seconds while one is still queued or running.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import { deleteDownload, fetchDownloads, type DownloadRow } from '$lib/downloads';
	import { videoUrl } from '$lib/routes';

	// How often the list reloads while a download is under way.
	const POLL_MS = 3000;
	const BYTES_PER_GB = 1024 ** 3;
	const BYTES_PER_MB = 1024 ** 2;

	let downloads = $state<DownloadRow[] | null>(null);
	let error = $state<string | null>(null);

	let active = $derived(
		(downloads ?? []).some(
			(download) => download.status === 'queued' || download.status === 'running'
		)
	);

	$effect(() => {
		void load();
	});

	$effect(() => {
		if (!active) return;
		const timer = setInterval(() => void load(), POLL_MS);
		return () => clearInterval(timer);
	});

	/** Loads the downloads again. */
	async function load() {
		try {
			downloads = await fetchDownloads();
			error = null;
		} catch (failure) {
			error = String(failure);
		}
	}

	/** Deletes a download with its file, hiding it at once. */
	async function remove(download: DownloadRow) {
		downloads = (downloads ?? []).filter((other) => other.jobId !== download.jobId);
		try {
			await deleteDownload(download.jobId);
		} catch (failure) {
			error = String(failure);
			await load();
		}
	}

	/** The card's last line: progress, size or what went wrong. */
	function statusNote(download: DownloadRow): string {
		if (download.status === 'queued') return 'Waiting to download';
		if (download.status === 'running')
			return `Downloading · ${Math.round(download.progress * 100)}%`;
		if (download.status === 'failed') return `Failed: ${download.error ?? 'unknown error'}`;
		if (download.status === 'cancelled') return 'Cancelled';
		if (download.sizeBytes === null) return 'Downloaded';
		return `Downloaded · ${formatSize(download.sizeBytes)}`;
	}

	/** A file size in GB or MB. */
	function formatSize(bytes: number): string {
		if (bytes >= BYTES_PER_GB) return `${(bytes / BYTES_PER_GB).toFixed(1)} GB`;
		return `${Math.round(bytes / BYTES_PER_MB)} MB`;
	}
</script>

{#if error}
	<p class="text-error text-sm">{error}</p>
{/if}

{#if downloads === null}
	<div class="flex justify-center py-10">
		<span class="loading loading-spinner"></span>
	</div>
{:else if downloads.length === 0}
	<p class="text-base-content/50 py-10 text-center text-sm">
		Nothing downloaded yet. Download a video from its page to keep it.
	</p>
{:else}
	<div class="card-grid">
		{#each downloads as download, position (download.jobId)}
			<SceneCard
				video={download.video}
				note={statusNote(download)}
				shown={{ surface: 'downloads', position }}
				actions={[
					{
						label: 'Delete download',
						icon: TrashIcon,
						destructive: true,
						run: () => remove(download)
					}
				]}
				onclick={() => goto(videoUrl(download.video.id))}
			/>
		{/each}
	</div>
{/if}
