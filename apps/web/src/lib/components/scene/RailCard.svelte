<script lang="ts">
	// Compact up-next card for the watch page rail: thumbnail on the left; title,
	// channel (with its avatar) and site · date on the right. Presentational — the
	// rail's FocusList handles click and Enter.
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
	let avatarUrl = $derived(cacheUrl(item.studioImagePath) || site?.iconUrl || null);
	let siteName = $derived.by(() => {
		if (!site) return null;
		if (site.displayName) return site.displayName;
		return site.name;
	});
	// The channel when known, else the site.
	let channelName = $derived(item.studioName || siteName);
</script>

<div class="group flex cursor-pointer gap-2">
	<div class="tv-card relative aspect-video w-[168px] shrink-0 rounded-lg!">
		{#if posterUrl}
			<img src={posterUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<span class="text-2xl opacity-10">🔞</span>
			</div>
		{/if}
		{#if durationLabel}
			<span
				class="absolute right-1 bottom-1 rounded bg-black/80 px-1 py-px text-[11px] font-medium text-white tabular-nums"
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
		<p class="line-clamp-2 text-[13px] leading-[18px] font-medium group-hover:text-white">
			{item.title}
		</p>
		{#if channelName}
			<p class="text-base-content/60 flex items-center gap-1.5 text-xs">
				<span class="bg-base-300 flex h-5 w-5 shrink-0 items-center justify-center overflow-hidden rounded-full">
					{#if avatarUrl}
						<img src={avatarUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
					{:else}
						<span class="text-[10px] font-semibold">{channelName.charAt(0).toUpperCase()}</span>
					{/if}
				</span>
				<span class="truncate">{channelName}</span>
			</p>
		{/if}
		<p class="text-base-content/50 flex gap-1 text-xs">
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
