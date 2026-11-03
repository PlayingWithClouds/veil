<!--
	Galleries: the top recommended galleries in a spotlight, recently viewed,
	favourites, newest releases, performers and the most common tags, then the
	rest of an endless feed (recommended, then the rest of the library); a top
	bar search asks the gallery sites.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import GallerySpotlight from '$lib/components/GallerySpotlight.svelte';
	import MediaOverview from '$lib/components/MediaOverview.svelte';
	import SiteMediaList from '$lib/components/SiteMediaList.svelte';
	import {
		fetchFavoriteGalleries,
		fetchGalleryListing,
		fetchRecentlyViewedGalleries,
		galleryFeed
	} from '$lib/galleries';
	import { galleryUrl } from '$lib/routes';

	const CARD_WIDTH = 180;
</script>

<SiteMediaList
	title="Galleries"
	kind="GALLERY"
	noun="galleries"
	load={fetchGalleryListing}
	pick={(batch) => batch.listed.galleries}
	cardWidth={CARD_WIDTH}
	feed={() => galleryFeed()}
	feedTitle="More for you"
	feedSurface="galleries"
>
	{#snippet spotlight(picks)}
		<GallerySpotlight
			galleries={picks}
			surface="galleries"
			onopen={(gallery) => goto(galleryUrl(gallery.id))}
		/>
	{/snippet}
	{#snippet card(gallery, shown)}
		<GalleryCard {gallery} {shown} onclick={() => goto(galleryUrl(gallery.id))} />
	{/snippet}
	{#snippet overview()}
		<MediaOverview
			kind="gallery"
			continueTitle="Recently viewed"
			peopleTitle="Performers"
			fetchContinue={fetchRecentlyViewedGalleries}
			fetchFavorites={fetchFavoriteGalleries}
			load={fetchGalleryListing}
			itemWidth={CARD_WIDTH}
		>
			{#snippet mediaCard(gallery)}
				<GalleryCard {gallery} onclick={() => goto(galleryUrl(gallery.id))} />
			{/snippet}
		</MediaOverview>
	{/snippet}
</SiteMediaList>
