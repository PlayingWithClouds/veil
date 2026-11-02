<!--
	The top of home: one video large (its clip plays muted while hovered), with
	why it's there, Play and Watchlist; beside it a short "Up next" list.
-->
<script lang="ts" module>
	import type { VideoCardRow } from '$lib/library';

	/** A video with the label saying why it's here ("Continue", "For you", …). */
	export interface FocusItem {
		video: VideoCardRow;
		label: string;
	}
</script>

<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { siteDisplayName, pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { addToWatchlist } from '$lib/curation';

	interface Props {
		focus: FocusItem;
		// A specific reason for the focus video, e.g. "Because you watched …".
		reason?: string | null;
		upNext: FocusItem[];
		onopen: (video: VideoCardRow) => void;
	}

	let { focus, reason = null, upNext, onopen }: Props = $props();

	let hovered = $state(false);
	let onWatchlist = $state(false);

	let video = $derived(focus.video);
	let previewSrc = $derived(video.preview?.uri ?? null);
	let siteName = $derived.by(() => {
		const sourceUrl = video.sources[0]?.url;
		if (!sourceUrl) return null;
		const plugin = pluginForUrl(sourceUrl, $sites);
		if (!plugin) return null;
		return siteDisplayName(plugin.name, $sites);
	});
	let meta = $derived(
		[video.studio?.name, siteName, formatClock(video.durationSeconds)].filter(Boolean).join(' · ')
	);
	// How much of a started video is left, for "Continue".
	let remaining = $derived.by(() => {
		const position = video.userState.progress;
		if (!position || !video.durationSeconds) return null;
		return formatClock(video.durationSeconds - position);
	});

	/** Puts the focus video on the watchlist. */
	async function saveForLater() {
		await addToWatchlist(video.id);
		onWatchlist = true;
	}

	$effect(() => {
		void video.id;
		onWatchlist = false;
	});
</script>

<section class="grid grid-cols-[2fr_1fr] gap-5 max-lg:grid-cols-1">
	<div
		class="tv-card bg-base-300 relative aspect-video overflow-hidden rounded-2xl"
		role="presentation"
		onpointerenter={() => (hovered = true)}
		onpointerleave={() => (hovered = false)}
	>
		{#if video.cover}
			<img src={cacheUrl(video.cover.uri)} alt="" class="absolute inset-0 h-full w-full object-cover" />
		{/if}
		{#if hovered && previewSrc}
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
		<div
			class="absolute inset-0 flex flex-col justify-end gap-2 bg-gradient-to-t from-black/85 via-black/30 to-transparent p-6"
		>
			<span class="text-primary text-xs font-semibold tracking-wide uppercase">
				{focus.label}{#if remaining}&nbsp;· {remaining} left{/if}
			</span>
			<button type="button" onclick={() => onopen(video)} class="text-left">
				<h2 class="line-clamp-2 max-w-3xl text-2xl font-bold text-white">{video.title}</h2>
			</button>
			{#if meta}
				<p class="text-sm text-white/70">{meta}</p>
			{/if}
			{#if reason}
				<p class="text-sm text-white/50">{reason}</p>
			{/if}
			<div class="mt-2 flex gap-2">
				<button
					type="button"
					class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold"
					onclick={() => onopen(video)}
				>
					<PlayIcon size={16} weight="fill" /> Play
				</button>
				<button
					type="button"
					class="flex h-10 items-center gap-2 rounded-full bg-white/15 px-4 text-sm font-medium text-white backdrop-blur disabled:opacity-60"
					disabled={onWatchlist}
					onclick={saveForLater}
				>
					{#if onWatchlist}
						<CheckIcon size={16} /> On watchlist
					{:else}
						<PlusIcon size={16} /> Watchlist
					{/if}
				</button>
			</div>
		</div>
	</div>

	{#if upNext.length > 0}
		<div class="flex flex-col gap-2">
			<h2 class="text-base-content/60 text-sm font-semibold">Up next</h2>
			{#each upNext as item (item.video.id)}
				<button
					type="button"
					class="hover:bg-base-200 flex items-center gap-3 rounded-xl p-1.5 text-left transition-colors"
					onclick={() => onopen(item.video)}
				>
					<span class="bg-base-300 relative aspect-video w-36 shrink-0 overflow-hidden rounded-lg">
						{#if item.video.cover}
							<img
								src={cacheUrl(item.video.cover.uri)}
								alt=""
								loading="lazy"
								class="h-full w-full object-cover"
							/>
						{/if}
					</span>
					<span class="flex min-w-0 flex-col gap-0.5">
						<span class="text-primary text-[11px] font-semibold uppercase">{item.label}</span>
						<span class="line-clamp-2 text-sm leading-5 font-medium">{item.video.title}</span>
					</span>
				</button>
			{/each}
		</div>
	{/if}
</section>
