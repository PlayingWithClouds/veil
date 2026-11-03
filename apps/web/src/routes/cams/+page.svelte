<!--
	Cams: recordings running now, the user's rooms (followed or auto-recorded)
	and finished recordings, then every room live now on the cam sites, most
	viewers first, narrowed by gender, site and tag. A top bar search (?q=)
	or a tag (?tag=) hides the rows and narrows the live rooms.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import RecordingCard from '$lib/components/RecordingCard.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		CAM_GENDERS,
		camSites,
		fetchActiveRecordings,
		fetchCamRooms,
		fetchRecordedVideos,
		fetchYourRooms,
		popularRoomTags,
		type ActiveRecordingRow,
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

	let recordings = $state<ActiveRecordingRow[]>([]);
	let yourRooms = $state<CamRow[]>([]);
	let recordedVideos = $state<VideoCardRow[]>([]);

	let gender = $state('');
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

	$effect(() => {
		load(fetchYourRooms(), (found) => (yourRooms = found));
		load(fetchRecordedVideos(), (found) => (recordedVideos = found));
		refreshRecordings();
		const timer = setInterval(refreshRecordings, RECORDING_REFRESH_MS);
		return () => clearInterval(timer);
	});

	// Starts the live rooms over whenever a filter changes.
	$effect(() => {
		void [gender, selectedSites, tag, query];
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
			const found = await fetchCamRooms(
				{ plugins: selectedSites, gender, tag, text: query },
				PAGE_SIZE,
				rooms.length
			);
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
				<CamCard cam={room} onclick={() => goto(camUrl(room.id))} />
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
	{/if}

	<section class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center gap-3">
			<h2 class="text-lg font-semibold">Live now</h2>
			<div class="bg-base-200 flex gap-1 rounded-full p-1" role="group" aria-label="Gender">
				{#each CAM_GENDERS as choice (choice.value)}
					<button
						type="button"
						class="rounded-full px-3 py-1 text-sm font-medium transition-colors"
						class:bg-base-100={gender === choice.value}
						class:shadow-sm={gender === choice.value}
						class:opacity-60={gender !== choice.value}
						aria-pressed={gender === choice.value}
						onclick={() => (gender = choice.value)}
					>
						{choice.label}
					</button>
				{/each}
			</div>
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
				<CamCard cam={room} onclick={() => goto(camUrl(room.id))} />
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
