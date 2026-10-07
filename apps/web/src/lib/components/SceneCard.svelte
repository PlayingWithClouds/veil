<script lang="ts" module>
	// The shared scene card for grids (history, library, subscriptions, tags, …).
	// Same look as Home's cards: landscape banner with a hover preview (cycling
	// frame strip + optional clip), duration badge bottom-right, a resume bar for
	// partially-watched scenes, and a YouTube-style caption under it.

	export interface SceneCardData {
		// Local scene id, when the scene is already ingested. Enables the resume bar
		// lookup; absent for fresh plugin results that haven't been scraped yet.
		id?: string;
		title: string;
		bannerUrl: string | null;
		previewImages?: string[];
		previewVideo?: string | null;
		// Source plugin icon, the caption avatar when there is no channel or site.
		providerIconUrl?: string | null;
		durationSeconds?: number | null;
		// Channel/studio and its image, known once the scene's page was visited.
		channelName?: string | null;
		channelImageUrl?: string | null;
		// The page the scene came from; its host names the site in the caption.
		sourceUrl?: string | null;
		// Release date, shown as "3 weeks ago".
		date?: string | null;
		// True when the card matched a tag filter via its studio/performer rather
		// than its own tags; rendered as a muted corner badge.
		inheritedTag?: boolean;
		// True for feed entries that arrived since the last visit; shows a "New" badge.
		isNew?: boolean;
	}

	// Formats a duration as a clock: "9:05", "12:34", "1:02:03".
	export function formatClock(seconds: number | null | undefined): string | null {
		if (!seconds || seconds <= 0) return null;
		const total = Math.round(seconds);
		const hours = Math.floor(total / 3600);
		const minutes = Math.floor((total % 3600) / 60);
		const secs = total % 60;
		const mm = hours > 0 ? String(minutes).padStart(2, '0') : String(minutes);
		const ss = String(secs).padStart(2, '0');
		if (hours > 0) return `${hours}:${mm}:${ss}`;
		return `${mm}:${ss}`;
	}
</script>

<script lang="ts">
	import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
	import CardCaption from './CardCaption.svelte';
	import CardHoverPlayer from './CardHoverPlayer.svelte';
	import { centerPreview } from '$lib/centerPreview';
	import { watchProgress } from '$lib/stores/watchProgress';
	import { quickSaveEntity } from '$lib/collectionActions';
	import { formatRelativeTime } from '$lib/time';
	import { pluginForUrl } from '$lib/search';
	import { sites } from '$lib/stores/sites';

	interface Props {
		item: SceneCardData;
		onclick: () => void;
	}

	let { item, onclick }: Props = $props();

	/** Quick-saves to the last-used collection without opening the card. */
	function save() {
		if (item.id) quickSaveEntity('scene', item.id, item.title);
	}

	let frames = $derived(item.previewImages ?? []);
	// Hover and focus tracked apart: the hover player covering the card takes
	// focus from its button without the pointer leaving.
	let hovered = $state(false);
	let focused = $state(false);

	/** Starts the hover preview for a mouse or pen; a finger passing over while scrolling doesn't count. */
	function onPointerEnter(event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		hovered = true;
	}
	let active = $derived(hovered || focused);
	// Touch screens: the card rests in the middle of the screen (centerPreview).
	let centered = $state(false);
	// The hover player covering the card replaces its preview.
	let hoverPlaying = $state(false);
	let previewing = $derived((active || centered) && !hoverPlaying);
	let frameIndex = $state(0);
	// Length of the preview clip once it loaded, for the hover player's delay.
	let clipSeconds = $state(0);

	// Cycle preview frames only while hovered/focused; torn down when inactive.
	$effect(() => {
		if (!previewing || frames.length <= 1) {
			frameIndex = 0;
			return;
		}
		const timer = setInterval(() => {
			frameIndex = (frameIndex + 1) % frames.length;
		}, 700);
		return () => clearInterval(timer);
	});

	let imageSrc = $derived(previewing && frames.length > 0 ? frames[frameIndex] : item.bannerUrl);
	let showVideo = $derived(previewing && Boolean(item.previewVideo));

	let durationLabel = $derived(formatClock(item.durationSeconds));
	let site = $derived.by(() => {
		if (!item.sourceUrl) return null;
		return pluginForUrl(item.sourceUrl, $sites);
	});
	let ageLabel = $derived.by(() => {
		if (!item.date) return null;
		return formatRelativeTime(item.date);
	});

	// Resume fraction from the shared watch-progress store, when this scene has an
	// unfinished watch. Prefers the card's own duration, falling back to the
	// duration snapshot stored with the watch.
	let resumeFraction = $derived.by(() => {
		if (!item.id) return 0;
		const progress = $watchProgress.get(item.id);
		if (!progress) return 0;
		const duration = item.durationSeconds || progress.durationSeconds;
		if (!duration || duration <= 0) return 0;
		const fraction = progress.progressSeconds / duration;
		if (fraction <= 0) return 0;
		return Math.min(fraction, 1);
	});
