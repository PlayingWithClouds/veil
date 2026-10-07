<script lang="ts">
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

	let query = $state('');
	let results = $state<PluginSearchResult[]>([]);
	let loading = $state(false);

	// Advanced DB filters. When a structured dimension is set, the page switches
	// from live plugin search to DB-backed filtered results.
	let filter = $state<SceneFilter>(emptyFilter());
	let filterMode = $derived(hasStructuredFilter(filter));
	let dbResults = $state<SceneRow[]>([]);

	async function runDbFilter(text: string) {
		const token = ++requestToken;
		loading = true;
		const rows = await fetchFilteredScenes({ ...filter, search: text || null }, 60, 0);
		if (token !== requestToken) return;
		dbResults = rows;
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
		loading = true;
		const seen = new Set<string>();
		currentStream = pluginSearchStream(
			text,
			activePluginNames,
			(result) => {
				if (token !== requestToken) return;
				if (seen.has(result.externalId)) return;
				seen.add(result.externalId);
				results = [...results, result];
			},
			() => {
				if (token !== requestToken) return;
				loading = false;
			}
		);
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
				{/if}
			{:else if !query.trim()}
				<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
					<MagnifyingGlassIcon size={48} class="opacity-20" />
					<p>Type to search, or set a filter.</p>
				</div>
			{:else}
				<NsfwGrid
					bind:this={grid}
					items={results}
					{loading}
					ariaLabel="Search results"
					onleftedge={() => keyboard?.focus()}
				/>
			{/if}
		</div>
	</div>
</div>
