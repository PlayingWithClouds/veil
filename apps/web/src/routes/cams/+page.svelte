<!--
	Cams: recordings running now, the user's rooms (followed or auto-recorded),
	live rooms picked by their taste and finished recordings, then every room
	live now on the cam sites, sorted by viewers or show start and narrowed by
	gender, site and tag. A top bar search (?q=) or a tag (?tag=) hides the
	rows and narrows the live rooms. Shown rooms refresh their picture and
	audience every minute, in place.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { CamSort } from '@playingwithclouds/veil-sdk/generated';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import RecordingCard from '$lib/components/RecordingCard.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		CAM_GENDERS,
		CAM_SORTS,
		camSites,
		fetchActiveRecordings,
		fetchCamRooms,
		fetchRecommendedCams,
		fetchRecordedVideos,
		fetchYourRooms,
		popularRoomTags,
		type ActiveRecordingRow,
		type CamFilter,
		type CamRow
	} from '$lib/cams';
	import type { VideoCardRow } from '$lib/library';
	import { sites } from '$lib/stores/sites';
	import { camUrl, camsUrl, videoUrl } from '$lib/routes';

	const PAGE_SIZE = 48;
	const CARD_WIDTH = 220;
	const VIDEO_WIDTH = 280;
	const POPULAR_TAG_COUNT = 16;
	// Running recordings change length and end; the row follows them.
	const RECORDING_REFRESH_MS = 10_000;
	// The sites' pictures and audiences change; shown rooms follow them.
	const ROOM_REFRESH_MS = 60_000;
	// Rooms refreshed at most, from the top: one listing answers them.
	const ROOM_REFRESH_LIMIT = 192;

	let recordings = $state<ActiveRecordingRow[]>([]);
	let yourRooms = $state<CamRow[]>([]);
	let forYou = $state<CamRow[]>([]);
	let recordedVideos = $state<VideoCardRow[]>([]);

	let gender = $state('');
	let sort = $state<CamSort>('viewers');
	let selectedSites = $state<string[]>([]);
	let rooms = $state<CamRow[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	// The tags of the unnarrowed rooms, so picking one keeps the chip row in place.
	let popularTags = $state<string[]>([]);
	// Bumped on every filter change so a slow answer for old filters is dropped.
	let generation = 0;

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let tag = $derived(page.url.searchParams.get('tag') ?? '');
	let showOverview = $derived(query === '' && tag === '');
	let siteChoices = $derived(camSites($sites));
	let tagChoices = $derived(tagChipsWith(popularTags, tag));
	let recordingIds = $derived(recordings.map((recording) => recording.camModel?.id));

	$effect(() => {
		load(fetchYourRooms(), (found) => (yourRooms = found));
		load(fetchRecommendedCams(), (found) => (forYou = found));
		load(fetchRecordedVideos(), (found) => (recordedVideos = found));
		refreshRecordings();
		const recordingTimer = setInterval(refreshRecordings, RECORDING_REFRESH_MS);
		const roomTimer = setInterval(refreshShownRooms, ROOM_REFRESH_MS);
		return () => {
			clearInterval(recordingTimer);
			clearInterval(roomTimer);
		};
	});

	// Starts the live rooms over whenever a filter changes.
	$effect(() => {
		void [gender, sort, selectedSites, tag, query];
		untrack(restart);
	});

	/** Fills a row when its data arrives; a row that fails stays hidden and says why in the console. */
	function load<T>(request: Promise<T>, fill: (found: T) => void) {
		request.then(fill, (failure) => console.warn('cams row failed to load', failure));
	}

	/** Reloads the recordings running now. */
	function refreshRecordings() {
		load(fetchActiveRecordings(), (found) => (recordings = found));
	}

	/** The live rooms' filter as the controls and the URL set it. */
	function currentFilter(): CamFilter {
		return { plugins: selectedSites, gender, tag, text: query, sort };
	}

	/** The popular tags, with the active one first when it isn't among them. */
	function tagChipsWith(popular: string[], active: string): string[] {
		if (active === '' || popular.includes(active)) return popular;
		return [active, ...popular];
	}

	/** Clears the live rooms and loads their first page for the current filters. */
	function restart() {
		generation += 1;
		rooms = [];
		total = 0;
		exhausted = false;
		loading = false;
		error = null;
		void loadMore();
	}

	/** The rooms of a page not shown yet, each once (a site can list a room twice). */
	function unseenRooms(pageRooms: CamRow[]): CamRow[] {
		const known = new Set(rooms.map((room) => room.id));
		const unseen: CamRow[] = [];
		for (const room of pageRooms) {
			if (known.has(room.id)) continue;
			known.add(room.id);
			unseen.push(room);
		}
		return unseen;
	}

	/** Appends the next page of live rooms. */
	async function loadMore() {
		if (loading || exhausted) return;
		loading = true;
		const requested = generation;
		try {
			const found = await fetchCamRooms(currentFilter(), PAGE_SIZE, rooms.length);
			if (requested !== generation) return;
			const before = rooms.length;
			rooms = [...rooms, ...unseenRooms(found.rooms)];
			total = found.total;
			exhausted = found.rooms.length < PAGE_SIZE || rooms.length === before;
			if (tag === '') popularTags = popularRoomTags(rooms, POPULAR_TAG_COUNT);
		} catch (failure) {
			if (requested !== generation) return;
			error = String(failure);
			exhausted = true;
		} finally {
			if (requested === generation) loading = false;
		}
	}

	/**
	 * Gives the shown rooms their current picture and audience without moving
	 * them, so the grid doesn't reorder under the user.
	 */
	async function refreshShownRooms() {
		if (document.hidden || rooms.length === 0 || loading) return;
		const requested = generation;
		const limit = Math.min(rooms.length, ROOM_REFRESH_LIMIT);
		const found = await fetchCamRooms(currentFilter(), limit, 0);
		if (requested !== generation) return;
		const fresh = new Map(found.rooms.map((room) => [room.id, room]));
		rooms = rooms.map((room) => fresh.get(room.id) ?? room);
		total = found.total;
	}

	/** Adds or removes one site from the filter. */
	function toggleSite(name: string) {
		if (selectedSites.includes(name)) {
			selectedSites = selectedSites.filter((entry) => entry !== name);
			return;
		}
		selectedSites = [...selectedSites, name];
	}

	/** Narrows to a tag, or clears it when it's the active one; the search stays. */
	function toggleTag(name: string) {
		const params = new URLSearchParams(page.url.searchParams);
		if (tag === name) {
			params.delete('tag');
		} else {
			params.set('tag', name);
		}
		const search = params.toString();
		if (search === '') {
			goto(camsUrl(), { keepFocus: true, noScroll: true });
			return;
		}
		goto(`${camsUrl()}?${search}`, { keepFocus: true, noScroll: true });
	}

	/** Opens the room a recording is made from. */
	function openRecording(recording: ActiveRecordingRow) {
		if (!recording.camModel) return;
		goto(camUrl(recording.camModel.id));
	}
</script>

<svelte:head><title>Cams · Veil</title></svelte:head>

{#snippet roomCard(room: CamRow, markLive: boolean)}
	<CamCard
		cam={room}
		{markLive}
		recording={recordingIds.includes(room.id)}
		onrecordingchange={refreshRecordings}
		onclick={() => goto(camUrl(room.id))}
	/>
{/snippet}

{#snippet segmented<Value extends string>(
	label: string,
	choices: { value: Value; label: string }[],
	current: Value,
	choose: (value: Value) => void
)}
	<div class="bg-base-200 flex gap-1 rounded-full p-1" role="group" aria-label={label}>
		{#each choices as choice (choice.value)}
			<button
				type="button"
				class="rounded-full px-3 py-1 text-sm font-medium whitespace-nowrap transition-colors"
				class:bg-base-100={current === choice.value}
				class:shadow-sm={current === choice.value}
				class:opacity-60={current !== choice.value}
				aria-pressed={current === choice.value}
				onclick={() => choose(choice.value)}
			>
				{choice.label}
			</button>
		{/each}
	</div>
{/snippet}

<div class="flex flex-col gap-8">
	<PageHeader
		title="Cams"
		description={total > 0 ? `${total.toLocaleString()} rooms live` : null}
	/>

	{#if !showOverview}
		<div class="-mt-4 flex items-center gap-3">
			<span class="text-base-content/70 text-sm">
				{#if query !== ''}
					Rooms matching <span class="text-base-content font-medium">“{query}”</span>
				{:else}
					Rooms tagged <span class="text-base-content font-medium">{tag}</span>
				{/if}
			</span>
			<button
				type="button"
				class="bg-base-200 hover:bg-base-300 flex h-7 items-center gap-1 rounded-full px-3 text-xs"
				onclick={() => goto(camsUrl())}
			>
				<XIcon size={12} /> Clear
			</button>
		</div>
	{/if}

	{#if showOverview}
		<Shelf
			title="Recording now"
			items={recordings}
			key={(recording) => recording.camModel?.id ?? recording.startedAt}
			itemWidth={CARD_WIDTH}
		>
			{#snippet card(recording)}
				<RecordingCard {recording} onclick={() => openRecording(recording)} />
			{/snippet}
		</Shelf>

		<Shelf title="Your rooms" items={yourRooms} key={(room) => room.id} itemWidth={CARD_WIDTH}>
			{#snippet card(room)}
				{@render roomCard(room, true)}
			{/snippet}
		</Shelf>

		<Shelf title="For you" items={forYou} key={(room) => room.id} itemWidth={CARD_WIDTH}>
			{#snippet card(room)}
				{@render roomCard(room, false)}
			{/snippet}
		</Shelf>

		<Shelf
			title="Recordings"
			items={recordedVideos}
			key={(video) => video.id}
			itemWidth={VIDEO_WIDTH}
		>
			{#snippet card(video, position)}
				<SceneCard
					{video}
					shown={{ surface: 'cams', position }}
					onclick={() => goto(videoUrl(video.id))}
				/>
			{/snippet}
		</Shelf>

		{#if yourRooms.length === 0 && recordings.length === 0}
			<p class="text-base-content/50 -mt-4 text-sm">
				Follow a room from its ⋮ menu to hear when it goes live, or auto-record it to keep its
				shows.
			</p>
		{/if}
	{/if}

	<section class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center gap-3">
			<h2 class="text-lg font-semibold">Live now</h2>
			{@render segmented('Gender', CAM_GENDERS, gender, (value) => (gender = value))}
			{@render segmented('Sort', CAM_SORTS, sort, (value) => (sort = value))}
			{#if siteChoices.length > 1}
				{#each siteChoices as site (site.name)}
					<button
						type="button"
						class="rounded-full px-3 py-1 text-sm whitespace-nowrap hover:brightness-125"
						class:bg-base-200={!selectedSites.includes(site.name)}
						class:bg-primary={selectedSites.includes(site.name)}
						class:text-primary-content={selectedSites.includes(site.name)}
						aria-pressed={selectedSites.includes(site.name)}
						onclick={() => toggleSite(site.name)}
					>
						{site.displayName ?? site.name}
					</button>
				{/each}
			{/if}
		</div>

		{#if tagChoices.length > 0}
			<div
				class="no-scrollbar fade-right-edge -mx-1 flex items-center gap-2 overflow-x-auto px-1 py-0.5"
			>
				{#each tagChoices as name (name)}
					<button
						type="button"
						class="shrink-0 rounded-full px-3 py-1 text-sm whitespace-nowrap hover:brightness-125"
						class:bg-base-200={tag !== name}
						class:bg-primary={tag === name}
						class:text-primary-content={tag === name}
						aria-pressed={tag === name}
						onclick={() => toggleTag(name)}
					>
						{name}
					</button>
				{/each}
			</div>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		<div
			class="grid gap-x-4 gap-y-6"
			style:grid-template-columns="repeat(auto-fill, minmax({CARD_WIDTH}px, 1fr))"
		>
			{#each rooms as room (room.id)}
				{@render roomCard(room, false)}
			{/each}
		</div>

		{#if !loading && !error && rooms.length === 0}
			{#if siteChoices.length === 0}
				<p class="text-base-content/50 text-sm">
					No installed plugin lists cam rooms. Install one on the plugins page.
				</p>
			{:else}
				<p class="text-base-content/50 text-sm">No room live matches this.</p>
			{/if}
		{/if}
	</section>

	<ScrollSentinel onvisible={loadMore} loading={loading && rooms.length > 0} done={exhausted} />
</div>
