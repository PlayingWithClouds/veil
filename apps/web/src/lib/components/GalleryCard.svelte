<script module lang="ts">
	/** Width / height of the cover. Fixed so the masonry knows the card's height before the image loads. */
	export const GALLERY_COVER_ASPECT = 3 / 4;
</script>

<script lang="ts">
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import { fadeInImage } from '$lib/fadeInImage';
	import { quickSaveEntity } from '$lib/collectionActions';
	import type { GalleryHit } from '$lib/explore';
	import type { SearchPlugin } from '$lib/search';

	interface Props {
		gallery: GalleryHit;
		// The card tag best matching the user's taste, shown top-left. Null hides it.
		preferredTag?: string | null;
		// The site the gallery came from, shown under the title.
		site?: SearchPlugin | null;
		onclick: () => void;
	}

	let { gallery, preferredTag = null, site = null, onclick }: Props = $props();

	let imageCountLabel = $derived.by(() => {
		if (gallery.imageCount <= 0) return null;
		return `${gallery.imageCount} images`;
	});

	let saved = $state(false);
	/** Quick-saves the gallery to the default collection. */
	function save() {
		quickSaveEntity('gallery', gallery.id, gallery.title);
		saved = true;
	}
</script>

<div class="feed-card group flex flex-col gap-2">
	<div class="relative">
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

			{#if preferredTag}
				<span
					class="bg-primary/85 absolute top-1.5 left-1.5 rounded px-1.5 py-0.5 text-[10px] font-semibold text-white shadow"
					title="Matches your taste"
				>
					{preferredTag}
				</span>
			{/if}

			{#if gallery.imageCount > 0}
				<span
					class="absolute right-1.5 bottom-1.5 flex items-center gap-1 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white"
				>
					<ImagesIcon size={12} />
					{gallery.imageCount}
				</span>
			{/if}
		</button>

		<!-- Actions appear on hover or keyboard/TV focus only. -->
		<div
			class="absolute top-1.5 right-1.5 flex gap-1 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100"
		>
			<button
				type="button"
				onclick={save}
				aria-label="Add to watchlist"
				title="Add to watchlist"
				class="flex h-8 w-8 items-center justify-center rounded-md bg-black/70 text-white backdrop-blur-sm transition-colors hover:bg-black/90"
				class:text-primary={saved}
			>
				{#if saved}
					<CheckIcon size={16} weight="bold" />
				{:else}
					<PlusIcon size={16} />
				{/if}
			</button>
		</div>
	</div>

	<CardCaption title={gallery.title} {site} detail={imageCountLabel} {onclick} />
</div>
