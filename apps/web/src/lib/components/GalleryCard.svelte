<!--
	A gallery in a grid: portrait cover with the photo count, and the caption
	with its studio and site.
-->
<script lang="ts">
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import { fadeInImage } from '$lib/fadeInImage';
	import type { SitePlugin } from '$lib/plugins';

	/** What a gallery card shows. */
	interface GalleryCardData {
		title: string;
		coverUrl: string | null;
		imageCount: number;
		studioName?: string | null;
	}

	interface Props {
		gallery: GalleryCardData;
		// The site the gallery came from, shown under the title.
		site?: SitePlugin | null;
		onclick: () => void;
	}

	let { gallery, site = null, onclick }: Props = $props();

	let photoCountLabel = $derived.by(() => {
		if (gallery.imageCount === 1) return '1 photo';
		return `${gallery.imageCount} photos`;
	});
</script>

<div class="feed-card group flex flex-col gap-2">
	<button
		type="button"
		{onclick}
		class="tv-card bg-base-300 relative block aspect-[3/4] w-full overflow-hidden ring-1 ring-white/5 transition-transform duration-200 group-hover:scale-[1.02] group-hover:shadow-xl focus:outline-none"
	>
		{#if gallery.coverUrl}
			<img
				src={gallery.coverUrl}
				alt={gallery.title}
				loading="lazy"
				use:fadeInImage
				class="h-full w-full object-cover"
			/>
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<ImagesIcon size={40} class="opacity-20" />
			</div>
		{/if}

		{#if gallery.imageCount > 0}
			<span
				class="absolute right-1.5 bottom-1.5 flex items-center gap-1 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white"
			>
				<ImagesIcon size={12} />
				{photoCountLabel}
			</span>
		{/if}
	</button>

	<CardCaption title={gallery.title} byline={gallery.studioName} {site} {onclick} />
</div>
