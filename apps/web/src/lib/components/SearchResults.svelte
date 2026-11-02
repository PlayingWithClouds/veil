<!--
	A search across the sites, shown as it comes in and never reshuffled:
	each site's best four take the top twenty places in the order the sites
	answer, the rest follows sites taking turns (SearchLayout). Scrolling on
	asks the sites for their next page until none has more. Every six grid
	rows a row breaks the grid up: performers by that name, picks for you,
	your favourites. A site that answered the same search within six hours
	isn't asked again; the backend sends its answer from then.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { fetchPerformers, fetchVideosMatching, type MemberRow, type VideoCardRow } from '$lib/library';
	import { fetchRecommendedVideos } from '$lib/home';
	import { searchPage, type SearchBatchRow, type SearchPage } from '$lib/search';
	import { SearchLayout, TOP_PLACES } from '$lib/searchLayout';
	import { logEvent, resetShown } from '$lib/events';
	import { videoUrl } from '$lib/routes';
	import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';

	// Grid rows between two breaking rows, after the top places.
	const SECTION_ROWS = 6;
	// The grid's card width and gap (px), to count its columns.
	const CARD_MIN_WIDTH = 260;
	const GRID_GAP = 16;
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
		const columns = Math.max(1, Math.floor((width + GRID_GAP) / (CARD_MIN_WIDTH + GRID_GAP)));
		return columns * SECTION_ROWS;
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
		layout = new SearchLayout<VideoCardRow>();
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
		const running = searchPage('VIDEO', searchText, pageNumber, plugins, activeFilter, receive);
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
			if (favorites.length > 0) return { title: 'Your favourites', videos: favorites, performers: [] };
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
		<p class="text-base-content/50 -mt-4 text-sm">
			{sitesAnswered} of {sitesAsked} sites answered
		</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#each sections as section, index (index)}
		<div
			class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-x-4 gap-y-8 max-md:grid-cols-1"
		>
			{#each section as video, offset (video.id)}
				<SceneCard
					{video}
					shown={{ surface: 'search', position: sectionStart(index) + offset }}
					onclick={() => goto(videoUrl(video.id))}
				/>
			{/each}
		</div>
		{#if rowShown[index] && rows[index]}
			{@const row = rows[index]}
			<div class="bg-base-200/60 -mx-3 rounded-2xl p-3">
				{#if row.performers.length > 0}
					<Shelf
						title={row.title}
						items={row.performers}
						key={(performer) => performer.id}
						itemWidth={PERFORMER_WIDTH}
					>
						{#snippet card(performer)}
							<MemberCard kind="performer" member={performer} />
						{/snippet}
					</Shelf>
				{:else}
					<Shelf title={row.title} items={row.videos} key={(video) => video.id} itemWidth={VIDEO_WIDTH}>
						{#snippet card(video)}
							<SceneCard {video} onclick={() => goto(videoUrl(video.id))} />
						{/snippet}
					</Shelf>
				{/if}
			</div>
		{/if}
	{/each}

	{#if !searching && finished && videos.length === 0}
		<p class="text-base-content/50 text-sm">No site found anything for this search.</p>
	{:else if searching && videos.length === 0}
		<p class="text-base-content/50 text-sm">Asking the sites…</p>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={searching && videos.length > 0} done={finished} />
</div>