</script>

<div class="group flex w-full flex-col gap-3">
	<div
		class="relative"
		role="presentation"
		use:centerPreview={(isCentered) => (centered = isCentered)}
		onpointerenter={onPointerEnter}
		onpointerleave={() => (hovered = false)}
	>
		<button
			type="button"
			{onclick}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			class="tv-card relative block aspect-video w-full overflow-hidden ring-1 ring-white/5 transition-transform duration-200 group-hover:scale-[1.02] group-hover:shadow-xl focus:outline-none"
		>
			{#if imageSrc}
				<img src={imageSrc} alt={item.title} loading="lazy" class="h-full w-full object-cover" />
			{:else}
				<div class="flex h-full w-full items-center justify-center">
					<span class="text-5xl opacity-10">🔞</span>
				</div>
			{/if}

			{#if showVideo}
				<!-- Preview clip overlays the frame strip when the source provides one. -->
				<!-- svelte-ignore a11y_media_has_caption -->
				<video
					src={item.previewVideo ?? undefined}
					class="absolute inset-0 h-full w-full object-cover"
					onloadedmetadata={(event) => (clipSeconds = event.currentTarget.duration)}
					autoplay
					muted
					loop
					playsinline
				></video>
			{/if}

			{#if item.inheritedTag}
				<span
					class="absolute top-1.5 right-1.5 rounded bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white/60"
					title="Matched via studio or performer tag"
				>
					via studio/performer
				</span>
			{/if}

			{#if item.isNew}
				<span
					class="bg-primary text-primary-content absolute bottom-1.5 left-1.5 rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase"
				>
					New
				</span>
			{/if}

			{#if durationLabel}
				<span
					class="absolute right-1.5 bottom-1.5 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
					class:bottom-2.5={resumeFraction > 0}
				>
					{durationLabel}
				</span>
			{/if}

			<!-- Frame-position ticks, shown while previewing a multi-frame strip. -->
			{#if previewing && frames.length > 1 && !showVideo}
				<div class="absolute inset-x-2 bottom-2 flex gap-1">
					{#each frames as _, index (index)}
						<span
							class="h-0.5 flex-1 rounded-full bg-white/30"
							class:bg-white={index === frameIndex}
						></span>
					{/each}
				</div>
			{/if}

			<!-- Resume bar: how far playback got into this scene. -->
			{#if resumeFraction > 0}
				<div class="absolute inset-x-0 bottom-0 h-1 bg-white/30">
					<div class="bg-primary h-full" style:width="{resumeFraction * 100}%"></div>
				</div>
			{/if}
		</button>

		{#if hovered && item.id}
			<CardHoverPlayer
				sceneId={item.id}
				title={item.title}
				frameCount={frames.length}
				{clipSeconds}
				onshow={(shown) => (hoverPlaying = shown)}
				onopen={onclick}
			/>
		{/if}

		{#if item.id}
			<!-- Quick-save: appears on hover or keyboard/TV focus only. -->
			<button
				type="button"
				onclick={save}
				tabindex="-1"
				aria-label="Save to collection"
				title="Save to collection"
				class="absolute top-1.5 left-1.5 z-20 flex h-8 w-8 items-center justify-center rounded-md bg-black/70 text-white opacity-0 backdrop-blur-sm transition-opacity group-focus-within:opacity-100 group-hover:opacity-100 hover:bg-black/90"
			>
				<BookmarkSimpleIcon size={16} weight="fill" />
			</button>
		{/if}
	</div>

	<CardCaption
		title={item.title}
		byline={item.channelName}
		avatarUrl={item.channelImageUrl || item.providerIconUrl}
		{site}
		detail={ageLabel}
		{onclick}
	/>
</div>
