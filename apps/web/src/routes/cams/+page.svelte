<!--
	Cams: the top recommended rooms in a spotlight (one plays, the rest are
	listed beside it), the user's rooms (being recorded, followed or
	auto-recorded) as round pictures, more live rooms picked by their taste
	and finished recordings, then every room
	live now on the cam sites, sorted by viewers or show start and narrowed by
	gender, site and tag from a bar that sticks while the grid scrolls. A top
	bar search (?q=) or a tag (?tag=) hides the rows and narrows the live
	rooms. Shown rooms refresh their picture and audience every minute, in
	place; a filter change fades the old rooms until the new ones arrive.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { CamSort } from '@playingwithclouds/veil-sdk/generated';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import Shelf from '$lib/components/Shelf.svelte';
	import CamAvatar from '$lib/components/CamAvatar.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import CamSpotlight from '$lib/components/CamSpotlight.svelte';
	import Segmented from '$lib/components/Segmented.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		CAM_GENDERS,
		CAM_SORTS,
		CAMS_SURFACE,
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
	// The live grid's columns at most, however wide the window; its gap (gap-x-4).
	const MAX_COLUMNS = 5;
	const COLUMN_GAP = 16;
	// Cards at least CARD_WIDTH wide, and never narrower than a fifth of the row.
	const GRID_COLUMNS = `repeat(auto-fill, minmax(max(${CARD_WIDTH}px, calc((100% - ${(MAX_COLUMNS - 1) * COLUMN_GAP}px) / ${MAX_COLUMNS})), 1fr))`;
	const AVATAR_WIDTH = 96;
	// Picks in the spotlight; the rest of the recommendations follow in a row.
	const SPOTLIGHT_COUNT = 5;
	// Tags shown instead of a reason an earlier pick already gave.
	const REASON_TAG_COUNT = 3;
	const POPULAR_TAG_COUNT = 16;
	// Placeholder cards while the first rooms load.
	const SKELETON_COUNT = 12;
	// Cards of one page rise in this far apart, up to RISE_MAX_DELAY_MS.
	const RISE_STEP_MS = 30;
	const RISE_MAX_DELAY_MS = 450;
	// Running recordings change length and end; the row follows them.
	const RECORDING_REFRESH_MS = 10_000;
	// The sites' pictures and audiences change; shown rooms follow them.
	const ROOM_REFRESH_MS = 60_000;
	// Rooms refreshed at most, from the top: one listing answers them.
	const ROOM_REFRESH_LIMIT = 192;

	let recordings = $state<ActiveRecordingRow[]>([]);
	let yourRooms = $state<CamRow[]>([]);
	let forYou = $state<CamRow[]>([]);
	let forYouLoaded = $state(false);
	let recordedVideos = $state<VideoCardRow[]>([]);

	let gender = $state('');
	let sort = $state<CamSort>('viewers');
	let selectedSites = $state<string[]>([]);
	let rooms = $state<CamRow[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let exhausted = $state(false);
	// The shown rooms are for old filters: they fade until the new ones replace them.
	let stale = $state(false);
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
	let spotlight = $derived(forYou.slice(0, SPOTLIGHT_COUNT));
	let moreForYou = $derived(forYou.slice(SPOTLIGHT_COUNT));
	let shownReasons = $derived(distinctReasons(forYou));
	let roomStrip = $derived(stripOf(recordings, yourRooms));
	let liveCount = $derived(yourRooms.filter((room) => room.live).length);

	$effect(() => {
		load(fetchYourRooms(), (found) => (yourRooms = found));
		const recommended = fetchRecommendedCams();
		load(recommended, (found) => (forYou = found));
		recommended.finally(() => (forYouLoaded = true)).catch(() => {});
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

	/**
	 * The user's rooms as the strip shows them: rooms being recorded first
	 * (with how long), then the live ones, then the offline ones.
	 */
	function stripOf(
		running: ActiveRecordingRow[],
		own: CamRow[]
	): { room: CamRow; recordingSeconds: number | null }[] {
		const strip: { room: CamRow; recordingSeconds: number | null }[] = [];
		for (const recording of running) {
			if (!recording.camModel) continue;
			strip.push({ room: recording.camModel, recordingSeconds: recording.durationSeconds ?? 0 });
		}
		const recorded = strip.map((entry) => entry.room.id);
		const rest = own.filter((room) => !recorded.includes(room.id));
		for (const room of rest.filter((candidate) => candidate.live)) {
			strip.push({ room, recordingSeconds: null });
		}
		for (const room of rest.filter((candidate) => !candidate.live)) {
			strip.push({ room, recordingSeconds: null });
		}
		return strip;
	}

	/**
	 * Each pick's caption line by room id: its reason the first time a reason
	 * appears, its tags when an earlier pick already said the same.
	 */
	function distinctReasons(picks: CamRow[]): Record<string, string> {
		const lines: Record<string, string> = {};
		const given: string[] = [];
		for (const pick of picks) {
			const reason = pick.reason ?? '';
			if (reason !== '' && !given.includes(reason)) {
				given.push(reason);
				lines[pick.id] = reason;
				continue;
			}
			const tags = pick.tags.slice(0, REASON_TAG_COUNT).map((name) => `#${name}`);
			lines[pick.id] = tags.join(' ');
		}
		return lines;
	}

	/** "3 live · 1 recording", or null when neither. */
	function stripSubtitle(live: number, recordingCount: number): string | null {
		const parts: string[] = [];
		if (live > 0) parts.push(`${live} live`);
		if (recordingCount > 0) parts.push(`${recordingCount} recording`);
		if (parts.length === 0) return null;
		return parts.join(' · ');
	}

	/** How long a card waits before rising in, so a page of cards cascades. */
	function riseDelay(position: number): string {
		const step = (position % PAGE_SIZE) * RISE_STEP_MS;
		return `${Math.min(step, RISE_MAX_DELAY_MS)}ms`;
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

	/** Loads the first page of live rooms for the current filters; the old ones fade until it arrives. */
	function restart() {
		generation += 1;
		stale = rooms.length > 0;
		exhausted = false;
		loading = false;
		error = null;
		void loadMore();
	}

	/** The rooms of a page not in shown yet, each once (a site can list a room twice). */
	function unseenRooms(shown: CamRow[], pageRooms: CamRow[]): CamRow[] {
		const known = new Set(shown.map((room) => room.id));
		const unseen: CamRow[] = [];
		for (const room of pageRooms) {
			if (known.has(room.id)) continue;
			known.add(room.id);
			unseen.push(room);
		}
		return unseen;
	}

	/** Appends the next page of live rooms, or replaces stale ones with the first. */
	async function loadMore() {
		if (loading || exhausted) return;
		loading = true;
		const requested = generation;
		let shown = rooms;
		if (stale) {
			shown = [];
		}
		try {
			const found = await fetchCamRooms(currentFilter(), PAGE_SIZE, shown.length);
			if (requested !== generation) return;
			rooms = [...shown, ...unseenRooms(shown, found.rooms)];
			stale = false;
			total = found.total;
			exhausted = found.rooms.length < PAGE_SIZE || rooms.length === shown.length;
			if (tag === '') popularTags = popularRoomTags(rooms, POPULAR_TAG_COUNT);
		} catch (failure) {
			if (requested !== generation) return;
			rooms = shown;
			stale = false;
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
</script>

<svelte:head><title>Cams · Veil</title></svelte:head>

{#snippet roomCard(room: CamRow, markLive: boolean, position: number, reason: string | null)}
	<CamCard
		cam={room}
		{markLive}
		{reason}
		shown={{ surface: CAMS_SURFACE, position }}
		recording={recordingIds.includes(room.id)}
		onrecordingchange={refreshRecordings}
		onclick={() => goto(camUrl(room.id))}
	/>
{/snippet}

{#snippet skeletonCard()}
	<div class="flex flex-col gap-2" aria-hidden="true">
		<div class="skeleton aspect-video w-full"></div>
		<div class="skeleton h-3.5 w-2/3 rounded"></div>
		<div class="skeleton h-3 w-1/3 rounded"></div>
	</div>
{/snippet}

<div class="isolate flex flex-col gap-7">
	<div class="flex items-end justify-between gap-4 pt-2">
		<div class="min-w-0">
			<h1 class="text-3xl font-bold tracking-tight">Cams</h1>
			{#if total > 0}
				<p class="text-base-content/60 mt-1 flex items-center gap-2 text-sm">
					<span class="live-dot"></span>
					<span class="tabular-nums">{total.toLocaleString()}</span> rooms live now
				</p>
			{/if}
		</div>
	</div>

	{#if !showOverview}
		<div class="-mt-4 flex items-center gap-3">
			<span class="text-base-content/70 text-sm">
				{#if query !== ''}
					Rooms matching <span class="text-base-content font-medium">“{query}”</span>
				{:else}
					Rooms tagged <span class="text-base-content font-medium">#{tag}</span>
				{/if}
			</span>
			<button type="button" class="chip" onclick={() => goto(camsUrl())}>
				<XIcon size={12} /> Clear
			</button>
		</div>
	{/if}

	{#if showOverview}
		{#if spotlight.length > 0}
			<div class="rise-in">
				<CamSpotlight
					rooms={spotlight}
					surface={CAMS_SURFACE}
					onopen={(room) => goto(camUrl(room.id))}
				/>
			</div>
		{:else if !forYouLoaded}
			<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)]" aria-hidden="true">
				<div class="skeleton aspect-video rounded-2xl"></div>
				<div class="flex flex-col gap-2 max-lg:hidden">
					{#each { length: SPOTLIGHT_COUNT }, position (position)}
						<div class="skeleton flex-1 rounded-xl"></div>
					{/each}
				</div>
			</div>
		{/if}

		<Shelf
			title="Your rooms"
			subtitle={stripSubtitle(liveCount, recordings.length)}
			items={roomStrip}
			key={(entry) => entry.room.id}
			itemWidth={AVATAR_WIDTH}
		>
			{#snippet card(entry, position)}
				<div class="rise-in pt-2" style:--rise-delay={riseDelay(position)}>
					<CamAvatar
						cam={entry.room}
						recordingSeconds={entry.recordingSeconds}
						onclick={() => goto(camUrl(entry.room.id))}
					/>
				</div>
			{/snippet}
		</Shelf>

		<Shelf
			title="More for you"
			subtitle="from what you watch"
			items={moreForYou}
			key={(room) => room.id}
			itemWidth={CARD_WIDTH}
		>
			{#snippet card(room, position)}
				<div class="rise-in" style:--rise-delay={riseDelay(position)}>
					{@render roomCard(room, false, SPOTLIGHT_COUNT + position, shownReasons[room.id] ?? null)}
				</div>
			{/snippet}
		</Shelf>

		<Shelf
			title="Recordings"
			subtitle="finished shows"
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
		<div
			class="top-chrome bg-base-100/75 sticky z-20 -mx-4 flex flex-col gap-3 border-b border-white/5 px-4 py-3 backdrop-blur-xl max-md:-mx-3 max-md:px-3"
		>
			<div class="flex flex-wrap items-center gap-3">
				<h2 class="flex items-center gap-2 text-lg font-semibold">
					<span class="live-dot"></span> Live now
				</h2>
				<Segmented
					label="Gender"
					choices={CAM_GENDERS}
					current={gender}
					onchoose={(value) => (gender = value)}
				/>
				<Segmented
					label="Sort"
					choices={CAM_SORTS}
					current={sort}
					onchoose={(value) => (sort = value)}
				/>
				{#if siteChoices.length > 1}
					{#each siteChoices as site (site.name)}
						<button
							type="button"
							class="chip"
							class:chip-on={selectedSites.includes(site.name)}
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
							class="chip"
							class:chip-on={tag === name}
							aria-pressed={tag === name}
							onclick={() => toggleTag(name)}
						>
							<span class="opacity-50">#</span>{name}
							{#if tag === name}
								<XIcon size={12} />
							{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		<div
			class="grid gap-x-4 gap-y-6 transition-opacity duration-300"
			class:opacity-35={stale}
			class:pointer-events-none={stale}
			style:grid-template-columns={GRID_COLUMNS}
		>
			{#each rooms as room, position (room.id)}
				<div class="rise-in" style:--rise-delay={riseDelay(position)}>
					{@render roomCard(room, false, position, null)}
				</div>
			{:else}
				{#if loading}
					{#each { length: SKELETON_COUNT }, position (position)}
						{@render skeletonCard()}
					{/each}
				{/if}
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
