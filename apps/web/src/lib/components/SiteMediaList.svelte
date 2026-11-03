<!--
	A page of one media kind beyond videos (galleries, audio, stories). With a
	search from the top bar (?q=), what the sites that list the kind answer,
	as they answer. With a tag (?tag=), the library's media carrying it.
	Otherwise the feed's first items in a spotlight (when the page has one),
	the overview's rows, then the rest of an endless feed: the page's own
	(galleries: recommended, then the rest), else everything most recently
	added first, continued by new pages from the sites once the library is
	through. Cards rise in; placeholders stand in while the first load
	runs. "Browse a site" stores a page of one site's catalog.
-->
<script lang="ts" generics="Item extends { id: string }">
	import { untrack, type Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { tagLabel } from '$lib/components/TagChips.svelte';
	import { browse } from '$lib/library';
	import { fetchTagName, type MediaListing } from '$lib/media';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { searchPage, type SearchBatchRow, type SearchPage } from '$lib/search';
	import { resetShown, type Shown } from '$lib/events';
	import { endlessFeed, siteSource, type Feed, type Source } from '$lib/feed';
	import { listerNames } from '$lib/stores/sites';

	const PAGE_SIZE = 40;
	// Feed items in the spotlight, when the page has one.
	const SPOTLIGHT_COUNT = 5;
	// Placeholder cards while the first page loads.
	const SKELETON_COUNT = 12;
	// Cards of one page rise in this far apart, up to RISE_MAX_DELAY_MS.
	const RISE_STEP_MS = 25;
	const RISE_MAX_DELAY_MS = 450;

	interface Props {
		title: string;
		kind: 'GALLERY' | 'AUDIO' | 'STORY';
		// What the page lists, by name ("stories").
		noun: string;
		// A page of the library's items.
		load: (limit: number, offset: number, listing: MediaListing) => Promise<Item[]>;
		// The items of this kind in one site's search answer.
		pick: (batch: SearchBatchRow) => Item[];
		// Smallest card width in px, and the cards' picture shape ("3 / 4").
		cardWidth: number;
		cardAspect?: string;
		// The feed's first items, drawn above the overview's rows.
		spotlight?: Snippet<[Item[]]>;
		// A card, with where it's shown when that's logged (the feed, a search).
		card: Snippet<[Item, Shown | null]>;
		// Rows shown above everything when there's no search or tag.
		overview?: Snippet;
		// Starts the endless feed under the overview, instead of everything most
		// recently added first, then what browsePage brings from the sites.
		feed?: () => Feed<Item>;
		// A page of one site's catalog, stored as it comes, once the library is through.
		browsePage?: (plugin: string, limit: number, offset: number) => Promise<Item[]>;
		// The feed's heading and the surface its cards are logged on.
		feedTitle?: string;
		feedSurface?: string;
	}

	let {
		title,
		kind,
		noun,
		load,
		pick,
		cardWidth,
		cardAspect = '3 / 4',
		spotlight,
		card,
		overview,
		feed,
		browsePage,
		feedTitle = 'Recently added',
		feedSurface = 'library'
	}: Props = $props();

	let items = $state<Item[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let browsePlugin = $state('');
	let browsing = $state(false);
	let sitesAsked = $state(0);
	let sitesAnswered = $state(0);
	let searchPageNumber = $state(0);
	let tagName = $state<string | null>(null);

	// The running search page, the sites that answered and those with nothing more.
	let running: SearchPage | null = null;
	let searchSites = new Set<string>();
	let drySites = new Set<string>();
	// The overview's endless feed while it's shown.
	let activeFeed: Feed<Item> | null = null;
	// Counts loads started, so a late answer to an earlier one is dropped.
	let generation = 0;

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let tagId = $derived(page.url.searchParams.get('tag') ?? '');
	let kindSites = $derived(listers($sites, kind));
	let showOverview = $derived(query === '' && tagId === '');
	// How many of the feed's first items the spotlight takes from the grid.
	let spotlightTaken = $derived.by(() => {
		if (!showOverview || !spotlight) return 0;
		return Math.min(SPOTLIGHT_COUNT, items.length);
	});
	let gridItems = $derived(items.slice(spotlightTaken));

	$effect(() => {
		const searchText = query;
		const tag = tagId;
		untrack(() => void restart(searchText, tag));
		return () => running?.cancel();
	});

	/** Starts over: the library's first page (all, or one tag's), or the search's. */
	async function restart(searchText: string, tag: string) {
		running?.cancel();
		running = null;
		generation++;
		items = [];
		exhausted = false;
		error = null;
		tagName = null;
		activeFeed = null;
		if (tag !== '') void fetchTagName(tag).then((name) => (tagName = name));
		if (searchText === '' && tag === '') {
			resetShown(feedSurface);
			activeFeed = overviewFeed();
		}
		if (searchText === '') {
			await loadMore();
			return;
		}
		searchPageNumber = 0;
		searchSites = new Set();
		drySites = new Set();
		await runSearch(searchText, null);
	}

	/** Appends the next page of the library or of the search. */
	async function loadMore() {
		if (loading || exhausted) return;
		if (query !== '') {
			searchNextPage();
			return;
		}
		loading = true;
		const started = generation;
		try {
			const nextPage = await nextLibraryPage();
			if (started !== generation) return;
			items = [...items, ...nextPage.items];
			exhausted = nextPage.last;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** The feed under the overview: the page's own, else the library most recently added first, then the sites. */
	function overviewFeed(): Feed<Item> {
		if (feed) return feed();
		let sites: Source<Item> | undefined;
		if (browsePage) sites = siteSource(() => listerNames(kind), browsePage);
		return endlessFeed({ rest: (limit, offset) => load(limit, offset, {}), sites });
	}

	/** The next page of the feed, or of everything (one tag's) most recently added first. */
	async function nextLibraryPage(): Promise<{ items: Item[]; last: boolean }> {
		if (activeFeed) {
			const found = await activeFeed(PAGE_SIZE);
			return { items: found, last: found.length === 0 };
		}
		const listing: MediaListing = {};
		if (tagId !== '') listing.tagId = tagId;
		const found = await load(PAGE_SIZE, items.length, listing);
		return { items: found, last: found.length < PAGE_SIZE };
	}

	/** Where the card at position is shown, when the page logs it: the feed or a search. */
	function shownAt(position: number): Shown | null {
		if (query !== '') return { surface: 'search', position };
		if (activeFeed) return { surface: feedSurface, position };
		return null;
	}

	/** How long a card waits before rising in, so a page of cards cascades. */
	function riseDelay(position: number): string {
		return `${Math.min((position % PAGE_SIZE) * RISE_STEP_MS, RISE_MAX_DELAY_MS)}ms`;
	}

	/** Asks the sites with more results for their next page. */
	function searchNextPage() {
		const remaining = [...searchSites].filter((site) => !drySites.has(site));
		if (remaining.length === 0) {
			exhausted = true;
			return;
		}
		searchPageNumber++;
		void runSearch(query, remaining);
	}

	/** Asks the sites (only plugins, when given) for one page of the search. */
	async function runSearch(searchText: string, plugins: string[] | null) {
		loading = true;
		sitesAsked = 0;
		sitesAnswered = 0;
		const started = generation;
		const current = searchPage(kind, searchText, searchPageNumber, plugins, null, null, (batch) => {
			if (started === generation) receive(batch);
		});
		running = current;
		try {
			await current.done;
		} catch (failure) {
			error = String(failure);
		}
		if (running !== current) return;
		running = null;
		loading = false;
		exhausted = [...searchSites].every((site) => drySites.has(site));
	}

	/** Takes part of a site's answer, skipping what's shown already. */
	function receive(batch: SearchBatchRow) {
		sitesAsked = batch.sites;
		searchSites.add(batch.plugin);
		const shown = new Set(items.map((item) => item.id));
		const fresh = pick(batch).filter((item) => !shown.has(item.id));
		items = [...items, ...fresh];
		if (!batch.final) return;
		sitesAnswered++;
		if (batch.exhausted || batch.returned === 0) drySites.add(batch.plugin);
	}

	/** Stores a page of the picked site's catalog, then shows the overview again. */
	async function browseSite() {
		if (!browsePlugin) return;
		browsing = true;
		error = null;
		try {
			await browse(browsePlugin, kind);
			if (!showOverview) {
				goto(page.url.pathname);
				return;
			}
			await restart('', '');
		} catch (failure) {
			error = String(failure);
		} finally {
			browsing = false;
		}
	}
</script>

<svelte:head><title>{title} · Veil</title></svelte:head>

<div class="flex flex-col gap-8">
	<PageHeader {title}>
		{#snippet actions()}
			{#if kindSites.length > 0}
				<select bind:value={browsePlugin} class="chip pr-2" aria-label="Site to browse">
					<option value="">Browse a site…</option>
					{#each kindSites as site (site.name)}
						<option value={site.name}>{site.displayName ?? site.name}</option>
					{/each}
				</select>
				<button
					type="button"
					class="bg-primary text-primary-content h-8 rounded-full px-4 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-50"
					disabled={!browsePlugin || browsing}
					onclick={browseSite}
				>
					{browsing ? 'Loading…' : 'Browse'}
				</button>
			{/if}
		{/snippet}
	</PageHeader>

	{#if !showOverview}
		<div class="-mt-4 flex items-center gap-3">
			<span class="text-base-content/70 text-sm">
				{#if query !== ''}
					Searching the sites for <span class="text-base-content font-medium">“{query}”</span>
				{:else}
					Tagged <span class="text-base-content font-medium">{tagLabel(tagName ?? '…')}</span>
				{/if}
			</span>
			{#if query !== '' && loading && searchPageNumber === 0 && sitesAsked > 0}
				<span class="chip pointer-events-none">
					<span class="loading loading-spinner loading-xs"></span>
					{sitesAnswered} of {sitesAsked} answered
				</span>
			{/if}
			<button type="button" class="chip" onclick={() => goto(page.url.pathname)}>
				<XIcon size={12} /> Clear
			</button>
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if spotlight && spotlightTaken > 0}
		<div class="rise-in">
			{@render spotlight(items.slice(0, spotlightTaken))}
		</div>
	{/if}

	{#if showOverview && overview}
		{@render overview()}
	{/if}

	<section class="flex flex-col gap-4">
		{#if showOverview && gridItems.length > 0}
			<h2 class="text-lg font-semibold">{feedTitle}</h2>
		{/if}
		<div
			class="grid gap-x-4 gap-y-6"
			style:grid-template-columns="repeat(auto-fill, minmax({cardWidth}px, 1fr))"
		>
			{#each gridItems as item, position (item.id)}
				<div class="rise-in" style:--rise-delay={riseDelay(position)}>
					{@render card(item, shownAt(spotlightTaken + position))}
				</div>
			{:else}
				{#if loading && items.length === 0}
					{#each { length: SKELETON_COUNT }, position (position)}
						<div class="flex flex-col gap-2" aria-hidden="true">
							<div class="skeleton w-full rounded-xl" style:aspect-ratio={cardAspect}></div>
							<div class="skeleton h-3.5 w-3/4 rounded"></div>
							<div class="skeleton h-3 w-1/3 rounded"></div>
						</div>
					{/each}
				{/if}
			{/each}
		</div>
	</section>

	{#if !loading && items.length === 0}
		{#if query !== ''}
			<p class="text-base-content/50 text-sm">No site found {noun} for this search.</p>
		{:else if kindSites.length === 0}
			<p class="text-base-content/50 text-sm">
				No {noun} yet, and no installed plugin lists them. Install one on the plugins page.
			</p>
		{:else}
			<p class="text-base-content/50 text-sm">
				No {noun} yet. Search from the top bar or browse a site.
			</p>
		{/if}
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && items.length > 0} done={exhausted} />
</div>
