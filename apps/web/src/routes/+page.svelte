<!--
	Home: the newest videos in the library, narrowed by ?tag=, ?performer= or
	?studio=. A search (?q=, from the top bar) asks every site and shows what
	they returned. "Browse a site" stores a page of one site's catalog.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		fetchVideos,
		fetchTagName,
		search,
		browse,
		videoCard,
		type VideoCardRow,
		type MediaFilter
	} from '$lib/library';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { videoUrl } from '$lib/routes';

	const PAGE_SIZE = 40;

	let videos = $state<VideoCardRow[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let tagName = $state<string | null>(null);
	let browsing = $state(false);
	let browsePlugin = $state('');

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let filter = $derived<MediaFilter>({
		tagId: page.url.searchParams.get('tag') ?? undefined,
		performerId: page.url.searchParams.get('performer') ?? undefined,
		studioId: page.url.searchParams.get('studio') ?? undefined
	});
	let videoSites = $derived(listers($sites, 'VIDEO'));

	// A new search or filter starts the grid over.
	$effect(() => {
		void loadFirst(query, filter);
	});

	$effect(() => {
		const tagId = filter.tagId;
		tagName = null;
		if (tagId) fetchTagName(tagId).then((name) => (tagName = name));
	});

	/** Loads the first page: search results for a query, else the library. */
	async function loadFirst(searchText: string, currentFilter: MediaFilter) {
		videos = [];
		exhausted = false;
		error = null;
		loading = true;
		try {
			if (searchText) {
				const listed = await search('VIDEO', searchText);
				videos = listed.videos;
				exhausted = true;
				return;
			}
			const firstPage = await fetchVideos(currentFilter, PAGE_SIZE);
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
			const nextPage = await fetchVideos(filter, PAGE_SIZE, videos.length);
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
			await loadFirst('', filter);
		} catch (failure) {
			error = String(failure);
		} finally {
			browsing = false;
		}
	}
</script>

<svelte:head><title>Veil</title></svelte:head>

<div class="flex flex-col gap-5">
	<PageHeader
		title={query ? `Results for “${query}”` : 'Videos'}
		description={query ? 'From every site that can search videos' : null}
	>
		{#snippet actions()}
			{#if !query && videoSites.length > 0}
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
			{/if}
		{/snippet}
	</PageHeader>

	{#if tagName || filter.performerId || filter.studioId}
		<div class="flex flex-wrap gap-2">
			{#if tagName}
				<button
					type="button"
					class="bg-base-200 flex items-center gap-1.5 rounded-full px-3 py-1 text-sm"
					onclick={() => goto('/')}
				>
					Tag: {tagName}
					<XIcon size={14} />
				</button>
			{/if}
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	<div
		class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-x-4 gap-y-8 max-md:grid-cols-1"
	>
		{#each videos as video (video.id)}
			<SceneCard item={videoCard(video)} onclick={() => goto(videoUrl(video.id))} />
		{/each}
	</div>

	{#if loading && videos.length === 0}
		<p class="text-base-content/50 text-sm">Loading…</p>
	{:else if !loading && videos.length === 0}
		<p class="text-base-content/50 text-sm">
			{#if query}
				No site found anything for this search.
			{:else}
				The library is empty. Browse a site above, or search the sites from the top bar.
			{/if}
		</p>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && videos.length > 0} done={exhausted} />
</div>
