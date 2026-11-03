<!--
	A titled row of cards that scrolls sideways, with an optional grey
	subtitle, ‹ › buttons (only while the row overflows, shown on hover) and
	an optional "See all" link. Renders nothing while it has no items, so
	empty rows vanish.
-->
<script lang="ts" generics="Item">
	import type { Snippet } from 'svelte';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';

	interface Props {
		title: string;
		// A short grey note beside the title, e.g. "3 live".
		subtitle?: string | null;
		items: Item[];
		// A stable key per item.
		key: (item: Item) => string;
		// Width of one card in px; cards keep it on every screen.
		itemWidth: number;
		// The page listing everything this row samples.
		href?: string | null;
		// Draws one item, given its 0-based position in the row.
		card: Snippet<[Item, number]>;
	}

	let { title, subtitle = null, items, key, itemWidth, href = null, card }: Props = $props();

	let scroller = $state<HTMLElement>();
	let atStart = $state(true);
	let atEnd = $state(false);

	/** Tracks whether the row can scroll further either way, to dim the buttons. */
	function updateEnds() {
		if (!scroller) return;
		atStart = scroller.scrollLeft <= 0;
		atEnd = scroller.scrollLeft + scroller.clientWidth >= scroller.scrollWidth - 1;
	}

	/** Scrolls the row one view's width back (-1) or forward (1). */
	function page(direction: number) {
		scroller?.scrollBy({ left: direction * scroller.clientWidth * 0.9, behavior: 'smooth' });
	}

	$effect(() => {
		void items.length;
		updateEnds();
	});

	// A resized window can make the row fit, or overflow.
	$effect(() => {
		if (!scroller) return;
		const observer = new ResizeObserver(updateEnds);
		observer.observe(scroller);
		return () => observer.disconnect();
	});
</script>

{#if items.length > 0}
	<section class="group/shelf flex flex-col gap-3">
		<div class="flex items-center justify-between gap-3">
			<h2 class="flex items-baseline gap-2 text-lg font-semibold">
				{title}
				{#if subtitle}
					<span class="text-base-content/45 text-sm font-normal">{subtitle}</span>
				{/if}
			</h2>
			<div class="flex items-center gap-1">
				{#if href}
					<a {href} class="text-primary mr-2 text-sm">See all</a>
				{/if}
				{#if !(atStart && atEnd)}
					{@render arrows()}
				{/if}
			</div>
		</div>
		<div
			bind:this={scroller}
			onscroll={updateEnds}
			class="no-scrollbar -mx-1 flex snap-x gap-4 overflow-x-auto px-1 pb-2"
		>
			{#each items as item, position (key(item))}
				<div class="shrink-0 snap-start" style:width="{itemWidth}px">
					{@render card(item, position)}
				</div>
			{/each}
		</div>
	</section>
{/if}

{#snippet arrows()}
	<div
		class="flex items-center gap-1 opacity-0 transition-opacity group-focus-within/shelf:opacity-100 group-hover/shelf:opacity-100"
	>
		<button
			type="button"
			aria-label="Scroll back"
			class="hover:bg-base-200 flex h-8 w-8 items-center justify-center rounded-full disabled:opacity-25"
			disabled={atStart}
			onclick={() => page(-1)}
		>
			<CaretLeftIcon size={18} />
		</button>
		<button
			type="button"
			aria-label="Scroll forward"
			class="hover:bg-base-200 flex h-8 w-8 items-center justify-center rounded-full disabled:opacity-25"
			disabled={atEnd}
			onclick={() => page(1)}
		>
			<CaretRightIcon size={18} />
		</button>
	</div>
{/snippet}
