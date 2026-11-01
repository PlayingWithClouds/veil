<!--
	The videos and galleries of one performer or studio, as two grids. Each
	loads its first page; the rest is on home filtered to the record.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import {
		fetchVideos,
		fetchGalleries,
		videoCard,
		type MediaFilter,
		type VideoCardRow,
		type GalleryCardRow
	} from '$lib/library';
	import { cacheUrl } from '$lib/img';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { videoUrl, galleryUrl } from '$lib/routes';

	interface Props {
		filter: MediaFilter;
		// Home narrowed to this record, for "All videos".
		allVideosHref: string;
	}

	let { filter, allVideosHref }: Props = $props();

	const PAGE_SIZE = 24;

	let videos = $state<VideoCardRow[]>([]);
	let galleries = $state<GalleryCardRow[]>([]);

	$effect(() => {
		const current = filter;
		fetchVideos(current, PAGE_SIZE).then((found) => (videos = found));
		fetchGalleries(current, PAGE_SIZE).then((found) => (galleries = found));
	});
</script>

{#if videos.length > 0}
	<section class="flex flex-col gap-3">
		<div class="flex items-baseline justify-between">
			<h2 class="text-lg font-semibold">Videos</h2>
			{#if videos.length === PAGE_SIZE}
				<a href={allVideosHref} class="text-primary text-sm">All videos</a>
			{/if}
		</div>
		<div
			class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-x-4 gap-y-6 max-md:grid-cols-1"
		>
			{#each videos as video (video.id)}
				<SceneCard item={videoCard(video)} onclick={() => goto(videoUrl(video.id))} />
			{/each}
		</div>
	</section>
{/if}

{#if galleries.length > 0}
	<section class="flex flex-col gap-3">
		<h2 class="text-lg font-semibold">Galleries</h2>
		<div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-x-4 gap-y-6">
			{#each galleries as gallery (gallery.id)}
				<GalleryCard
					gallery={{
						title: gallery.title,
						coverUrl: cacheUrl(gallery.cover?.uri),
						imageCount: gallery.imageCount,
						studioName: gallery.studio?.name
					}}
					site={pluginForUrl(gallery.sources[0]?.url ?? '', $sites)}
					onclick={() => goto(galleryUrl(gallery.id))}
				/>
			{/each}
		</div>
	</section>
{/if}

{#if videos.length === 0 && galleries.length === 0}
	<p class="text-base-content/50 text-sm">Nothing of this in the library yet.</p>
{/if}
