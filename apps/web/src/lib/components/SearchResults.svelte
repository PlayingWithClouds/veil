<!--
	A search across the sites, shown as it comes in and never reshuffled:
	each site's best four take the top twenty places in the order the sites
	answer, the rest follows sites taking turns (SearchLayout). Scrolling on
	asks the sites for their next page until none has more. Every thirty
	results a row breaks the grid up: performers by that name, picks for you,
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

	// Results between two breaking rows, after the top places.
	const SECTION = 30;
	// Videos in a breaking row.
	const ROW_VIDEOS = 8;
	const VIDEO_WIDTH = 260;
	const PERFORMER_WIDTH = 140;
	// How long results wait below the fold for slower sites before they're let in.
	const RELEASE_DELAY_MS = 1500;

	/** A row between two sections of results. */
	interface BreakRow {
		title: string;
		videos: VideoCardRow[];
		performers: MemberRow[];
	}

	let { query }: { query: string } = $props();

	let videos = $state<VideoCardRow[]>([]);
	let rows = $state<(BreakRow | null)[]>([]);
	// Whether the row after each section shows; decided once, when the next section starts.
	let rowShown = $state<boolean[]>([]);
	let sitesAsked = $state(0);
	let sitesAnswered = $state(0);
	let searching = $state(false);
	let finished = $state(false);
	let error = $state<string | null>(null);

	let layout = new SearchLayout<VideoCardRow>();
	let current: SearchPage | null = null;
	// Counts searches started, so late answers to an earlier one are dropped.
	let generation = 0;
	let pageNumber = $state(0);
	let sites = new Set<string>();
	let exhausted = new Set<string>();
	let newThisPage = new Map<string, number>();
	let releaseTimer: ReturnType<typeof setTimeout> | null = null;
	let recommendedPool: Promise<VideoCardRow[]> | null = null;
	let favoritePool: Promise<VideoCardRow[]> | null = null;

	let sections = $derived(splitSections(videos));

	$effect(() => {
		const searchText = query;
		untrack(() => void start(searchText));
		return stop;
	});

	/** The results cut into the top places and sections of SECTION. */
	function splitSections(all: VideoCardRow[]): VideoCardRow[][] {
		const result = [all.slice(0, TOP_PLACES)];
		for (let start = TOP_PLACES; start < all.length; start += SECTION) {
			result.push(all.slice(start, start + SECTION));
		}
		return result;
	}

	/** Where section index starts in the results. */
	function sectionStart(index: number): number {
		if (index === 0) return 0;
		return TOP_PLACES + (index - 1) * SECTION;
	}

	/** Stops listening to the running page and the release timer. */
	function stop() {
		current?.cancel();
		current = null;
		if (releaseTimer) clearTimeout(releaseTimer);
		releaseTimer = null;
	}

	/** Starts a search over: empty results, page 0. */
	async function start(searchText: string) {
		stop();
		generation++;
		resetShown('search');
		logEvent({ kind: 'search', data: { text: searchText } });
		layout = new SearchLayout<VideoCardRow>();
		videos = [];
		rows = [];
		rowShown = [];
		pageNumber = 0;
		sites = new Set();
		exhausted = new Set();
		recommendedPool = null;
		favoritePool = null;
		finished = false;
		error = null;
		prepareRow(0);
		await runPage(searchText, null);
	}

	/** Asks the sites for one page and shows what they answer. */
	async function runPage(searchText: string, plugins: string[] | null) {
		searching = true;
		sitesAsked = 0;
		sitesAnswered = 0;
		newThisPage = new Map();
		const running = searchPage('VIDEO', searchText, pageNumber, plugins, receive);
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
		show(layout.add(batch.plugin, batch.listed.videos));
		if (layout.hasWaiting) scheduleRelease();
		if (!batch.final) return;
		sitesAnswered++;
		// A site that had nothing new this page won't on the next either.
		if (batch.exhausted || newThisPage.get(batch.plugin) === 0) exhausted.add(batch.plugin);
	}

	/** Lets waiting results in after a moment, once the top places are settled. */
	function scheduleRelease() {
		if (releaseTimer || videos.length < TOP_PLACES) return;
		releaseTimer = setTimeout(() => {
			releaseTimer = null;
			show(layout.release());
		}, RELEASE_DELAY_MS);
	}

	/** Appends results, settling each breaking row as the section after it starts. */
	function show(items: VideoCardRow[]) {
		for (const item of items) {
			const index = sectionIndexStartingAt(videos.length);
			if (index > 0) settleRow(index - 1);
			videos.push(item);
		}
	}

	/** The section starting at position, or -1 when none does. */
	function sectionIndexStartingAt(position: number): number {
		if (position < TOP_PLACES) return position === 0 ? 0 : -1;
		if ((position - TOP_PLACES) % SECTION !== 0) return -1;
		return (position - TOP_PLACES) / SECTION + 1;
	}

	/** Decides whether row index shows (only when ready in time) and prepares the next. */
	function settleRow(index: number) {
		const row = rows[index];
		rowShown[index] = Boolean(row && (row.videos.length > 0 || row.performers.length > 0));
		prepareRow(index + 1);
	}

	/** Loads the row after section index, ahead of it being needed. */
	async function prepareRow(index: number) {
		if (rows[index] !== undefined) return;
		rows[index] = null;
		const startedFor = generation;
		const row = await loadRow(index);
		// A row loaded for an earlier search stays out of this one.
		if (startedFor === generation) rows[index] = row;
	}

	/** The row after section index: performers first, then picks and favourites. */
	async function loadRow(index: number): Promise<BreakRow> {
		if (index === 0) {
			const performers = await fetchPerformers(query, 12);
			return { title: 'Performers', videos: [], performers };
		}
		if (index % 3 === 2) {
			favoritePool ??= fetchVideosMatching({ text: query, favorite: true }, 60, 0);
			return { title: 'Your favourites', videos: takeUnseen(await favoritePool), performers: [] };
		}
		recommendedPool ??= fetchRecommendedVideos(60, { text: query }).then((picks) =>
			picks.map((pick) => pick.video)
		);
		return { title: 'Picked for you', videos: takeUnseen(await recommendedPool), performers: [] };
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

<div class="flex flex-col gap-8">
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
