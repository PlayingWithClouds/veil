<!--
	The user's own collections as stacks, and a field to start a new one,
	which opens once it's created.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import CollectionCard from '$lib/components/CollectionCard.svelte';
	import { createCollection, fetchCollections, type CollectionCardRow } from '$lib/curation';
	import { collectionUrl } from '$lib/routes';

	let collections = $state<CollectionCardRow[] | null>(null);
	let error = $state<string | null>(null);
	let name = $state('');
	let creating = $state(false);

	$effect(() => {
		fetchCollections()
			.then((found) => (collections = found))
			.catch((failure) => (error = String(failure)));
	});

	/** Creates a collection by the typed name and opens it. */
	async function create(event: SubmitEvent) {
		event.preventDefault();
		const trimmed = name.trim();
		if (trimmed === '') return;
		creating = true;
		try {
			const id = await createCollection(trimmed);
			await goto(collectionUrl(id));
		} catch (failure) {
			error = String(failure);
			creating = false;
		}
	}
</script>

<form class="flex max-w-md gap-2" onsubmit={create}>
	<input class="input w-full" bind:value={name} placeholder="New collection" />
	<button type="submit" class="btn" disabled={creating || name.trim() === ''}>
		<PlusIcon size={16} />
		Create
	</button>
</form>

{#if error}
	<p class="text-error text-sm">{error}</p>
{/if}

{#if collections === null}
	<div class="flex justify-center py-10">
		<span class="loading loading-spinner"></span>
	</div>
{:else if collections.length === 0}
	<p class="text-base-content/50 py-10 text-center text-sm">
		No collections yet. Name one above, then add videos to it.
	</p>
{:else}
	<div class="card-grid">
		{#each collections as collection (collection.id)}
			<CollectionCard {collection} onclick={() => goto(collectionUrl(collection.id))} />
		{/each}
	</div>
{/if}
