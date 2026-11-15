<!--
	The top recommended galleries at the head of the galleries page: a row of
	large gallery cards. Cards stay near the size most covers are scraped at,
	so low-resolution covers aren't blown up. The first pick's cover lights the
	app's shared glow.
-->
<script lang="ts">
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import { cacheUrl } from '$lib/img';
	import type { GalleryCardRow } from '$lib/library';

	interface Props {
		// The picks, best first.
		galleries: GalleryCardRow[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		onopen: (gallery: GalleryCardRow) => void;
	}

	let { galleries, surface, onopen }: Props = $props();

	// Portrait covers: about five across a desktop window.
	const CARD_WIDTH = 280;

	showAmbientGlow(() => cacheUrl(galleries[0]?.cover?.uri));
</script>

<Shelf
	title="Top picks for you"
	items={galleries}
	key={(gallery) => gallery.id}
	itemWidth={CARD_WIDTH}
>
	{#snippet card(gallery, position)}
		<GalleryCard {gallery} shown={{ surface, position }} onclick={() => onopen(gallery)} />
	{/snippet}
</Shelf>
