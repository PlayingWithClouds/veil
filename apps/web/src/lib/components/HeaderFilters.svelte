<!--
	The library filter under the search box: the chip row and, from its Filters
	button, the panel dropping down over the page. The filter lives in home's
	URL (libraryFilter); off home, every chip and option leads to home with it.
	Home's rows and grid read the same URL, so recommendations follow it too.
-->
<script lang="ts">
	import { page } from '$app/state';
	import FilterChips from './FilterChips.svelte';
	import FilterPanel from './FilterPanel.svelte';
	import { fetchPinnedFilters, type SavedFilterChip } from '$lib/home';
	import {
		filterFromParams,
		fetchTagCounts,
		fetchFilterNames,
		fetchRecentSearches,
		type TagCountRow
	} from '$lib/libraryFilter';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';

	// Panel settings counted on the Filters button.
	const PANEL_PARAMS = ['len', 'res', 'rating', 'unwatched', 'fav', 'site', 'sort'];

	let panelOpen = $state(false);
	let root = $state<HTMLElement | null>(null);
	let pinnedFilters = $state<SavedFilterChip[]>([]);
	let recentSearches = $state<string[]>([]);
	let tagCounts = $state<TagCountRow[]>([]);
	let filterNames = $state(new Map<string, string>());

	// Home's URL: the current one on home, else home without a filter.
	let base = $derived.by(() => {
		if (page.url.pathname === '/') return page.url;
		return new URL('/', page.url.origin);
	});
	// Serialised so the lookups rerun only when the filter itself changes.
	let filterKey = $derived(JSON.stringify(filterFromParams(base.searchParams)));
	let panelCount = $derived(PANEL_PARAMS.filter((name) => base.searchParams.has(name)).length);
	let videoSites = $derived(listers($sites, 'VIDEO'));

	$effect(() => {
		fetchTagCounts(JSON.parse(filterKey))
			.then((found) => (tagCounts = found))
			.catch(() => (tagCounts = []));
	});

	$effect(() => {
		void filterKey;
		fetchFilterNames(base.searchParams)
			.then((found) => (filterNames = found))
			.catch(() => (filterNames = new Map()));
	});

	$effect(() => {
		fetchPinnedFilters()
			.then((found) => (pinnedFilters = found))
			.catch(() => (pinnedFilters = []));
		fetchRecentSearches()
			.then((found) => (recentSearches = found))
			.catch(() => (recentSearches = []));
	});

	/** Closes the panel on a press outside the filter. */
	function onWindowPointerDown(event: PointerEvent) {
		if (!panelOpen || root?.contains(event.target as Node)) return;
		panelOpen = false;
	}

	/** Closes the panel on Escape. */
	function onWindowKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') panelOpen = false;
	}
</script>

<svelte:window onpointerdown={onWindowPointerDown} onkeydown={onWindowKeydown} />

<div bind:this={root} class="relative min-w-0">
	<FilterChips
		url={base}
		names={filterNames}
		{tagCounts}
		{pinnedFilters}
		{recentSearches}
		{panelCount}
		{panelOpen}
		ontogglepanel={() => (panelOpen = !panelOpen)}
	/>
	{#if panelOpen}
		<div
			class="z-modal bg-base-100 absolute top-full left-1/2 mt-2 w-[min(64rem,calc(100vw-2rem))] -translate-x-1/2 rounded-2xl shadow-2xl"
		>
			<FilterPanel url={base} sites={videoSites} {tagCounts} />
		</div>
	{/if}
</div>
