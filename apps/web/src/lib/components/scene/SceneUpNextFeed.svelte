<!--
	The phone watch page's up next: one column of Home's scene cards, with
	chips to pick the source: everything (related first, then the personal
	feed), only related scenes, only the personal feed, or more of the first
	performer. The mixed and personal feeds page in as their end scrolls into view.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import { entityFilter, feedItemToSceneHit, fetchEntitySceneHits, type SceneHit } from '$lib/explore';
	import { fetchSearchPlugins, pluginForUrl, type PluginSearchResult, type SearchPlugin } from '$lib/search';
	import { sceneUrl } from '$lib/routes';
	import type { PerformerRef } from '$lib/sceneDetail';
	import type { RelatedFeed } from './relatedFeed.svelte';

	type Source = 'all' | 'related' | 'forYou' | 'performer';

	// One card, whichever source it came from.
	type Card = {
		key: string;
		scene: SceneHit;
		sourceUrl: string;
		// Stored in the library; site results aren't until opened.
		actionable: boolean;
		open: () => void;
	};

	interface Props {
		feed: RelatedFeed;
		// Scene on screen, left out of the performer's list.
		sceneId: string;
		// Credited performers; the first gets a "more of" chip.
		performers: PerformerRef[];
		// Source URL of the item being stored before navigation.
		openingUrl: string | null;
		onopen: (item: PluginSearchResult) => void;
	}

	let { feed, sceneId, performers, openingUrl, onopen }: Props = $props();

	let performer = $derived.by(() => {
		if (performers.length === 0) return null;
		return performers[0];
	});

	const SKELETON_CARDS = 2;
	const PERFORMER_PAGE = 40;

	let source = $state<Source>('all');
	let sites = $state<SearchPlugin[]>([]);
	let sentinel = $state<HTMLElement | null>(null);
	let performerScenes = $state.raw<SceneHit[]>([]);
	let performerLoading = $state(false);
	let performerLoadedFor = '';

	let cards = $derived(cardsFor(source));
	let loading = $derived.by(() => {
		if (source === 'performer') return performerLoading;
		return feed.loading;
	});
	let pages = $derived(source === 'all' || source === 'forYou');

	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	// Another scene starts on everything again.
	$effect(() => {
		sceneId;
		source = 'all';
		performerScenes = [];
		performerLoadedFor = '';
	});

	// Re-observing after each page makes a still-visible sentinel load the next one.
	$effect(() => {
		if (!sentinel || !pages || feed.loading || feed.done) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) feed.loadMore();
			},
			{ rootMargin: '800px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});

	/** Switches the source, loading the performer's scenes the first time. */
	function choose(next: Source) {
		source = next;
		if (next === 'performer') loadPerformerScenes();
	}

	/** Loads the first performer's library scenes, once per performer. */
	async function loadPerformerScenes() {
		if (!performer || performerLoadedFor === performer.id) return;
		const id = performer.id;
		performerLoadedFor = id;
		performerLoading = true;
		const hits = await fetchEntitySceneHits(entityFilter('performer', id), PERFORMER_PAGE, 0);
		if (performerLoadedFor !== id) return;
		performerScenes = hits.filter((hit) => hit.id !== sceneId);
		performerLoading = false;
	}

	/** The cards for a source. */
	function cardsFor(selected: Source): Card[] {
		if (selected === 'performer') return performerScenes.map(hitCard);
		if (selected === 'related') return feed.related.map(feedCard);
		if (selected === 'forYou') return feed.recommendations.map(feedCard);
		return feed.items.map(feedCard);
	}

	/** A card for a related or personal feed item. */
	function feedCard(item: PluginSearchResult): Card {
		return {
			key: item.externalId,
			scene: feedItemToSceneHit(item),
			sourceUrl: item.sourceUrl,
			actionable: Boolean(item.sceneId || item.dbId),
			open: () => onopen(item)
		};
	}

	/** A card for a library scene of the performer. */
	function hitCard(hit: SceneHit): Card {
		return {
			key: hit.id,
			scene: hit,
			sourceUrl: hit.sourceUrl,
			actionable: true,
			open: () => goto(sceneUrl(hit.id))
		};
	}

	/** The chips; the performer's only when the scene credits one. */
	function chips(): { key: Source; label: string }[] {
		const list: { key: Source; label: string }[] = [
			{ key: 'all', label: 'All' },
			{ key: 'related', label: 'Related' },
			{ key: 'forYou', label: 'For you' }
		];
		if (performer) list.push({ key: 'performer', label: `More of ${performer.name}` });
		return list;
	}
</script>

<section class="flex flex-col gap-3" aria-label="Up next">
	<div class="flex flex-col gap-2">
		<h2 class="text-base font-semibold">Up next</h2>
		<div class="no-scrollbar fade-right-edge -mx-3 flex gap-2 overflow-x-auto px-3 pr-8">
			{#each chips() as chip (chip.key)}
				<button
					type="button"
					aria-pressed={source === chip.key}
					onclick={() => choose(chip.key)}
					class="upnext-chip bg-base-200 flex h-8 shrink-0 items-center rounded-lg px-3 text-sm font-medium whitespace-nowrap"
					class:upnext-chip-active={source === chip.key}
				>
					{chip.label}
				</button>
			{/each}
		</div>
	</div>

	{#if cards.length === 0 && loading}
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
	{:else if cards.length === 0}
		<p class="text-base-content/40 py-6 text-sm">Nothing here yet.</p>
	{/if}
	<div class="flex flex-col gap-6">
		{#each cards as card (card.key)}
			<ExploreSceneCard
				scene={card.scene}
				site={pluginForUrl(card.sourceUrl, sites)}
				actionable={card.actionable}
				opening={openingUrl === card.sourceUrl}
				onclick={card.open}
			/>
		{/each}
	</div>
	{#if loading && cards.length > 0}
		<div class="flex justify-center py-4">
			<span class="loading loading-sm text-base-content/40"></span>
		</div>
	{/if}
	<div bind:this={sentinel} class="h-1 shrink-0"></div>
</section>

<style>
	.upnext-chip {
		transition: background-color 150ms;
	}

	.upnext-chip-active {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
