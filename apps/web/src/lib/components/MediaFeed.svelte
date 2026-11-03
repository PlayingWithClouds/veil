<!--
	An endless grid of one kind of media: pages of a feed (see feed.ts) as
	the end comes into view, each card logged as shown on a surface.
-->
<script lang="ts" generics="Item extends { id: string }">
	import { untrack, type Snippet } from 'svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { resetShown, type Shown } from '$lib/events';
	import type { Feed } from '$lib/feed';

	interface Props {
		// Starts the feed; called once (key the component to start over).
		feed: () => Feed<Item>;
		// Smallest card width in px.
		cardWidth: number;
		surface: string;
		card: Snippet<[Item, Shown]>;
		// Shown when the feed has nothing at all.
		empty: string;
	}

	let { feed, cardWidth, surface, card, empty }: Props = $props();

	const PAGE_SIZE = 40;

	let items = $state<Item[]>([]);
	let loading = $state(false);
	let done = $state(false);
	let error = $state<string | null>(null);

	let running: Feed<Item> | null = null;

	// Starts once; the page keys this component to start over.
	$effect(() => {
		untrack(() => {
			resetShown(surface);
			running = feed();
			void loadMore();
		});
	});

	/** Appends the feed's next page. */
	async function loadMore() {
		if (!running || loading || done) return;
		loading = true;
		try {
			const nextPage = await running(PAGE_SIZE);
			items = [...items, ...nextPage];
			done = nextPage.length === 0;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}
</script>

<div
	class="grid gap-x-4 gap-y-6"
	style:grid-template-columns="repeat(auto-fill, minmax({cardWidth}px, 1fr))"
>
	{#each items as item, position (item.id)}
		{@render card(item, { surface, position })}
	{/each}
</div>
{#if error}
	<p class="text-error text-sm">{error}</p>
{/if}
{#if done && items.length === 0}
	<p class="text-base-content/50 text-sm">{empty}</p>
{/if}
<ScrollSentinel onvisible={loadMore} loading={loading && items.length > 0} {done} />
{#if loading && items.length === 0}
	<p class="text-base-content/50 text-sm">Loading…</p>
{/if}
