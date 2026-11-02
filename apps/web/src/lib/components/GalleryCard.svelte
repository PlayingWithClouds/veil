<!--
	A gallery in a grid: portrait cover with the photo count, and the caption
	with its studio and site.
-->
<script lang="ts">
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import { cacheUrl } from '$lib/img';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import type { GalleryCardRow } from '$lib/library';
	import { logChosen, type Shown } from '$lib/events';
	import { impression } from '$lib/impression';

	interface Props {
		gallery: GalleryCardRow;
		// Where the card is shown; logs it as shown there and, on click, chosen.
		shown?: Shown | null;
		onclick: () => void;
	}

	let { gallery, shown = null, onclick }: Props = $props();

	let subject = $derived({ kind: 'gallery', id: gallery.id } as const);

	/** Logs the pick from its surface, then opens the gallery. */
	function open() {
		if (shown) logChosen(subject, shown);
		onclick();
	}

	let site = $derived.by(() => {
		const sourceUrl = gallery.sources[0]?.url;
		if (!sourceUrl) return null;
		return pluginForUrl(sourceUrl, $sites);
	});
	let photoCountLabel = $derived.by(() => {
		if (gallery.imageCount === 1) return '1 photo';
		return `${gallery.imageCount} photos`;
	});
</script>

<div class="feed-card group flex flex-col gap-2" use:impression={{ subject, shown }}>
	<CoverFrame
		src={cacheUrl(gallery.cover?.uri)}
		alt={gallery.title}
		aspect="3 / 4"
		placeholder={ImagesIcon}
		onclick={open}
	>
		{#if gallery.imageCount > 0}
			<CardBadge corner="bottom-right"><ImagesIcon size={12} /> {photoCountLabel}</CardBadge>
		{/if}
	</CoverFrame>

	<CardCaption title={gallery.title} byline={gallery.studio?.name} {site} onclick={open} />
</div>
