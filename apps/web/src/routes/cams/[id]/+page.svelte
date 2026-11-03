<!--
	One cam room: its live stream, played as soon as the page opens, its
	picture lighting the app's shared glow; a header with the room, its audience,
	performer and site, Follow, Record (or the running recording and Stop)
	and a ⋮ menu (link a performer, block); its topic and tags; a recording
	card (start from back in the live buffer, quality, auto-record); when
	it's usually on; and similar rooms live now. Beside it (under it on
	narrow screens) her videos: the room's recordings and the library's
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
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import type { SheetAction } from '$lib/components/ActionSheet.svelte';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import Segmented from '$lib/components/Segmented.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import Toggle from '$lib/components/Toggle.svelte';
	import UsualHours from '$lib/components/UsualHours.svelte';
	import VideoListRow from '$lib/components/VideoListRow.svelte';
	import RecordPickerDialog from '$lib/components/RecordPickerDialog.svelte';
	import {
		CAMS_SURFACE,
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
		trackRoomWatch,
		uniqueVideos,
		watchCam,
		type ActiveRecordingRow,
		type CamRow,
		type CamShowRow
	} from '$lib/cams';
	import { fetchPerformers, type VideoCardRow } from '$lib/library';
	import type { SearchPage } from '$lib/search';
	import { siteDisplayName } from '$lib/plugins';
	import { block } from '$lib/recordActions';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { camUrl, camsUrl, performerUrl, videoUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';

	const ROOM_WIDTH = 220;
	// A running recording's length and size, and one just started or stopped.
	const RECORDING_REFRESH_MS = 5_000;
	// The rewind choices as the segmented control takes them (string values).
	const REWIND_CHOICES = RECORD_REWINDS.map((rewind) => ({
		value: String(rewind.seconds),
		label: rewind.label
	}));

	const megabytes = new Intl.NumberFormat(undefined, {
		style: 'unit',
		unit: 'megabyte',
		maximumFractionDigits: 0
	});
	const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact' });

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
	let rewind = $state('0');
	let pickingPerformer = $state(false);
	let menuAnchor = $state<HTMLElement | null>(null);
	// An action waiting for the backend; its buttons are disabled meanwhile.
	let busy = $state(false);
	let playerBox = $state<HTMLElement | null>(null);

	let id = $derived(page.params.id ?? '');
	let quality = $derived(cam?.recordQuality ?? 'source');
	let followed = $derived($followedCams.has(id));
	let topic = $derived(roomTopic(cam?.live?.subject));
	let picture = $derived(cacheUrl(cam?.live?.thumbnail?.uri ?? cam?.thumbnail?.uri));
	let audience = $derived(cam ? audienceLine(cam) : '');

	showAmbientGlow(() => picture);

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

	// Logs the room's watch while its stream is open, for the taste.
	$effect(() => {
		if (!stream) return;
		return trackRoomWatch(
			untrack(() => id),
			streamPlaying
		);
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

	/** "2.1K watching · on for 1 hour", or when an offline room was last seen. */
	function audienceLine(room: CamRow): string {
		if (!room.live) {
			if (room.lastSeenOnlineAt) return `Offline · seen ${dayjs(room.lastSeenOnlineAt).fromNow()}`;
			return 'Offline';
		}
		const audienceText = `${compactNumber.format(room.live.viewers)} watching`;
		if (!room.live.startedAt) return audienceText;
		return `${audienceText} · on for ${dayjs(room.live.startedAt).fromNow(true)}`;
	}

	/** Whether the room's stream is playing now, on screen. */
	function streamPlaying(): boolean {
		const video = playerBox?.querySelector('video');
		if (!video) return false;
		return (
			!video.paused && video.readyState >= HTMLMediaElement.HAVE_FUTURE_DATA && !document.hidden
		);
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
			await startCamRecording(id, quality, Number(rewind));
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
	function changeQuality(chosen: string) {
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

	/** The ⋮ menu: link or change the performer, block the room. */
	function menuActions(room: CamRow): SheetAction[] {
		return [
			{
				label: room.performer ? 'Change performer' : 'Link a performer',
				icon: UserIcon,
				run: () => (pickingPerformer = true)
			},
			{
				label: 'Block this room',
				icon: ProhibitIcon,
				run: () => {
					block('camModel', room.id, room.displayName);
					goto(camsUrl());
				},
				destructive: true
			}
		];
	}

	/** The cams page narrowed to one tag. */
	function tagUrl(tag: string): string {
		return `${camsUrl()}?${new URLSearchParams({ tag })}`;
	}
</script>

<svelte:head><title>{cam?.displayName ?? 'Cam'} · Veil</title></svelte:head>

<div
	class="mx-auto grid w-full max-w-[1700px] gap-6 pt-2 xl:grid-cols-[minmax(0,1fr)_380px]"
>

	<div class="flex min-w-0 flex-col gap-5">
		{#if stream && cam}
			<div
				class="overflow-hidden rounded-2xl shadow-2xl ring-1 shadow-black/50 ring-white/10"
				bind:this={playerBox}
			>
				<VideoPlayer
					src={stream.url}
					mimeType={stream.mimeType}
					title={cam.displayName}
					poster={picture}
					live
					inline
					square={$isCompact}
					onclose={() => (stream = null)}
				/>
			</div>
		{:else}
			<div
				class="bg-base-300 relative flex aspect-video w-full items-center justify-center overflow-hidden rounded-2xl ring-1 ring-white/10"
			>
				{#if picture}
					<img
						src={picture}
						alt=""
						class="absolute inset-0 h-full w-full object-cover opacity-40 grayscale"
					/>
				{/if}
				<span class="relative flex items-center gap-2 text-sm text-white/80">
					{#if loading}
						<span class="loading loading-spinner"></span>
					{:else if cam && !cam.live}
						<VideoCameraIcon size={18} />
						{cam.displayName} is offline{#if cam.lastSeenOnlineAt}, last seen {dayjs(
								cam.lastSeenOnlineAt
							).fromNow()}{/if}.
					{:else if !cam}
						This room isn't known.
					{/if}
				</span>
			</div>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		{#if cam}
			<header class="flex flex-wrap items-center justify-between gap-4">
				<div class="flex min-w-0 items-center gap-3.5">
					<span
						class="bg-base-300 block h-12 w-12 shrink-0 overflow-hidden rounded-full"
						class:live-ring={cam.live}
						class:grayscale={!cam.live}
					>
						{#if picture}
							<img src={picture} alt="" class="h-full w-full object-cover" />
						{/if}
					</span>
					<div class="flex min-w-0 flex-col">
						<h1 class="truncate text-2xl font-bold tracking-tight">{cam.displayName}</h1>
						<p class="text-base-content/60 flex flex-wrap items-center gap-x-1.5 text-sm">
							{#if cam.live}
								<span class="live-dot mr-0.5" style:width="6px" style:height="6px"></span>
							{/if}
							<span>{audience}</span>
							<span aria-hidden="true">·</span>
							<span>{siteDisplayName(cam.plugin, $sites)}</span>
							{#if cam.performer}
								<span aria-hidden="true">·</span>
								<a href={performerUrl(cam.performer.id)} class="text-base-content hover:underline">
									{cam.performer.name}
								</a>
							{/if}
						</p>
					</div>
				</div>

				<div class="flex shrink-0 items-center gap-2">
					<button
						type="button"
						class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-50"
						class:glass-pill-on={followed}
						disabled={busy}
						onclick={() => void act(() => toggleCamFollow(id))}
						title="Notifies you when the room goes live"
					>
						{#if followed}
							<BellRingingIcon size={16} weight="fill" /> Following
						{:else}
							<BellIcon size={16} /> Follow
						{/if}
					</button>
					{#if recording}
						<button
							type="button"
							class="bg-error flex h-10 items-center gap-2 rounded-full px-4 text-sm font-semibold text-black transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-50"
							disabled={busy}
							onclick={stop}
							title="Stops the recording; it becomes a video"
						>
							<StopIcon size={16} weight="fill" />
							<span class="tabular-nums"
								>REC {formatClock(recording.durationSeconds) ?? '0:00'}</span
							>
						</button>
					{:else}
						<button
							type="button"
							class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-50"
							disabled={busy || !cam.live}
							onclick={record}
						>
							<RecordIcon size={16} weight="fill" class="text-error" /> Record
						</button>
					{/if}
					<button
						type="button"
						class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
						aria-label="More"
						onclick={(event) => (menuAnchor = event.currentTarget)}
					>
						<DotsThreeVerticalIcon size={20} weight="bold" />
					</button>
				</div>
			</header>

			{#if topic || cam.tags.length > 0}
				<div class="flex flex-col gap-3">
					{#if topic}
						<p class="text-base-content/70 text-sm">{topic}</p>
					{/if}
					{#if cam.tags.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each cam.tags as tag (tag)}
								<a href={tagUrl(tag)} class="chip">
									<span class="opacity-50">#</span>{tag}
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{/if}

			<div class="grid gap-4 lg:grid-cols-2">
				<section
					class="bg-base-300/60 flex flex-col gap-4 rounded-2xl p-5 ring-1 ring-white/5 backdrop-blur"
				>
					<h2 class="flex items-center justify-between gap-2 text-base font-semibold">
						Recording
						{#if recording}
							<span class="text-error flex items-center gap-1.5 text-xs font-semibold tabular-nums">
								<RecordIcon size={12} weight="fill" />
								{formatClock(recording.durationSeconds) ?? '0:00'} · {megabytes.format(
									recording.sizeBytes / 1_000_000
								)}
							</span>
						{/if}
					</h2>
					{#if !recording}
						<div class="flex flex-wrap items-center justify-between gap-3">
							<span class="text-base-content/60 text-sm">Start from</span>
							<Segmented
								label="Start the recording from"
								choices={REWIND_CHOICES}
								current={rewind}
								onchoose={(value) => (rewind = value)}
							/>
						</div>
					{/if}
					<div class="flex flex-wrap items-center justify-between gap-3">
						<span class="text-base-content/60 text-sm">Quality</span>
						<Segmented
							label="Recording quality"
							choices={RECORD_QUALITIES}
							current={quality}
							onchoose={changeQuality}
						/>
					</div>
					<div class="flex items-center justify-between gap-3">
						<span class="flex flex-col">
							<span class="text-sm">Auto-record</span>
							<span class="text-base-content/50 text-xs">Records whenever the room goes live</span>
						</span>
						<Toggle
							checked={cam.autoRecord}
							disabled={busy}
							ariaLabel="Auto-record"
							onchange={changeAutoRecord}
						/>
					</div>
				</section>

				{#if shows.length > 0}
					<section
						class="bg-base-300/60 flex flex-col gap-4 rounded-2xl p-5 ring-1 ring-white/5 backdrop-blur"
					>
						<h2 class="flex items-baseline gap-2 text-base font-semibold">
							Usually on
							<span class="text-base-content/45 text-sm font-normal">
								from {shows.length}
								{shows.length === 1 ? 'show' : 'shows'} seen
							</span>
						</h2>
						<UsualHours {shows} />
					</section>
				{/if}
			</div>

			<div class="pt-2">
				<Shelf
					title="Similar rooms"
					subtitle="live now"
					items={similar}
					key={(room) => room.id}
					itemWidth={ROOM_WIDTH}
				>
					{#snippet card(room, position)}
						<div class="rise-in" style:--rise-delay="{Math.min(position * 30, 300)}ms">
							<CamCard
								cam={room}
								shown={{ surface: CAMS_SURFACE, position }}
								onclick={() => goto(camUrl(room.id))}
							/>
						</div>
					{/snippet}
				</Shelf>
			</div>
		{/if}
	</div>

	{#if cam}
		<!-- Beside the player it stays put and scrolls on its own: the page's
		     scroller is <main>, which runs under the floating top bar. -->
		<aside
			class="no-scrollbar xl:top-chrome flex flex-col gap-3 xl:sticky xl:max-h-[calc(100dvh-var(--chrome-offset)-1rem)] xl:self-start xl:overflow-y-auto xl:overscroll-y-contain xl:pt-2 xl:pb-4"
		>
			<div class="flex items-center justify-between gap-3">
				<h2 class="flex items-baseline gap-2 text-lg font-semibold">
					Her videos
					{#if videos.length > 0}
						<span class="text-base-content/45 text-sm font-normal">{videos.length}</span>
					{/if}
				</h2>
				<button
					type="button"
					class="chip"
					disabled={finding !== null}
					onclick={findShows}
					title="Asks every site of recorded cam shows for {cam.username}"
				>
					{#if finding}
						<span class="loading loading-spinner loading-xs"></span> Searching
					{:else}
						<MagnifyingGlassIcon size={14} /> Find her shows
					{/if}
				</button>
			</div>
			{#if finding && sitesAsked > 0}
				<p class="text-base-content/50 text-xs">
					Asking {sitesAsked}
					{sitesAsked === 1 ? 'site' : 'sites'} · {sitesAnswered} answered
				</p>
			{/if}
			{#if videos.length > 0}
				<div
					class="flex flex-col gap-1 max-xl:grid max-xl:grid-cols-[repeat(auto-fill,minmax(320px,1fr))]"
				>
					{#each videos as video, position (video.id)}
						<div class="rise-in" style:--rise-delay="{Math.min(position * 30, 300)}ms">
							<VideoListRow
								{video}
								shown={{ surface: 'cam', position }}
								onclick={() => goto(videoUrl(video.id))}
							/>
						</div>
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

{#if menuAnchor && cam}
	<CardMenu
		anchor={menuAnchor}
		title={cam.displayName}
		subtitle={audience}
		imageUrl={picture}
		actions={menuActions(cam)}
		onclose={() => (menuAnchor = null)}
	/>
{/if}

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
