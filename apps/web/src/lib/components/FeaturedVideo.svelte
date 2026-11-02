<!--
	The top recommendation at the head of home: what it is on the left (why
	it's picked, title, length, quality, site, who is in it, tags, a line of
	its description, Play and Watchlist), its picture on the right, scrubbed
	on hover like a card (its clip plays when it has no frames). Starts from
	the card row and fills in from the video's page once that is read.
-->
<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import dayjs from 'dayjs';
	import ScrubPreview from './ScrubPreview.svelte';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { siteDisplayName, pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { addToWatchlist } from '$lib/curation';
	import { qualityLabel } from '$lib/videoCard';
	import { performerUrl, tagFilterUrl } from '$lib/routes';
	import { logChosen, type Shown } from '$lib/events';
	import { impression } from '$lib/impression';
	import type { VideoCardRow, VideoDetail } from '$lib/library';

	interface Props {
		video: VideoCardRow;
		// The same video read from its page: credits with pictures, tags, description.
		detail?: VideoDetail | null;
		// Why the engine picked it, e.g. "Because you watched …".
		reason?: string | null;
		shown?: Shown | null;
		onopen: () => void;
	}

	let { video, detail = null, reason = null, shown = null, onopen }: Props = $props();

	// Performers and tags named on the banner; the video's page lists them all.
	const SHOWN_PERFORMERS = 4;
	const SHOWN_TAGS = 6;

	let hovered = $state(false);
	let onWatchlist = $state(false);

	let subject = $derived({ kind: 'video', id: video.id } as const);
	let current = $derived<VideoCardRow>(detail ?? video);
	let frames = $derived(current.previewFrames ?? []);
	let canScrub = $derived(frames.length > 0 || current.trickplay !== null);
	let previewSrc = $derived(current.preview?.uri ?? null);
	let performers = $derived.by(() => {
		if (detail) return detail.performers.slice(0, SHOWN_PERFORMERS);
		return video.performers
			.slice(0, SHOWN_PERFORMERS)
			.map((performer) => ({ ...performer, image: null }));
	});
	let tags = $derived(detail?.tags.slice(0, SHOWN_TAGS) ?? []);
	let facts = $derived(
		[
			formatClock(current.durationSeconds),
			qualityLabel(current.width, current.height),
			current.studio?.name,
			siteOf(current),
			current.date ? dayjs(current.date).fromNow() : null
		].filter(Boolean)
	);

	/** The display name of the site the video came from. */
	function siteOf(row: VideoCardRow): string | null {
		const sourceUrl = row.sources[0]?.url;
		if (!sourceUrl) return null;
		const plugin = pluginForUrl(sourceUrl, $sites);
		if (!plugin) return null;
		return siteDisplayName(plugin.name, $sites);
	}

	/** Logs the pick from home, then opens the video. */
	function open() {
		if (shown) logChosen(subject, shown);
		onopen();
	}

	/** Puts the video on the watchlist. */
	async function saveForLater() {
		await addToWatchlist(video.id);
		onWatchlist = true;
	}

	$effect(() => {
		void video.id;
		onWatchlist = false;
	});
</script>

<section
	class="bg-base-200 grid h-[clamp(300px,40vh,420px)] grid-cols-[minmax(0,1fr)_minmax(0,1.45fr)] overflow-hidden rounded-2xl max-lg:h-auto max-lg:grid-cols-1"
	use:impression={{ subject, shown }}
>
	<div class="flex min-w-0 flex-col justify-center gap-3 p-8 max-lg:order-2 max-lg:p-5">
		<p class="text-base-content/60 truncate text-xs font-semibold tracking-wide uppercase">
			Top pick for you{#if reason}<span class="normal-case">&nbsp;· {reason}</span>{/if}
		</p>
		<button type="button" onclick={open} class="text-left">
			<h2 class="line-clamp-2 text-3xl leading-tight font-bold">{video.title}</h2>
		</button>
		{#if facts.length > 0}
			<p class="text-base-content/70 flex flex-wrap items-center gap-x-2 text-sm">
				{#each facts as fact, position (position)}
					{#if position > 0}<span aria-hidden="true">·</span>{/if}
					<span>{fact}</span>
				{/each}
			</p>
		{/if}
		{#if performers.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each performers as performer (performer.id)}
					<a
						href={performerUrl(performer.id)}
						class="bg-base-300 hover:bg-base-content/15 flex items-center gap-2 rounded-full py-1 pr-3 pl-1 text-sm"
					>
						<span class="bg-base-100 h-6 w-6 shrink-0 overflow-hidden rounded-full">
							{#if performer.image}
								<img
									src={cacheUrl(performer.image.uri)}
									alt=""
									class="h-full w-full object-cover"
								/>
							{/if}
						</span>
						{performer.name}
					</a>
				{/each}
			</div>
		{/if}
		{#if tags.length > 0}
			<div class="flex flex-wrap gap-1.5">
				{#each tags as tag (tag.id)}
					<a
						href={tagFilterUrl(tag.id)}
						class="border-base-content/15 text-base-content/70 hover:text-base-content rounded-full border px-2.5 py-0.5 text-xs"
					>
						{tag.name}
					</a>
				{/each}
			</div>
		{/if}
		{#if detail?.details}
			<p class="text-base-content/55 line-clamp-2 text-sm">{detail.details}</p>
		{/if}
		<div class="mt-1 flex gap-2">
			<button
				type="button"
				class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold"
				onclick={open}
			>
				<PlayIcon size={16} weight="fill" /> Play
			</button>
			<button
				type="button"
				class="bg-base-300 flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium disabled:opacity-60"
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

	<button
		type="button"
		onclick={open}
		onpointerenter={() => (hovered = true)}
		onpointerleave={() => (hovered = false)}
		class="tv-card relative min-h-0 overflow-hidden max-lg:order-1 max-lg:aspect-video"
		aria-label="Play {video.title}"
	>
		{#if current.cover}
			<img
				src={cacheUrl(current.cover.uri)}
				alt=""
				class="absolute inset-0 h-full w-full object-cover"
			/>
		{/if}
		{#if canScrub}
			<ScrubPreview
				{frames}
				trickplay={current.trickplay}
				durationSeconds={current.durationSeconds}
			/>
		{:else if hovered && previewSrc}
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
	</button>
</section>
