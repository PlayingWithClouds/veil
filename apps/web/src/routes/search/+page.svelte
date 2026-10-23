<script lang="ts">
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import SpinnerGapIcon from 'phosphor-svelte/lib/SpinnerGapIcon';
	import { goto } from '$app/navigation';
	import SearchKeyboard from '$lib/components/SearchKeyboard.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
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
		pluginSearchStream,
		fetchSearchPlugins,
		type PluginSearchResult,
		type SearchPlugin,
		type SearchStream
	} from '$lib/search';
	import {
		FIRST_RESULTS_WINDOW_MS,
		FROZEN_VISIBLE_COUNT,
		SORT_OPTIONS,
		insertArrival,
		sortResults,
		type SortMode
	} from '$lib/searchOrder';

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

	// Each plugin's share of one page of live results, and the size of one
	// page of filtered (stored) results.
	const PER_PLUGIN = 20;
	const DB_PAGE_SIZE = 60;

	let query = $state('');
	let results = $state<PluginSearchResult[]>([]);
	let loading = $state(false);
	// Paging of the live search: the next page's offset, the keys shown so
	// far, and whether a page came back with nothing new.
	let liveOffset = 0;
	let seen = new Set<string>();
	let liveDone = $state(false);
	let dbDone = $state(false);

	// Ordering of the live results. The server scores every result as it
	// streams in; the first second of arrivals is held back and sorted
	// together, so the grid opens in order. From then on a late result is
	// sorted in below the first FROZEN_VISIBLE_COUNT results (and below
	// everything shown when a further page loads), so the screen never
	// reshuffles under the reader's eyes.
	let sortMode = $state<SortMode>('relevance');
	let heldResults: PluginSearchResult[] = [];
	let firstResultsTimer: ReturnType<typeof setTimeout> | null = null;
	let firstResultsShown = false;
	let frozenCount = 0;

	/** Shows the held first results, sorted. */
	function showFirstResults() {
		if (firstResultsTimer) clearTimeout(firstResultsTimer);
		firstResultsTimer = null;
		if (firstResultsShown) return;
		firstResultsShown = true;
		results = sortResults([...results, ...heldResults], sortMode);
		heldResults = [];
		frozenCount = Math.min(results.length, FROZEN_VISIBLE_COUNT);
	}

	/** Takes one streamed result: held during the first second, then sorted in. */
	function receiveResult(result: PluginSearchResult) {
		if (!firstResultsShown) {
			heldResults.push(result);
			applyReplacement(result);
			if (!firstResultsTimer) {
				firstResultsTimer = setTimeout(showFirstResults, FIRST_RESULTS_WINDOW_MS);
			}
			return;
		}
		results = insertArrival(results, result, sortMode, frozenCount);
	}

	/** Drops a held result that a better copy of the same video supersedes. */
	function applyReplacement(result: PluginSearchResult) {
		if (!result.replaces) return;
		heldResults = heldResults.filter(
			(held) => held === result || held.sourceUrl !== result.replaces
		);
	}

	/** Re-sorts everything shown when the reader picks another order. */
	function changeSort(mode: SortMode) {
		sortMode = mode;
		results = sortResults(results, mode);
		heldResults = sortResults(heldResults, mode);
		frozenCount = Math.min(results.length, FROZEN_VISIBLE_COUNT);
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
	let currentStream: SearchStream | null = null;

	let keyboard = $state<{ focus: () => void; focusFirst: () => void }>();
	let grid = $state<{ focusFirst: () => void }>();

	function runSearch(text: string) {
		const token = ++requestToken;
		if (currentStream) {
			currentStream.close();
			currentStream = null;
		}
		if (!text.trim()) {
			results = [];
			loading = false;
			return;
		}
		// No sources selected means nothing to search.
		if (searchPlugins.length > 0 && activePluginNames.length === 0) {
			results = [];
			loading = false;
			return;
		}
		results = [];
		heldResults = [];
		if (firstResultsTimer) clearTimeout(firstResultsTimer);
		firstResultsTimer = null;
		firstResultsShown = false;
		frozenCount = 0;
		seen = new Set();
		liveOffset = 0;
		liveDone = false;
		streamNextPage(text, token);
	}

	/** Streams the sites' next page of results; a page adding nothing ends the list. */
	function streamNextPage(text: string, token: number) {
		loading = true;
		// A further page is below what the reader has scrolled through.
		frozenCount = results.length;
		let added = 0;
		currentStream = pluginSearchStream(
			text,
			activePluginNames,
			(result) => {
				if (token !== requestToken) return;
				if (seen.has(result.externalId)) return;
				seen.add(result.externalId);
				added++;
				receiveResult(result);
			},
			() => {
				if (token !== requestToken) return;
				showFirstResults();
				loading = false;
				currentStream = null;
				if (added === 0) liveDone = true;
			},
			{ offset: liveOffset, limit: PER_PLUGIN }
		);
		liveOffset += PER_PLUGIN;
	}

	/** Loads the next page of live results when the end of the grid comes near. */
	function loadMoreLive() {
		if (loading || liveDone || !query.trim()) return;
		streamNextPage(query, requestToken);
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

	function focusResults() {
		if (results.length === 0) return;
		grid?.focusFirst();
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
		{#if loading}
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
							onclick={() => changeSort(option.mode)}
							aria-pressed={sortMode === option.mode}
							class="rounded-full border px-3 py-1 text-sm font-medium transition-colors
								{sortMode === option.mode
								? 'border-primary bg-primary/15 text-primary'
								: 'border-base-300 text-base-content/40 hover:text-base-content/70'}"
						>
							{option.label}
						</button>
					{/each}
				</div>
				<NsfwGrid
					bind:this={grid}
					items={results}
					loading={loading && results.length === 0}
					ariaLabel="Search results"
					onleftedge={() => keyboard?.focus()}
				/>
				{#if results.length > 0}
					<ScrollSentinel onvisible={loadMoreLive} {loading} done={liveDone} />
				{/if}
			{/if}
		</div>
	</div>
</div>
