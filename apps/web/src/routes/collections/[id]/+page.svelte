<!--
	One collection: its name, details and its videos and galleries in order.
-->
<script lang="ts">
	import { page } from '$app/state';
	import CollectionGrid from '$lib/components/CollectionGrid.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { fetchCollection, type CollectionPage } from '$lib/curation';

	let collection = $state<CollectionPage | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		const id = page.params.id ?? '';
		loading = true;
		error = null;
		fetchCollection(id)
			.then((found) => (collection = found))
			.catch((failure) => (error = String(failure)))
			.finally(() => (loading = false));
	});
</script>

<svelte:head><title>{collection?.name ?? 'Collection'} · Veil</title></svelte:head>

{#if error}
	<p class="text-error text-sm">{error}</p>
{:else if loading}
	<div class="flex justify-center py-10">
		<span class="loading loading-spinner"></span>
	</div>
{:else if collection}
	<div class="flex flex-col gap-6">
		<PageHeader
			title={collection.name}
			description={collection.details ?? `${collection.itemCount} items`}
		/>
		{#key collection.id}
			<CollectionGrid {collection} surface="collection" empty="Nothing in this collection yet." />
		{/key}
	</div>
{:else}
	<p class="text-base-content/50 py-10 text-center text-sm">This collection is gone.</p>
{/if}
