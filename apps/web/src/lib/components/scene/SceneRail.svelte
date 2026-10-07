<script lang="ts">
	// The watch page's up-next rail: compact cards in a vertical FocusList (D-pad
	// navigable), paging in more of the feed as its end scrolls into view.
	import FocusList from '$lib/components/FocusList.svelte';
	import { fetchSearchPlugins, pluginForUrl, type PluginSearchResult, type SearchPlugin } from '$lib/search';
	import type { RelatedFeed } from './relatedFeed.svelte';
	import RailCard from './RailCard.svelte';

	type Direction = 'left' | 'right' | 'up' | 'down';

	interface Props {
		feed: RelatedFeed;
		// Source URL of the item being ingested before navigation.
		openingUrl: string | null;
		onopen: (item: PluginSearchResult) => void;
		onedge?: (direction: Direction) => void;
	}

	let { feed, openingUrl, onopen, onedge }: Props = $props();

	// Sites, to label each card by its source URL.
	let sites = $state<SearchPlugin[]>([]);
	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	let list = $state<{ focusFirst: () => void } | undefined>();
	let sentinel = $state<HTMLElement | null>(null);

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

	// Infinite scroll: the rail flows with the page, so watch against the viewport.
	// Re-observing after each page makes a still-visible sentinel load the next one.
	$effect(() => {
		if (!sentinel || feed.loading || feed.done) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) feed.loadMore();
			},
			{ rootMargin: '400px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});
</script>

<aside class="flex flex-col gap-2" aria-label="Up next">
	{#if feed.items.length === 0 && !feed.loading}
		<p class="text-base-content/40 text-sm">Nothing here yet.</p>
	{/if}
	<FocusList
		bind:this={list}
		items={feed.items}
		orientation="vertical"
		ariaLabel="Up next"
		class="[--tv-gap:0.5rem]"
		onselect={(item) => onopen(item)}
		onedge={handleEdge}
	>
		{#snippet item(entry)}
			<RailCard
				item={entry}
				site={pluginForUrl(entry.sourceUrl, sites)}
				opening={openingUrl === entry.sourceUrl}
			/>
		{/snippet}
	</FocusList>
	{#if feed.loading}
		<div class="flex justify-center py-4">
			<span class="loading loading-sm text-base-content/40"></span>
		</div>
	{/if}
	<div bind:this={sentinel} class="h-1 shrink-0"></div>
</aside>
