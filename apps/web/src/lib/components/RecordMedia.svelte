<!--
	The videos and galleries of one performer or studio: the galleries as a
	row, then every video, a page more as you scroll. The page can add
	entries to each video's ⋮ menu (videoActions), e.g. taking the performer
	off it, which can drop the video from the list.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import FunnelSimpleIcon from 'phosphor-svelte/lib/FunnelSimpleIcon';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import type { SheetAction } from '$lib/components/ActionSheet.svelte';
	import {
		fetchVideos,
		fetchGalleries,
		type MediaFilter,
		type VideoCardRow,
		type GalleryCardRow
	} from '$lib/library';
	import { videoUrl, galleryUrl } from '$lib/routes';

	interface Props {
		filter: MediaFilter;
		// Home narrowed to this record, for "All videos" (filters and sorting).
		allVideosHref: string;
		// Extra ⋮ entries per video; remove drops the video from the list.
		videoActions?: (video: VideoCardRow, remove: () => void) => SheetAction[];
	}

	let { filter, allVideosHref, videoActions }: Props = $props();

	const PAGE_SIZE = 24;
	const GALLERY_WIDTH = 160;

	let videos = $state<VideoCardRow[]>([]);
	let galleries = $state<GalleryCardRow[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let loaded = $state(false);

	// Only the filter restarts the list: the load reads the list's own state,
	// which would otherwise restart it every time a page arrives.
	$effect(() => {
		const current = filter;
		untrack(() => restart(current));
	});

	/** Empties the list and loads the filter's first page and galleries. */
	function restart(current: MediaFilter) {
		videos = [];
		exhausted = false;
		loaded = false;
		void loadMore(current);
		fetchGalleries(current, 60).then((found) => (galleries = found));
	}

	/** Appends the next page of videos. */
	async function loadMore(current: MediaFilter = filter) {
		if (loading || exhausted) return;
		loading = true;
		try {
			const page = await fetchVideos(current, PAGE_SIZE, videos.length);
			videos = [...videos, ...page];
			exhausted = page.length < PAGE_SIZE;
		} finally {
			loading = false;
			loaded = true;
		}
	}

	/** Drops a video from the list, e.g. once the performer is taken off it. */
	function remove(video: VideoCardRow) {
		videos = videos.filter((other) => other.id !== video.id);
	}
</script>

<Shelf title="Galleries" items={galleries} key={(gallery) => gallery.id} itemWidth={GALLERY_WIDTH}>
	{#snippet card(gallery)}
		<GalleryCard {gallery} onclick={() => goto(galleryUrl(gallery.id))} />
	{/snippet}
</Shelf>

{#if videos.length > 0}
	<section class="flex flex-col gap-3">
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-lg font-semibold">Videos</h2>
			<a href={allVideosHref} class="chip chip-sm">
				<FunnelSimpleIcon size={14} /> Filter and sort
			</a>
		</div>
		<div class="card-grid">
			{#each videos as video, position (video.id)}
				<div class="rise-in" style:--rise-delay="{Math.min((position % PAGE_SIZE) * 25, 400)}ms">
					<SceneCard
						{video}
						actions={videoActions?.(video, () => remove(video)) ?? []}
						onclick={() => goto(videoUrl(video.id))}
					/>
				</div>
			{/each}
		</div>
		<ScrollSentinel onvisible={() => loadMore()} {loading} done={exhausted} />
	</section>
{/if}

{#if loaded && videos.length === 0 && galleries.length === 0}
	<p class="text-base-content/50 text-sm">Nothing of this in the library yet.</p>
{/if}
