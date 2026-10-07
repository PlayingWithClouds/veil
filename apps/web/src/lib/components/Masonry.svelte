<!--
	Left-to-right masonry. Each item goes to the currently shortest column in
	arrival order, so the grid reads row by row like a normal grid while mixed
	aspect ratios pack tightly. Placement is append-only: an item never moves once
	placed. Only a change in column count (resize) or a list that drops placed
	items (new search) lays everything out again.
-->
<script lang="ts" generics="Item">
	import { untrack, type Snippet } from 'svelte';

	interface Props {
		items: Item[];
		keyOf: (item: Item) => string;
		// Width / height of the item's media, used to estimate its height before
		// it renders.
		aspectOf: (item: Item) => number;
		// Typical height of everything under the media (title, meta), in pixels.
		// Only an estimate: rendered columns are measured before each placement.
		captionHeight: number;
		minColumnWidth?: number;
		// Wide screens get wider columns instead of more of them.
		maxColumns?: number;
		gap?: number;
		children: Snippet<[Item]>;
	}

	let {
		items,
		keyOf,
		aspectOf,
		captionHeight,
		minColumnWidth = 280,
		maxColumns = 4,
		gap = 16,
		children
	}: Props = $props();

	let containerWidth = $state(0);
	let columnCount = $derived(
		Math.min(maxColumns, Math.max(1, Math.floor((containerWidth + gap) / (minColumnWidth + gap))))
	);
	let columns = $state.raw<Item[][]>([]);
	let columnElements = $state<HTMLElement[]>([]);

	// Placement bookkeeping; not rendered, so kept outside reactivity.
	let placedKeys = new Set<string>();
	let placedColumnCount = 0;
	let estimatedHeights: number[] = [];

	$effect(() => {
		const count = columnCount;
		const list = items;
		untrack(() => place(list, count));
	});

	/** Appends unplaced items to the shortest columns, restarting the layout when needed. */
	function place(list: Item[], count: number) {
		if (needsRelayout(list, count)) {
			resetLayout(count);
		} else {
			syncMeasuredHeights();
		}
		const columnWidth = (containerWidth - gap * (count - 1)) / count;
		const nextColumns = columns.map((column) => [...column]);
		for (const item of list) {
			const key = keyOf(item);
			if (placedKeys.has(key)) continue;
			const target = shortestColumn();
			nextColumns[target].push(item);
			estimatedHeights[target] += columnWidth / aspectOf(item) + captionHeight + gap;
			placedKeys.add(key);
		}
		columns = nextColumns;
	}

	/** Whether the column count changed or placed items vanished from the list. */
	function needsRelayout(list: Item[], count: number): boolean {
		if (count !== placedColumnCount) return true;
		const currentKeys = new Set(list.map(keyOf));
		for (const key of placedKeys) {
			if (!currentKeys.has(key)) return true;
		}
		return false;
	}

	/** Clears every column for a fresh layout with count columns. */
	function resetLayout(count: number) {
		placedKeys = new Set();
		placedColumnCount = count;
		estimatedHeights = new Array(count).fill(0);
		columns = Array.from({ length: count }, () => []);
	}

	/** Replaces estimates with real column heights where they have rendered. */
	function syncMeasuredHeights() {
		for (let index = 0; index < estimatedHeights.length; index++) {
			const element = columnElements[index];
			if (element) estimatedHeights[index] = element.offsetHeight;
		}
	}

	/** Index of the shortest column; the leftmost wins ties. */
	function shortestColumn(): number {
		let shortest = 0;
		for (let index = 1; index < estimatedHeights.length; index++) {
			if (estimatedHeights[index] < estimatedHeights[shortest]) shortest = index;
		}
		return shortest;
	}
</script>

<div bind:clientWidth={containerWidth} class="flex items-start" style:gap="{gap}px">
	{#each columns as column, columnIndex (columnIndex)}
		<div bind:this={columnElements[columnIndex]} class="flex min-w-0 flex-1 flex-col" style:gap="{gap}px">
			{#each column as item (keyOf(item))}
				{@render children(item)}
			{/each}
		</div>
	{/each}
</div>
