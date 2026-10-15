<script lang="ts">
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import { goto } from '$app/navigation';
	import { cacheUrl } from '$lib/img';
	import TagChips from '$lib/components/TagChips.svelte';
	import RatingButtons from '$lib/components/RatingButtons.svelte';
	import OCounterButton from '$lib/components/OCounterButton.svelte';
	import FavoriteButton from '$lib/components/FavoriteButton.svelte';
	import GalleryViewer from '$lib/components/GalleryViewer.svelte';
	import FaceIdentifyDialog from '$lib/components/FaceIdentifyDialog.svelte';
	import AlikeMatchesDialog from '$lib/components/AlikeMatchesDialog.svelte';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import { notifications } from '$lib/stores/notifications';
	import { findGalleryVideos, linkGalleryVideo, type AlikeCandidate } from '$lib/alike';
	import { sceneUrl } from '$lib/routes';
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

	// Viewer: index of the open image, or null when closed.
	let viewerIndex = $state<number | null>(null);

	// The gallery image "Who is this?" is asked about, or null when closed.
	let identifyImageId = $state<string | null>(null);

	// "Find the video": matches searched live across the sites.
	let videoSearchOpen = $state(false);
	let videoMatches = $state<AlikeCandidate[]>([]);
	let videoSearching = $state(false);

	/** Opens the matches dialog and searches the sites for the set's video. */
	async function findVideo() {
		if (!gallery) return;
		videoSearchOpen = true;
		videoSearching = true;
		videoMatches = await findGalleryVideos(gallery.id);
		videoSearching = false;
	}

	/** Links the picked match as the gallery's video and opens it. */
	async function pickVideo(candidate: AlikeCandidate) {
		if (!gallery) return;
		try {
			const sceneId = await linkGalleryVideo(gallery.id, candidate.plugin, candidate.sourceUrl);
			gallery = { ...gallery, video: { id: sceneId, title: candidate.title } };
			videoSearchOpen = false;
			await goto(sceneUrl(sceneId));
		} catch (error) {
			console.error('linkGalleryVideo failed', error);
			notifications.push('Could not open that video', 'error');
		}
	}
</script>

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

		<div class="flex flex-wrap items-center gap-2">
			<RatingButtons mediaId={gallery.id} />
			<OCounterButton mediaId={gallery.id} />
			<FavoriteButton mediaId={gallery.id} />
			<button
				type="button"
				class="flex items-center gap-1.5 rounded-full bg-white/10 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20"
				title="Add to collection"
				onclick={() => gallery && openMoveDialog('gallery', gallery.id, gallery.title)}
			>
				<ListPlusIcon size={18} />
				Save
			</button>
			{#if gallery.video}
				<a
					href={sceneUrl(gallery.video.id)}
					class="flex items-center gap-1.5 rounded-full bg-white px-4 py-2.5 text-sm font-semibold text-black transition-colors hover:bg-white/80"
					title={gallery.video.title}
				>
					<PlayIcon size={18} weight="fill" />
					Watch the video
				</a>
			{/if}
			<button
				type="button"
				class="flex items-center gap-1.5 rounded-full bg-white/10 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20"
				onclick={findVideo}
			>
				<FilmStripIcon size={18} />
				Find the video
			</button>
		</div>

		{#if images.length === 0}
			<div
				class="border-base-300 flex flex-col items-center gap-3 rounded-xl border border-dashed py-24 text-center"
			>
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
						onclick={() => (viewerIndex = index)}
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

	<AlikeMatchesDialog
		open={videoSearchOpen}
		heading="Find the video"
		subtitle={gallery.title}
		searchingText="Searching every site for this set's video…"
		emptyText="No matching videos found."
		candidates={videoMatches}
		loading={videoSearching}
		onClose={() => (videoSearchOpen = false)}
		pick={pickVideo}
	/>

	{#if viewerIndex !== null}
		<GalleryViewer
			{images}
			bind:index={viewerIndex}
			title={gallery.title}
			onclose={() => (viewerIndex = null)}
			onidentify={(image) => (identifyImageId = image.id)}
		/>
	{/if}

	<FaceIdentifyDialog
		open={identifyImageId !== null}
		subtitle={gallery.title}
		mediaId={gallery.id}
		imageId={identifyImageId}
		onClose={() => (identifyImageId = null)}
	/>
{/if}
