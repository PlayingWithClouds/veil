<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import { activeCamRecordings } from '$lib/stores/camRecordings';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import CamCard from '$lib/components/CamCard.svelte';
	import RecordingNowCard from '$lib/components/cams/RecordingNowCard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		fetchCamRooms,
		fetchCamSources,
		fetchFollowedCamModels,
		popularRoomTags,
		type CamModel,
		type CamSource
	} from '$lib/cams';
	import { camUrl } from '$lib/routes';
	import { pluginIconUrl } from '$lib/server';

	const PAGE_SIZE = 48;
	const SEARCH_DEBOUNCE_MS = 300;
	const POPULAR_TAG_COUNT = 16;

	const GENDERS = [
		{ value: '', label: 'All' },
		{ value: 'female', label: 'Female' },
		{ value: 'male', label: 'Male' },
		{ value: 'couple', label: 'Couples' },
		{ value: 'trans', label: 'Trans' }
	];

	let sources = $state<CamSource[]>([]);
	let followed = $state<CamModel[]>([]);

	let selectedSources = $state<string[]>([]);
	let gender = $state('');
	// A tag chip on a room page links here with ?tag=.
	let tag = $state(page.url.searchParams.get('tag') || '');
	let query = $state('');
	let typedQuery = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	let rooms = $state<CamModel[]>([]);
	let totalCount = $state(0);
	let loading = $state(false);
	let done = $state(false);
	let failed = $state(false);
	// Bumped on every filter change so a slow response for old filters is dropped.
	let generation = 0;

	let sourcesByName = $derived(new Map(sources.map((source) => [source.name, source])));

	// The tags most rooms carry right now, as one-click filters. Taken from the
	// rooms listed without a tag filter, so picking one keeps the row in place.
	let popularTags = $state<string[]>([]);
	$effect(() => {
		if (tag || rooms.length === 0) return;
		popularTags = popularRoomTags(rooms, POPULAR_TAG_COUNT);
	});
	let tagOutsidePopular = $derived(tag !== '' && !popularTags.includes(tag));

	/** Filters by a tag chip, or clears the filter when it is the active one. */
	function toggleTag(name: string) {
		if (tag === name) {
			tag = '';
			return;
		}
		tag = name;
	}

	$effect(() => {
		fetchCamSources().then((loaded) => {
			sources = loaded.map((source) => ({ ...source, iconUrl: pluginIconUrl(source.iconUrl) }));
		});
		fetchFollowedCamModels().then((loaded) => {
			followed = loaded;
		});
	});

	// Reload the grid from the top whenever a filter changes.
	$effect(() => {
		void [selectedSources, gender, tag, query];
		untrack(reload);
	});

	/** Clears the grid and loads its first page for the current filters. */
	function reload() {
		generation += 1;
		rooms = [];
		totalCount = 0;
		done = false;
		failed = false;
		loading = false;
		void loadMore();
	}

	/** The page's rooms not shown yet, each once (a site can list a room twice). */
	function unseenRooms(pageRooms: CamModel[]): CamModel[] {
		const known = new Set(rooms.map((room) => room.id));
		const unseen: CamModel[] = [];
		for (const room of pageRooms) {
			if (known.has(room.id)) continue;
			known.add(room.id);
			unseen.push(room);
		}
		return unseen;
	}

	/** Appends the next page of rooms. */
	async function loadMore() {
		if (loading || done) return;
		loading = true;
		const requested = generation;
		try {
			const page = await fetchCamRooms(
				{ sources: selectedSources, tag, query, gender },
				PAGE_SIZE,
				rooms.length
			);
			if (requested !== generation) return;
			const before = rooms.length;
			rooms = [...rooms, ...unseenRooms(page.rooms)];
			totalCount = page.totalCount;
			done = page.rooms.length < PAGE_SIZE || rooms.length === before;
		} catch {
			if (requested !== generation) return;
			failed = true;
			done = true;
		} finally {
			if (requested === generation) loading = false;
		}
	}

	/** Adds or removes one site from the filter. */
	function toggleSource(name: string) {
		if (selectedSources.includes(name)) {
			selectedSources = selectedSources.filter((entry) => entry !== name);
			return;
		}
		selectedSources = [...selectedSources, name];
	}

	/** Applies the typed search after a short pause. */
	function onSearchInput() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => (query = typedQuery.trim()), SEARCH_DEBOUNCE_MS);
	}

	/** Filters the grid to one tag and scrolls to the top. */
	function filterByTag(name: string) {
		tag = name;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}
</script>

