<!--
	A collection's videos and galleries in its order, each video's ⋮ menu
	able to take it out again. Used by the watchlist, the queue and every
	collection's page.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import { entryId, removeFromCollection, type CollectionPage } from '$lib/curation';
	import { galleryUrl, videoUrl } from '$lib/routes';

	interface Props {
		collection: CollectionPage;
		// Where the cards are shown, for the impression log.
		surface: string;
		// Shown while the collection is empty.
		empty: string;
	}

	let { collection, surface, empty }: Props = $props();

	// Entries taken out on this page, hidden until the collection is loaded again.
	let removed = $state<Set<string>>(new Set());
	let error = $state<string | null>(null);

	let entries = $derived(collection.entries.filter((entry) => !removed.has(entryId(entry))));

	/** Takes an entry out of the collection, hiding it at once. */
	async function remove(entry: (typeof entries)[number]) {
		const id = entryId(entry);
		removed = new Set([...removed, id]);
		try {
			await removeFromCollection(collection.id, entry);
		} catch (failure) {
			removed = new Set([...removed].filter((other) => other !== id));
			error = String(failure);
		}
	}
</script>

{#if error}
	<p class="text-error text-sm">{error}</p>
{/if}

{#if entries.length === 0}
	<p class="text-base-content/50 py-10 text-center text-sm">{empty}</p>
{:else}
	<div class="card-grid">
		{#each entries as entry, position (entryId(entry))}
			{#if entry.kind === 'video'}
				<SceneCard
					video={entry.video}
					shown={{ surface, position }}
					actions={[
						{ label: `Remove from ${collection.name}`, icon: XIcon, run: () => remove(entry) }
					]}
					onclick={() => goto(videoUrl(entry.video.id))}
				/>
			{:else}
				<GalleryCard
					gallery={entry.gallery}
					shown={{ surface, position }}
					onclick={() => goto(galleryUrl(entry.gallery.id))}
				/>
			{/if}
		{/each}
	</div>
{/if}
