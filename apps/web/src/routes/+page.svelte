<!--
	Home. Without a search: the featured video and rows (HomeRows) above the
	library's videos, sites taking turns. The filter under the search box
	(HeaderFilters; on phones above the rows) narrows recommendations and grid
	through the URL (libraryFilter: ?tag=, ?not=, ?performer=, ?studio=,
	?site=, ?len=, ?res=, ?rating=, ?unwatched, ?fav, ?sort=). ?saved= plays a
	saved filter in the grid alone. A search (?q=, from the top bar) asks every
	site and shows what they returned. "Browse a site" stores a page of one
	site's catalog.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import HomeRows from '$lib/components/HomeRows.svelte';
	import HeaderFilters from '$lib/components/HeaderFilters.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { fetchVideosMatching, search, browse, type VideoCardRow } from '$lib/library';
	import { fetchSavedFilter, type SavedFilterDetail } from '$lib/home';
	import { filterFromParams, isFiltered } from '$lib/libraryFilter';
	import { logEvent, resetShown } from '$lib/events';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { isCompact } from '$lib/stores/viewport';
	import { videoUrl } from '$lib/routes';
	import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';

	const PAGE_SIZE = 40;

	let videos = $state<VideoCardRow[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let savedFilter = $state<SavedFilterDetail | null>(null);
	let browsing = $state(false);
	let browsePlugin = $state('');

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let savedId = $derived(page.url.searchParams.get('saved') ?? '');
	// Serialised so the grid reloads only when the filter itself changes.
	let filterKey = $derived(JSON.stringify(filterFromParams(page.url.searchParams)));
	let filterInput = $derived<FilterInput>(JSON.parse(filterKey));
	let filtered = $derived(isFiltered(page.url.searchParams) || Boolean(savedId));
	let showRows = $derived(!query && !savedId);
	let gridSurface = $derived(query ? 'search' : 'library');
	let videoSites = $derived(listers($sites, 'VIDEO'));
	let title = $derived.by(() => {
		if (query) return `Results for “${query}”`;
		if (savedFilter) return savedFilter.name;
		return 'Videos';
	});
	let gridTitle = $derived.by(() => {
		if (filtered) return 'All matching videos';
		return 'Latest videos';
	});

	// A new search or filter starts the grid over.
	$effect(() => {
		void loadFirst(query, JSON.parse(filterKey), savedId);
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

	{#if !query && $isCompact}
		<HeaderFilters />
	{/if}

	{#if showRows}
		<HomeRows filter={filterInput} {filtered} />
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-lg font-semibold">{gridTitle}</h2>
			{@render browsePicker()}
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
