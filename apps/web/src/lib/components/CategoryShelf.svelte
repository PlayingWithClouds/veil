<!--
	A row of the categories holding the most of one kind of media (groups
	left out: their categories show instead), with "See all" to the
	categories page. Hidden while none holds anything.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import Shelf from '$lib/components/Shelf.svelte';
	import CategoryCard from '$lib/components/CategoryCard.svelte';
	import { categoryKindOf, fetchCategoryTiles, type CategoryTileRow } from '$lib/categories';
	import { categoriesUrl, categoryUrl } from '$lib/routes';
	import type { MediaKind } from '@playingwithclouds/veil-sdk/generated';

	interface Props {
		kind: MediaKind;
	}

	let { kind }: Props = $props();

	const SHOWN = 16;
	const TILE_WIDTH = 220;

	let tiles = $state<CategoryTileRow[]>([]);

	let noun = $derived(categoryKindOf(kind).noun);

	$effect(() => {
		fetchCategoryTiles(kind).then(
			(found) => (tiles = fullest(found)),
			(failure) => console.warn('categories row failed to load', failure)
		);
	});

	/** The categories (not groups) holding the most, most first. */
	function fullest(found: CategoryTileRow[]): CategoryTileRow[] {
		return found
			.filter((tile) => tile.category.parent !== null && tile.count > 0)
			.sort((left, right) => right.count - left.count)
			.slice(0, SHOWN);
	}
</script>

<Shelf
	title="Categories"
	items={tiles}
	key={(tile) => tile.category.id}
	itemWidth={TILE_WIDTH}
	href={categoriesUrl(kind)}
>
	{#snippet card(tile)}
		<CategoryCard {tile} {noun} onclick={() => goto(categoryUrl(tile.category.id, kind))} />
	{/snippet}
</Shelf>
