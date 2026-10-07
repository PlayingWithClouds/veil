<script lang="ts">
	import { goto } from '$app/navigation';
	import SpinnerGapIcon from 'phosphor-svelte/lib/SpinnerGapIcon';
	import FocusGrid from './FocusGrid.svelte';
	import SceneCard from './SceneCard.svelte';
	import { sidebarFocus } from '$lib/stores/focus';
	import { notifications } from '$lib/stores/notifications';
	import { ingestPluginResult, type PluginSearchResult } from '$lib/search';
	import { toSceneCard } from '$lib/nsfw';

	interface Props {
		items: PluginSearchResult[];
		loading?: boolean;
		ariaLabel?: string;
		// When set, left-edge navigation hands focus here (e.g. the search keyboard)
		// instead of opening the sidebar.
		onleftedge?: () => void;
	}

	let { items, loading = false, ariaLabel = 'Adult content', onleftedge }: Props = $props();

	const GRID_CLASS = 'grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4';

	// Source URL currently being scraped into a library record (null = idle).
	let openingUrl = $state<string | null>(null);

	let grid = $state<{ focus: () => void; focusFirst: () => void }>();
	export function focusFirst() {
		grid?.focusFirst();
	}

	// Adult results aren't library records yet: scrape the source into a canonical
	// record on select, then open its detail page.
	async function openResult(result: PluginSearchResult) {
		// Locally-stored scenes are already ingested: open the detail page directly.
		if (result.sceneId) {
			goto(`/scene/${result.sceneId}`);
			return;
		}
		if (openingUrl !== null) return;
		openingUrl = result.sourceUrl;
		try {
			const mediaId = await ingestPluginResult(result.plugin, result.sourceUrl, result.posterUrl);
			goto(`/scene/${mediaId}`);
		} catch {
			openingUrl = null;
			notifications.push('Could not open this title', 'error');
		}
	}

	function onGridEdge(direction: 'left' | 'right' | 'up' | 'down') {
		if (direction !== 'left') return;
		if (onleftedge) {
			onleftedge();
			return;
		}
		sidebarFocus.openSidebar();
	}
</script>

{#if items.length === 0 && !loading}
	<div class="flex flex-col items-center gap-3 py-24 text-center">
		<span class="text-6xl opacity-10">🔞</span>
		<p class="text-base-content/40">No results.</p>
	</div>
{:else}
	<FocusGrid
		bind:this={grid}
		{items}
		{ariaLabel}
		gridClass={GRID_CLASS}
		onselect={(item) => openResult(item)}
		onedge={onGridEdge}
	>
		{#snippet item(entry)}
			<div class="relative">
				<SceneCard
					item={toSceneCard(entry)}
					onclick={() => openResult(entry)}
				/>
				{#if openingUrl === entry.sourceUrl}
					<div
						class="absolute inset-0 flex items-center justify-center rounded-xl bg-black/60"
						aria-hidden="true"
					>
						<SpinnerGapIcon size={28} class="animate-spin text-white" />
					</div>
				{/if}
			</div>
		{/snippet}
	</FocusGrid>
{/if}

{#if loading}
	<div class="flex justify-center py-8">
		<span class="text-base-content/40 text-sm">Loading…</span>
	</div>
{/if}
