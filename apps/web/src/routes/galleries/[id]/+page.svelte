<!--
	One gallery. Opening it reads its pages, which list the images; a click
	shows an image in the lightbox, arrow keys step through them.
-->
<script lang="ts">
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import { openGallery, type GalleryDetail } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl } from '$lib/routes';

	let gallery = $state<GalleryDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let openIndex = $state<number | null>(null);

	let id = $derived(page.params.id ?? '');
	let openImage = $derived(openIndex === null ? null : (gallery?.images[openIndex] ?? null));

	$effect(() => {
		void load(id);
	});

	/** Opens the gallery, reading its pages when needed. */
	async function load(galleryId: string) {
		loading = true;
		error = null;
		try {
			gallery = await openGallery(galleryId);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Steps the lightbox with the arrow keys. */
	function onKeydown(event: KeyboardEvent) {
		if (openIndex === null || !gallery) return;
		if (event.key === 'ArrowRight') openIndex = Math.min(openIndex + 1, gallery.images.length - 1);
		if (event.key === 'ArrowLeft') openIndex = Math.max(openIndex - 1, 0);
	}
</script>

<svelte:window onkeydown={onKeydown} />
<svelte:head><title>{gallery?.title ?? 'Gallery'} · Veil</title></svelte:head>

<div class="flex flex-col gap-4 pt-2">
	{#if loading}
		<p class="text-base-content/50 flex items-center gap-2 text-sm">
			<span class="loading loading-spinner"></span> Reading the page…
		</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if gallery}
		<div class="flex flex-col gap-2">
			<h1 class="text-xl font-semibold">{gallery.title}</h1>
			<div class="text-base-content/60 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
				{#if gallery.studio}
					<a
						href={studioUrl(gallery.studio.id)}
						class="text-base-content font-medium hover:underline">{gallery.studio.name}</a
					>
				{/if}
				{#if gallery.date}<span>{gallery.date}</span>{/if}
				<span>{gallery.imageCount} photos</span>
				{#each gallery.sources as source (source.url)}
					<a
						href={source.url}
						target="_blank"
						rel="noreferrer"
						class="hover:text-base-content flex items-center gap-1"
					>
						<ArrowSquareOutIcon size={13} />
						{siteDisplayName(source.plugin, $sites)}
					</a>
				{/each}
			</div>
			{#if gallery.performers.length > 0}
				<div class="flex flex-wrap gap-2 text-sm">
					{#each gallery.performers as performer (performer.id)}
						<a
							href={performerUrl(performer.id)}
							class="bg-base-200 hover:bg-base-300 rounded-full px-3 py-1">{performer.name}</a
						>
					{/each}
				</div>
			{/if}
			<TagChips tags={gallery.tags} />
		</div>

		<div class="columns-2 gap-3 sm:columns-3 lg:columns-4 xl:columns-5">
			{#each gallery.images as image, index (image.id)}
				<button
					type="button"
					class="mb-3 block w-full overflow-hidden rounded-lg"
					onclick={() => (openIndex = index)}
				>
					<img src={cacheUrl(image.asset.uri)} alt="" loading="lazy" class="w-full" />
				</button>
			{/each}
		</div>

		{#if !loading && gallery.images.length === 0}
			<p class="text-base-content/50 text-sm">The site listed no images for this gallery.</p>
		{/if}
	{/if}
</div>

{#if openImage}
	<Lightbox
		src={cacheUrl(openImage.asset.uri) ?? ''}
		alt={gallery?.title}
		onclose={() => (openIndex = null)}
	/>
{/if}
