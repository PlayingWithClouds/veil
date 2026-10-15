<!--
	Galleries tab of a performer/studio page: the galleries credited to it, two
	columns on phones, loading more as you scroll.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import GalleryCard from './GalleryCard.svelte';
	import ScrollSentinel from './ScrollSentinel.svelte';
	import { entityFilter, fetchEntityGalleries, type GalleryHit } from '$lib/explore';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { galleryUrl } from '$lib/routes';

	interface Props {
		entityKind: 'performer' | 'studio';
		entityId: string;
	}

	let { entityKind, entityId }: Props = $props();

	const PAGE_SIZE = 40;

	let galleries = $state.raw<GalleryHit[]>([]);
	let loading = $state(true);
	let loadingMore = $state(false);
	let done = $state(false);
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
		done = false;
		fetchEntityGalleries(entityFilter(entityKind, id), PAGE_SIZE).then((found) => {
			if (requestedFor !== id) return;
			galleries = found;
			done = found.length < PAGE_SIZE;
			loading = false;
		});
	});

	/** Appends the next page of galleries. */
	async function loadMore() {
		if (loading || loadingMore || done) return;
		const id = entityId;
		loadingMore = true;
		const next = await fetchEntityGalleries(
			entityFilter(entityKind, id),
			PAGE_SIZE,
			galleries.length
		);
		loadingMore = false;
		if (requestedFor !== id) return;
		galleries = [...galleries, ...next];
		done = next.length < PAGE_SIZE;
	}
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
	<ScrollSentinel onvisible={loadMore} loading={loadingMore} {done} />
{/if}
