<script lang="ts">
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import { cacheUrl } from '$lib/img';
	import TagChips from '$lib/components/TagChips.svelte';
	import { ensureGalleryImages, type GalleryDetail } from '$lib/galleries';

	let { data } = $props();
	// Live copy, replaced once a stub's images have been fetched.
	let gallery = $state<GalleryDetail | null>(data.gallery);
	let images = $derived(gallery?.images ?? []);
	let fetchingImages = $state(false);

	// A gallery found through search/browse is a stub without images; opening it
	// fetches its page from the site.
	$effect(() => {
		if (!data.gallery || data.gallery.images.length > 0) return;
		let cancelled = false;
		fetchingImages = true;
		ensureGalleryImages(data.id)
			.then((fresh) => {
				if (!cancelled && fresh) gallery = fresh;
			})
			.catch((error) => console.error('ensureGalleryImages failed', data.id, error))
			.finally(() => {
				if (!cancelled) fetchingImages = false;
			});
		return () => {
			cancelled = true;
		};
	});

	// Lightbox: index of the open image, or null when closed.
	let lightboxIndex = $state<number | null>(null);
	let open = $derived(lightboxIndex !== null);

	function openAt(index: number) {
		lightboxIndex = index;
	}
	function closeLightbox() {
		lightboxIndex = null;
	}
	function step(delta: number) {
		if (lightboxIndex === null || images.length === 0) return;
		lightboxIndex = (lightboxIndex + delta + images.length) % images.length;
	}

	function onKeydown(event: KeyboardEvent) {
		if (lightboxIndex === null) return;
		if (event.key === 'Escape') closeLightbox();
		if (event.key === 'ArrowRight') step(1);
		if (event.key === 'ArrowLeft') step(-1);
	}
</script>

<svelte:window onkeydown={onKeydown} />

{#if !gallery}
	<div class="flex h-64 items-center justify-center">
		<span class="text-base-content/40">Gallery not found</span>
	</div>
{:else}
	<div class="flex w-full flex-col gap-6">
		<div class="flex items-center gap-3">
			<button
				type="button"
				class="btn btn-square btn-ghost btn-sm"
				aria-label="Back"
				onclick={() => history.back()}
			>
				<ArrowLeftIcon size={18} />
			</button>
			<div class="min-w-0">
				<h1 class="truncate text-2xl font-bold">{gallery.title}</h1>
				<p class="text-base-content/40 flex items-center gap-1.5 text-sm">
					<ImagesIcon size={14} />
					{images.length} images
					{#if gallery.date}<span class="opacity-30">·</span><span>{gallery.date}</span>{/if}
				</p>
				{#if gallery.tags.length > 0}
					<div class="mt-1">
						<TagChips tags={gallery.tags} />
					</div>
				{/if}
			</div>
			<span class="badge badge-error ml-auto">18+</span>
		</div>

		{#if images.length === 0}
			<div class="border-base-300 flex flex-col items-center gap-3 rounded-xl border border-dashed py-24 text-center">
				<span class="text-base-content/20"><ImagesIcon size={48} /></span>
				{#if fetchingImages}
					<span class="loading loading-spinner"></span>
					<p class="text-base-content/50 text-sm">Fetching images…</p>
				{:else}
					<p class="text-base-content/50 text-sm">Couldn't load this gallery's images.</p>
				{/if}
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
				{#each images as image, index (image.id)}
					<button
						type="button"
						class="bg-base-200 aspect-square overflow-hidden rounded-lg transition-transform hover:scale-[1.02]"
						onclick={() => openAt(index)}
					>
						<img
							src={cacheUrl(image.filePath)}
							alt="{gallery.title} {index + 1}"
							loading="lazy"
							class="h-full w-full object-cover"
						/>
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#if open && lightboxIndex !== null}
		<div
			class="fixed inset-0 z-[60] flex items-center justify-center bg-black/90"
			role="presentation"
			onclick={closeLightbox}
		>
			<button
				type="button"
				class="absolute right-4 top-4 flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
				onclick={closeLightbox}
				aria-label="Close"
			>
				<XIcon size={20} weight="bold" />
			</button>
			<button
				type="button"
				class="absolute left-4 flex h-12 w-12 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
				onclick={(event) => {
					event.stopPropagation();
					step(-1);
				}}
				aria-label="Previous"
			>
				<CaretLeftIcon size={24} weight="bold" />
			</button>
			<img
				src={cacheUrl(images[lightboxIndex].filePath)}
				alt="{gallery.title} {lightboxIndex + 1}"
				class="max-h-[92vh] max-w-[92vw] object-contain"
				onclick={(event) => event.stopPropagation()}
			/>
			<button
				type="button"
				class="absolute right-4 flex h-12 w-12 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
				onclick={(event) => {
					event.stopPropagation();
					step(1);
				}}
				aria-label="Next"
			>
				<CaretRightIcon size={24} weight="bold" />
			</button>
			<span
				class="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-black/70 px-3 py-1 text-xs text-white"
			>
				{lightboxIndex + 1} / {images.length}
			</span>
		</div>
	{/if}
{/if}
