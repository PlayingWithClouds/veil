<script lang="ts">
	// Up-next card for the watch page rail: a large thumbnail (60% of the card)
	// on the left; title, channel and site · date on the right. Presentational —
	// the rail's FocusList handles click and Enter.
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import type { PluginSearchResult, SearchPlugin } from '$lib/search';
	import { formatAbsoluteDate, formatRelativeDate } from './format';

	interface Props {
		item: PluginSearchResult;
		// The site the item came from, matched by its source URL.
		site: SearchPlugin | null;
		// True while the item is being ingested before navigation.
		opening: boolean;
	}

	let { item, site, opening }: Props = $props();

	let posterUrl = $derived(cacheUrl(item.posterUrl));
	let durationLabel = $derived(formatClock(item.durationSeconds));
	let relativeDate = $derived(formatRelativeDate(item.date));
	let siteName = $derived.by(() => {
		if (!site) return null;
		if (site.displayName) return site.displayName;
		return site.name;
	});
	// The channel when known, else the site.
	let channelName = $derived(item.studioName || siteName);
</script>

<div class="group flex cursor-pointer gap-3">
	<div class="tv-card relative aspect-video w-[60%] min-w-[168px] shrink-0 rounded-xl!">
		{#if posterUrl}
			<img src={posterUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<span class="text-2xl opacity-10">🔞</span>
			</div>
		{/if}
		{#if durationLabel}
			<span
				class="absolute right-1.5 bottom-1.5 rounded bg-black/80 px-1 py-px text-xs font-medium text-white tabular-nums"
			>
				{durationLabel}
			</span>
		{/if}
		{#if opening}
			<div class="absolute inset-0 flex items-center justify-center bg-black/60" aria-hidden="true">
				<span class="loading loading-sm text-white"></span>
			</div>
		{/if}
	</div>
	<div class="flex min-w-0 flex-col gap-1">
		<p class="line-clamp-3 text-[15px] leading-[21px] font-medium group-hover:text-white">
			{item.title}
		</p>
		{#if channelName}
			<p class="text-base-content/60 truncate text-[13px]">{channelName}</p>
		{/if}
		<p class="text-base-content/50 flex gap-1 text-[13px]">
			{#if item.studioName && siteName}
				<span class="truncate">{siteName}</span>
			{/if}
			{#if item.studioName && siteName && relativeDate}
				<span aria-hidden="true">·</span>
			{/if}
			{#if relativeDate}
				<span class="shrink-0" title={formatAbsoluteDate(item.date)}>{relativeDate}</span>
			{/if}
		</p>
	</div>
</div>
