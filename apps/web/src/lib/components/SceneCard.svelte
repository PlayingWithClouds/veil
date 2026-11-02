<!--
	The shared video card for grids and rows: landscape cover, quality top left,
	the user's rating and favourite top right, duration bottom right, a resume
	bar for partly watched videos and a Watched mark for finished ones. Hovering
	scrubs through the preview frames; a video without frames plays its hover
	clip instead. Under it a YouTube-style caption with the performers and a ⋮
	menu (watchlist, favourite, not interested, block). The card hides when the
	video, its studio or a performer is blocked or dismissed on this page.
-->
<script lang="ts">
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import StarIcon from 'phosphor-svelte/lib/StarIcon';
	import CardCaption from './CardCaption.svelte';
	import CardBadge from './CardBadge.svelte';
	import CardMenu from './CardMenu.svelte';
	import ScrubPreview from './ScrubPreview.svelte';
	import dayjs from 'dayjs';
	import { formatClock } from '$lib/time';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import type { VideoCardRow } from '$lib/library';
	import { logChosen, type Shown } from '$lib/events';
	import { impression } from '$lib/impression';
	import { hiddenRecords, hiddenKey } from '$lib/userState';
	import { qualityLabel, videoCardActions } from '$lib/videoCard';
	import type { SheetAction } from './ActionSheet.svelte';

	interface Props {
		video: VideoCardRow;
		// A quiet last caption line, e.g. why the video was recommended.
		note?: string | null;
		// Where the card is shown; logs it as shown there and, on click, chosen.
		shown?: Shown | null;
		// Entries the page adds to the top of the ⋮ menu, e.g. "Not Mia Malkova".
		actions?: SheetAction[];
		onclick: () => void;
	}

	let { video, note = null, shown = null, actions = [], onclick }: Props = $props();

	let subject = $derived({ kind: 'video', id: video.id } as const);

	/** Logs the pick from its surface, then opens the video. */
	function open() {
		if (shown) logChosen(subject, shown);
		onclick();
	}

	// Hover and focus tracked apart: the hover player covering the card takes
	// focus from its button without the pointer leaving.
	let hovered = $state(false);
	let focused = $state(false);

	/** Starts the hover preview for a mouse or pen; a finger passing over while scrolling doesn't count. */
	function onPointerEnter(event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		hovered = true;
	}
	let previewing = $derived(hovered || focused);

	let coverSrc = $derived(cacheUrl(video.cover?.uri));
	// The image URL that failed to load; the placeholder shows instead of its alt text.
	let failedSrc = $state<string | null>(null);
	let frames = $derived(video.previewFrames ?? []);
	let canScrub = $derived(frames.length > 0 || video.trickplay !== null);
	let previewSrc = $derived(video.preview?.uri ?? null);
	let showVideo = $derived(previewing && previewSrc !== null && !canScrub);

	// Favourite changed from the menu; the row itself is the parent's.
	let favoriteOverride = $state<boolean | null>(null);
	let favorite = $derived(favoriteOverride ?? video.userState.favorite);
	let menuAnchor = $state<HTMLElement | null>(null);
	let menuActions = $derived([
		...actions,
		...videoCardActions(
			{ ...video, userState: { ...video.userState, favorite } },
			{ onfavorite: (value) => (favoriteOverride = value) }
		)
	]);

	let hidden = $derived.by(() => {
		const keys = $hiddenRecords;
		if (keys.size === 0) return false;
		if (keys.has(hiddenKey('video', video.id))) return true;
		if (video.studio && keys.has(hiddenKey('studio', video.studio.id))) return true;
		return video.performers.some((performer) => keys.has(hiddenKey('performer', performer.id)));
	});

	let durationLabel = $derived(formatClock(video.durationSeconds));
	let quality = $derived(qualityLabel(video.width, video.height));
	let rating = $derived(video.userState.rating);
	let site = $derived.by(() => {
		const sourceUrl = video.sources[0]?.url;
		if (!sourceUrl) return null;
		return pluginForUrl(sourceUrl, $sites);
	});
	let ageLabel = $derived.by(() => {
		if (!video.date) return null;
		return dayjs(video.date).fromNow();
	});
	// How far the user got, 0..1; null when not started or finished.
	let resume = $derived.by(() => {
		const position = video.userState.progress;
		if (!position || !video.durationSeconds) return null;
		return Math.min(position / video.durationSeconds, 1);
	});
	let watched = $derived(video.userState.viewCount > 0 && resume === null);
	// With performers on their own line, the studio joins the site line.
	let byline = $derived.by(() => {
		if (video.performers.length > 0) return null;
		return video.studio?.name ?? null;
	});
	let siteName = $derived.by(() => {
		if (video.performers.length === 0 || !video.studio) return null;
		const siteLabel = site?.displayName ?? site?.name;
		if (!siteLabel) return video.studio.name;
		return `${video.studio.name} · ${siteLabel}`;
	});
</script>

{#if !hidden}
	<div class="group flex w-full flex-col gap-3" use:impression={{ subject, shown }}>
		<div
			class="relative"
			role="presentation"
			onpointerenter={onPointerEnter}
			onpointerleave={() => (hovered = false)}
		>
			<button
				type="button"
				onclick={open}
				onfocus={() => (focused = true)}
				onblur={() => (focused = false)}
				class="tv-card relative block aspect-video w-full overflow-hidden ring-1 ring-white/5 transition-transform duration-200 group-hover:scale-[1.02] group-hover:shadow-xl focus:outline-none"
			>
				{#if coverSrc && coverSrc !== failedSrc}
					<img
						src={coverSrc}
						alt={video.title}
						loading="lazy"
						onerror={() => (failedSrc = coverSrc)}
						class="h-full w-full object-cover"
					/>
				{:else}
					<div class="flex h-full w-full items-center justify-center">
						<span class="text-5xl opacity-10">🔞</span>
					</div>
				{/if}

				{#if showVideo}
					<!-- svelte-ignore a11y_media_has_caption -->
					<video
						src={previewSrc}
						class="absolute inset-0 h-full w-full object-cover"
						autoplay
						muted
						loop
						playsinline
					></video>
				{/if}

				{#if quality}
					<CardBadge corner="top-left">{quality}</CardBadge>
				{/if}
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
				{#if watched}
					<CardBadge corner="bottom-left">Watched</CardBadge>
				{/if}
				{#if durationLabel}
					<CardBadge corner="bottom-right">{durationLabel}</CardBadge>
				{/if}

				{#if resume && !previewing}
					<span class="absolute inset-x-0 bottom-0 h-1 bg-white/25">
						<span class="bg-primary block h-full" style:width="{resume * 100}%"></span>
					</span>
				{/if}

				{#if canScrub}
					<ScrubPreview
						{frames}
						trickplay={video.trickplay}
						durationSeconds={video.durationSeconds}
					/>
				{/if}
			</button>
		</div>

		<CardCaption
			title={video.title}
			{byline}
			performers={video.performers}
			avatarUrl={cacheUrl(video.studio?.image?.uri)}
			{site}
			{siteName}
			detail={ageLabel}
			{note}
			onclick={open}
			onmore={(button) => (menuAnchor = button)}
		/>
	</div>

	{#if menuAnchor}
		<CardMenu
			anchor={menuAnchor}
			title={video.title}
			subtitle={video.studio?.name}
			imageUrl={coverSrc}
			actions={menuActions}
			onclose={() => (menuAnchor = null)}
		/>
	{/if}
{/if}
