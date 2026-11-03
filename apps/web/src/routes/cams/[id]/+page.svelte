<!--
	One cam room: its live stream, played as soon as the page opens; its
	performer (linked from the library), topic and tags; follow, recording
	(now or from back in the live buffer) and auto-record at the room's
	quality; when it's usually on; its finished recordings; and similar rooms
	live now.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import Toggle from '$lib/components/Toggle.svelte';
	import UsualHours from '$lib/components/UsualHours.svelte';
	import RecordPickerDialog from '$lib/components/RecordPickerDialog.svelte';
	import {
		RECORD_QUALITIES,
		RECORD_REWINDS,
		fetchCam,
		fetchCamRecordings,
		fetchCamShows,
		fetchRoomRecording,
		fetchSimilarRooms,
		followedCams,
		linkCamPerformer,
		loadFollowedCams,
		roomTopic,
		setCamAutoRecord,
		startCamRecording,
		stopCamRecording,
		toggleCamFollow,
		watchCam,
		type ActiveRecordingRow,
		type CamRow,
		type CamShowRow
	} from '$lib/cams';
	import { fetchPerformers, type VideoCardRow } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { camUrl, camsUrl, performerUrl, videoUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';

	const VIDEO_WIDTH = 280;
	const ROOM_WIDTH = 220;
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
	let recording = $state<ActiveRecordingRow | null>(null);
	let recordings = $state<VideoCardRow[]>([]);
	let shows = $state<CamShowRow[]>([]);
	let similar = $state<CamRow[]>([]);
	let rewindSeconds = $state(0);
	let pickingPerformer = $state(false);
	// An action waiting for the backend; its buttons are disabled meanwhile.
	let busy = $state(false);

	let id = $derived(page.params.id ?? '');
	let quality = $derived(cam?.recordQuality ?? 'source');
	let followed = $derived($followedCams.has(id));
	let topic = $derived(roomTopic(cam?.live?.subject));

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

	/** Loads the room and its extras, and while it's live, its stream. */
	async function load(camId: string) {
		cam = null;
		stream = null;
		error = null;
		recording = null;
		recordings = [];
		shows = [];
		similar = [];
		loading = true;
		void loadExtras(camId);
		try {
			cam = await fetchCam(camId);
			if (cam) void loadSimilar(cam);
			if (cam?.live) stream = await watchCam(camId);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Loads what the controls, the schedule and the recordings section show. */
	async function loadExtras(camId: string) {
		try {
			const [running, finished, pastShows] = await Promise.all([
				fetchRoomRecording(camId),
				fetchCamRecordings(camId),
				fetchCamShows(camId),
				loadFollowedCams()
			]);
			if (camId !== id) return;
			recording = running;
			recordings = finished;
			shows = pastShows;
		} catch (failure) {
			console.warn('cam room details failed to load', failure);
		}
	}

	/** Loads the rooms live now that share the room's tags. */
	async function loadSimilar(room: CamRow) {
		try {
			const found = await fetchSimilarRooms(room);
			if (room.id === id) similar = found;
		} catch (failure) {
			console.warn('similar rooms failed to load', failure);
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

	/** Links the room to a library performer, or unlinks it with null. */
	function linkPerformer(performerId: string | null) {
		pickingPerformer = false;
		void act(async () => {
			cam = await linkCamPerformer(id, performerId);
		});
	}

	/** The cams page narrowed to one tag. */
	function tagUrl(tag: string): string {
		return `${camsUrl()}?${new URLSearchParams({ tag })}`;
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
				{cam.displayName} is offline{#if cam.lastSeenOnlineAt}, last seen {dayjs(
						cam.lastSeenOnlineAt
					).fromNow()}{/if}.
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
			<div class="flex min-w-0 flex-col gap-1.5">
				<h1 class="text-xl font-semibold">{cam.displayName}</h1>
				<p class="text-base-content/60 flex flex-wrap items-center gap-x-1.5 text-sm">
					{#if cam.performer}
						<a href={performerUrl(cam.performer.id)} class="text-base-content hover:underline">
							{cam.performer.name}
						</a>
						<span aria-hidden="true">·</span>
					{/if}
					<span>{siteDisplayName(cam.plugin, $sites)}</span>
					{#if cam.live}
						<span aria-hidden="true">·</span>
						<span>{cam.live.viewers.toLocaleString()} watching</span>
					{/if}
					{#if cam.live?.startedAt}
						<span aria-hidden="true">·</span>
						<span>on for {dayjs(cam.live.startedAt).fromNow(true)}</span>
					{/if}
				</p>
				{#if topic}
					<p class="text-base-content/80 text-sm">{topic}</p>
				{/if}
				{#if cam.tags.length > 0}
					<div class="flex flex-wrap gap-1.5 pt-1">
						{#each cam.tags as tag (tag)}
							<a
								href={tagUrl(tag)}
								class="bg-base-200 rounded-full px-2.5 py-0.5 text-xs hover:brightness-125"
							>
								#{tag}
							</a>
						{/each}
					</div>
				{/if}
			</div>
			<div class="flex shrink-0 items-center gap-2">
				<button
					type="button"
					class="btn btn-sm"
					disabled={busy}
					onclick={() => (pickingPerformer = true)}
					title="The same person in the library; her recordings credit them"
				>
					<UserIcon size={16} />
					{cam.performer ? 'Change performer' : 'Link a performer'}
				</button>
				<button
					type="button"
					class="btn btn-sm"
					class:btn-primary={!followed}
					disabled={busy}
					onclick={() => void act(() => toggleCamFollow(id))}
					title="Notifies you when the room goes live"
				>
					{#if followed}
						<BellRingingIcon size={16} /> Following
					{:else}
						<BellIcon size={16} /> Follow
					{/if}
				</button>
			</div>
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

		{#if shows.length > 0}
			<section class="flex flex-col gap-3 pt-4">
				<h2 class="text-lg font-semibold">
					Usually on
					<span class="text-base-content/50 text-sm font-normal">
						· from {shows.length}
						{shows.length === 1 ? 'show' : 'shows'} seen
					</span>
				</h2>
				<UsualHours {shows} />
			</section>
		{/if}

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

		<div class="pt-4">
			<Shelf title="Similar rooms" items={similar} key={(room) => room.id} itemWidth={ROOM_WIDTH}>
				{#snippet card(room)}
					<CamCard cam={room} onclick={() => goto(camUrl(room.id))} />
				{/snippet}
			</Shelf>
		</div>
	{/if}
</div>

{#if pickingPerformer && cam}
	<RecordPickerDialog
		kind="performer"
		title={cam.performer ? 'Change the performer' : 'Link a performer'}
		placeholder="Find a performer by name…"
		search={(text) => fetchPerformers(text, 24)}
		taken={cam.performer ? [cam.performer.id] : []}
		clearLabel={cam.performer ? 'Unlink the performer' : undefined}
		onclear={() => linkPerformer(null)}
		onpick={(performer) => linkPerformer(performer.id)}
		onclose={() => (pickingPerformer = false)}
	/>
{/if}
