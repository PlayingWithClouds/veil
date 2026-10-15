<script lang="ts">
	import { untrack } from 'svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
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

	// Each plugin's share of one page of the live listing, and the size of
	// one page of filtered (stored) results.
	const PER_PLUGIN = 20;
	const DB_PAGE_SIZE = 60;

	let items = $state<PluginSearchResult[]>([]);
	let loading = $state(true);
	let liveDone = $state(false);
	let liveOffset = 0;
	let seen = new Set<string>();
	let stream: SearchStream | null = null;

	// Advanced filters: when a structured dimension is set, swap the plugin stream
	// for DB-backed filtered results.
	let filter = $state<SceneFilter>(emptyFilter());
	let filterMode = $derived(hasStructuredFilter(filter));
	let dbResults = $state<SceneRow[]>([]);
	let dbLoading = $state(false);
	let dbDone = $state(false);

	/** Loads the first page of filtered results. */
	async function runDbFilter() {
		dbLoading = true;
		dbResults = await fetchFilteredScenes(
			{ ...filter, search: category?.query || null },
			DB_PAGE_SIZE,
			0
		);
		dbDone = dbResults.length < DB_PAGE_SIZE;
		dbLoading = false;
	}

	/** Appends the next page of filtered results. */
	async function loadMoreDb() {
		if (dbLoading || dbDone) return;
		dbLoading = true;
		const next = await fetchFilteredScenes(
			{ ...filter, search: category?.query || null },
			DB_PAGE_SIZE,
			dbResults.length
		);
		dbResults = [...dbResults, ...next];
		dbDone = next.length < DB_PAGE_SIZE;
		dbLoading = false;
	}

	/** Streams the sites' next page for the category; a page adding nothing ends the feed. */
	function loadNextLivePage() {
		if (!category || (loading && stream) || liveDone) return;
		loading = true;
		let added = 0;
		stream = pluginSearchStream(
			category.query,
			[],
			(result) => {
				if (seen.has(result.externalId)) return;
				seen.add(result.externalId);
				added++;
				items = [...items, result];
			},
			() => {
				loading = false;
				stream = null;
				if (added === 0) liveDone = true;
			},
			{ offset: liveOffset, limit: PER_PLUGIN }
		);
		liveOffset += PER_PLUGIN;
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
		items = [];
		seen = new Set();
		liveOffset = 0;
		liveDone = false;
		stream = null;
		untrack(loadNextLivePage);
		return () => {
			stream?.close();
			stream = null;
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
			<div class="text-base-content/40 py-24 text-center text-sm">
				No scenes match these filters.
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-4">
				{#each dbResults as scene (scene.id)}
					<SceneCard item={sceneToCard(scene)} onclick={() => goto(`/scene/${scene.id}`)} />
				{/each}
			</div>
			<ScrollSentinel onvisible={loadMoreDb} loading={dbLoading} done={dbDone} />
		{/if}
	{:else}
		<NsfwGrid
			{items}
			loading={loading && items.length === 0}
			ariaLabel={category?.label ?? 'Category'}
		/>
		{#if items.length > 0}
			<ScrollSentinel onvisible={loadNextLivePage} {loading} done={liveDone} />
		{/if}
	{/if}
</div>
