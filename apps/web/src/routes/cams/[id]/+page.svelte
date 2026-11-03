<!--
	One cam room: its live stream, played as soon as the page opens; its
	performer (linked from the library), topic and tags; follow, recording
	(now or from back in the live buffer) and auto-record at the room's
	quality; when it's usually on; and similar rooms live now. Beside it (under
	it on narrow screens) her videos: the room's recordings and the library's
	videos of her, which "Find her shows" adds to by asking every site of
	recorded cam shows for her username.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
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
		fetchCamShows,
		fetchRoomRecording,
		fetchRoomVideos,
		fetchSimilarRooms,
		findRoomShows,
		followedCams,
		linkCamPerformer,
		loadFollowedCams,
		roomTopic,
		setCamAutoRecord,
		startCamRecording,
		stopCamRecording,
		toggleCamFollow,
		uniqueVideos,
		watchCam,
		type ActiveRecordingRow,
		type CamRow,
		type CamShowRow
	} from '$lib/cams';
	import { fetchPerformers, type VideoCardRow } from '$lib/library';
	import type { SearchPage } from '$lib/search';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { camUrl, camsUrl, performerUrl, videoUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';

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
	let videos = $state<VideoCardRow[]>([]);
	// The running search of the cam show sites, and how far it got.
	// Raw, so it stays the very object findShows compares it with.
	let finding = $state.raw<SearchPage | null>(null);
	let sitesAsked = $state(0);
	let sitesAnswered = $state(0);
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

	// Reloads for a new room only: load reads state (the running search) it
	// must not rerun on.
	$effect(() => {
		const camId = id;
		untrack(() => void load(camId));
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

	// A search still running when the page leaves the room stops.
	$effect(() => {
		return () => finding?.cancel();
	});

	/** Loads the room and its extras, and while it's live, its stream. */
	async function load(camId: string) {
		finding?.cancel();
		finding = null;
		cam = null;
		stream = null;
		error = null;
		recording = null;
		videos = [];
		shows = [];
		similar = [];
		loading = true;
		void loadExtras(camId);
		try {
			cam = await fetchCam(camId);
			if (cam) {
				void loadVideos(cam);
				void loadSimilar(cam);
			}
			if (cam?.live) stream = await watchCam(camId);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Loads what the controls and the schedule show. */
	async function loadExtras(camId: string) {
		try {
			const [running, pastShows] = await Promise.all([
				fetchRoomRecording(camId),
				fetchCamShows(camId),
				loadFollowedCams()
			]);
			if (camId !== id) return;
			recording = running;
			shows = pastShows;
		} catch (failure) {
			console.warn('cam room details failed to load', failure);
		}
	}

	/** Loads her videos in the library, keeping what a running search found. */
	async function loadVideos(room: CamRow) {
		try {
			const found = await fetchRoomVideos(room);
			if (room.id === id) videos = uniqueVideos([...found, ...videos]);
		} catch (failure) {
			console.warn('room videos failed to load', failure);
		}
	}

	/** Asks every site of recorded cam shows for her; what they find joins her videos. */
	function findShows() {
		if (!cam) return;
		const room = cam;
		finding?.cancel();
		sitesAsked = 0;
		sitesAnswered = 0;
		const search = findRoomShows(
			room,
			(found) => {
				if (room.id === id) videos = uniqueVideos([...videos, ...found]);
			},
			(asked, answered) => {
				sitesAsked = asked;
				sitesAnswered = answered;
			}
		);
		finding = search;
		search.done
			.catch((failure) => (error = String(failure)))
			.finally(() => {
				if (finding === search) finding = null;
			});
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

	/** Rereads the running recording; when one ends, its video joins her videos. */
	async function refreshRecording(camId: string) {
		const running = await fetchRoomRecording(camId);
		if (camId !== id) return;
		const ended = recording !== null && running === null;
		recording = running;
		if (ended && cam) await loadVideos(cam);
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

<div class="mx-auto grid w-full max-w-[1700px] gap-6 pt-2 xl:grid-cols-[minmax(0,1fr)_380px]">
	<div class="flex min-w-0 flex-col gap-4">
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

			<div class="pt-4">
				<Shelf title="Similar rooms" items={similar} key={(room) => room.id} itemWidth={ROOM_WIDTH}>
					{#snippet card(room)}
						<CamCard cam={room} onclick={() => goto(camUrl(room.id))} />
					{/snippet}
				</Shelf>
			</div>
		{/if}
	</div>

	{#if cam}
		<aside class="flex flex-col gap-4">
			<div class="flex items-center justify-between gap-3">
				<h2 class="text-lg font-semibold">Her videos</h2>
				<button
					type="button"
					class="btn btn-sm"
					disabled={finding !== null}
					onclick={findShows}
					title="Asks every site of recorded cam shows for {cam.username}"
				>
					<MagnifyingGlassIcon size={16} /> Find her shows
				</button>
			</div>
			{#if finding && sitesAsked > 0}
				<p class="text-base-content/60 text-sm">
					Asking {sitesAsked}
					{sitesAsked === 1 ? 'site' : 'sites'} · {sitesAnswered} answered
				</p>
			{/if}
			{#if videos.length > 0}
				<div
					class="grid grid-cols-1 gap-x-4 gap-y-6 max-xl:grid-cols-[repeat(auto-fill,minmax(260px,1fr))]"
				>
					{#each videos as video, position (video.id)}
						<SceneCard
							{video}
							shown={{ surface: 'cam', position }}
							onclick={() => goto(videoUrl(video.id))}
						/>
					{/each}
				</div>
			{:else if !finding}
				<p class="text-base-content/50 text-sm">
					No videos of {cam.displayName} yet. Record her, or find her shows on the cam show sites.
				</p>
			{/if}
		</aside>
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
