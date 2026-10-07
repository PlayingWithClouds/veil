<script lang="ts">
	import { goto } from '$app/navigation';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import FolderSimpleIcon from 'phosphor-svelte/lib/FolderSimpleIcon';
	import RobotIcon from 'phosphor-svelte/lib/RobotIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { cacheUrl } from '$lib/img';
	import { collectionUrl } from '$lib/routes';
	import { fetchCollections, createCollection, type Collection } from '$lib/collections';

	let collections = $state<Collection[]>([]);
	let loading = $state(true);
	let newName = $state('');
	let creating = $state(false);

	let userCollections = $derived(collections.filter((entry) => entry.origin === 'user'));
	let scrapedCollections = $derived(collections.filter((entry) => entry.origin === 'scraped'));

	$effect(() => {
		load();
	});

	async function load() {
		loading = true;
		collections = await fetchCollections();
		loading = false;
	}

	async function create() {
		const name = newName.trim();
		if (!name || creating) return;
		creating = true;
		const collection = await createCollection(name);
		creating = false;
		newName = '';
		if (collection) collections = [collection, ...collections];
	}
</script>

{#snippet collectionGrid(entries: Collection[])}
	<div class="grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
		{#each entries as collection (collection.id)}
			<button
				type="button"
				class="flex flex-col gap-2 text-left"
				onclick={() => goto(collectionUrl(collection.id))}
			>
				<div
					class="tv-card bg-base-200 flex aspect-video w-full items-center justify-center overflow-hidden rounded-xl"
				>
					{#if cacheUrl(collection.coverPath)}
						<img
							src={cacheUrl(collection.coverPath)}
							alt={collection.name}
							loading="lazy"
							class="h-full w-full object-cover"
						/>
					{:else}
						<FolderSimpleIcon size={32} class="opacity-15" />
					{/if}
				</div>
				<span class="line-clamp-1 text-sm font-medium">{collection.name}</span>
				<span class="text-base-content/40 text-xs">{collection.itemCount} items</span>
			</button>
		{/each}
	</div>
{/snippet}

<div class="flex flex-col gap-6">
	<PageHeader title="Collections" description="Your playlists and the series sites group scenes into." />

	<!-- Create collection -->
	<form
		class="flex max-w-md items-center gap-2"
		onsubmit={(event) => {
			event.preventDefault();
			create();
		}}
	>
		<input
			type="text"
			bind:value={newName}
			placeholder="New collection name"
			maxlength={60}
			class="bg-base-200 border-base-300 focus:border-base-content/30 h-11 flex-1 rounded-lg border px-3 text-sm outline-none"
		/>
		<button
			type="submit"
			disabled={!newName.trim() || creating}
			class="bg-primary text-primary-content flex h-11 items-center gap-1 rounded-lg px-4 text-sm font-semibold disabled:opacity-40"
		>
			<PlusIcon size={16} weight="bold" /> Create
		</button>
	</form>

	{#if loading}
		<div class="flex justify-center py-16">
			<span class="loading loading-spinner text-base-content/40"></span>
		</div>
	{:else if collections.length === 0}
		<div class="flex flex-col items-center gap-3 py-20 text-center">
			<span class="text-6xl opacity-10">📁</span>
			<p class="text-base-content/40">No collections yet. Create one above or save a scene.</p>
		</div>
	{:else}
		{#if userCollections.length > 0}
			{@render collectionGrid(userCollections)}
		{/if}

		{#if scrapedCollections.length > 0}
			<div class="mt-2 flex items-center gap-2">
				<RobotIcon size={18} class="text-base-content/40" />
				<h2 class="text-base-content/70 text-lg font-semibold">From sources</h2>
			</div>
			{@render collectionGrid(scrapedCollections)}
		{/if}
	{/if}
</div>
