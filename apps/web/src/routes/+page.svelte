<!--
	Home. Without a search or filter: the focus video and rows (HomeRows) above
	the latest videos in the library, sites taking turns. The chip row and the
	Filters panel narrow the grid through the URL (libraryFilter: ?tag=, ?not=,
	?performer=, ?studio=, ?site=, ?len=, ?res=, ?rating=, ?unwatched, ?fav,
	?sort=), which then shows just the grid; ?saved= plays a saved filter. A
	search (?q=, from the top bar) asks every site and shows what they returned.
	"Browse a site" stores a page of one site's catalog.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import HomeRows from '$lib/components/HomeRows.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import FilterChips from '$lib/components/FilterChips.svelte';
	import FilterPanel from '$lib/components/FilterPanel.svelte';
	import { fetchVideosMatching, search, browse, type VideoCardRow } from '$lib/library';
	import {
		fetchPinnedFilters,
		fetchSavedFilter,
		type SavedFilterChip,
		type SavedFilterDetail
	} from '$lib/home';
	import {
		filterFromParams,
		isFiltered,
		fetchTagCounts,
		fetchFilterNames,
		fetchRecentSearches,
		type TagCountRow
	} from '$lib/libraryFilter';
	import { logEvent, resetShown } from '$lib/events';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { videoUrl } from '$lib/routes';
	import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';

	const PAGE_SIZE = 40;
	// Panel settings counted on the Filters button.
	const PANEL_PARAMS = ['len', 'res', 'rating', 'unwatched', 'fav', 'site', 'sort'];

	let videos = $state<VideoCardRow[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let savedFilter = $state<SavedFilterDetail | null>(null);
	let pinnedFilters = $state<SavedFilterChip[]>([]);
	let recentSearches = $state<string[]>([]);
	let tagCounts = $state<TagCountRow[]>([]);
	let filterNames = $state(new Map<string, string>());
	let panelOpen = $state(false);
	let browsing = $state(false);
	let browsePlugin = $state('');

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let savedId = $derived(page.url.searchParams.get('saved') ?? '');
	// Serialised so the grid reloads only when the filter itself changes.
	let filterKey = $derived(JSON.stringify(filterFromParams(page.url.searchParams)));
	let filtered = $derived(isFiltered(page.url.searchParams) || Boolean(savedId));
	let showRows = $derived(!query && !filtered);
	let gridSurface = $derived(query ? 'search' : 'library');
	let videoSites = $derived(listers($sites, 'VIDEO'));
	let panelCount = $derived(PANEL_PARAMS.filter((name) => page.url.searchParams.has(name)).length);
	let title = $derived.by(() => {
		if (query) return `Results for “${query}”`;
		if (savedFilter) return savedFilter.name;
		return 'Videos';
	});

	// A new search or filter starts the grid over.
	$effect(() => {
		void loadFirst(query, JSON.parse(filterKey), savedId);
	});

	$effect(() => {
		if (query) return;
		const filter: FilterInput = JSON.parse(filterKey);
		fetchTagCounts(filter)
			.then((found) => (tagCounts = found))
			.catch(() => (tagCounts = []));
	});

	$effect(() => {
		void filterKey;
		fetchFilterNames(page.url.searchParams)
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

	/** Fetches a page of the grid: the saved filter's videos, else the URL filter's. */
	async function fetchPage(filter: FilterInput, offset: number): Promise<VideoCardRow[]> {
		if (savedFilter) return fetchVideosMatching(savedFilter.filter, PAGE_SIZE, offset);
		return fetchVideosMatching(filter, PAGE_SIZE, offset);
	}

	/** Loads the first page: search results for a query, else the library. */
	async function loadFirst(searchText: string, filter: FilterInput, savedFilterId: string) {
		resetShown('search');
		resetShown('library');
		videos = [];
		exhausted = false;
		error = null;
		loading = true;
		try {
			if (searchText) {
				logEvent({ kind: 'search', data: { text: searchText } });
				const listed = await search('VIDEO', searchText);
				videos = listed.videos;
				exhausted = true;
				return;
			}
			savedFilter = null;
			if (savedFilterId) savedFilter = await fetchSavedFilter(savedFilterId);
			const firstPage = await fetchPage(filter, 0);
			videos = firstPage;
			exhausted = firstPage.length < PAGE_SIZE;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Appends the next page of the library grid. */
	async function loadMore() {
		if (loading || exhausted || query) return;
		loading = true;
		try {
			const nextPage = await fetchPage(JSON.parse(filterKey), videos.length);
			videos = [...videos, ...nextPage];
			exhausted = nextPage.length < PAGE_SIZE;
		} finally {
			loading = false;
		}
	}

	/** Stores a page of the picked site's catalog, then shows the library. */
	async function browseSite() {
		if (!browsePlugin) return;
		browsing = true;
		error = null;
		try {
			await browse(browsePlugin, 'VIDEO');
			await loadFirst('', JSON.parse(filterKey), savedId);
		} catch (failure) {
			error = String(failure);
		} finally {
			browsing = false;
		}
	}
</script>

<svelte:head><title>Veil</title></svelte:head>

{#snippet browsePicker()}
	{#if videoSites.length > 0}
		<div class="flex items-center gap-2">
			<select
				bind:value={browsePlugin}
				class="bg-base-200 h-9 rounded-lg px-2 text-sm"
				aria-label="Site to browse"
			>
				<option value="">Browse a site…</option>
				{#each videoSites as site (site.name)}
					<option value={site.name}>{site.displayName ?? site.name}</option>
				{/each}
			</select>
			<button
				type="button"
				class="bg-primary text-primary-content h-9 rounded-lg px-3 text-sm font-medium disabled:opacity-50"
				disabled={!browsePlugin || browsing}
				onclick={browseSite}
			>
				{browsing ? 'Loading…' : 'Browse'}
			</button>
		</div>
	{/if}
{/snippet}

<div class="flex flex-col gap-8 pt-2">
	{#if !showRows}
		<PageHeader {title} description={query ? 'From every site that can search videos' : null}>
			{#snippet actions()}
				{#if !query}
					{@render browsePicker()}
				{/if}
			{/snippet}
		</PageHeader>
	{/if}

	{#if showRows}
		<HomeRows />
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-lg font-semibold">Latest videos</h2>
			{@render browsePicker()}
		</div>
	{/if}

	{#if !query}
		<div class="-mt-4 flex flex-col gap-4">
			<FilterChips
				url={page.url}
				names={filterNames}
				{tagCounts}
				{pinnedFilters}
				{recentSearches}
				{panelCount}
				{panelOpen}
				ontogglepanel={() => (panelOpen = !panelOpen)}
			/>
			{#if panelOpen}
				<FilterPanel url={page.url} sites={videoSites} {tagCounts} />
			{/if}
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	<div
		class="-mt-2 grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-x-4 gap-y-8 max-md:grid-cols-1"
	>
		{#each videos as video, position (video.id)}
			<SceneCard
				{video}
				shown={{ surface: gridSurface, position }}
				onclick={() => goto(videoUrl(video.id))}
			/>
		{/each}
	</div>

	{#if loading && videos.length === 0}
		<p class="text-base-content/50 text-sm">Loading…</p>
	{:else if !loading && videos.length === 0}
		<p class="text-base-content/50 text-sm">
			{#if query}
				No site found anything for this search.
			{:else if filtered}
				Nothing matches this filter.
			{:else}
				The library is empty. Browse a site, or search the sites from the top bar.
			{/if}
		</p>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && videos.length > 0} done={exhausted} />
</div>
