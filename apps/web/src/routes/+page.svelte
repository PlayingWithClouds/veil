<script lang="ts">
	import { browser } from '$app/environment';
	import { beforeNavigate, goto, replaceState } from '$app/navigation';
	import { page } from '$app/stores';
	import { tick, untrack } from 'svelte';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import { clearHomeFeed, homeFeedKey, loadHomeFeed, saveHomeFeed } from '$lib/homeFeedCache';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import HomeFeed from '$lib/components/home/HomeFeed.svelte';
	import SiteChips from '$lib/components/SiteChips.svelte';
	import MediaFilterChip from '$lib/components/MediaFilterChip.svelte';
	import FaceSearch from '$lib/components/FaceSearch.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import { searchQuery } from '$lib/stores/searchQuery';
	import { mediaFilter } from '$lib/stores/mediaFilter';
	import HomeSearch from '$lib/components/search/HomeSearch.svelte';
	import { LiveSearch } from '$lib/liveSearch.svelte';
	import { sceneUrl } from '$lib/routes';
	import {
		carriesContent,
		fetchSearchPlugins,
		parseContentKind,
		supportsMediaFilter,
		type ContentKind,
		type SearchPlugin
	} from '$lib/search';
	import {
		recordClicked,
		resetShownImpressions,
		type ImpressionSurface,
		type ShownEvent
	} from '$lib/recommendations';
	import {
		searchGalleries,
		fetchContinueWatching,
		fetchRecommendationShelves,
		fetchRecommendedSceneHits,
		fetchPreferredTags,
		bestPreferredTag,
		gridItemKey,
		gridItemSiteKey,
		mixBySite,
		type GalleryHit,
		type RecommendationShelf,
		type SceneHit,
		type GridItem
	} from '$lib/explore';

	const SEARCH_DEBOUNCE_MS = 300;
	const SCENE_PAGE_SIZE = 40;
	const GALLERY_PAGE_SIZE = 30;
	// How many titled rows the feed shows between its chunks, and cards per row.
	const SHELF_COUNT = 6;
	const SHELF_SIZE = 12;
	// Placeholder cards while the first page loads.
	const SKELETON_CARDS = 6;
	// Impression surface for recommendation-engine feedback.
	const IMPRESSION_SURFACE: ImpressionSurface = 'home';

	// The query lives in the top bar. It is seeded from the URL so returning here
	// (back button, or Enter from another page) restores that search.
	const initialQuery = $page.url.searchParams.get('q');
	if (initialQuery !== null) searchQuery.set(initialQuery);
	let query = $derived($searchQuery);
	let selectedPlugins = $state<string[]>([]);
	let isSearching = $derived(query.trim().length > 0);
	// The kind of content the search asks for (?content=, e.g. from a cam room's
	// "Past shows"), null for anything.
	let contentKind = $state<ContentKind | null>(
		parseContentKind($page.url.searchParams.get('content'))
	);

	// Mirror the query and kind into the URL (shallow, no navigation) so they
	// survive a round-trip through a scene page. Compared with window.location,
	// not $page.url: shallow replaceState leaves page.url as it was, and reading
	// $page here would re-run this effect on every replaceState.
	$effect(() => {
		if (!browser) return;
		const url = new URL(window.location.href);
		const trimmed = query.trim();
		if (trimmed) url.searchParams.set('q', trimmed);
		else url.searchParams.delete('q');
		if (trimmed && contentKind) url.searchParams.set('content', contentKind);
		else url.searchParams.delete('content');
		if (url.href !== window.location.href) replaceState(url, {});
	});

	// The search on screen while there is a query (HomeSearch renders it).
	const liveSearch = new LiveSearch();

	// The grid only ever grows while a feed is showing; see appendItems.
	let gridItems = $state.raw<GridItem[]>([]);
	let shownKeys = new Set<string>();
	let loading = $state(false);
	let sceneOffset = 0;
	let galleryOffset = 0;
	let sceneFeedDone = $state(false);
	let galleryFeedDone = false;
	let sentinel = $state<HTMLElement | null>(null);

	// Candidate source of each recommended scene, keyed by scene id, for the
	// impressions sent back to the engine. Empty while searching.
	let recommendationSources = new Map<string, string>();
	// Place of each recommended scene in the feed (0-based, across pages).
	let recommendationPositions = new Map<string, number>();
	// When the engine ranked the feed on screen; the cache expires from here.
	let feedBuiltAt = 0;
	// Set by the refresh button: the next first page asks the engine to re-rank
	// instead of serving its recent ranking.
	let refreshRequested = false;

	// The preferred tag order, for the taste badge on each card.
	let preferredTags = $state<string[]>([]);
	$effect(() => {
		fetchPreferredTags().then((tags) => (preferredTags = tags));
	});

	// Sites, to label each card and to apply the site filter to recommendations.
	let sites = $state<SearchPlugin[]>([]);
	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	// The selected sites that offer the chosen media type; none left = all sites.
	let activeSources = $derived.by(() => {
		if (sites.length === 0) return selectedPlugins;
		return selectedPlugins.filter((name) => {
			const site = sites.find((plugin) => plugin.name === name);
			return site !== undefined && supportsMediaFilter(site, $mediaFilter);
		});
	});

	// The sites a search can ask: enabled and offering the media type shown.
	let searchableSites = $derived(
		sites.filter((site) => site.enabled && supportsMediaFilter(site, $mediaFilter))
	);

	// The sites a search is limited to: the picked ones carrying the kind asked
	// for, else every site carrying it; without a kind, the picked ones (none = all).
	let searchSources = $derived.by(() => {
		if (contentKind === null) return activeSources;
		const carrying = searchableSites.filter((site) => carriesContent(site, contentKind));
		const picked = carrying.filter((site) => activeSources.includes(site.name));
		if (picked.length > 0) return picked.map((site) => site.name);
		return carrying.map((site) => site.name);
	});

	let requestToken = 0;

	/** Whether a card has something to show: galleries without a cover would be empty boxes. */
	function hasPoster(item: GridItem): boolean {
		if (item.kind === 'gallery') return Boolean(item.gallery.coverUrl);
		return true;
	}

	/** Appends items that aren't shown yet; shown items never move or disappear. */
	function appendItems(items: GridItem[]) {
		const fresh = items.filter((item) => hasPoster(item) && !shownKeys.has(gridItemKey(item)));
		if (fresh.length === 0) return;
		for (const item of fresh) shownKeys.add(gridItemKey(item));
		gridItems = [...gridItems, ...fresh];
	}

	/** Clears the grid for a new search, source or media selection. */
	function resetResults() {
		gridItems = [];
		shownKeys = new Set();
		recommendationSources = new Map();
		recommendationPositions = new Map();
		resetShownImpressions();
		sceneOffset = 0;
		galleryOffset = 0;
		sceneFeedDone = $mediaFilter === 'images';
		galleryFeedDone = $mediaFilter === 'videos';
	}

	/**
	 * Interleaves the feed's scenes and galleries so neither clumps together, and
	 * mixes the galleries' sites. The recommended scenes keep their ranking, which
	 * already caps how often a site repeats.
	 */
	function interleave(scenes: SceneHit[], galleries: GalleryHit[]): GridItem[] {
		galleries = mixBySite(galleries, (gallery) => gridItemSiteKey({ kind: 'gallery', gallery }));
		const merged: GridItem[] = [];
		const longest = Math.max(scenes.length, galleries.length);
		for (let index = 0; index < longest; index++) {
			if (index < scenes.length) merged.push({ kind: 'scene', scene: scenes[index] });
			if (index < galleries.length) merged.push({ kind: 'gallery', gallery: galleries[index] });
		}
		return merged;
	}

	/** Next page of the recommendation feed. */
	async function nextScenePage(): Promise<SceneHit[]> {
		if (sceneFeedDone) return [];
		const refresh = refreshRequested && sceneOffset === 0;
		refreshRequested = false;
		const recommended = await fetchRecommendedSceneHits(
			SCENE_PAGE_SIZE,
			sceneOffset,
			refresh,
			activeSources
		);
		recommended.forEach((entry, index) => {
			recommendationSources.set(entry.scene.id, entry.source);
			recommendationPositions.set(entry.scene.id, sceneOffset + index);
		});
		sceneOffset += recommended.length;
		sceneFeedDone = recommended.length < SCENE_PAGE_SIZE;
		return recommended.map((entry) => entry.scene);
	}

	/** Next page of the feed's galleries, newest first. */
	async function nextGalleryPage(): Promise<GalleryHit[]> {
		if (galleryFeedDone) return [];
		const galleries = await searchGalleries('', activeSources, GALLERY_PAGE_SIZE, galleryOffset);
		galleryOffset += galleries.length;
		galleryFeedDone = galleries.length < GALLERY_PAGE_SIZE;
		return galleries;
	}

	/** The element that scrolls the page (the layout's main column). */
	function scrollContainer(): HTMLElement | null {
		if (!sentinel) return null;
		return sentinel.closest('main');
	}

	/** Brings back a cached feed and its scroll position; false when there is none. */
	async function restoreCachedFeed(token: number): Promise<boolean> {
		const cached = loadHomeFeed(homeFeedKey($mediaFilter, activeSources));
		if (!cached) return false;
		feedBuiltAt = cached.savedAt;
		recommendationSources = new Map(cached.recommendationSources);
		recommendationPositions = new Map(cached.recommendationPositions);
		sceneOffset = cached.sceneOffset;
		galleryOffset = cached.galleryOffset;
		sceneFeedDone = cached.sceneFeedDone;
		galleryFeedDone = cached.galleryFeedDone;
		appendItems(cached.gridItems);
		loading = false;
		// Wait for the grid to lay the cards out before scrolling back.
		await tick();
		requestAnimationFrame(() => {
			if (token !== requestToken) return;
			const container = scrollContainer();
			if (container) container.scrollTop = cached.scrollTop;
		});
		return true;
	}

	// Remember the recommendation feed when leaving, so coming back restores it.
	beforeNavigate(() => {
		if (isSearching || gridItems.length === 0) return;
		saveHomeFeed({
			key: homeFeedKey($mediaFilter, activeSources),
			savedAt: feedBuiltAt,
			gridItems,
			sceneOffset,
			galleryOffset,
			sceneFeedDone,
			galleryFeedDone,
			recommendationSources,
			recommendationPositions,
			scrollTop: scrollContainer()?.scrollTop ?? 0
		});
	});

	/** Throws away the cached feed and asks the engine to rank again. */
	async function refreshRecommendations() {
		clearHomeFeed();
		refreshRequested = true;
		scrollContainer()?.scrollTo({ top: 0 });
		await runSearch();
	}

	/** Pull-to-refresh: re-ranks the feed, or re-runs the search while searching. */
	async function pullRefresh() {
		if (isSearching) {
			await runSearch();
			return;
		}
		await refreshRecommendations();
	}

	/** Runs the search (HomeSearch shows it), or shows the recommendation feed without a query. */
	async function runSearch() {
		const token = ++requestToken;
		resetResults();
		if ($mediaFilter === 'faces') {
			liveSearch.stop();
			loading = false;
			return;
		}
		const text = query.trim();
		if (text) {
			loading = false;
			liveSearch.run(text, searchSources, searchedSiteNames(), contentKind);
			return;
		}
		liveSearch.stop();
		loading = true;
		loadFeedExtras();
		if (await restoreCachedFeed(token)) return;
		feedBuiltAt = Date.now();

		const [galleryHits, sceneHits] = await Promise.all([nextGalleryPage(), nextScenePage()]);
		if (token !== requestToken) return;

		appendItems(interleave(sceneHits, galleryHits));
		loading = false;
	}

	/** The sites a search asks, for their status in the chips: the picked ones, else every shown one. */
	function searchedSiteNames(): string[] {
		if (searchSources.length > 0) return searchSources;
		return searchableSites.map((site) => site.name);
	}

	// Stop the live search when leaving Home.
	$effect(() => {
		return () => liveSearch.stop();
	});

	// Debounce query/source/media changes into one search. The first run is
	// immediate so a cached feed is back before the page paints empty.
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;
	let firstRun = true;
	$effect(() => {
		// Referenced so the effect re-runs on any of these.
		query;
		activeSources;
		searchSources;
		$mediaFilter;
		if (debounceTimer) clearTimeout(debounceTimer);
		if (firstRun) {
			firstRun = false;
			untrack(() => runSearch());
			return;
		}
		debounceTimer = setTimeout(() => runSearch(), SEARCH_DEBOUNCE_MS);
		return () => {
			if (debounceTimer) clearTimeout(debounceTimer);
		};
	});

	/** Infinite scroll: appends the next page of the feed's scenes and galleries. */
	async function loadMore() {
		if (isSearching || loading || (sceneFeedDone && galleryFeedDone)) return;
		const token = requestToken;
		loading = true;
		const [galleryPage, scenePage] = await Promise.all([nextGalleryPage(), nextScenePage()]);
		if (token !== requestToken) return;
		appendItems(interleave(scenePage, galleryPage));
		loading = false;
	}

	$effect(() => {
		if (!sentinel) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) loadMore();
			},
			{ rootMargin: '800px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	// Each site's search state for the chips, only while searching.
	let chipStatus = $derived.by(() => {
		if (!isSearching) return undefined;
		return liveSearch.sites;
	});

	/** The card's taste badge: its tag that ranks highest in the user's preferences. */
	function badgeFor(tags: string[]): string | null {
		return bestPreferredTag(tags, preferredTags);
	}

	// The feed's scenes and galleries, apart: the scenes fill the grid, the
	// galleries the shelves between its chunks.
	let feedScenes = $derived(
		gridItems.flatMap((item) => {
			if (item.kind === 'scene') return [item.scene];
			return [];
		})
	);
	let feedGalleries = $derived(
		gridItems.flatMap((item) => {
			if (item.kind === 'gallery') return [item.gallery];
			return [];
		})
	);

	// The engine's titled rows and the unfinished watches, loaded with the feed.
	let shelves = $state<RecommendationShelf[]>([]);
	let continueScenes = $state<SceneHit[]>([]);

	/** Loads the shelves and Continue watching for the feed. */
	async function loadFeedExtras() {
		const [rows, unfinished] = await Promise.all([
			fetchRecommendationShelves(SHELF_COUNT, SHELF_SIZE),
			fetchContinueWatching(SHELF_SIZE)
		]);
		shelves = rows;
		continueScenes = unfinished;
	}

	/** The impression for a recommended scene card; empty id (skipped) while searching. */
	function impressionFor(scene: SceneHit): ShownEvent {
		const source = recommendationSources.get(scene.id);
		const position = recommendationPositions.get(scene.id);
		if (source === undefined || position === undefined) {
			return { mediaId: '', surface: IMPRESSION_SURFACE, position: 0 };
		}
		return { mediaId: scene.id, source, surface: IMPRESSION_SURFACE, position };
	}

	/** Opens a scene, telling the engine when it was a recommendation. */
	function openScene(scene: SceneHit) {
		const event = impressionFor(scene);
		if (event.mediaId) recordClicked(event);
		goto(sceneUrl(scene.id));
	}
</script>

<div class="flex flex-col gap-6 max-md:gap-4">
	{#if $isCompact}
		<PullToRefresh onrefresh={pullRefresh} />
	{/if}
	{#if $mediaFilter === 'faces'}
		{#if $isCompact}
			<div class="flex">
				<MediaFilterChip />
			</div>
		{/if}
		<FaceSearch />
	{:else}
		<SiteChips bind:selectedPlugins status={chipStatus} content={isSearching ? contentKind : null}>
			{#snippet leading()}
				{#if $isCompact}
					<MediaFilterChip />
				{/if}
			{/snippet}
		</SiteChips>
	{/if}

	{#if $mediaFilter === 'faces'}
		<!-- The face search above replaces the feed and the live search. -->
	{:else if isSearching}
		<HomeSearch
			search={liveSearch}
			{sites}
			sources={searchSources}
			{searchableSites}
			content={contentKind}
			oncontent={(kind) => (contentKind = kind)}
		/>
	{:else}
		<!-- The recommendation feed (HomeFeed lays it out). -->
		<section class="flex flex-col gap-3">
			{#if !$isCompact}
				<!-- Phones refresh by pulling down instead. -->
				<div class="flex items-center justify-between gap-3">
					<h2 class="text-lg font-semibold">For you</h2>
					<button
						type="button"
						onclick={refreshRecommendations}
						disabled={loading}
						class="bg-base-200 hover:bg-base-300 flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-colors disabled:opacity-50"
					>
						<ArrowClockwiseIcon size={16} weight="bold" />
						Refresh
					</button>
				</div>
			{/if}
			{#if gridItems.length === 0 && loading}
				<div class="skeleton-grid grid gap-6">
					{#each { length: SKELETON_CARDS } as _, index (index)}
						<div class="flex flex-col gap-3">
							<div class="skeleton aspect-video w-full rounded-[var(--tv-radius)]"></div>
							<div class="flex gap-3">
								<div class="skeleton h-9 w-9 shrink-0 rounded-full"></div>
								<div class="flex flex-1 flex-col gap-2 pt-1">
									<div class="skeleton h-3.5 w-11/12 rounded"></div>
									<div class="skeleton h-3 w-1/2 rounded"></div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{:else if gridItems.length === 0 && continueScenes.length === 0 && !loading}
				<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
					<MagnifyingGlassIcon size={48} class="opacity-20" />
					<p>Nothing here yet. Search for something to get started.</p>
				</div>
			{:else}
				<HomeFeed
					scenes={feedScenes}
					galleries={feedGalleries}
					{shelves}
					{continueScenes}
					{sites}
					scenesDone={sceneFeedDone}
					galleriesOnly={$mediaFilter === 'images'}
					{impressionFor}
					{badgeFor}
					onopen={openScene}
				/>
			{/if}

			{#if loading && gridItems.length > 0}
				<div class="flex justify-center py-6">
					<span class="loading loading-spinner text-base-content/40"></span>
				</div>
			{/if}
			<div bind:this={sentinel} class="h-1"></div>
		</section>
	{/if}
</div>

<style>
	/* Same column sizing as the feed grid, so the feed doesn't jump in. */
	.skeleton-grid {
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	}
</style>
