<!--
	Our categories, grouped (Body, Hair, Acts, …), each tile with how much of
	one kind of media (?kind=, videos by default) it holds and a picture.
	Empty ones hide until "Show empty". New categories go into a group or
	start their own; "Add missing defaults" brings back deleted defaults.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import ArrowCounterClockwiseIcon from 'phosphor-svelte/lib/ArrowCounterClockwiseIcon';
	import CategoryCard from '$lib/components/CategoryCard.svelte';
	import CategoryEditor from '$lib/components/CategoryEditor.svelte';
	import CategoryKindTabs from '$lib/components/CategoryKindTabs.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import {
		addDefaultCategories,
		categoryKindOf,
		fetchCategoryTiles,
		groupTiles,
		type CategoryDetail,
		type CategoryTileRow,
		type TileGroup
	} from '$lib/categories';
	import { showUndo } from '$lib/stores/undoToast';
	import { categoriesUrl, categoryUrl } from '$lib/routes';

	let tiles = $state<CategoryTileRow[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let showEmpty = $state(false);
	// The group a new category goes into ('' = its own group); null while the editor is closed.
	let creatingIn = $state<string | null>(null);
	let restoring = $state(false);

	let current = $derived(categoryKindOf(page.url.searchParams.get('kind')));
	let groups = $derived(groupTiles(tiles));
	// Groups with categories inside, and the categories standing on their own.
	let filledGroups = $derived(groups.filter((group) => group.children.length > 0));
	let standalone = $derived(
		groups.filter((group) => group.children.length === 0).map((group) => group.parent)
	);

	$effect(() => {
		void load(current.kind);
	});

	/** Loads the tiles counting one kind of media. */
	async function load(kind: typeof current.kind) {
		loading = true;
		error = null;
		try {
			tiles = await fetchCategoryTiles(kind);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** The tiles of a group to show: all, or only those holding something. */
	function visible(group: TileGroup): CategoryTileRow[] {
		if (showEmpty) return group.children;
		return group.children.filter((tile) => tile.count > 0);
	}

	/** Opens a category's page. */
	function open(tile: CategoryTileRow) {
		goto(categoryUrl(tile.category.id, current.kind));
	}

	/** Goes to the new category's page. */
	function created(category: CategoryDetail) {
		creatingIn = null;
		goto(categoryUrl(category.id, current.kind));
	}

	/** What adding the defaults did. */
	function restoredMessage(added: number): string {
		if (added === 0) return 'Every default category is there';
		if (added === 1) return 'Added 1 category';
		return `Added ${added} categories`;
	}

	/** Adds back the defaults the library lacks, and says how many. */
	async function restoreDefaults() {
		restoring = true;
		try {
			showUndo({ message: restoredMessage(await addDefaultCategories()) });
			await load(current.kind);
		} catch (failure) {
			error = String(failure);
		} finally {
			restoring = false;
		}
	}
</script>

<svelte:head><title>Categories · Veil</title></svelte:head>

<div class="flex flex-col gap-8">
	<PageHeader title="Categories">
		{#snippet actions()}
			<button type="button" class="btn btn-sm" disabled={restoring} onclick={restoreDefaults}>
				<ArrowCounterClockwiseIcon size={16} /> Add missing defaults
			</button>
			<button type="button" class="btn btn-sm btn-primary" onclick={() => (creatingIn = '')}>
				<PlusIcon size={16} /> New category
			</button>
		{/snippet}
	</PageHeader>

	<div class="-mt-4 flex flex-wrap items-center justify-between gap-3">
		<CategoryKindTabs {current} hrefOf={(entry) => categoriesUrl(entry.kind)} />
		<label class="flex items-center gap-2 text-sm">
			<input type="checkbox" class="toggle toggle-sm" bind:checked={showEmpty} />
			Show empty
		</label>
	</div>

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if loading && tiles.length === 0}
		<p class="text-base-content/50 text-sm">Loading…</p>
	{/if}

	{#each filledGroups as group (group.parent.category.id)}
		{@const shown = visible(group)}
		{#if shown.length > 0 || showEmpty}
			<section class="flex flex-col gap-3">
				<div class="flex items-center gap-3">
					<a
						href={categoryUrl(group.parent.category.id, current.kind)}
						class="text-lg font-semibold hover:underline">{group.parent.category.name}</a
					>
					<span class="text-base-content/50 text-sm">{group.parent.count} {current.noun}</span>
					<button
						type="button"
						class="btn btn-ghost btn-xs ml-auto"
						onclick={() => (creatingIn = group.parent.category.id)}
					>
						<PlusIcon size={14} /> Add
					</button>
				</div>
				<div class="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-3">
					{#each shown as tile (tile.category.id)}
						<CategoryCard {tile} noun={current.noun} onclick={() => open(tile)} />
					{/each}
				</div>
			</section>
		{/if}
	{/each}

	{#if standalone.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">More</h2>
			<div class="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-3">
				{#each standalone as tile (tile.category.id)}
					{#if tile.count > 0 || showEmpty}
						<CategoryCard {tile} noun={current.noun} onclick={() => open(tile)} />
					{/if}
				{/each}
			</div>
		</section>
	{/if}

	{#if !loading && tiles.length > 0 && !showEmpty && tiles.every((tile) => tile.count === 0)}
		<p class="text-base-content/50 text-sm">
			No {current.noun} in any category yet. They fill as the library gets tagged {current.noun}.
		</p>
	{/if}
</div>

{#if creatingIn !== null}
	<CategoryEditor parentId={creatingIn} onsaved={created} onclose={() => (creatingIn = null)} />
{/if}
