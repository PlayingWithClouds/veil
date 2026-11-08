<!--
	A search across the sites, shown as it comes in and never reshuffled:
	each site's best four take the top twenty places in the order the sites
	answer, the rest follows sites taking turns (SearchLayout). The first
	five fill a spotlight (the best match in a large frame, lighting the
	shared glow), the rest a grid at most five across; a line on top says
	how many sites answered while they still are. Scrolling on
	asks the sites for their next page until none has more. Every six grid
	rows a row breaks the grid up: performers by that name, picks for you,
	your favourites. A site that answered the same search within six hours
	isn't asked again; the backend sends its answer from then.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import VideoSpotlight from '$lib/components/VideoSpotlight.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import {
		fetchPerformers,
		fetchVideosMatching,
		type MemberRow,
		type VideoCardRow
	} from '$lib/library';
	import { fetchRecommendedVideos } from '$lib/home';
	import { searchPage, sitePreferences, type SearchBatchRow, type SearchPage } from '$lib/search';
	import { SearchLayout, TOP_PLACES } from '$lib/searchLayout';
	import { logEvent, resetShown } from '$lib/events';
	import { pluginForUrl, siteDisplayName } from '$lib/plugins';
	import { sites as installedSites } from '$lib/stores/sites';
	import { videoUrl } from '$lib/routes';
	import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';

	// Grid rows between two breaking rows, after the top places.
	const SECTION_ROWS = 6;
	// The grid's card width, gap (px) and most columns (card-grid), to count its columns.
	const CARD_MIN_WIDTH = 260;
	const GRID_GAP = 16;
	const MAX_COLUMNS = 5;
	// Results in the spotlight, before the grid.
	const SPOTLIGHT_COUNT = 5;
	// Placeholder cards while the sites are asked.
	const SKELETON_COUNT = 10;
	// Cards rise in this far apart, up to RISE_MAX_DELAY_MS.
	const RISE_STEP_MS = 30;
	const RISE_MAX_DELAY_MS = 450;
	// Videos in a breaking row.
	const ROW_VIDEOS = 8;
	const VIDEO_WIDTH = 260;
	const PERFORMER_WIDTH = 140;
	// How long results wait below the fold for slower sites before they're let in.
	const RELEASE_DELAY_MS = 1500;
	// How long results wait for a breaking row still loading before it's skipped.
	const ROW_WAIT_MS = 1500;

	/** A row between two sections of results. */
	interface BreakRow {
		title: string;
		videos: VideoCardRow[];
		performers: MemberRow[];
	}

	interface Props {
		query: string;
		// The header's filter, narrowing the results; null without one.
		filter?: FilterInput | null;
	}

	let { query, filter = null }: Props = $props();

	// The filter of the running search, as it started.
	let activeFilter: FilterInput | null = null;
	let filterActive = false;

	let videos = $state<VideoCardRow[]>([]);
	let rows = $state<(BreakRow | null)[]>([]);
	// Whether the row after each section shows; decided once, when the next section starts.
	let rowShown = $state<boolean[]>([]);
	let sitesAsked = $state(0);
	let sitesAnswered = $state(0);
	let searching = $state(false);
	let finished = $state(false);
	let error = $state<string | null>(null);
	let container = $state<HTMLElement>();
	// Results between two breaking rows: whole grid rows, counted when the search starts.
	let section = $state(SECTION_ROWS * 4);

	let layout = new SearchLayout<VideoCardRow>();
	let current: SearchPage | null = null;
	// Counts searches started, so late answers to an earlier one are dropped.
	let generation = 0;
	let pageNumber = $state(0);
	let sites = new Set<string>();
	let exhausted = new Set<string>();
	let newThisPage = new Map<string, number>();
	let returnedThisPage = new Map<string, number>();
	let releaseTimer: ReturnType<typeof setTimeout> | null = null;
	// Results waiting to be appended, and whether they're being appended now.
	let queued: VideoCardRow[] = [];
	let draining = false;
	// Each breaking row's load, by the section it follows.
	let rowLoads = new Map<number, Promise<void>>();
	let recommendedPool: Promise<VideoCardRow[]> | null = null;
	let favoritePool: Promise<VideoCardRow[]> | null = null;

	let sections = $derived(splitSections(videos));
	let spotlight = $derived(
		videos.slice(0, SPOTLIGHT_COUNT).map((video) => ({ video, reason: siteOf(video) }))
	);

	$effect(() => {
		const searchText = query;
		const searchFilter = filter;
		untrack(() => void start(searchText, searchFilter));
		return stop;
	});

	/** The results cut into the top places and sections. */
	function splitSections(all: VideoCardRow[]): VideoCardRow[][] {
		const result = [all.slice(0, TOP_PLACES)];
		for (let start = TOP_PLACES; start < all.length; start += section) {
			result.push(all.slice(start, start + section));
		}
		return result;
	}

	/** Where section index starts in the results. */
	function sectionStart(index: number): number {
		if (index === 0) return 0;
		return TOP_PLACES + (index - 1) * section;
	}

	/** SECTION_ROWS full rows of the grid at its current width. */
	function sectionSize(): number {
		const width = container?.clientWidth ?? 0;
		const fitting = Math.floor((width + GRID_GAP) / (CARD_MIN_WIDTH + GRID_GAP));
		const columns = Math.min(MAX_COLUMNS, Math.max(1, fitting));
		return columns * SECTION_ROWS;
	}

	/** "On xVideos": the site a result came from, for the spotlight. */
	function siteOf(video: VideoCardRow): string | null {
		const sourceUrl = video.sources[0]?.url;
		if (!sourceUrl) return null;
		const plugin = pluginForUrl(sourceUrl, $installedSites);
		if (!plugin) return null;
		return `On ${siteDisplayName(plugin.name, $installedSites)}`;
	}

	/** How long a card waits before rising in, so arriving results cascade. */
	function riseDelay(offset: number): string {
		return `${Math.min((offset % 10) * RISE_STEP_MS, RISE_MAX_DELAY_MS)}ms`;
	}

	/** Stops listening to the running page and the release timer. */
	function stop() {
		current?.cancel();
		current = null;
		if (releaseTimer) clearTimeout(releaseTimer);
		releaseTimer = null;
	}

	/** Starts a search over: empty results, page 0. */
	async function start(searchText: string, searchFilter: FilterInput | null) {
		activeFilter = searchFilter;
		filterActive = searchFilter !== null && Object.keys(searchFilter).length > 0;
		stop();
		generation++;
		resetShown('search');
		logEvent({ kind: 'search', data: { text: searchText } });
		const started = generation;
		layout = new SearchLayout<VideoCardRow>();
		// The sites are asked at once; their preference applies from when it arrives.
		void sitePreferences().then((preference) => {
			if (started === generation) layout.prefer(preference);
		});
		section = sectionSize();
		videos = [];
		rows = [];
		rowShown = [];
		rowLoads = new Map();
		queued = [];
		draining = false;
		pageNumber = 0;
		sites = new Set();
		exhausted = new Set();
		recommendedPool = null;
		favoritePool = null;
		finished = false;
		error = null;
		prepareRow(0);
		prepareRow(1);
		await runPage(searchText, null);
	}

	/** Asks the sites for one page and shows what they answer. */
	async function runPage(searchText: string, plugins: string[] | null) {
		searching = true;
		sitesAsked = 0;
		sitesAnswered = 0;
		newThisPage = new Map();
		returnedThisPage = new Map();
		const running = searchPage(
			'VIDEO',
			searchText,
			pageNumber,
			plugins,
			activeFilter,
			null,
			receive
		);
		current = running;
		try {
			await running.done;
		} catch (failure) {
			error = String(failure);
		}
		if (current !== running) return;
		current = null;
		show(layout.release());
		searching = false;
		finished = remainingSites().length === 0;
	}

	/** Takes part of a site's answer. */
	function receive(batch: SearchBatchRow) {
		sitesAsked = batch.sites;
		sites.add(batch.plugin);
		const fresh = batch.listed.videos.filter((video) => !layout.has(video.id));
		newThisPage.set(batch.plugin, (newThisPage.get(batch.plugin) ?? 0) + fresh.length);
		returnedThisPage.set(batch.plugin, (returnedThisPage.get(batch.plugin) ?? 0) + batch.returned);
		show(layout.add(batch.plugin, batch.listed.videos));
		if (layout.hasWaiting) scheduleRelease();
		if (!batch.final) return;
		sitesAnswered++;
		if (batch.exhausted || siteRanDry(batch.plugin)) exhausted.add(batch.plugin);
	}

	/**
	 * Whether a site's page brought nothing: no results at all, or (without a
	 * filter, which can hide a whole page) only ones already shown.
	 */
	function siteRanDry(site: string): boolean {
		if (returnedThisPage.get(site) === 0) return true;
		return !filterActive && newThisPage.get(site) === 0;
	}

	/** Lets waiting results in after a moment, once the top places are settled. */
	function scheduleRelease() {
		if (releaseTimer || videos.length < TOP_PLACES) return;
		releaseTimer = setTimeout(() => {
			releaseTimer = null;
			show(layout.release());
		}, RELEASE_DELAY_MS);
	}

	/** Queues results to append, in order. */
	function show(items: VideoCardRow[]) {
		queued.push(...items);
		void drain();
	}

	/**
	 * Appends queued results, settling each breaking row as the section after
	 * it starts. A row still loading holds the results back a moment, so it
	 * isn't skipped just because a stored answer came in faster.
	 */
	async function drain() {
		if (draining) return;
		draining = true;
		const startedFor = generation;
		while (queued.length > 0) {
			const index = sectionIndexStartingAt(videos.length);
			if (index > 0 && rowShown[index - 1] === undefined) {
				await Promise.race([rowLoads.get(index - 1), wait(ROW_WAIT_MS)]);
				if (startedFor !== generation) return;
				settleRow(index - 1);
			}
			const next = queued.shift();
			if (next) videos.push(next);
		}
		draining = false;
	}

	/** Resolves after ms. */
	function wait(ms: number): Promise<void> {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	/** The section starting at position, or -1 when none does. */
	function sectionIndexStartingAt(position: number): number {
		if (position < TOP_PLACES) return position === 0 ? 0 : -1;
		if ((position - TOP_PLACES) % section !== 0) return -1;
		return (position - TOP_PLACES) / section + 1;
	}

	/** Decides whether row index shows (only when ready by now) and prepares those after. */
	function settleRow(index: number) {
		const row = rows[index];
		rowShown[index] = Boolean(row && (row.videos.length > 0 || row.performers.length > 0));
		prepareRow(index + 1);
		prepareRow(index + 2);
	}

	/** Starts loading the row after section index, ahead of it being needed. */
	function prepareRow(index: number) {
		if (rowLoads.has(index)) return;
		rows[index] = null;
		const startedFor = generation;
		rowLoads.set(
			index,
			loadRow(index).then((row) => {
				// A row loaded for an earlier search stays out of this one.
				if (startedFor === generation) rows[index] = row;
			})
		);
	}

	/**
	 * The row after section index: performers first, then picks and every
	 * third your favourites; with no favourites left, picks again.
	 */
	async function loadRow(index: number): Promise<BreakRow> {
		if (index === 0) {
			const performers = await fetchPerformers(query, 12);
			return { title: 'Performers', videos: [], performers };
		}
		if (index % 3 === 2) {
			favoritePool ??= fetchVideosMatching({ text: query, favorite: true }, 60, 0);
			const favorites = takeUnseen(await favoritePool);
			if (favorites.length > 0)
				return { title: 'Your favourites', videos: favorites, performers: [] };
		}
		recommendedPool ??= fetchPicks();
		return { title: 'Picked for you', videos: takeUnseen(await recommendedPool), performers: [] };
	}

	/**
	 * Picks for the rows: those matching the search first, then picks of any
	 * kind, since the matching ones soon all are results themselves.
	 */
	async function fetchPicks(): Promise<VideoCardRow[]> {
		const [matching, general] = await Promise.all([
			fetchRecommendedVideos(60, { text: query }),
			fetchRecommendedVideos(120)
		]);
		return [...matching, ...general].map((pick) => pick.video);
	}

	/** The next few videos of a pool that aren't on the page yet, now counted as shown. */
	function takeUnseen(pool: VideoCardRow[]): VideoCardRow[] {
		const taken = pool.filter((video) => !layout.has(video.id)).slice(0, ROW_VIDEOS);
		for (const video of taken) layout.claim(video.id);
		return taken;
	}

	/** The sites that may still have more results. */
	function remainingSites(): string[] {
		return [...sites].filter((site) => !exhausted.has(site));
	}

	/** Asks the remaining sites for their next page. */
	function loadMore() {
		if (searching || finished) return;
		const remaining = remainingSites();
		if (remaining.length === 0) {
			finished = true;
			return;
		}
		pageNumber++;
		void runPage(query, remaining);
	}
</script>

<div class="flex flex-col gap-8" bind:this={container}>
	{#if searching && pageNumber === 0 && sitesAsked > 0}
		<div class="-mt-4 flex items-center gap-3">
			<span class="chip pointer-events-none">
				<span class="loading loading-spinner loading-xs"></span>
				{sitesAnswered} of {sitesAsked} sites answered
			</span>
			<span class="h-1 w-40 overflow-hidden rounded-full bg-white/10">
				<span
					class="bg-primary block h-full rounded-full transition-[width] duration-300"
					style:width="{(sitesAnswered / sitesAsked) * 100}%"
				></span>
			</span>
		</div>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if spotlight.length > 0}
		<div class="rise-in">
			<VideoSpotlight
				picks={spotlight}
				surface="search"
				label="Best match"
				onopen={(video) => goto(videoUrl(video.id))}
			/>
		</div>
	{/if}

	{#each sections as section, index (index)}
		{@const gridVideos = index === 0 ? section.slice(SPOTLIGHT_COUNT) : section}
		{@const skipped = index === 0 ? SPOTLIGHT_COUNT : 0}
		{#if gridVideos.length > 0}
			<div class="card-grid">
				{#each gridVideos as video, offset (video.id)}
					<div class="rise-in" style:--rise-delay={riseDelay(offset)}>
						<SceneCard
							{video}
							shown={{ surface: 'search', position: sectionStart(index) + skipped + offset }}
							onclick={() => goto(videoUrl(video.id))}
						/>
					</div>
				{/each}
			</div>
		{/if}
		{#if rowShown[index] && rows[index]}
			{@const row = rows[index]}
			<div class="-mx-3 rounded-2xl bg-white/[0.03] p-4 ring-1 ring-white/5">
				{#if row.performers.length > 0}
					<Shelf
						title={row.title}
						subtitle="named “{query}”"
						items={row.performers}
						key={(performer) => performer.id}
						itemWidth={PERFORMER_WIDTH}
					>
						{#snippet card(performer)}
							<MemberCard kind="performer" member={performer} />
						{/snippet}
					</Shelf>
				{:else}
					<Shelf
						title={row.title}
						items={row.videos}
						key={(video) => video.id}
						itemWidth={VIDEO_WIDTH}
					>
						{#snippet card(video)}
							<SceneCard {video} onclick={() => goto(videoUrl(video.id))} />
						{/snippet}
					</Shelf>
				{/if}
			</div>
		{/if}
	{/each}

	{#if !searching && finished && videos.length === 0}
		<div class="flex flex-col items-center gap-3 py-16 text-center">
			<span class="flex h-14 w-14 items-center justify-center rounded-full bg-white/5">
				<MagnifyingGlassIcon size={26} class="opacity-50" />
			</span>
			<p class="font-medium">No site found anything for “{query}”</p>
			<p class="text-base-content/50 text-sm">Try fewer words, or another spelling of the name.</p>
		</div>
	{:else if searching && videos.length === 0}
		<div class="flex flex-col gap-8" aria-hidden="true">
			<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)]">
				<div class="skeleton aspect-video rounded-2xl"></div>
				<div class="flex flex-col gap-2 max-lg:hidden">
					{#each { length: SPOTLIGHT_COUNT }, position (position)}
						<div class="skeleton flex-1 rounded-xl"></div>
					{/each}
				</div>
			</div>
			<div class="card-grid">
				{#each { length: SKELETON_COUNT }, position (position)}
					<div class="flex flex-col gap-2">
						<div class="skeleton aspect-video w-full rounded-xl"></div>
						<div class="skeleton h-3.5 w-3/4 rounded"></div>
						<div class="skeleton h-3 w-1/3 rounded"></div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={searching && videos.length > 0} done={finished} />
</div>
