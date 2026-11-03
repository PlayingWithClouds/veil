<!--
	One cam room: its live stream, played as soon as the page opens, then
	follow, recording (now or from back in the live buffer), auto-record with
	the room's quality, and the room's finished recordings.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import Toggle from '$lib/components/Toggle.svelte';
	import {
		RECORD_QUALITIES,
		RECORD_REWINDS,
		fetchCam,
		fetchCamFollow,
		fetchCamRecordings,
		fetchRoomRecording,
		followCam,
		setCamAutoRecord,
		startCamRecording,
		stopCamRecording,
		unfollowCam,
		watchCam,
		type ActiveRecordingRow,
		type CamRow
	} from '$lib/cams';
	import type { VideoCardRow } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { videoUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';

	const VIDEO_WIDTH = 280;
	// A running recording's length and size, and one just started or stopped.
	const RECORDING_REFRESH_MS = 5_000;

	const megabytes = new Intl.NumberFormat(undefined, {
		style: 'unit',
		unit: 'megabyte',
		maximumFractionDigits: 0
	});

	let cam = $state<CamRow | null>(null);
	let stream = $state<{ url: string; mimeType: string } | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let followId = $state<string | null>(null);
	let recording = $state<ActiveRecordingRow | null>(null);
	let recordings = $state<VideoCardRow[]>([]);
	let rewindSeconds = $state(0);
	// An action waiting for the backend; its buttons are disabled meanwhile.
	let busy = $state(false);

	let id = $derived(page.params.id ?? '');
	let quality = $derived(cam?.recordQuality ?? 'source');

	$effect(() => {
		void load(id);
	});

	$effect(() => {
		const roomId = id;
		const timer = setInterval(() => {
			refreshRecording(roomId).catch((failure) =>
				console.warn('recording refresh failed', failure)
			);
		}, RECORDING_REFRESH_MS);
		return () => clearInterval(timer);
	});

	/** Loads the room, its follow, recording and recordings, and while it's live, its stream. */
	async function load(camId: string) {
		cam = null;
		stream = null;
		error = null;
		followId = null;
		recording = null;
		recordings = [];
		loading = true;
		void loadExtras(camId);
		try {
			cam = await fetchCam(camId);
			if (cam?.live) stream = await watchCam(camId);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Loads what the controls and the recordings section show. */
	async function loadExtras(camId: string) {
		try {
			const [follow, running, finished] = await Promise.all([
				fetchCamFollow(camId),
				fetchRoomRecording(camId),
				fetchCamRecordings(camId)
			]);
			if (camId !== id) return;
			followId = follow;
			recording = running;
			recordings = finished;
		} catch (failure) {
			console.warn('cam room details failed to load', failure);
		}
	}

	/** Rereads the running recording; when one ends, its video joins the recordings. */
	async function refreshRecording(camId: string) {
		const running = await fetchRoomRecording(camId);
		if (camId !== id) return;
		const ended = recording !== null && running === null;
		recording = running;
		if (!ended) return;
		const finished = await fetchCamRecordings(camId);
		if (camId === id) recordings = finished;
	}

	/** Runs an action of the controls, showing what failed. */
	async function act(action: () => Promise<void>) {
		busy = true;
		error = null;
		try {
			await action();
		} catch (failure) {
			error = String(failure);
		} finally {
			busy = false;
		}
	}

	/** Follows the room, or stops following it. */
	function toggleFollow() {
		void act(async () => {
			if (followId) {
				await unfollowCam(followId);
				followId = null;
				return;
			}
			followId = await followCam(id);
		});
	}

	/** Starts recording at the room's quality, from the chosen point back. */
	function record() {
		void act(async () => {
			await startCamRecording(id, quality, rewindSeconds);
			await refreshRecording(id);
		});
	}

	/** Stops the recording; the backend files it as a video. */
	function stop() {
		void act(async () => {
			await stopCamRecording(id);
			await refreshRecording(id);
		});
	}

	/** Turns auto-record on or off. */
	function changeAutoRecord(on: boolean) {
		void act(async () => {
			cam = await setCamAutoRecord(id, on, quality);
		});
	}

	/** Keeps the room's quality, used by its recordings and auto-record. */
	function changeQuality(event: Event) {
		const chosen = (event.currentTarget as HTMLSelectElement).value;
		void act(async () => {
			cam = await setCamAutoRecord(id, cam?.autoRecord ?? false, chosen);
		});
	}
</script>

<svelte:head><title>{cam?.displayName ?? 'Cam'} · Veil</title></svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-4 pt-2">
	{#if stream && cam}
		<VideoPlayer
			src={stream.url}
			mimeType={stream.mimeType}
			title={cam.displayName}
			poster={cacheUrl(cam.live?.thumbnail?.uri ?? cam.thumbnail?.uri)}
			live
			inline
			square={$isCompact}
			onclose={() => (stream = null)}
		/>
	{:else}
		<div
			class="bg-base-200 text-base-content/70 flex aspect-video w-full items-center justify-center rounded-xl text-sm"
		>
			{#if loading}
				<span class="loading loading-spinner"></span>
			{:else if cam && !cam.live}
				{cam.displayName} is offline.
			{:else if !cam}
				This room isn't known.
			{/if}
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if cam}
		<div class="flex flex-wrap items-start justify-between gap-4">
			<div class="flex min-w-0 flex-col gap-1">
				<h1 class="text-xl font-semibold">{cam.displayName}</h1>
				<p class="text-base-content/60 text-sm">
					{siteDisplayName(cam.plugin, $sites)}
					{#if cam.live}· {cam.live.viewers} watching{/if}
				</p>
				{#if cam.live?.subject}
					<p class="text-base-content/80 text-sm">{cam.live.subject}</p>
				{/if}
			</div>
			<button
				type="button"
				class="btn btn-sm"
				class:btn-primary={followId === null}
				disabled={busy}
				onclick={toggleFollow}
				title="Notifies you when the room goes live"
			>
				{#if followId}
					<BellRingingIcon size={16} /> Following
				{:else}
					<BellIcon size={16} /> Follow
				{/if}
			</button>
		</div>

		<div class="bg-base-200 flex flex-wrap items-center gap-x-6 gap-y-3 rounded-xl p-4">
			{#if recording}
				<div class="flex items-center gap-3">
					<span class="text-error flex items-center gap-1.5 text-sm font-medium">
						<RecordIcon size={16} weight="fill" /> REC
					</span>
					<span class="text-sm tabular-nums">
						{formatClock(recording.durationSeconds) ?? '0:00'} · {megabytes.format(
							recording.sizeBytes / 1_000_000
						)}
					</span>
					<button type="button" class="btn btn-sm" disabled={busy} onclick={stop}>
						<StopIcon size={16} weight="fill" /> Stop
					</button>
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<select
						bind:value={rewindSeconds}
						class="bg-base-300 h-8 rounded-lg px-2 text-sm"
						aria-label="Start the recording from"
						disabled={!cam.live}
					>
						{#each RECORD_REWINDS as rewind (rewind.seconds)}
							<option value={rewind.seconds}>{rewind.label}</option>
						{/each}
					</select>
					<button
						type="button"
						class="btn btn-sm btn-primary"
						disabled={busy || !cam.live}
						onclick={record}
					>
						<RecordIcon size={16} weight="fill" /> Record
					</button>
				</div>
			{/if}

			<label class="flex items-center gap-2 text-sm">
				Quality
				<select
					value={quality}
					class="bg-base-300 h-8 rounded-lg px-2 text-sm"
					disabled={busy}
					onchange={changeQuality}
				>
					{#each RECORD_QUALITIES as choice (choice.value)}
						<option value={choice.value}>{choice.label}</option>
					{/each}
				</select>
			</label>

			<div class="flex items-center gap-2 text-sm">
				<Toggle
					checked={cam.autoRecord}
					disabled={busy}
					ariaLabel="Auto-record"
					onchange={changeAutoRecord}
				/>
				<span>
					Auto-record
					<span class="text-base-content/50">· records whenever the room goes live</span>
				</span>
			</div>
		</div>

		{#if recordings.length > 0}
			<section class="flex flex-col gap-4 pt-4">
				<h2 class="text-lg font-semibold">Recordings</h2>
				<div
					class="grid gap-x-4 gap-y-6"
					style:grid-template-columns="repeat(auto-fill, minmax({VIDEO_WIDTH}px, 1fr))"
				>
					{#each recordings as video, position (video.id)}
						<SceneCard
							{video}
							shown={{ surface: 'cam', position }}
							onclick={() => goto(videoUrl(video.id))}
						/>
					{/each}
				</div>
			</section>
		{/if}
	{/if}
</div>
