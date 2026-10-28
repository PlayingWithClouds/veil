<script lang="ts">
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import SpinnerGapIcon from 'phosphor-svelte/lib/SpinnerGapIcon';
	import { goto } from '$app/navigation';
	import SearchKeyboard from '$lib/components/SearchKeyboard.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import SearchGalleryResults from '$lib/components/SearchGalleryResults.svelte';
	import SceneFilters from '$lib/components/SceneFilters.svelte';
	import SubscribeSearchButton from '$lib/components/SubscribeSearchButton.svelte';
	import {
		fetchFilteredScenes,
		sceneToCard,
		emptyFilter,
		hasStructuredFilter,
		type SceneFilter,
		type SceneRow
	} from '$lib/filters';
	import {
		fetchSearchPlugins,
		type PluginSearchResult,
		type SearchMatchedEntity,
		type SearchPlugin
	} from '$lib/search';
	import { LiveSearch } from '$lib/liveSearch.svelte';
	import { performerUrl, studioUrl, tagUrl } from '$lib/routes';
	import { mediaFilter } from '$lib/stores/mediaFilter';
	import { SORT_OPTIONS } from '$lib/searchOrder';

	// Search-capable source plugins for the badges. Toggling one excludes it.
	let searchPlugins = $state<SearchPlugin[]>([]);
	let disabledPlugins = $state<Set<string>>(new Set());

	let activePluginNames = $derived(
		searchPlugins.filter((plugin) => !disabledPlugins.has(plugin.name)).map((plugin) => plugin.name)
	);

	/** Sources a subscription to this search should use: empty (all) unless some are excluded. */
	function subscriptionSources(): string[] {
		if (disabledPlugins.size === 0) return [];
		return activePluginNames;
	}

	function pluginLabel(plugin: SearchPlugin): string {
		return plugin.displayName || plugin.name;
	}

	function togglePlugin(name: string) {
		const next = new Set(disabledPlugins);
		if (next.has(name)) next.delete(name);
		else next.add(name);
		disabledPlugins = next;
		runSearch(query);
	}

	// The size of one page of filtered (stored) results.
	const DB_PAGE_SIZE = 60;

	let query = $state('');
	// The live search, ranked by the backend's score (see LiveSearch). Videos
	// are the main list, library matches first; galleries are ranked apart in
	// their own section, so they never take a video's place.
	const search = new LiveSearch();
	let results = $derived([...search.library, ...search.videos]);
	let galleryResults = $derived(search.galleries);
	let matchedEntities = $derived(search.matched);
	// Loading of the filtered (stored) results.
	let loading = $state(false);
	let dbDone = $state(false);

	// The top bar's media toggle: 'images' makes the galleries the main
	// content, 'videos' hides them.
	let showVideos = $derived($mediaFilter !== 'images');
	let showGalleries = $derived($mediaFilter !== 'videos');
	let showGalleryShelf = $derived(showVideos && showGalleries && galleryResults.length > 0);

	const ENTITY_LABELS: Record<SearchMatchedEntity['kind'], string> = {
		performer: 'Performer',
		studio: 'Studio',
		tag: 'Tag'
	};

	/** The page of a recognised entity. */
	function entityUrl(entity: SearchMatchedEntity): string {
		switch (entity.kind) {
			case 'performer':
				return performerUrl(entity.id);
			case 'studio':
				return studioUrl(entity.id);
			case 'tag':
				return tagUrl(entity.id);
		}
	}

	/** Logs that the user opened a result of the current search. */
	function logOpened(result: PluginSearchResult) {
		search.logOpened(result);
	}

	// Advanced DB filters. When a structured dimension is set, the page switches
	// from live plugin search to DB-backed filtered results.
	let filter = $state<SceneFilter>(emptyFilter());
	let filterMode = $derived(hasStructuredFilter(filter));
	let dbResults = $state<SceneRow[]>([]);

	async function runDbFilter(text: string) {
		const token = ++requestToken;
		loading = true;
		const rows = await fetchFilteredScenes({ ...filter, search: text || null }, DB_PAGE_SIZE, 0);
		if (token !== requestToken) return;
		dbResults = rows;
		dbDone = rows.length < DB_PAGE_SIZE;
		loading = false;
	}

	/** Appends the next page of filtered results. */
	async function loadMoreDb() {
		if (loading || dbDone) return;
		const token = requestToken;
		loading = true;
		const rows = await fetchFilteredScenes(
			{ ...filter, search: query || null },
			DB_PAGE_SIZE,
			dbResults.length
		);
		if (token !== requestToken) return;
		dbResults = [...dbResults, ...rows];
		dbDone = rows.length < DB_PAGE_SIZE;
		loading = false;
	}

	// Re-run whichever mode is active. Called on filter commit and query changes.
	function refresh() {
		if (filterMode) runDbFilter(query);
		else runSearch(query);
	}

	let debounceTimer: ReturnType<typeof setTimeout> | null = null;
	let requestToken = 0;

	let keyboard = $state<{ focus: () => void; focusFirst: () => void }>();
	let grid = $state<{ focusFirst: () => void }>();
	let galleryGrid = $state<{ focusFirst: () => void }>();

	/** Starts the live search for text on the selected sources; nothing without text or sites. */
	function runSearch(text: string) {
		++requestToken;
		const noSources = searchPlugins.length > 0 && activePluginNames.length === 0;
		if (!text.trim() || noSources) {
			search.run('', [], []);
			return;
		}
		search.run(text, activePluginNames, activePluginNames);
	}

	/** Loads the next page of live results when the end of the grid comes near. */
	function loadMoreLive() {
		search.loadMore();
	}

	function scheduleSearch() {
		if (debounceTimer) clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => refresh(), 300);
	}

	function appendChar(char: string) {
		query += char;
		scheduleSearch();
	}
	function backspace() {
		query = query.slice(0, -1);
		scheduleSearch();
	}
	function clearQuery() {
		query = '';
		scheduleSearch();
	}

	/** Hands keyboard focus from the on-screen keyboard to the first result shown. */
	function focusResults() {
		if (showGalleryShelf || !showVideos) {
			galleryGrid?.focusFirst();
			return;
		}
		grid?.focusFirst();
	}

	/** Moves keyboard focus out of the gallery section: left to the keyboard, down to the videos. */
	function leaveGalleries(direction: 'left' | 'right' | 'up' | 'down') {
		if (direction === 'left') keyboard?.focus();
		if (direction === 'down' && showVideos) grid?.focusFirst();
	}

	/** Moves keyboard focus from the top of the videos up to the gallery shelf. */
	function enterGalleryShelf() {
		if (showGalleryShelf) galleryGrid?.focusFirst();
	}

	/** What the empty video grid says: "No videos." when galleries are shown above it. */
	function videosEmptyMessage(): string {
		if (showGalleryShelf) return 'No videos.';
		return 'No results.';
	}

	/** Whether the main list (videos, or galleries on the images toggle) has any result. */
	function hasMainResults(): boolean {
		if (showVideos) return results.length > 0;
		return galleryResults.length > 0;
	}

	// Focus the keyboard on entry so a TV remote can start typing immediately.
	$effect(() => {
		queueMicrotask(() => keyboard?.focusFirst());
	});

	$effect(() => {
		fetchSearchPlugins().then((found) => (searchPlugins = found));
	});
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-center gap-3">
		<h1 class="text-3xl font-bold">Search</h1>
		<span class="badge badge-error">18+</span>
	</div>

	<!-- Source badges: which plugins feed the search. Toggle to include/exclude. -->
	<div class="flex flex-wrap items-center gap-2">
		<span class="text-base-content/40 text-xs font-medium uppercase">Sources</span>
		{#if searchPlugins.length === 0}
			<span class="text-base-content/30 text-sm">No search plugins installed</span>
		{/if}
		{#each searchPlugins as plugin (plugin.name)}
			{@const on = !disabledPlugins.has(plugin.name)}
			<button
				type="button"
				onclick={() => togglePlugin(plugin.name)}
				aria-pressed={on}
				class="flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm font-medium transition-colors
					{on
					? 'border-primary bg-primary/15 text-primary'
					: 'border-base-300 text-base-content/40 hover:text-base-content/70'}"
			>
				{#if plugin.iconUrl}
					<img
						src={plugin.iconUrl}
						alt=""
						class="h-4 w-4 rounded-sm {on ? '' : 'opacity-40 grayscale'}"
					/>
				{/if}
				{pluginLabel(plugin)}
			</button>
		{/each}
	</div>

	<!-- Advanced filters -->
	<SceneFilters bind:filter oncommit={refresh} />

	<!-- Query display -->
	<div class="border-base-300 bg-base-200 flex items-center gap-3 rounded-xl border px-4 py-3">
		{#if loading || search.loading}
			<SpinnerGapIcon size={22} class="text-primary shrink-0 animate-spin" />
		{:else}
			<MagnifyingGlassIcon size={22} class="text-base-content/40 shrink-0" />
		{/if}
		{#if $isCompact}
			<!-- Phones type with their own keyboard instead of the on-screen one. -->
			<input
				type="search"
				enterkeyhint="search"
				bind:value={query}
				oninput={scheduleSearch}
				placeholder="Search adult sources…"
				class="placeholder:text-base-content/30 min-w-0 flex-1 bg-transparent text-lg outline-none"
			/>
		{:else}
			<span class="flex-1 truncate text-lg {query ? '' : 'text-base-content/30'}">
				{query || 'Search adult sources…'}
			</span>
		{/if}
		{#if query.trim() && !filterMode}
			<SubscribeSearchButton {query} sources={subscriptionSources()} />
		{/if}
	</div>

	<div class="flex flex-col gap-6 lg:flex-row">
		<!-- On-screen keyboard (top-left) -->
		<div class="w-full shrink-0 max-md:hidden lg:w-96">
			<SearchKeyboard
				bind:this={keyboard}
				oninput={appendChar}
				onbackspace={backspace}
				onclear={clearQuery}
				onedge={focusResults}
			/>
		</div>

		<!-- Results grid (right) -->
		<div class="min-w-0 flex-1">
			{#if filterMode}
				{#if dbResults.length === 0 && !loading}
					<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
						<MagnifyingGlassIcon size={48} class="opacity-20" />
						<p>No scenes match these filters.</p>
					</div>
				{:else}
					<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-4">
						{#each dbResults as scene (scene.id)}
							<SceneCard item={sceneToCard(scene)} onclick={() => goto(`/scene/${scene.id}`)} />
						{/each}
					</div>
					<ScrollSentinel onvisible={loadMoreDb} {loading} done={dbDone} />
				{/if}
			{:else if !query.trim()}
				<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
					<MagnifyingGlassIcon size={48} class="opacity-20" />
					<p>Type to search, or set a filter.</p>
				</div>
			{:else}
				<!-- Sort order of the live results; relevance is the server's score. -->
				<div class="mb-4 flex flex-wrap items-center gap-2">
					<span class="text-base-content/40 text-xs font-medium uppercase">Sort</span>
					{#each SORT_OPTIONS as option (option.mode)}
						<button
							type="button"
							onclick={() => search.changeSort(option.mode)}
							aria-pressed={search.sortMode === option.mode}
							class="rounded-full border px-3 py-1 text-sm font-medium transition-colors
								{search.sortMode === option.mode
								? 'border-primary bg-primary/15 text-primary'
								: 'border-base-300 text-base-content/40 hover:text-base-content/70'}"
						>
							{option.label}
						</button>
					{/each}
				</div>
				{#if matchedEntities.length > 0}
					<!-- Library entities the query names. -->
					<div class="mb-4 flex flex-wrap items-center gap-2">
						{#each matchedEntities as entity (entity.id)}
							<a
								href={entityUrl(entity)}
								class="border-base-300 hover:border-primary hover:text-primary rounded-full border px-3 py-1 text-sm font-medium transition-colors"
							>
								{#if entity.fuzzy}<span class="text-base-content/40">Did you mean</span>{/if}
								{ENTITY_LABELS[entity.kind]}: {entity.name}
							</a>
						{/each}
					</div>
				{/if}
				{#if showGalleryShelf}
					<section class="mb-6 flex flex-col gap-3" aria-label="Galleries">
						<h2 class="flex items-baseline gap-2 text-lg font-semibold">
							Galleries
							<span class="text-base-content/40 text-sm font-normal">{galleryResults.length}</span>
						</h2>
						<SearchGalleryResults
							bind:this={galleryGrid}
							items={galleryResults}
							sites={searchPlugins}
							layout="row"
							onedge={leaveGalleries}
						/>
					</section>
					<h2 class="mb-3 text-lg font-semibold">Videos</h2>
				{/if}
				{#if showVideos}
					<NsfwGrid
						bind:this={grid}
						onopen={logOpened}
						items={results}
						loading={search.loading && results.length === 0}
						ariaLabel="Search results"
						emptyMessage={videosEmptyMessage()}
						onleftedge={() => keyboard?.focus()}
						ontopedge={enterGalleryShelf}
					/>
				{:else}
					<SearchGalleryResults
						bind:this={galleryGrid}
						items={galleryResults}
						sites={searchPlugins}
						layout="grid"
						loading={search.loading && galleryResults.length === 0}
						onedge={leaveGalleries}
					/>
				{/if}
				{#if hasMainResults()}
					<ScrollSentinel onvisible={loadMoreLive} loading={search.loading} done={search.done} />
				{/if}
			{/if}
		</div>
	</div>
</div>
