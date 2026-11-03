<!--
	The top recommended videos at the head of home: one fills a large 16:9
	frame (its preview clip plays muted while the frame is on screen; without
	a clip, hovering scrubs its frames) with why it's picked, its title, what
	it is, Play and Watchlist; beside it every pick is listed, and pointing at
	one for a moment, or clicking it, puts it in the frame. Clicking the
	featured pick, or the frame, opens the video. Its cover lights the app's
	shared glow.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import ScrubPreview from '$lib/components/ScrubPreview.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import { addToWatchlist } from '$lib/curation';
	import { logChosen } from '$lib/events';
	import { fadeInImage } from '$lib/fadeInImage';
	import type { RecommendedVideo } from '$lib/home';
	import { cacheUrl } from '$lib/img';
	import { impression } from '$lib/impression';
	import type { VideoCardRow } from '$lib/library';
	import { formatClock } from '$lib/time';
	import { qualityLabel } from '$lib/videoCard';

	interface Props {
		// The picks, best first, each with why it's picked.
		picks: RecommendedVideo[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		onopen: (video: VideoCardRow) => void;
	}

	let { picks, surface, onopen }: Props = $props();

	// Pointing at a pick this long puts it in the frame, so passing over the list doesn't.
	const SELECT_DELAY_MS = 350;

	let selectedId = $state<string | null>(null);
	let frame = $state<HTMLElement | null>(null);
	let onScreen = $state(false);
	let hovered = $state(false);
	// Videos put on the watchlist from here.
	let saved = $state<string[]>([]);
	let hoverTimer: ReturnType<typeof setTimeout> | null = null;

	let selected = $derived(picks.find((pick) => pick.video.id === selectedId) ?? picks[0]);
	let video = $derived(selected.video);
	let cover = $derived(cacheUrl(video.cover?.uri));
	let previewSrc = $derived(video.preview?.uri ?? null);
	let canScrub = $derived((video.previewFrames ?? []).length > 0 || video.trickplay !== null);
	let facts = $derived(factsOf(video));

	showAmbientGlow(() => cover);

	// The clip plays only while the frame shows, so a scrolled-down page costs nothing.
	$effect(() => {
		if (!frame) return;
		const observer = new IntersectionObserver(
			(entries) => {
				onScreen = entries.some((entry) => entry.isIntersecting);
			},
			{ threshold: 0.25 }
		);
		observer.observe(frame);
		return () => observer.disconnect();
	});

	/** "12:34 · 1080p · Studio · 3 days ago", leaving out what's unknown. */
	function factsOf(row: VideoCardRow): string {
		const parts: string[] = [];
		const length = formatClock(row.durationSeconds);
		if (length) parts.push(length);
		const quality = qualityLabel(row.width, row.height);
		if (quality) parts.push(quality);
		if (row.studio) parts.push(row.studio.name);
		if (row.date) parts.push(dayjs(row.date).fromNow());
		return parts.join(' · ');
	}

	/** Logs the pick as chosen from the spotlight, then opens it. */
	function open(pick: RecommendedVideo) {
		logChosen({ kind: 'video', id: pick.video.id }, { surface, position: picks.indexOf(pick) });
		onopen(pick.video);
	}

	/** Puts a pick in the frame; clicking the one already featured opens it. */
	function choose(pick: RecommendedVideo) {
		cancelHover();
		if (pick.video.id === selected.video.id) {
			open(pick);
			return;
		}
		selectedId = pick.video.id;
	}

	/** Puts a pick in the frame after the pointer rests on it. */
	function hoverStart(pick: RecommendedVideo, event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		cancelHover();
		hoverTimer = setTimeout(() => (selectedId = pick.video.id), SELECT_DELAY_MS);
	}

	/** Forgets a pending hover selection. */
	function cancelHover() {
		if (hoverTimer) clearTimeout(hoverTimer);
		hoverTimer = null;
	}

	/** Puts the featured video on the watchlist. */
	async function saveForLater() {
		const videoId = video.id;
		await addToWatchlist(videoId);
		saved = [...saved, videoId];
	}
</script>

<section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)]">
	<div
		bind:this={frame}
		class="bg-base-300 relative aspect-video overflow-hidden rounded-2xl shadow-2xl ring-1 shadow-black/50 ring-white/10"
	>
		<button
			type="button"
			class="tv-card absolute inset-0 block h-full w-full"
			aria-label="Play {video.title}"
			onclick={() => open(selected)}
			onpointerenter={() => (hovered = true)}
			onpointerleave={() => (hovered = false)}
		>
			{#key video.id}
				{#if cover}
					<img src={cover} alt="" class="absolute inset-0 h-full w-full object-cover" />
				{/if}
				{#if previewSrc && onScreen}
					<video
						src={previewSrc}
						class="absolute inset-0 h-full w-full object-cover"
						autoplay
						muted
						loop
						playsinline
					></video>
				{:else if canScrub && hovered}
					<ScrubPreview
						frames={video.previewFrames ?? []}
						trickplay={video.trickplay}
						durationSeconds={video.durationSeconds}
					/>
				{/if}
			{/key}
		</button>

		<div
			class="pointer-events-none absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/85 via-black/40 to-transparent"
		></div>

		<div
			class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between gap-4 p-6 max-md:flex-col max-md:items-start max-md:p-4"
		>
			<div class="min-w-0">
				<p class="text-xs font-semibold tracking-wide text-white/90 uppercase">
					Top pick for you{#if selected.reason}<span class="text-white/65 normal-case">
							· {selected.reason}</span
						>{/if}
				</p>
				<h2 class="mt-1.5 line-clamp-2 text-3xl leading-tight font-bold text-white">
					{video.title}
				</h2>
				{#if facts}
					<p class="mt-1 text-sm text-white/70">{facts}</p>
				{/if}
			</div>
			<div class="pointer-events-auto flex shrink-0 gap-2">
				<button
					type="button"
					class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95"
					onclick={() => open(selected)}
				>
					<PlayIcon size={16} weight="fill" /> Play
				</button>
				<button
					type="button"
					class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-60"
					disabled={saved.includes(video.id)}
					onclick={saveForLater}
				>
					{#if saved.includes(video.id)}
						<CheckIcon size={16} /> On watchlist
					{:else}
						<PlusIcon size={16} /> Watchlist
					{/if}
				</button>
			</div>
		</div>
	</div>

	<ol class="flex flex-col gap-2 max-lg:grid max-lg:grid-cols-2 max-md:grid-cols-1">
		{#each picks as pick, position (pick.video.id)}
			<li
				class="min-h-0 lg:flex-1"
				use:impression={{
					subject: { kind: 'video', id: pick.video.id },
					shown: { surface, position }
				}}
			>
				<button
					type="button"
					class="flex h-full w-full items-center gap-3 rounded-xl p-2 text-left transition-colors hover:bg-white/5"
					class:spotlight-on={pick.video.id === video.id}
					aria-current={pick.video.id === video.id}
					onclick={() => choose(pick)}
					onpointerenter={(event) => hoverStart(pick, event)}
					onpointerleave={cancelHover}
				>
					<span
						class="bg-base-300 relative aspect-video h-full max-h-24 min-h-14 shrink-0 overflow-hidden rounded-lg"
					>
						{#if pick.video.cover}
							<img
								src={cacheUrl(pick.video.cover.uri)}
								alt=""
								loading="lazy"
								use:fadeInImage
								class="h-full w-full object-cover"
							/>
						{/if}
					</span>
					<span class="flex min-w-0 flex-col gap-0.5">
						<span class="line-clamp-2 text-sm leading-snug font-semibold">{pick.video.title}</span>
						{#if pick.reason}
							<span class="text-base-content/45 truncate text-xs">{pick.reason}</span>
						{/if}
					</span>
				</button>
			</li>
		{/each}
	</ol>
</section>

<style>
	.spotlight-on {
		background-color: color-mix(in oklab, var(--color-base-content) 8%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-base-content) 14%, transparent);
	}
</style>
