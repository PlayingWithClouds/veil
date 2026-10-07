<script lang="ts">
	import { goto } from '$app/navigation';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import TagSimpleIcon from 'phosphor-svelte/lib/TagSimpleIcon';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import { cacheUrl } from '$lib/img';
	import { fetchImages, type Image } from '$lib/images';
	import { sceneUrl, performerUrl, studioUrl, galleryUrl, collectionUrl } from '$lib/routes';
	import {
		fetchScenesForTag,
		fetchGalleriesForTag,
		fetchPerformersForTag,
		fetchStudiosForTag,
		fetchCollectionsForTag,
		type TaggedScenes,
		type TaggedGallery,
		type TaggedPerformer,
		type TaggedStudio,
		type TaggedCollection
	} from '$lib/tags';

	let { data } = $props();
	let tag = $derived(data.tag);

	let scenes = $state<TaggedScenes>({ direct: [], inherited: [] });
	let galleries = $state<TaggedGallery[]>([]);
	let images = $state<Image[]>([]);
	let performers = $state<TaggedPerformer[]>([]);
	let studios = $state<TaggedStudio[]>([]);
	let collections = $state<TaggedCollection[]>([]);
	let loading = $state(true);
	let lightboxImage = $state<Image | null>(null);

	let loadedFor = $state<string | null>(null);
	$effect(() => {
		if (loadedFor === data.id) return;
		loadedFor = data.id;
		load(data.id);
	});

	async function load(tagId: string) {
		loading = true;
		[scenes, galleries, images, performers, studios, collections] = await Promise.all([
			fetchScenesForTag(tagId),
			fetchGalleriesForTag(tagId),
			fetchImages({ tagId }, 30),
			fetchPerformersForTag(tagId),
			fetchStudiosForTag(tagId),
			fetchCollectionsForTag(tagId)
		]);
		loading = false;
	}

	let empty = $derived(
		!loading &&
			scenes.direct.length === 0 &&
			scenes.inherited.length === 0 &&
			galleries.length === 0 &&
			images.length === 0 &&
			performers.length === 0 &&
			studios.length === 0 &&
			collections.length === 0
	);
</script>

{#snippet tileGrid(entries: { id: string; label: string; image: string | null; caption: string; href: string }[])}
	<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
		{#each entries as entry (entry.id)}
			<button type="button" class="flex flex-col gap-2 text-left" onclick={() => goto(entry.href)}>
				<div
					class="tv-card bg-base-200 flex aspect-video w-full items-center justify-center overflow-hidden rounded-xl"
				>
					{#if cacheUrl(entry.image)}
						<img
							src={cacheUrl(entry.image)}
							alt={entry.label}
							loading="lazy"
							class="h-full w-full object-cover"
						/>
					{:else}
						<TagSimpleIcon size={28} class="opacity-15" />
					{/if}
				</div>
				<span class="line-clamp-1 text-sm font-medium">{entry.label}</span>
				<span class="text-base-content/40 text-xs">{entry.caption}</span>
			</button>
		{/each}
	</div>
{/snippet}

<div class="flex flex-col gap-8">
	<div class="flex flex-col gap-2">
		<div class="flex items-center gap-3">
			<button
				type="button"
				class="btn btn-square btn-ghost btn-sm"
				aria-label="Back to tags"
				onclick={() => goto('/tags')}
			>
				<ArrowLeftIcon size={18} />
			</button>
			<TagSimpleIcon size={26} weight="fill" class="text-base-content/70" />
			<h1 class="text-3xl font-bold">{tag?.name ?? 'Tag'}</h1>
			{#if tag?.category}
				<span class="badge">{tag.category}</span>
			{/if}
			{#if tag}
				<SubscribeButton kind="TAG" targetId={tag.id} targetName={tag.name} />
			{/if}
		</div>
		{#if tag?.aliases?.length}
			<p class="text-base-content/40 text-sm">Also known as: {tag.aliases.join(', ')}</p>
		{/if}
		{#if tag?.description}
			<p class="text-base-content/60 max-w-2xl text-sm">{tag.description}</p>
		{/if}
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<span class="loading loading-spinner text-base-content/40"></span>
		</div>
	{:else if empty}
		<div class="flex flex-col items-center gap-3 py-20 text-center">
			<span class="text-6xl opacity-10">🏷️</span>
			<p class="text-base-content/40">Nothing carries this tag yet.</p>
		</div>
	{:else}
		{#if scenes.direct.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Scenes</h2>
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each scenes.direct as scene (scene.id)}
						<SceneCard item={scene} onclick={() => scene.id && goto(sceneUrl(scene.id))} />
					{/each}
				</div>
			</section>
		{/if}

		{#if scenes.inherited.length > 0}
			<section class="flex flex-col gap-3">
				<div class="flex items-baseline gap-2">
					<h2 class="text-lg font-semibold">Scenes via studio / performer</h2>
					<span class="text-base-content/40 text-xs">tag carried by the studio or a performer</span>
				</div>
				<div class="grid grid-cols-1 gap-4 opacity-90 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each scenes.inherited as scene (scene.id)}
						<SceneCard item={scene} onclick={() => scene.id && goto(sceneUrl(scene.id))} />
					{/each}
				</div>
			</section>
		{/if}

		{#if galleries.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Galleries</h2>
				{@render tileGrid(
					galleries.map((gallery) => ({
						id: gallery.id,
						label: gallery.title,
						image: gallery.coverPath,
						caption: `${gallery.imageCount} images`,
						href: galleryUrl(gallery.id)
					}))
				)}
			</section>
		{/if}

		{#if images.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Images</h2>
				<div class="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8">
					{#each images as image (image.id)}
						<button
							type="button"
							class="bg-base-200 aspect-square overflow-hidden rounded-lg"
							onclick={() => (lightboxImage = image)}
						>
							{#if cacheUrl(image.filePath)}
								<img
									src={cacheUrl(image.filePath)}
									alt={image.title ?? ''}
									loading="lazy"
									class="h-full w-full object-cover transition-transform duration-200 hover:scale-105"
								/>
							{/if}
						</button>
					{/each}
				</div>
			</section>
		{/if}

		{#if performers.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Performers</h2>
				{@render tileGrid(
					performers.map((performer) => ({
						id: performer.id,
						label: performer.name,
						image: performer.imagePath,
						caption: `${performer.sceneCount} scenes`,
						href: performerUrl(performer.id)
					}))
				)}
			</section>
		{/if}

		{#if studios.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Studios</h2>
				{@render tileGrid(
					studios.map((studio) => ({
						id: studio.id,
						label: studio.name,
						image: studio.imagePath,
						caption: `${studio.sceneCount} scenes`,
						href: studioUrl(studio.id)
					}))
				)}
			</section>
		{/if}

		{#if collections.length > 0}
			<section class="flex flex-col gap-3">
				<h2 class="text-lg font-semibold">Collections</h2>
				{@render tileGrid(
					collections.map((collection) => ({
						id: collection.id,
						label: collection.name,
						image: collection.coverPath,
						caption: `${collection.itemCount} items`,
						href: collectionUrl(collection.id)
					}))
				)}
			</section>
		{/if}
	{/if}
</div>

{#if lightboxImage && cacheUrl(lightboxImage.filePath)}
	<Lightbox
		src={cacheUrl(lightboxImage.filePath) ?? ''}
		alt={lightboxImage.title ?? ''}
		onclose={() => (lightboxImage = null)}
	/>
{/if}
