<!--
	Gallery results of a live search, kept apart from the videos so they never
	take a video's place in the ranking. As a row (beside the videos) it is a
	horizontal shelf; as a grid it is the page's main content (media toggle on
	images).
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import SpinnerGapIcon from 'phosphor-svelte/lib/SpinnerGapIcon';
	import FocusGrid from './FocusGrid.svelte';
	import GalleryCard from './GalleryCard.svelte';
	import { pluginResultToGalleryHit } from '$lib/explore';
	import { pluginForUrl, type PluginSearchResult, type SearchPlugin } from '$lib/search';
	import { galleryUrl } from '$lib/routes';

	type Direction = 'left' | 'right' | 'up' | 'down';

	interface Props {
		/** Stored gallery results (each has a dbId), in the order to show. */
		items: PluginSearchResult[];
		/** The search plugins, to label each card with its site. */
		sites: SearchPlugin[];
		layout: 'row' | 'grid';
		loading?: boolean;
		ariaLabel?: string;
		/** Fired when keyboard navigation leaves the galleries at an outer edge. */
		onedge?: (direction: Direction) => void;
	}

	let {
		items,
		sites,
		layout,
		loading = false,
		ariaLabel = 'Gallery results',
		onedge
	}: Props = $props();

	const ROW_CLASS =
		'no-scrollbar auto-cols-[9rem] grid-flow-col gap-4 overflow-x-auto p-1 sm:auto-cols-[11rem]';
	const GRID_CLASS = 'grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5';

	let grid = $state<{ focusFirst: () => void }>();

	/** Focuses the first gallery. */
	export function focusFirst() {
		grid?.focusFirst();
	}

	/** The grid classes of the chosen layout. */
	function layoutClass(): string {
		if (layout === 'row') return ROW_CLASS;
		return GRID_CLASS;
	}

	/** Opens the gallery's page. */
	function openGallery(result: PluginSearchResult) {
		if (!result.dbId) return;
		goto(galleryUrl(result.dbId));
	}
</script>

{#if items.length === 0 && !loading}
	<div class="flex flex-col items-center gap-3 py-24 text-center">
		<span class="text-6xl opacity-10">🔞</span>
		<p class="text-base-content/40">No galleries.</p>
	</div>
{:else}
	<FocusGrid
		bind:this={grid}
		{items}
		{ariaLabel}
		gridClass={layoutClass()}
		onselect={(item) => openGallery(item)}
		{onedge}
	>
		{#snippet item(entry)}
			<GalleryCard
				gallery={pluginResultToGalleryHit(entry, entry.dbId || '')}
				site={pluginForUrl(entry.sourceUrl, sites)}
				onclick={() => openGallery(entry)}
			/>
		{/snippet}
	</FocusGrid>
{/if}

{#if loading && items.length === 0}
	<div class="flex justify-center py-8">
		<SpinnerGapIcon size={22} class="text-base-content/40 animate-spin" />
	</div>
{/if}
