<!--
	Home. Without a search: the featured video and rows (HomeRows) above the
	library's videos, endless: unfiltered, new pages from the video sites
	follow once the library is through. The filter under the search box
	(HeaderFilters; on phones above the rows) narrows recommendations and grid
	through the URL (libraryFilter: ?tag=, ?not=, ?performer=, ?studio=,
	?site=, ?len=, ?res=, ?rating=, ?unwatched, ?fav, ?sort=). ?saved= plays a
	saved filter in the grid alone. A search (?q=, from the top bar) asks every
	site, streamed and endless (SearchResults). "Browse a site" stores a page of one
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
	import SearchResults from '$lib/components/SearchResults.svelte';
	import { fetchVideosMatching, browse, browseVideos, type VideoCardRow } from '$lib/library';
	import { endlessFeed, siteSource, type Feed, type Source } from '$lib/feed';
	import { listerNames } from '$lib/stores/sites';
	import { fetchSavedFilter, type SavedFilterDetail } from '$lib/home';
	import { filterFromParams, isFiltered } from '$lib/libraryFilter';
	import { resetShown } from '$lib/events';
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
	// The grid's feed for the current filter; replaced when it changes.
	let grid: Feed<VideoCardRow> | null = null;
	let browsing = $state(false);
	let browsePlugin = $state('');

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let savedId = $derived(page.url.searchParams.get('saved') ?? '');
	// Serialised so the grid reloads only when the filter itself changes.
	let filterKey = $derived(JSON.stringify(filterFromParams(page.url.searchParams)));
	let filterInput = $derived<FilterInput>(JSON.parse(filterKey));
	let filtered = $derived(isFiltered(page.url.searchParams) || Boolean(savedId));
	let showRows = $derived(!query && !savedId);
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

	/**
	 * The grid's feed: the saved filter's videos, else the URL filter's; with
	 * neither, new videos from the sites once the library is through (the
	 * sites can't apply a filter).
	 */
	function gridFeed(filter: FilterInput): Feed<VideoCardRow> {
		const shownFilter = savedFilter ? savedFilter.filter : filter;
		let sites: Source<VideoCardRow> | undefined;
		if (!savedFilter && !filtered) sites = siteSource(() => listerNames('VIDEO'), browseVideos);
		return endlessFeed({
			rest: (limit, offset) => fetchVideosMatching(shownFilter, limit, offset),
			sites
		});
	}

	/** Starts the grid over; a search has its own results. */
	async function loadFirst(searchText: string, filter: FilterInput, savedFilterId: string) {
		resetShown('library');
		videos = [];
		exhausted = false;
		error = null;
		grid = null;
		if (searchText) return;
		try {
			savedFilter = null;
			if (savedFilterId) savedFilter = await fetchSavedFilter(savedFilterId);
			grid = gridFeed(filter);
			await loadMore();
		} catch (failure) {
			error = String(failure);
		}
	}

	/** Appends the grid's next page. */
	async function loadMore() {
		if (!grid || loading || exhausted || query) return;
		const current = grid;
		loading = true;
		try {
			const nextPage = await current(PAGE_SIZE);
			if (current !== grid) return;
			videos = [...videos, ...nextPage];
			exhausted = nextPage.length === 0;
		} catch (failure) {
			error = String(failure);
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

	{#if query}
		<!-- Keyed by the filter, so changing it starts the search over. -->
		{#key filterKey}
			<SearchResults {query} filter={filtered ? filterInput : null} />
		{/key}
	{:else}
		<div
			class="-mt-2 grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-x-4 gap-y-8 max-md:grid-cols-1"
		>
			{#each videos as video, position (video.id)}
				<SceneCard
					{video}
					shown={{ surface: 'library', position }}
					onclick={() => goto(videoUrl(video.id))}
				/>
			{/each}
		</div>

		{#if loading && videos.length === 0}
			<p class="text-base-content/50 text-sm">Loading…</p>
		{:else if !loading && videos.length === 0}
			<p class="text-base-content/50 text-sm">
				{#if filtered}
					Nothing matches this filter.
				{:else}
					The library is empty. Browse a site, or search the sites from the top bar.
				{/if}
			</p>
		{/if}

		<ScrollSentinel onvisible={loadMore} loading={loading && videos.length > 0} done={exhausted} />
	{/if}
</div>
