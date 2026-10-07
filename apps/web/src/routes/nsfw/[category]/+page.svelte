<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import SceneFilters from '$lib/components/SceneFilters.svelte';
	import { pluginSearchStream, type PluginSearchResult, type SearchStream } from '$lib/search';
	import { categoryBySlug } from '$lib/nsfw';
	import {
		fetchFilteredScenes,
		sceneToCard,
		emptyFilter,
		hasStructuredFilter,
		type SceneFilter,
		type SceneRow
	} from '$lib/filters';

	let slug = $derived($page.params.category);
	let category = $derived(slug ? categoryBySlug(slug) : undefined);

	let items = $state<PluginSearchResult[]>([]);
	let loading = $state(true);

	// Advanced filters: when a structured dimension is set, swap the plugin stream
	// for DB-backed filtered results.
	let filter = $state<SceneFilter>(emptyFilter());
	let filterMode = $derived(hasStructuredFilter(filter));
	let dbResults = $state<SceneRow[]>([]);
	let dbLoading = $state(false);

	async function runDbFilter() {
		dbLoading = true;
		dbResults = await fetchFilteredScenes({ ...filter, search: category?.query || null }, 60, 0);
		dbLoading = false;
	}

	// Reload whenever the category slug changes (client-side navigation between
	// categories reuses this component). Skipped while filter mode is active.
	$effect(() => {
		const current = category;
		if (!current) {
			goto('/nsfw');
			return;
		}
		if (filterMode) {
			runDbFilter();
			return;
		}
		loading = true;
		items = [];
		const seen = new Set<string>();
		const stream: SearchStream = pluginSearchStream(
			current.query,
			[],
			(result) => {
				if (seen.has(result.externalId)) return;
				seen.add(result.externalId);
				items = [...items, result];
			},
			() => {
				loading = false;
			}
		);
		return () => {
			stream.close();
		};
	});
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-center gap-3">
		<button
			type="button"
			class="btn btn-square btn-ghost btn-sm"
			aria-label="Back to browse"
			onclick={() => goto('/nsfw')}
		>
			<ArrowLeftIcon size={18} />
		</button>
		<h1 class="text-3xl font-bold">{category?.label ?? 'Category'}</h1>
		<span class="badge badge-error">18+</span>
	</div>

	<SceneFilters bind:filter oncommit={runDbFilter} />

	{#if filterMode}
		{#if dbResults.length === 0 && !dbLoading}
			<div class="text-base-content/40 py-24 text-center text-sm">No scenes match these filters.</div>
		{:else}
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-4">
				{#each dbResults as scene (scene.id)}
					<SceneCard item={sceneToCard(scene)} onclick={() => goto(`/scene/${scene.id}`)} />
				{/each}
			</div>
		{/if}
	{:else}
		<NsfwGrid {items} {loading} ariaLabel={category?.label ?? 'Category'} />
	{/if}
</div>
