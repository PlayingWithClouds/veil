<script lang="ts">
	// The watch page's up-next rail: compact cards in a vertical FocusList (D-pad
	// navigable), paging in more of the feed as its end scrolls into view. Chips
	// above filter it by why items are there (site's picks, cast, similar, for you).
	import FocusList from '$lib/components/FocusList.svelte';
	import { fetchSearchPlugins, pluginForUrl, type PluginSearchResult, type SearchPlugin } from '$lib/search';
	import type { RelatedFeed } from './relatedFeed.svelte';
	import RailCard from './RailCard.svelte';
	import { relatedImpression } from './relatedImpressions';
	import { impression, recordClicked } from '$lib/recommendations';
	import { upNextGroup, upNextGroupLabel, upNextReason, type UpNextGroup } from './upNextReason';

	type Direction = 'left' | 'right' | 'up' | 'down';

	interface Props {
		feed: RelatedFeed;
		// Source URL of the item being ingested before navigation.
		openingUrl: string | null;
		// The watched scene's source URL, so cards from its own site go unnamed.
		sceneSourceUrl: string;
		// The watched scene's performer names, for "With <name>" reasons.
		cast: string[];
		onopen: (item: PluginSearchResult) => void;
		onedge?: (direction: Direction) => void;
	}

	let { feed, openingUrl, sceneSourceUrl, cast, onopen, onedge }: Props = $props();

	// Chip order above the list.
	const GROUP_ORDER: UpNextGroup[] = ['site', 'cast', 'similar', 'feed'];

	// Sites, to label each card by its source URL.
	let sites = $state<SearchPlugin[]>([]);
	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});
	let sceneSite = $derived(pluginForUrl(sceneSourceUrl, sites));

	// The picked filter chip; null shows everything.
	let selectedGroup = $state<UpNextGroup | null>(null);
	// The groups the current items fall into; chips only show when there are two or more.
	let groups = $derived(GROUP_ORDER.filter((group) => feed.items.some((item) => upNextGroup(item) === group)));
	let visibleItems = $derived.by(() => {
		if (selectedGroup === null) return feed.items;
		return feed.items.filter((item) => upNextGroup(item) === selectedGroup);
	});
	// Only the personal feed pages in more; the related groups are complete.
	let pagesMore = $derived(selectedGroup === null || selectedGroup === 'feed');

	// A new scene starts unfiltered.
	$effect(() => {
		sceneSourceUrl;
		selectedGroup = null;
	});

	/** Whether an item comes from the watched scene's site. */
	function onSceneSite(item: PluginSearchResult): boolean {
		if (!sceneSite) return false;
		return pluginForUrl(item.sourceUrl, sites)?.name === sceneSite.name;
	}

	let list = $state<{ focusFirst: () => void } | undefined>();
	let sentinel = $state<HTMLElement | null>(null);
	let rail = $state<HTMLElement | null>(null);

	/** The rail when it is its own scroll container, else null for the viewport. */
	function scrollRoot(): HTMLElement | null {
		if (!rail) return null;
		if (getComputedStyle(rail).overflowY === 'visible') return null;
		return rail;
	}

	/** Logs the click with the item's place in the rail, then opens it. */
	function openItem(item: PluginSearchResult, position: number) {
		recordClicked(relatedImpression(item, position));
		onopen(item);
	}

	/** Moves D-pad focus onto the first rail card. */
	export function focusFirst() {
		list?.focusFirst();
	}

	/** Pages in more items when D-pad focus runs off the bottom; other edges go to the caller. */
	function handleEdge(direction: Direction) {
		if (direction === 'down') {
			feed.loadMore();
			return;
		}
		onedge?.(direction);
	}

	// Infinite scroll, watched against the rail itself where it scrolls on its
	// own (beside the player), else against the viewport (below the player).
	// Re-observing after each page makes a still-visible sentinel load the next one.
	$effect(() => {
		if (!sentinel || feed.loading || feed.done || !pagesMore) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) feed.loadMore();
			},
			{ root: scrollRoot(), rootMargin: '400px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});
</script>

<!-- Beside the player (lg) the rail scrolls on its own, held at the top of the
     page, so the player stays in place while the list moves. -->
<aside
	bind:this={rail}
	class="no-scrollbar flex flex-col gap-2 lg:sticky lg:top-0 lg:max-h-[calc(100dvh-3.5rem)] lg:self-start lg:overflow-y-auto lg:overscroll-y-contain lg:pb-4"
	aria-label="Up next"
>
	{#if groups.length > 1}
		<div class="no-scrollbar mb-1 flex shrink-0 gap-2 overflow-x-auto" role="toolbar" aria-label="Filter up next">
			{@render chip(null, 'All')}
			{#each groups as group (group)}
				{@render chip(group, upNextGroupLabel(group, cast))}
			{/each}
		</div>
	{/if}
	{#if feed.items.length === 0 && !feed.loading}
		<p class="text-base-content/40 text-sm">Nothing here yet.</p>
	{/if}
	<FocusList
		bind:this={list}
		items={visibleItems}
		orientation="vertical"
		ariaLabel="Up next"
		class="[--tv-gap:0.75rem]"
		onselect={openItem}
		onedge={handleEdge}
	>
		{#snippet item(entry, position, focused)}
			<div use:impression={relatedImpression(entry, position)}>
				<RailCard
					item={entry}
					site={pluginForUrl(entry.sourceUrl, sites)}
					opening={openingUrl === entry.sourceUrl}
					{focused}
					reason={upNextReason(entry, cast)}
					sameSite={onSceneSite(entry)}
				/>
			</div>
		{/snippet}
	</FocusList>
	{#if feed.loading && pagesMore}
		<div class="flex justify-center py-4">
			<span class="loading loading-sm text-base-content/40"></span>
		</div>
	{/if}
	<div bind:this={sentinel} class="h-1 shrink-0"></div>
</aside>

{#snippet chip(group: UpNextGroup | null, label: string)}
	<button
		type="button"
		class="bg-base-200 ring-base-content/15 h-8 shrink-0 rounded-lg px-3 text-sm font-semibold whitespace-nowrap capitalize ring-1 ring-inset transition-colors hover:bg-white/15"
		class:bg-white={selectedGroup === group}
		class:text-black={selectedGroup === group}
		class:hover:bg-white={selectedGroup === group}
		aria-pressed={selectedGroup === group}
		onclick={() => (selectedGroup = group)}
	>
		{label}
	</button>
{/snippet}
