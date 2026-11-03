<!--
	One gallery. Opening it reads its pages, which list the images, and logs
	it as viewed. Its cover lights the shared glow; a header shows the cover,
	title, studio, photo count, date and performers, with the slideshow, the
	linked video, favourite, rating and the ⋮ menu, and the sites it's on.
	Below: the description, the tags, the photos in justified rows (a click opens the
	lightbox), related galleries, then an endless feed of more galleries
	(recommended, then the rest of the library).
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
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
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';

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

	showAmbientGlow(() => cover);

	$effect(() => {
		void load(id);
	});

	/** How long a card waits before rising in, so a row of cards cascades. */
	function riseDelay(position: number): string {
		return `${Math.min((position % FEED_PAGE) * 25, 400)}ms`;
	}

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
		<div class="flex gap-7 max-md:flex-col" aria-label="Reading the page…">
			<div class="skeleton aspect-[3/4] w-56 shrink-0 rounded-3xl max-md:w-40"></div>
			<div class="flex flex-1 flex-col justify-end gap-3">
				<div class="skeleton h-9 w-2/3 rounded-lg"></div>
				<div class="skeleton h-4 w-1/4 rounded"></div>
				<div class="skeleton h-10 w-80 rounded-full"></div>
			</div>
		</div>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if !loading && !gallery && !error}
		<p class="text-base-content/50 text-sm">This gallery isn't in the library.</p>
	{/if}

	{#if gallery}
		<header class="flex gap-7 max-md:flex-col max-md:items-center max-md:text-center">
			<button
				type="button"
				class="group bg-base-300 relative flex aspect-[3/4] w-56 shrink-0 items-center justify-center overflow-hidden rounded-3xl shadow-2xl ring-1 shadow-black/50 ring-white/10 max-md:w-44"
				aria-label="Open the first photo"
				disabled={photos.length === 0}
				onclick={() => openPhoto(0)}
			>
				{#if cover}
					<img
						src={cover}
						alt={gallery.title}
						class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
					/>
				{:else}
					<ImagesIcon size={56} class="text-base-content/20" />
				{/if}
			</button>

			<div class="flex min-w-0 flex-1 flex-col justify-end gap-4">
				<div class="flex flex-col gap-1">
					<h1 class="text-4xl leading-tight font-bold tracking-tight max-md:text-2xl">
						{gallery.title}
					</h1>
					<p
						class="text-base-content/60 flex flex-wrap items-center gap-x-2 text-sm max-md:justify-center"
					>
						{#if gallery.studio}
							<a href={studioUrl(gallery.studio.id)} class="text-base-content hover:underline">
								{gallery.studio.name}
							</a>
							<span aria-hidden="true">·</span>
						{/if}
						<span>
							{gallery.imageCount}
							{gallery.imageCount === 1 ? 'photo' : 'photos'}
						</span>
						{#if gallery.date}
							<span aria-hidden="true">·</span>
							<span>{gallery.date}</span>
						{/if}
					</p>
				</div>

				{#if gallery.performers.length > 0}
					<div class="flex flex-wrap gap-2 max-md:justify-center">
						{#each gallery.performers as performer (performer.id)}
							<a href={performerUrl(performer.id)} class="chip pl-1">
								<span
									class="flex h-6 w-6 items-center justify-center overflow-hidden rounded-full bg-white/10 text-xs font-semibold"
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

				<div class="flex flex-wrap items-center gap-2 max-md:justify-center">
					<button
						type="button"
						class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-50"
						disabled={photos.length === 0}
						onclick={() => openPhoto(0, true)}
					>
						<PlayIcon size={16} weight="fill" /> Slideshow
					</button>
					{#if gallery.video}
						<a
							href={videoUrl(gallery.video.id)}
							class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
							title={gallery.video.title}
						>
							<FilmSlateIcon size={16} /> Watch the video
						</a>
					{/if}
					<button
						type="button"
						class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
						aria-pressed={gallery.userState.favorite}
						onclick={toggleFavorite}
					>
						<span class="flex" class:text-error={gallery.userState.favorite}>
							<HeartIcon size={18} weight={gallery.userState.favorite ? 'fill' : 'regular'} />
						</span>
						Favourite
					</button>
					<div class="glass-pill flex h-10 items-center rounded-full px-3">
						<RatingStars rating={gallery.userState.rating} onrate={rate} />
					</div>
					<button
						type="button"
						class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
						aria-label="More"
						onclick={(event) => (menuAnchor = event.currentTarget)}
					>
						<DotsThreeVerticalIcon size={20} weight="bold" />
					</button>
				</div>

				{#if gallery.sources.length > 0}
					<div class="flex flex-wrap items-center gap-2 max-md:justify-center">
						<span class="text-base-content/45 mr-1 text-xs font-semibold tracking-wide uppercase">
							Found on
						</span>
						{#each gallery.sources as source (source.url)}
							<a href={source.url} target="_blank" rel="noreferrer" class="chip chip-sm">
								{siteDisplayName(source.plugin, $sites)}
								<ArrowSquareOutIcon size={12} class="opacity-60" />
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</header>

		{#if gallery.details}
			<section class="flex max-w-prose flex-col items-start gap-2">
				<p
					class="text-base-content/85 text-[15px] leading-relaxed whitespace-pre-line"
					class:line-clamp-6={!fullDetails}
				>
					{gallery.details}
				</p>
				{#if gallery.details.length > LONG_DETAILS}
					<button type="button" class="chip chip-sm" onclick={() => (fullDetails = !fullDetails)}>
						{fullDetails ? 'Show less' : 'Show more'}
					</button>
				{/if}
			</section>
		{/if}

		<TagChips tags={gallery.tags} />

		{#if photos.length > 0}
			<section class="flex flex-col gap-4">
				<div class="flex items-center justify-between gap-3">
					<h2 class="flex items-baseline gap-2 text-lg font-semibold">
						Photos
						<span class="text-base-content/45 text-sm font-normal">{photos.length}</span>
					</h2>
					<button type="button" class="chip chip-sm" onclick={() => openPhoto(0, true)}>
						<PlayIcon size={12} weight="fill" /> Slideshow
					</button>
				</div>
				<div class="rise-in">
					<PhotoGrid {photos} onopen={(index) => openPhoto(index)} />
				</div>
			</section>
		{:else if !loading}
			<p class="text-base-content/50 text-sm">The site listed no images for this gallery.</p>
		{/if}
	{/if}

	{#if related.length > 0}
		<section class="flex flex-col gap-4">
			<h2 class="flex items-baseline gap-2 text-lg font-semibold">
				Related galleries
				<span class="text-base-content/45 text-sm font-normal">same people, studio or look</span>
			</h2>
			<div
				class="grid gap-x-4 gap-y-6"
				style:grid-template-columns="repeat(auto-fill, minmax({CARD_WIDTH}px, 1fr))"
			>
				{#each related as other, position (other.id)}
					<div class="rise-in" style:--rise-delay={riseDelay(position)}>
						<GalleryCard
							gallery={other}
							shown={{ surface: 'related', position }}
							onclick={() => goto(galleryUrl(other.id))}
						/>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	{#if more.length > 0}
		<section class="flex flex-col gap-4">
			<h2 class="flex items-baseline gap-2 text-lg font-semibold">
				More galleries
				<span class="text-base-content/45 text-sm font-normal">for you</span>
			</h2>
			<div
				class="grid gap-x-4 gap-y-6"
				style:grid-template-columns="repeat(auto-fill, minmax({CARD_WIDTH}px, 1fr))"
			>
				{#each more as other, position (other.id)}
					<div class="rise-in" style:--rise-delay={riseDelay(position)}>
						<GalleryCard
							gallery={other}
							shown={{ surface: FEED_SURFACE, position }}
							onclick={() => goto(galleryUrl(other.id))}
						/>
					</div>
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
