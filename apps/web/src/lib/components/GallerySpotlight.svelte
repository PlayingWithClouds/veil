<!--
	The top recommended galleries at the head of the galleries page: one
	fills a large frame (its portrait cover over a blurred copy of itself)
	with its title, photo count, studio and performers, and Open; beside it
	every pick is listed, and pointing at one for a moment, or clicking it,
	puts it in the frame. Clicking the featured pick, or the frame, opens the
	gallery. Its cover lights the app's shared glow.
-->
<script lang="ts">
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import { logChosen } from '$lib/events';
	import { fadeInImage } from '$lib/fadeInImage';
	import { cacheUrl } from '$lib/img';
	import { impression } from '$lib/impression';
	import type { GalleryCardRow } from '$lib/library';

	interface Props {
		// The picks, best first.
		galleries: GalleryCardRow[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		onopen: (gallery: GalleryCardRow) => void;
	}

	let { galleries, surface, onopen }: Props = $props();

	// Pointing at a pick this long puts it in the frame, so passing over the list doesn't.
	const SELECT_DELAY_MS = 350;

	let selectedId = $state<string | null>(null);
	let hoverTimer: ReturnType<typeof setTimeout> | null = null;

	let selected = $derived(galleries.find((gallery) => gallery.id === selectedId) ?? galleries[0]);
	let cover = $derived(cacheUrl(selected.cover?.uri));
	let facts = $derived(factsOf(selected));

	showAmbientGlow(() => cover);

	/** "120 photos · Studio · Performer, Performer", leaving out what's unknown. */
	function factsOf(gallery: GalleryCardRow): string {
		const parts: string[] = [];
		if (gallery.imageCount > 0) parts.push(photoCount(gallery.imageCount));
		if (gallery.studio) parts.push(gallery.studio.name);
		if (gallery.performers.length > 0) {
			parts.push(gallery.performers.map((performer) => performer.name).join(', '));
		}
		return parts.join(' · ');
	}

	/** "1 photo", "120 photos". */
	function photoCount(count: number): string {
		if (count === 1) return '1 photo';
		return `${count} photos`;
	}

	/** Logs the pick as chosen from the spotlight, then opens it. */
	function open(gallery: GalleryCardRow) {
		logChosen(
			{ kind: 'gallery', id: gallery.id },
			{ surface, position: galleries.indexOf(gallery) }
		);
		onopen(gallery);
	}

	/** Puts a pick in the frame; clicking the one already featured opens it. */
	function choose(gallery: GalleryCardRow) {
		cancelHover();
		if (gallery.id === selected.id) {
			open(gallery);
			return;
		}
		selectedId = gallery.id;
	}

	/** Puts a pick in the frame after the pointer rests on it. */
	function hoverStart(gallery: GalleryCardRow, event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		cancelHover();
		hoverTimer = setTimeout(() => (selectedId = gallery.id), SELECT_DELAY_MS);
	}

	/** Forgets a pending hover selection. */
	function cancelHover() {
		if (hoverTimer) clearTimeout(hoverTimer);
		hoverTimer = null;
	}
</script>

<section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)]">
	<div
		class="bg-base-300 relative aspect-video overflow-hidden rounded-2xl shadow-2xl ring-1 shadow-black/50 ring-white/10 max-md:aspect-[4/5]"
	>
		<button
			type="button"
			class="tv-card absolute inset-0 block h-full w-full"
			aria-label="Open {selected.title}"
			onclick={() => open(selected)}
		>
			{#key selected.id}
				{#if cover}
					<img
						src={cover}
						alt=""
						aria-hidden="true"
						class="absolute inset-0 h-full w-full scale-110 object-cover opacity-60 blur-2xl"
					/>
					<img
						src={cover}
						alt=""
						class="spotlight-cover absolute inset-y-0 right-[8%] h-full object-contain py-6 drop-shadow-2xl max-md:inset-x-0 max-md:right-0 max-md:mx-auto max-md:pb-32"
					/>
				{:else}
					<span class="absolute inset-0 flex items-center justify-center">
						<ImagesIcon size={48} class="opacity-20" />
					</span>
				{/if}
			{/key}
		</button>

		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-r from-black/80 via-black/30 to-transparent max-md:bg-gradient-to-t"
		></div>

		<div
			class="pointer-events-none absolute inset-y-0 left-0 flex max-w-md flex-col justify-end gap-2 p-6 max-md:inset-x-0 max-md:max-w-none max-md:p-4"
		>
			<p class="text-xs font-semibold tracking-wide text-white/90 uppercase">Top pick for you</p>
			<h2 class="line-clamp-3 text-3xl leading-tight font-bold text-white">{selected.title}</h2>
			{#if facts}
				<p class="text-sm text-white/70">{facts}</p>
			{/if}
			<div class="pointer-events-auto mt-2 flex gap-2">
				<button
					type="button"
					class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95"
					onclick={() => open(selected)}
				>
					<ImagesIcon size={16} weight="fill" /> Open
				</button>
			</div>
		</div>
	</div>

	<ol class="flex flex-col gap-2 max-lg:grid max-lg:grid-cols-2 max-md:grid-cols-1">
		{#each galleries as gallery, position (gallery.id)}
			<li
				class="min-h-0 lg:flex-1"
				use:impression={{
					subject: { kind: 'gallery', id: gallery.id },
					shown: { surface, position }
				}}
			>
				<button
					type="button"
					class="flex h-full w-full items-center gap-3 rounded-xl p-2 text-left transition-colors hover:bg-white/5"
					class:spotlight-on={gallery.id === selected.id}
					aria-current={gallery.id === selected.id}
					onclick={() => choose(gallery)}
					onpointerenter={(event) => hoverStart(gallery, event)}
					onpointerleave={cancelHover}
				>
					<span
						class="bg-base-300 relative aspect-[3/4] h-full max-h-28 min-h-16 shrink-0 overflow-hidden rounded-lg"
					>
						{#if gallery.cover}
							<img
								src={cacheUrl(gallery.cover.uri)}
								alt=""
								loading="lazy"
								use:fadeInImage
								class="h-full w-full object-cover"
							/>
						{/if}
					</span>
					<span class="flex min-w-0 flex-col gap-0.5">
						<span class="line-clamp-2 text-sm leading-snug font-semibold">{gallery.title}</span>
						<span class="text-base-content/45 truncate text-xs">
							{photoCount(gallery.imageCount)}{#if gallery.studio}
								· {gallery.studio.name}{/if}
						</span>
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

	/* The featured cover settles in when the pick changes. */
	.spotlight-cover {
		animation: feed-card-enter 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
	}

	@media (prefers-reduced-motion: reduce) {
		.spotlight-cover {
			animation: none;
		}
	}
</style>
