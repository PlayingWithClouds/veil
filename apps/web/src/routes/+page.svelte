<script lang="ts">
	import { browser } from '$app/environment';
	import { beforeNavigate, goto, replaceState } from '$app/navigation';
	import { page } from '$app/stores';
	import { tick, untrack } from 'svelte';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import { clearHomeFeed, homeFeedKey, loadHomeFeed, saveHomeFeed } from '$lib/homeFeedCache';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import GalleryCard, { GALLERY_COVER_ASPECT } from '$lib/components/GalleryCard.svelte';
	import { CARD_CAPTION_HEIGHT } from '$lib/components/CardCaption.svelte';
	import Masonry from '$lib/components/Masonry.svelte';
	import PerformerCard from '$lib/components/PerformerCard.svelte';
	import SiteChips from '$lib/components/SiteChips.svelte';
	import MediaFilterChip from '$lib/components/MediaFilterChip.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import { searchQuery } from '$lib/stores/searchQuery';
	import { mediaFilter } from '$lib/stores/mediaFilter';
	import SubscribeSearchButton from '$lib/components/SubscribeSearchButton.svelte';
	import { sceneUrl, galleryUrl, performerUrl, studioUrl } from '$lib/routes';
	import { cacheUrl } from '$lib/img';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { impression, recordClicked, resetShownImpressions } from '$lib/recommendations';
	import {
		searchPerformers,
		searchStudios,
		searchGalleries,
		searchScenes,
		fetchRecommendedSceneHits,
		fetchPreferredTags,
		bestPreferredTag,
		pluginSearchStream,
		streamResultToGridItem,
		gridItemKey,
		type SearchStream,
		type PerformerHit,
		type StudioHit,
		type GalleryHit,
		type SceneHit,
		type GridItem
	} from '$lib/explore';

	const SEARCH_DEBOUNCE_MS = 300;
	const SCENE_PAGE_SIZE = 40;
	const GALLERY_PAGE_SIZE = 30;
	const SCENE_ASPECT = 16 / 9;
	// Placeholder cards while the first page loads.
	const SKELETON_CARDS = 6;
	// Impression surface for recommendation-engine feedback.
	const IMPRESSION_SURFACE = 'home';

	// The query lives in the top bar. It is seeded from the URL so returning here
	// (back button, or Enter from another page) restores that search.
	const initialQuery = $page.url.searchParams.get('q');
	if (initialQuery !== null) searchQuery.set(initialQuery);
	let query = $derived($searchQuery);
	let selectedPlugins = $state<string[]>([]);
	let isSearching = $derived(query.trim().length > 0);

	// Mirror the query into the URL (shallow, no navigation) so it survives a
	// round-trip through a scene page.
	$effect(() => {
		if (!browser) return;
		const url = new URL($page.url);
		const trimmed = query.trim();
		if (trimmed) url.searchParams.set('q', trimmed);
		else url.searchParams.delete('q');
		if (url.href !== $page.url.href) replaceState(url, {});
	});

	let performers = $state<PerformerHit[]>([]);
	let studios = $state<StudioHit[]>([]);
	// The grid only ever grows while a feed is showing; see appendItems.
	let gridItems = $state.raw<GridItem[]>([]);
	let shownKeys = new Set<string>();
	let loading = $state(false);
	let pluginsSearching = $state(false);
	let sceneOffset = 0;
	let galleryOffset = 0;
	let sceneFeedDone = false;
	let galleryFeedDone = false;
	let sentinel = $state<HTMLElement | null>(null);

	// Candidate source of each recommended scene, keyed by scene id, for the
	// impressions sent back to the engine. Empty while searching.
	let recommendationSources = new Map<string, string>();
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

	let requestToken = 0;
	let seedStream: SearchStream | null = null;

	/** Appends items that aren't shown yet; shown items never move or disappear. */
	function appendItems(items: GridItem[]) {
		const fresh = items.filter((item) => !shownKeys.has(gridItemKey(item)));
		if (fresh.length === 0) return;
		for (const item of fresh) shownKeys.add(gridItemKey(item));
		gridItems = [...gridItems, ...fresh];
	}

	/** Clears the grid for a new search, source or media selection. */
	function resetResults() {
		if (seedStream) {
			seedStream.close();
			seedStream = null;
		}
		pluginsSearching = false;
		gridItems = [];
		shownKeys = new Set();
		recommendationSources = new Map();
		resetShownImpressions();
		sceneOffset = 0;
		galleryOffset = 0;
		sceneFeedDone = $mediaFilter === 'images';
		galleryFeedDone = $mediaFilter === 'videos';
	}

	/** Interleaves scenes and galleries so neither clumps together. */
	function interleave(scenes: SceneHit[], galleries: GalleryHit[]): GridItem[] {
		const merged: GridItem[] = [];
		const longest = Math.max(scenes.length, galleries.length);
		for (let index = 0; index < longest; index++) {
			if (index < scenes.length) merged.push({ kind: 'scene', scene: scenes[index] });
			if (index < galleries.length) merged.push({ kind: 'gallery', gallery: galleries[index] });
		}
		return merged;
	}

	/** Whether a recommended scene came from one of the selected sites (none selected = all). */
	function matchesSelectedSites(scene: SceneHit): boolean {
		if (selectedPlugins.length === 0) return true;
		const site = pluginForUrl(scene.sourceUrl, sites);
		if (!site) return false;
		return selectedPlugins.includes(site.name);
	}

	/** Next page of scenes: search matches, or the recommendation feed when there is no query. */
	async function nextScenePage(text: string): Promise<SceneHit[]> {
		if (sceneFeedDone) return [];
		if (text) {
			const scenes = await searchScenes(text, selectedPlugins, SCENE_PAGE_SIZE, sceneOffset);
			sceneOffset += scenes.length;
			sceneFeedDone = scenes.length < SCENE_PAGE_SIZE;
			return scenes;
		}
		const refresh = refreshRequested && sceneOffset === 0;
		refreshRequested = false;
		const recommended = await fetchRecommendedSceneHits(SCENE_PAGE_SIZE, sceneOffset, refresh);
		sceneOffset += recommended.length;
		sceneFeedDone = recommended.length < SCENE_PAGE_SIZE;
		for (const entry of recommended) recommendationSources.set(entry.scene.id, entry.source);
		return recommended.map((entry) => entry.scene).filter(matchesSelectedSites);
	}

	/** Next page of galleries matching the query (newest first when there is none). */
	async function nextGalleryPage(text: string): Promise<GalleryHit[]> {
		if (galleryFeedDone) return [];
		const galleries = await searchGalleries(text, selectedPlugins, GALLERY_PAGE_SIZE, galleryOffset);
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
		const cached = loadHomeFeed(homeFeedKey($mediaFilter, selectedPlugins));
		if (!cached) return false;
		performers = [];
		studios = [];
		feedBuiltAt = cached.savedAt;
		recommendationSources = new Map(cached.recommendationSources);
		sceneOffset = cached.sceneOffset;
		galleryOffset = cached.galleryOffset;
		sceneFeedDone = cached.sceneFeedDone;
		galleryFeedDone = cached.galleryFeedDone;
		appendItems(cached.gridItems);
		loading = false;
		// Wait for the masonry to lay the cards out before scrolling back.
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
			key: homeFeedKey($mediaFilter, selectedPlugins),
			savedAt: feedBuiltAt,
			gridItems,
			sceneOffset,
			galleryOffset,
			sceneFeedDone,
			galleryFeedDone,
			recommendationSources,
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

	/** Shows stored matches (or recommendations) first, then streams in live results from the sites. */
	async function runSearch() {
		const token = ++requestToken;
		resetResults();
		loading = true;
		const text = query.trim();
		if (!text && (await restoreCachedFeed(token))) return;
		feedBuiltAt = Date.now();

		const [performerHits, studioHits, galleryHits, sceneHits] = await Promise.all([
			searchPerformers(text),
			searchStudios(text),
			nextGalleryPage(text),
			nextScenePage(text)
		]);
		if (token !== requestToken) return;

		performers = performerHits;
		studios = studioHits;
		appendItems(interleave(sceneHits, galleryHits));
		loading = false;

		if (text) streamFromPlugins(text, selectedPlugins, token);
	}

	/** Whether the media toggle lets this grid item through. */
	function allowedByMediaFilter(item: GridItem): boolean {
		if (item.kind === 'scene') return $mediaFilter !== 'images';
		return $mediaFilter !== 'videos';
	}

	/** Appends each live plugin result as it arrives. */
	function streamFromPlugins(text: string, sources: string[], token: number) {
		pluginsSearching = true;
		seedStream = pluginSearchStream(
			text,
			sources,
			(result) => {
				// DB hits are already covered by the source-filtered listing above.
				if (token !== requestToken || !result.plugin) return;
				const item = streamResultToGridItem(result);
				if (item && allowedByMediaFilter(item)) appendItems([item]);
			},
			() => {
				if (token !== requestToken) return;
				seedStream = null;
				pluginsSearching = false;
			}
		);
	}

	// Debounce query/source/media changes into one search. The first run is
	// immediate so a cached feed is back before the page paints empty.
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;
	let firstRun = true;
	$effect(() => {
		// Referenced so the effect re-runs on any of these.
		query;
		selectedPlugins;
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

	/** Infinite scroll: appends the next page of scenes and galleries. */
	async function loadMore() {
		if (loading || (sceneFeedDone && galleryFeedDone)) return;
		const token = requestToken;
		loading = true;
		const text = query.trim();
		const [galleryPage, scenePage] = await Promise.all([nextGalleryPage(text), nextScenePage(text)]);
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

	/** The card's taste badge: its tag that ranks highest in the user's preferences. */
	function badgeFor(tags: string[]): string | null {
		return bestPreferredTag(tags, preferredTags);
	}

	/** Width / height of a grid item's media. */
	function aspectOf(item: GridItem): number {
		if (item.kind === 'scene') return SCENE_ASPECT;
		return GALLERY_COVER_ASPECT;
	}

	/** The site a record came from, by its page URL. */
	function siteOf(sourceUrl: string): SearchPlugin | null {
		return pluginForUrl(sourceUrl, sites);
	}

	/** The impression for a recommended scene card; empty id (skipped) while searching. */
	function impressionFor(scene: SceneHit) {
		const source = recommendationSources.get(scene.id);
		if (source === undefined) return { mediaId: '' };
		return { mediaId: scene.id, source, surface: IMPRESSION_SURFACE };
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
	<SiteChips bind:selectedPlugins>
		{#snippet leading()}
			{#if $isCompact}
				<MediaFilterChip />
			{/if}
		{/snippet}
	</SiteChips>

	<!-- Performers row. -->
	{#if performers.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Performers</h2>
			<div class="no-scrollbar flex gap-4 overflow-x-auto pb-1">
				{#each performers as performer (performer.id)}
					<div class="w-28 shrink-0">
						<PerformerCard {performer} onclick={() => goto(performerUrl(performer.id))} />
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Studios row. -->
	{#if studios.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Studios</h2>
			<div class="no-scrollbar flex gap-4 overflow-x-auto pb-1">
				{#each studios as studio (studio.id)}
					<button
						type="button"
						class="flex w-32 shrink-0 flex-col items-center gap-2 text-center"
						onclick={() => goto(studioUrl(studio.id))}
					>
						<div class="tv-card bg-base-300 flex aspect-video w-full items-center justify-center overflow-hidden rounded-lg">
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
						</div>
						<span class="line-clamp-1 text-xs font-medium">{studio.name}</span>
					</button>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Galleries + scenes, left-to-right masonry. -->
	<section class="flex flex-col gap-3">
		{#if isSearching}
			<div class="flex items-center justify-between gap-3">
				<h2 class="flex items-center gap-3 text-lg font-semibold">
					Results
					{#if pluginsSearching}
						<span class="text-base-content/40 flex items-center gap-2 text-sm font-normal">
							<span class="loading loading-spinner loading-xs"></span>
							Searching sites…
						</span>
					{/if}
				</h2>
				<SubscribeSearchButton {query} sources={selectedPlugins} />
			</div>
		{:else if !$isCompact}
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
		{:else if gridItems.length === 0 && !loading && !pluginsSearching}
			<div class="text-base-content/30 flex flex-col items-center gap-3 py-24 text-center">
				<MagnifyingGlassIcon size={48} class="opacity-20" />
				<p>{isSearching ? 'Nothing matches your search.' : 'Nothing here yet. Search for something to get started.'}</p>
			</div>
		{:else}
			<Masonry items={gridItems} keyOf={gridItemKey} {aspectOf} captionHeight={CARD_CAPTION_HEIGHT} gap={24}>
				{#snippet children(item: GridItem)}
					{#if item.kind === 'scene'}
						<div use:impression={impressionFor(item.scene)}>
							<ExploreSceneCard
								scene={item.scene}
								site={siteOf(item.scene.sourceUrl)}
								preferredTag={badgeFor(item.scene.tags)}
								onclick={() => openScene(item.scene)}
							/>
						</div>
					{:else}
						<GalleryCard
							gallery={item.gallery}
							site={siteOf(item.gallery.sourceUrl)}
							preferredTag={badgeFor(item.gallery.tags)}
							onclick={() => goto(galleryUrl(item.gallery.id))}
						/>
					{/if}
				{/snippet}
			</Masonry>
		{/if}

		{#if loading && gridItems.length > 0}
			<div class="flex justify-center py-6">
				<span class="loading loading-spinner text-base-content/40"></span>
			</div>
		{/if}
		<div bind:this={sentinel} class="h-1"></div>
	</section>
</div>

<style>
	/* Same column sizing as the masonry below, so the feed doesn't jump in. */
	.skeleton-grid {
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	}
</style>
