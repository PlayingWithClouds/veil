<script lang="ts">
	// Home while searching. From the top: the performers and studios the query
	// names, other name matches, then the results in the order the backend ranks
	// them (or the picked sort): stored scenes in their own row, site videos in
	// an even 16:9 grid, and the galleries as a shelf of portrait covers after
	// the first two rows (or as the grid itself on the images filter).
	import { goto } from '$app/navigation';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import PerformerCard from '$lib/components/PerformerCard.svelte';
	import SubscribeSearchButton from '$lib/components/SubscribeSearchButton.svelte';
	import SearchEntityStrip from './SearchEntityStrip.svelte';
	import { cacheUrl } from '$lib/img';
	import { galleryUrl, performerUrl, sceneUrl, studioUrl } from '$lib/routes';
	import {
		pluginResultToGalleryHit,
		searchPerformers,
		searchStudios,
		streamResultToGridItem,
		type PerformerHit,
		type SceneHit,
		type StudioHit
	} from '$lib/explore';
	import type { LiveSearch } from '$lib/liveSearch.svelte';
	import ContentKindPicker from './ContentKindPicker.svelte';
	import {
		pluginForUrl,
		type ContentKind,
		type PluginSearchResult,
		type SearchPlugin
	} from '$lib/search';
	import { SORT_OPTIONS } from '$lib/searchOrder';
	import { mediaFilter } from '$lib/stores/mediaFilter';

	interface Props {
		search: LiveSearch;
		// Search plugins, to label each card with its site.
		sites: SearchPlugin[];
		// Sites the search is limited to (empty = all), for the subscribe button.
		sources: string[];
		// The sites a search can ask, for the kind picker.
		searchableSites: SearchPlugin[];
		// The kind of content asked for, null for anything.
		content: ContentKind | null;
		oncontent: (kind: ContentKind | null) => void;
	}

	let { search, sites, sources, searchableSites, content, oncontent }: Props = $props();

	// Video grid geometry; the gallery shelf goes after VIDEO_ROWS_BEFORE_SHELF rows.
	const VIDEO_CARD_MIN_WIDTH = 280;
	const GRID_GAP = 24;
	const VIDEO_ROWS_BEFORE_SHELF = 2;

	let showVideos = $derived($mediaFilter !== 'images');
	let showGalleries = $derived($mediaFilter !== 'videos');

	// The grid's width decides its column count, and so where two rows end.
	let gridWidth = $state(0);
	let columns = $derived(
		Math.max(1, Math.floor((gridWidth + GRID_GAP) / (VIDEO_CARD_MIN_WIDTH + GRID_GAP)))
	);
	let videosBeforeShelf = $derived(search.videos.slice(0, columns * VIDEO_ROWS_BEFORE_SHELF));
	let videosAfterShelf = $derived(search.videos.slice(columns * VIDEO_ROWS_BEFORE_SHELF));
	let gallerySection = $derived(showGalleries && search.galleries.length > 0);

	// Performers and studios whose name contains the query, apart from the ones
	// the entity strip shows.
	let performers = $state<PerformerHit[]>([]);
	let studios = $state<StudioHit[]>([]);
	let matchedIds = $derived(new Set(search.matched.map((entity) => entity.id)));
	let otherPerformers = $derived(performers.filter((performer) => !matchedIds.has(performer.id)));
	let otherStudios = $derived(studios.filter((studio) => !matchedIds.has(studio.id)));

	$effect(() => {
		const query = search.query;
		performers = [];
		studios = [];
		if (!query) return;
		Promise.all([searchPerformers(query), searchStudios(query)]).then(
			([foundPerformers, foundStudios]) => {
				if (search.query !== query) return;
				performers = foundPerformers;
				studios = foundStudios;
			}
		);
	});

	let hasResults = $derived.by(() => {
		if (showVideos && (search.library.length > 0 || search.videos.length > 0)) return true;
		return gallerySection;
	});
	let searchingSites = $derived(
		[...search.sites.values()].some((site) => site.state === 'searching')
	);

	// Infinite scroll: the next page loads as the end comes near. Re-observing
	// after each page makes a still-visible sentinel load the one after.
	let sentinel = $state<HTMLElement | null>(null);
	$effect(() => {
		if (!sentinel || search.loading || search.done) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) search.loadMore();
			},
			{ rootMargin: '800px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	/** The scene card data of a result. */
	function sceneOf(result: PluginSearchResult): SceneHit | null {
		const item = streamResultToGridItem(result);
		if (!item || item.kind !== 'scene') return null;
		return item.scene;
	}

	/** Opens a video result, logging the click for the search ranking. */
	function openScene(result: PluginSearchResult) {
		search.logOpened(result);
		if (result.dbId) goto(sceneUrl(result.dbId));
	}

	/** Opens a gallery result, logging the click. */
	function openGallery(result: PluginSearchResult) {
		search.logOpened(result);
		if (result.dbId) goto(galleryUrl(result.dbId));
	}
</script>

<div class="flex flex-col gap-6">
	<ContentKindPicker sites={searchableSites} selected={content} onselect={oncontent} />

	<SearchEntityStrip entities={search.matched} />

	{#if otherPerformers.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Performers</h2>
			<div class="no-scrollbar flex gap-4 overflow-x-auto pb-1">
				{#each otherPerformers as performer (performer.id)}
					<div class="w-28 shrink-0">
						<PerformerCard {performer} onclick={() => goto(performerUrl(performer.id))} />
					</div>
				{/each}
			</div>
		</section>
	{/if}

	{#if otherStudios.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Studios</h2>
			<div class="no-scrollbar flex gap-4 overflow-x-auto pb-1">
				{#each otherStudios as studio (studio.id)}
					<a
						href={studioUrl(studio.id)}
						class="flex w-32 shrink-0 flex-col items-center gap-2 text-center"
					>
						<span
							class="tv-card bg-base-300 flex aspect-video w-full items-center justify-center overflow-hidden rounded-lg"
						>
							{#if cacheUrl(studio.imagePath)}
								<img
									src={cacheUrl(studio.imagePath)}
									alt={studio.name}
									loading="lazy"
									class="h-full w-full object-contain"
								/>
							{:else}
								<span class="text-2xl opacity-20">🎬</span>
							{/if}
						</span>
						<span class="line-clamp-1 text-xs font-medium">{studio.name}</span>
					</a>
				{/each}
			</div>
		</section>
	{/if}

	<div class="flex flex-wrap items-center justify-between gap-3">
		<div class="flex flex-wrap items-center gap-2">
			<h2 class="mr-2 flex items-center gap-3 text-lg font-semibold">
				Results
				{#if searchingSites || (search.loading && !hasResults)}
					<span class="loading loading-spinner loading-xs text-base-content/40"></span>
				{/if}
			</h2>
			{#each SORT_OPTIONS as option (option.mode)}
				<button
					type="button"
					class="sort-chip"
					class:sort-chip-active={search.sortMode === option.mode}
					aria-pressed={search.sortMode === option.mode}
					onclick={() => search.changeSort(option.mode)}
				>
					{option.label}
				</button>
			{/each}
		</div>
		<SubscribeSearchButton query={search.query} {sources} />
	</div>

	{#if showVideos && search.library.length > 0}
		<section class="flex flex-col gap-3" aria-label="In your library">
			<h3 class="flex items-baseline gap-2 font-semibold">
				In your library
				<span class="text-base-content/40 text-sm font-normal">{search.library.length}</span>
			</h3>
			<div class="no-scrollbar -mx-1 flex gap-5 overflow-x-auto px-1 pb-1">
				{#each search.library as result (result.dbId)}
					{@const scene = sceneOf(result)}
					{#if scene}
						<div class="w-72 shrink-0">
							<ExploreSceneCard
								{scene}
								site={pluginForUrl(result.sourceUrl, sites)}
								onclick={() => openScene(result)}
							/>
						</div>
					{/if}
				{/each}
			</div>
		</section>
	{/if}

	{#if !hasResults && !search.loading && !searchingSites}
		<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
			<MagnifyingGlassIcon size={48} class="opacity-20" />
			<p>Nothing matches your search.</p>
		</div>
	{:else if showVideos}
		<div class="flex flex-col gap-8" bind:clientWidth={gridWidth}>
			{#if search.videos.length > 0}
				{@render videoGrid(videosBeforeShelf)}
			{/if}
			{#if gallerySection}
				<section class="flex flex-col gap-3" aria-label="Galleries">
					<h3 class="flex items-baseline gap-2 font-semibold">
						Galleries
						<span class="text-base-content/40 text-sm font-normal">{search.galleries.length}</span>
					</h3>
					<div class="no-scrollbar -mx-1 flex gap-4 overflow-x-auto px-1 pb-1">
						{#each search.galleries as result (result.dbId)}
							<div class="w-44 shrink-0">
								{@render galleryCard(result)}
							</div>
						{/each}
					</div>
				</section>
			{/if}
			{#if videosAfterShelf.length > 0}
				{@render videoGrid(videosAfterShelf)}
			{/if}
		</div>
	{:else}
		<div class="gallery-grid grid gap-6">
			{#each search.galleries as result (result.dbId)}
				{@render galleryCard(result)}
			{/each}
		</div>
	{/if}

	{#if search.loading && hasResults}
		<div class="flex justify-center py-6">
			<span class="loading loading-spinner text-base-content/40"></span>
		</div>
	{/if}
	<div bind:this={sentinel} class="h-1"></div>
</div>

{#snippet videoGrid(results: PluginSearchResult[])}
	<div class="video-grid grid gap-x-6 gap-y-8">
		{#each results as result (result.dbId)}
			{@const scene = sceneOf(result)}
			{#if scene}
				<ExploreSceneCard
					{scene}
					site={pluginForUrl(result.sourceUrl, sites)}
					onclick={() => openScene(result)}
				/>
			{/if}
		{/each}
	</div>
{/snippet}

{#snippet galleryCard(result: PluginSearchResult)}
	<GalleryCard
		gallery={pluginResultToGalleryHit(result, result.dbId ?? '')}
		site={pluginForUrl(result.sourceUrl, sites)}
		onclick={() => openGallery(result)}
	/>
{/snippet}

<style>
	.video-grid {
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	}

	.gallery-grid {
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
	}

	.sort-chip {
		height: 2rem;
		padding-inline: 0.75rem;
		border-radius: 0.5rem;
		background-color: var(--color-base-200);
		font-size: 14px;
		font-weight: 500;
		transition: background-color 150ms;
	}

	.sort-chip:hover {
		background-color: var(--color-base-300);
	}

	.sort-chip-active,
	.sort-chip-active:hover {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