<div class="flex flex-col gap-6">
	<PageHeader
		title="Cams"
		description={totalCount > 0 ? `${totalCount.toLocaleString()} rooms live now` : 'Live rooms'}
	/>

	{#if $activeCamRecordings.length > 0}
		<section class="flex flex-col gap-3" aria-label="Recording now" aria-live="polite">
			<h2 class="flex items-center gap-2 text-lg font-semibold">
				<RecordIcon size={18} weight="fill" class="text-red-500" />
				Recording now
			</h2>
			<div class="no-scrollbar -mx-4 flex gap-4 overflow-x-auto px-4">
				{#each $activeCamRecordings as active (active.id)}
					<RecordingNowCard recording={active} />
				{/each}
			</div>
		</section>
	{/if}

	{#if followed.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Following</h2>
			<div class="no-scrollbar -mx-4 flex gap-4 overflow-x-auto px-4">
				{#each followed as camModel (camModel.id)}
					<div class="w-64 shrink-0">
						<CamCard
							{camModel}
							source={sourcesByName.get(camModel.pluginName)}
							onclick={() => goto(camUrl(camModel.id))}
							ontag={filterByTag}
						/>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<div class="flex flex-col gap-3">
		<div class="no-scrollbar flex items-center gap-2 overflow-x-auto">
			<label
				class="bg-base-200 border-base-300 focus-within:border-base-content/30 mr-1 flex h-9 w-64 shrink-0 items-center gap-2 rounded-lg border px-3"
			>
				<MagnifyingGlassIcon size={16} class="text-base-content/40" />
				<input
					type="text"
					bind:value={typedQuery}
					oninput={onSearchInput}
					placeholder="Search rooms"
					class="min-w-0 flex-1 bg-transparent text-sm outline-none"
				/>
			</label>
			{#each GENDERS as option (option.value)}
				<button
					type="button"
					aria-pressed={gender === option.value}
					onclick={() => (gender = option.value)}
					class="flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium transition-colors"
					class:bg-base-content={gender === option.value}
					class:text-base-100={gender === option.value}
					class:bg-base-200={gender !== option.value}
					class:hover:bg-base-300={gender !== option.value}
				>
					{option.label}
				</button>
			{/each}
			{#if sources.length > 1}
				<span class="bg-base-300 mx-1 h-5 w-px shrink-0"></span>
				{#each sources as source (source.name)}
					<button
						type="button"
						aria-pressed={selectedSources.includes(source.name)}
						onclick={() => toggleSource(source.name)}
						class="flex h-8 shrink-0 items-center gap-1.5 rounded-lg px-3 text-sm font-medium transition-colors"
						class:bg-base-content={selectedSources.includes(source.name)}
						class:text-base-100={selectedSources.includes(source.name)}
						class:bg-base-200={!selectedSources.includes(source.name)}
						class:hover:bg-base-300={!selectedSources.includes(source.name)}
					>
						{#if source.iconUrl}
							<img src={source.iconUrl} alt="" class="h-4 w-4 rounded-sm" />
						{/if}
						{source.displayName}
					</button>
				{/each}
			{/if}
		</div>

		{#if popularTags.length > 0 || tag}
			<div class="no-scrollbar flex items-center gap-2 overflow-x-auto" aria-label="Popular tags">
				{#if tagOutsidePopular}
					<button
						type="button"
						class="bg-base-content text-base-100 flex h-7 shrink-0 items-center gap-1.5 rounded-full px-3 text-[13px] font-medium"
						onclick={() => (tag = '')}
						aria-label="Remove tag filter {tag}"
					>
						#{tag}
						<XIcon size={12} weight="bold" />
					</button>
				{/if}
				{#each popularTags as popular (popular)}
					<button
						type="button"
						aria-pressed={tag === popular}
						onclick={() => toggleTag(popular)}
						class="flex h-7 shrink-0 items-center gap-1 rounded-full px-3 text-[13px] transition-colors"
						class:bg-base-content={tag === popular}
						class:text-base-100={tag === popular}
						class:font-medium={tag === popular}
						class:text-base-content={tag !== popular}
						class:hover:bg-base-200={tag !== popular}
					>
						#{popular}
						{#if tag === popular}
							<XIcon size={11} weight="bold" />
						{/if}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#if rooms.length === 0 && !loading}
		<div class="flex flex-col items-center gap-3 py-20 text-center">
			<VideoCameraIcon size={56} class="text-base-content/10" />
			{#if failed}
				<p class="text-base-content/40">The cam sites did not answer. Try again in a moment.</p>
				<button type="button" class="btn btn-sm" onclick={reload}>Retry</button>
			{:else if sources.length === 0}
				<p class="text-base-content/40">No cam plugin is installed. Add one in Plugins.</p>
			{:else}
				<p class="text-base-content/40">No rooms match these filters.</p>
			{/if}
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
			{#each rooms as camModel (camModel.id)}
				<CamCard
					{camModel}
					source={sourcesByName.get(camModel.pluginName)}
					onclick={() => goto(camUrl(camModel.id))}
					ontag={filterByTag}
				/>
			{/each}
		</div>
		<ScrollSentinel onvisible={loadMore} {loading} {done} />
	{/if}
</div>
