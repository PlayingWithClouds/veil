<!--
	A gallery in a grid: portrait cover with the photo count and a favourite
	or rating badge, and the caption with its performers, studio, site and a
	⋮ menu (favourite, not interested, block). The card hides when the
	gallery, its studio or a performer is blocked or dismissed.
-->
<script lang="ts">
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import StarIcon from 'phosphor-svelte/lib/StarIcon';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import { cacheUrl } from '$lib/img';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import type { GalleryCardRow } from '$lib/library';
	import { logChosen, type Shown } from '$lib/events';
	import { impression } from '$lib/impression';
	import { recordActions } from '$lib/recordActions';
	import { hiddenRecords, hiddenKey } from '$lib/userState';

	interface Props {
		gallery: GalleryCardRow;
		// Where the card is shown; logs it as shown there and, on click, chosen.
		shown?: Shown | null;
		onclick: () => void;
	}

	let { gallery, shown = null, onclick }: Props = $props();

	let subject = $derived({ kind: 'gallery', id: gallery.id } as const);

	// Favourite changed from the menu; the row itself is the parent's.
	let favoriteOverride = $state<boolean | null>(null);
	let favorite = $derived(favoriteOverride ?? gallery.userState.favorite);
	let rating = $derived(gallery.userState.rating);
	let menuAnchor = $state<HTMLElement | null>(null);
	let menuActions = $derived(
		recordActions(
			{
				kind: 'gallery',
				id: gallery.id,
				favorite,
				performers: gallery.performers,
				studio: gallery.studio,
				sourceUrl: gallery.sources[0]?.url ?? null
			},
			(value) => (favoriteOverride = value)
		)
	);

	let hidden = $derived.by(() => {
		const keys = $hiddenRecords;
		if (keys.size === 0) return false;
		if (keys.has(hiddenKey('gallery', gallery.id))) return true;
		if (gallery.studio && keys.has(hiddenKey('studio', gallery.studio.id))) return true;
		return gallery.performers.some((performer) => keys.has(hiddenKey('performer', performer.id)));
	});

	/** Logs the pick from its surface, then opens the gallery. */
	function open() {
		if (shown) logChosen(subject, shown);
		onclick();
	}

	let cover = $derived(cacheUrl(gallery.cover?.uri));
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

{#if !hidden}
	<div class="feed-card group flex flex-col gap-2" use:impression={{ subject, shown }}>
		<CoverFrame
			src={cover}
			alt={gallery.title}
			aspect="3 / 4"
			placeholder={ImagesIcon}
			onclick={open}
		>
			{#if rating !== null || favorite}
				<CardBadge corner="top-right">
					{#if favorite}
						<HeartIcon size={12} weight="fill" class="text-error" />
					{/if}
					{#if rating !== null}
						<StarIcon size={12} weight="fill" class="text-warning" />
						{Math.round(rating * 10) / 10}
					{/if}
				</CardBadge>
			{/if}
			{#if gallery.imageCount > 0}
				<CardBadge corner="bottom-right"><ImagesIcon size={12} /> {photoCountLabel}</CardBadge>
			{/if}
		</CoverFrame>

		<CardCaption
			title={gallery.title}
			byline={gallery.studio?.name}
			performers={gallery.performers}
			{site}
			onclick={open}
			onmore={(button) => (menuAnchor = button)}
		/>
	</div>

	{#if menuAnchor}
		<CardMenu
			anchor={menuAnchor}
			title={gallery.title}
			subtitle={gallery.studio?.name}
			imageUrl={cover}
			actions={menuActions}
			onclose={() => (menuAnchor = null)}
		/>
	{/if}
{/if}
