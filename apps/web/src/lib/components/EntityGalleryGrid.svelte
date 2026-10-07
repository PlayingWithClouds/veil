<!--
	Galleries tab of a performer/studio page: the galleries credited to it, two
	columns on phones.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import GalleryCard from './GalleryCard.svelte';
	import { entityFilter, fetchEntityGalleries, type GalleryHit } from '$lib/explore';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { galleryUrl } from '$lib/routes';

	interface Props {
		entityKind: 'performer' | 'studio';
		entityId: string;
	}

	let { entityKind, entityId }: Props = $props();

	let galleries = $state.raw<GalleryHit[]>([]);
	let loading = $state(true);
	let sites = $state<SearchPlugin[]>([]);

	$effect(() => {
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	// Answers for a previous entity are dropped.
	let requestedFor = '';
	$effect(() => {
		const id = entityId;
		requestedFor = id;
		loading = true;
		fetchEntityGalleries(entityFilter(entityKind, id), 60).then((found) => {
			if (requestedFor !== id) return;
			galleries = found;
			loading = false;
		});
	});
</script>

{#if loading}
	<div class="grid grid-cols-2 gap-3 md:grid-cols-4">
		{#each { length: 4 } as _, index (index)}
			<div class="skeleton aspect-[3/4] rounded-[var(--tv-radius)]"></div>
		{/each}
	</div>
{:else if galleries.length === 0}
	<div class="text-base-content/40 flex flex-col items-center gap-3 py-16 text-center text-sm">
		<ImagesIcon size={40} class="opacity-30" />
		No galleries yet.
	</div>
{:else}
	<div class="grid grid-cols-2 gap-3 md:grid-cols-4">
		{#each galleries as gallery (gallery.id)}
			<GalleryCard
				{gallery}
				site={pluginForUrl(gallery.sourceUrl, sites)}
				onclick={() => goto(galleryUrl(gallery.id))}
			/>
		{/each}
	</div>
{/if}
