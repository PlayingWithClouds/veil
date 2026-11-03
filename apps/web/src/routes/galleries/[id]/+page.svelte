<!--
	One gallery. Opening it reads its pages, which list the images, and logs
	it as viewed. A banner over the blurred cover shows the cover, title,
	studio, performers, date, photo count and sites, with favourite, rating,
	the slideshow, the linked video and the ⋮ menu. Below: the description,
	the tags folded away, the photos in justified rows (a click opens the
	lightbox), related galleries, then an endless feed of more galleries
	(recommended, then the rest of the library).
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import PhotoGrid from '$lib/components/PhotoGrid.svelte';
	import RatingStars from '$lib/components/RatingStars.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import { openGallery, type GalleryCardRow, type GalleryDetail } from '$lib/library';
	import { fetchRelatedGalleries, galleryFeed } from '$lib/galleries';
	import type { Feed } from '$lib/feed';
	import { recordActions } from '$lib/recordActions';
	import { setFavorite, setRating } from '$lib/userState';
	import { logEvent, resetShown } from '$lib/events';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { galleryUrl, performerUrl, studioUrl, videoUrl } from '$lib/routes';

	const RELATED_COUNT = 12;
	const FEED_PAGE = 24;
	const CARD_WIDTH = 180;
	// Longer descriptions are cut until "Show more".
	const LONG_DETAILS = 280;
	const FEED_SURFACE = 'galleries';

	let gallery = $state<GalleryDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let openIndex = $state<number | null>(null);
	let slideshow = $state(false);
	let fullDetails = $state(false);
	let menuAnchor = $state<HTMLElement | null>(null);
	let related = $state<GalleryCardRow[]>([]);
	let more = $state<GalleryCardRow[]>([]);
	let moreLoading = $state(false);
	let moreDone = $state(false);

	// The feed under the gallery, once related galleries are known (it leaves them out).
	let feed: Feed<GalleryCardRow> | null = null;

	let id = $derived(page.params.id ?? '');
	let cover = $derived(cacheUrl(gallery?.cover?.uri));
	let photos = $derived(
		(gallery?.images ?? []).map((image) => ({
			id: image.id,
			src: cacheUrl(image.asset.uri) ?? '',
			width: image.width,
			height: image.height
		}))
	);
	let menuActions = $derived.by(() => {
		if (!gallery) return [];
		const shown = gallery;
		return recordActions(
			{
				kind: 'gallery',
				id: shown.id,
				favorite: shown.userState.favorite,
				performers: shown.performers,
				studio: shown.studio,
				sourceUrl: shown.sources[0]?.url ?? null
			},
			(favorite) => (shown.userState.favorite = favorite)
		);
	});

	$effect(() => {
		void load(id);
	});

	/** Opens the gallery, reading its pages when needed, and starts what's shown under it. */
	async function load(galleryId: string) {
		gallery = null;
		error = null;
		loading = true;
		openIndex = null;
		fullDetails = false;
		void loadBelow(galleryId);
		try {
			const found = await openGallery(galleryId);
			if (galleryId !== id) return;
			gallery = found;
			if (found) logEvent({ kind: 'view', subject: { kind: 'gallery', id: found.id } });
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Loads the related galleries, then the feed's first page without them. */
	async function loadBelow(galleryId: string) {
		related = [];
		more = [];
		moreDone = false;
		feed = null;
		resetShown(FEED_SURFACE);
		const found = await fetchRelatedGalleries(galleryId, RELATED_COUNT).catch(() => []);
		if (galleryId !== id) return;
		related = found;
		feed = galleryFeed([galleryId, ...found.map((other) => other.id)]);
		await loadMore();
	}

	/** Appends the feed's next page. */
	async function loadMore() {
		if (!feed || moreLoading || moreDone) return;
		const current = feed;
		moreLoading = true;
		try {
			const nextPage = await current(FEED_PAGE);
			if (current !== feed) return;
			more = [...more, ...nextPage];
			moreDone = nextPage.length === 0;
		} finally {
			moreLoading = false;
		}
	}

	/** Flips the favourite mark. */
	async function toggleFavorite() {
		if (!gallery) return;
		const favorite = !gallery.userState.favorite;
		gallery.userState.favorite = favorite;
		await setFavorite('gallery', gallery.id, favorite);
	}

	/** Rates the gallery, or clears its rating. */
	async function rate(rating: number | null) {
		if (!gallery) return;
		gallery.userState.rating = rating;
		await setRating('gallery', gallery.id, rating);
	}

	/** Opens the lightbox at a photo, playing the slideshow when asked. */
	function openPhoto(index: number, play = false) {
		slideshow = play;
		openIndex = index;
	}
</script>

<svelte:head><title>{gallery?.title ?? 'Gallery'} · Veil</title></svelte:head>

<div class="mx-auto flex w-full max-w-[1700px] flex-col gap-8 pt-2">
	{#if loading && !gallery}
		<p class="text-base-content/50 flex items-center gap-2 text-sm">
			<span class="loading loading-spinner"></span> Reading the page…
		</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if !loading && !gallery && !error}
		<p class="text-base-content/50 text-sm">This gallery isn't in the library.</p>
	{/if}

	{#if gallery}
		<section class="bg-base-200 relative overflow-hidden rounded-3xl">
			{#if cover}
				<img
					src={cover}
					alt=""
					aria-hidden="true"
					class="absolute inset-0 h-full w-full scale-125 object-cover opacity-40 blur-3xl"
				/>
			{/if}
			<div
				class="from-base-200 via-base-200/85 to-base-200/50 absolute inset-0 bg-linear-to-r"
			></div>

			<div class="relative flex gap-8 p-8 max-md:flex-col max-md:gap-5 max-md:p-5">
				<button
					type="button"
					class="bg-base-300 flex aspect-[3/4] w-56 shrink-0 items-center justify-center self-start overflow-hidden rounded-xl shadow-2xl max-md:w-40"
					aria-label="Open the first photo"
					disabled={photos.length === 0}
					onclick={() => openPhoto(0)}
				>
					{#if cover}
						<img src={cover} alt={gallery.title} class="h-full w-full object-cover" />
					{:else}
						<ImagesIcon size={56} class="text-base-content/20" />
					{/if}
				</button>

				<div class="flex min-w-0 flex-1 flex-col gap-4">
					<div class="flex flex-col gap-1.5">
						<h1 class="text-3xl leading-tight font-bold tracking-tight max-md:text-2xl">
							{gallery.title}
						</h1>
						{#if gallery.studio}
							<a
								href={studioUrl(gallery.studio.id)}
								class="text-base-content/70 hover:text-base-content self-start hover:underline"
								>{gallery.studio.name}</a
							>
						{/if}
					</div>

					{#if gallery.performers.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each gallery.performers as performer (performer.id)}
								<a
									href={performerUrl(performer.id)}
									class="bg-base-content/10 hover:bg-base-content/20 flex items-center gap-2 rounded-full py-1 pr-3 pl-1 text-sm"
								>
									<span
										class="bg-base-300 flex h-7 w-7 items-center justify-center overflow-hidden rounded-full text-xs font-semibold"
									>
										{#if performer.image}
											<img
												src={cacheUrl(performer.image.uri)}
												alt=""
												class="h-full w-full object-cover"
											/>
										{:else}
											{performer.name.charAt(0)}
										{/if}
									</span>
									{performer.name}
								</a>
							{/each}
						</div>
					{/if}

					<div class="flex flex-wrap items-center gap-2 text-xs">
						<span class="bg-base-content/10 rounded-full px-3 py-1">
							{gallery.imageCount}
							{gallery.imageCount === 1 ? 'photo' : 'photos'}
						</span>
						{#if gallery.date}
							<span class="bg-base-content/10 rounded-full px-3 py-1">{gallery.date}</span>
						{/if}
						{#each gallery.sources as source (source.url)}
							<a
								href={source.url}
								target="_blank"
								rel="noreferrer"
								class="bg-base-content/10 hover:bg-base-content/20 flex items-center gap-1 rounded-full px-3 py-1"
							>
								{siteDisplayName(source.plugin, $sites)}
								<ArrowSquareOutIcon size={12} />
							</a>
						{/each}
					</div>

					<div class="mt-auto flex flex-wrap items-center gap-2 pt-2">
						<button
							type="button"
							class="btn btn-sm btn-primary"
							disabled={photos.length === 0}
							onclick={() => openPhoto(0, true)}
						>
							<PlayIcon size={16} weight="fill" /> Slideshow
						</button>
						{#if gallery.video}
							<a href={videoUrl(gallery.video.id)} class="btn btn-sm" title={gallery.video.title}>
								<FilmSlateIcon size={16} /> Watch the video
							</a>
						{/if}
						<button
							type="button"
							class="btn btn-sm"
							aria-pressed={gallery.userState.favorite}
							onclick={toggleFavorite}
						>
							<HeartIcon
								size={16}
								weight={gallery.userState.favorite ? 'fill' : 'regular'}
								class={gallery.userState.favorite ? 'text-error' : ''}
							/>
							{gallery.userState.favorite ? 'Favourite' : 'Add to favourites'}
						</button>
						<RatingStars rating={gallery.userState.rating} onrate={rate} />
						<button
							type="button"
							class="btn btn-sm btn-square"
							aria-label="More"
							onclick={(event) => (menuAnchor = event.currentTarget)}
						>
							<DotsThreeVerticalIcon size={18} weight="bold" />
						</button>
					</div>
				</div>
			</div>
		</section>

		{#if gallery.details}
			<section class="flex max-w-prose flex-col items-start gap-2">
				<h2 class="text-lg font-semibold">About</h2>
				<p
					class="text-base-content/85 leading-relaxed whitespace-pre-line"
					class:line-clamp-6={!fullDetails}
				>
					{gallery.details}
				</p>
				{#if gallery.details.length > LONG_DETAILS}
					<button
						type="button"
						class="text-base-content/60 hover:text-base-content text-sm"
						onclick={() => (fullDetails = !fullDetails)}
					>
						{fullDetails ? 'Show less' : 'Show more'}
					</button>
				{/if}
			</section>
		{/if}

		{#if gallery.tags.length > 0}
			<details class="group">
				<summary class="flex cursor-pointer list-none items-center gap-2 text-lg font-semibold">
					Tags
					<span class="text-base-content/50 text-sm font-normal">{gallery.tags.length}</span>
					<CaretDownIcon size={14} class="transition-transform group-open:rotate-180" />
				</summary>
				<div class="pt-3">
					<TagChips tags={gallery.tags} />
				</div>
			</details>
		{/if}

		{#if photos.length > 0}
			<PhotoGrid {photos} onopen={(index) => openPhoto(index)} />
		{:else if !loading}
			<p class="text-base-content/50 text-sm">The site listed no images for this gallery.</p>
		{/if}
	{/if}

	{#if related.length > 0}
		<section class="flex flex-col gap-4">
			<h2 class="text-lg font-semibold">Related galleries</h2>
			<div
				class="grid gap-x-4 gap-y-6"
				style:grid-template-columns="repeat(auto-fill, minmax({CARD_WIDTH}px, 1fr))"
			>
				{#each related as other, position (other.id)}
					<GalleryCard
						gallery={other}
						shown={{ surface: 'related', position }}
						onclick={() => goto(galleryUrl(other.id))}
					/>
				{/each}
			</div>
		</section>
	{/if}

	{#if more.length > 0}
		<section class="flex flex-col gap-4">
			<h2 class="text-lg font-semibold">More galleries</h2>
			<div
				class="grid gap-x-4 gap-y-6"
				style:grid-template-columns="repeat(auto-fill, minmax({CARD_WIDTH}px, 1fr))"
			>
				{#each more as other, position (other.id)}
					<GalleryCard
						gallery={other}
						shown={{ surface: FEED_SURFACE, position }}
						onclick={() => goto(galleryUrl(other.id))}
					/>
				{/each}
			</div>
		</section>
	{/if}
	<ScrollSentinel onvisible={loadMore} loading={moreLoading && more.length > 0} done={moreDone} />
</div>

{#if gallery && openIndex !== null}
	<Lightbox
		images={photos}
		bind:index={openIndex}
		bind:playing={slideshow}
		alt={gallery.title}
		onclose={() => (openIndex = null)}
	/>
{/if}

{#if gallery && menuAnchor}
	<CardMenu
		anchor={menuAnchor}
		title={gallery.title}
		subtitle={gallery.studio?.name}
		imageUrl={cover}
		actions={menuActions}
		onclose={() => (menuAnchor = null)}
	/>
{/if}
