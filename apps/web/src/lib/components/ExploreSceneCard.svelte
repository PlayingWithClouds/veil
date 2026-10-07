<script lang="ts">
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import SpinnerGapIcon from 'phosphor-svelte/lib/SpinnerGapIcon';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import CardHoverPlayer from '$lib/components/CardHoverPlayer.svelte';
	import { centerPreview } from '$lib/centerPreview';
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import { quickSaveEntity } from '$lib/collectionActions';
	import { notifications } from '$lib/stores/notifications';
	import { ensureStreamsCached, downloadDefault } from '$lib/cardActions';
	import { sceneByline, type SceneHit } from '$lib/explore';
	import { formatRelativeTime } from '$lib/time';
	import type { SearchPlugin } from '$lib/search';
	import { longPress } from '$lib/longPress';
	import { fadeInImage } from '$lib/fadeInImage';
	import { isCompact } from '$lib/stores/viewport';
	import { clearVerdict } from '$lib/rating';
	import { dismissedScenes, openSceneSheet, setSceneDismissed } from '$lib/stores/sceneSheet';

	interface Props {
		scene: SceneHit;
		// The card tag best matching the user's taste, shown top-left. Null hides it.
		preferredTag?: string | null;
		// The site the scene came from, shown under the title.
		site?: SearchPlugin | null;
		// False for site results not stored yet: no id to save, download or
		// hover-play, so only the preview and the click work.
		actionable?: boolean;
		// True while the click is still preparing the scene (e.g. storing a site result).
		opening?: boolean;
		onclick: () => void;
	}

	let {
		scene,
		preferredTag = null,
		site = null,
		actionable = true,
		opening = false,
		onclick
	}: Props = $props();

	// Warm scene detail/streams after a deliberate hover so opening (or downloading)
	// is instant. A quick pass-over shouldn't trigger a fetch.
	const HOVER_PREFETCH_MS = 1400;

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

	let frames = $derived(scene.previewImages ?? []);
	let imageSrc = $derived(previewing && frames.length > 0 ? frames[frameIndex] : scene.bannerUrl);
	let showVideo = $derived(previewing && Boolean(scene.previewVideo));
	let durationLabel = $derived(formatClock(scene.durationSeconds));
	// "3 weeks ago" from the release date. When it was stored is not shown: it
	// reads like an upload date but only says when a search found it.
	let ageLabel = $derived.by(() => {
		if (!scene.date) return null;
		return formatRelativeTime(scene.date);
	});

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

	// Prefetch the streams after a deliberate hover; leaving cancels it.
	$effect(() => {
		if (!active || !actionable) return;
		const timer = setTimeout(() => {
			ensureStreamsCached(scene.id).catch(() => {});
		}, HOVER_PREFETCH_MS);
		return () => clearTimeout(timer);
	});

	/** Opens the phone action sheet for this scene. */
	function openActions() {
		if (!actionable) return;
		const parts: string[] = [];
		const byline = sceneByline(scene);
		if (byline) parts.push(byline);
		if (site) parts.push(site.displayName || site.name);
		openSceneSheet({
			sceneId: scene.id,
			title: scene.title,
			imageUrl: scene.bannerUrl,
			subtitle: parts.join(' · ') || null
		});
	}

	/** Brings back a card marked "Not interested". */
	function undoDismiss() {
		setSceneDismissed(scene.id, false);
		clearVerdict(scene.id).catch(() => {});
	}

	let saved = $state(false);
	/** Quick-saves the scene to the default collection. */
	function save() {
		quickSaveEntity('scene', scene.id, scene.title);
		saved = true;
	}

	let downloading = $state(false);
	/** Starts a download at the default resolution. */
	async function download() {
		downloading = true;
		try {
			await downloadDefault(scene.id, scene.title);
			notifications.push(`Download started: ${scene.title}`, 'success');
		} catch (error) {
			notifications.push(String(error), 'error');
		} finally {
			downloading = false;
		}
	}
</script>

{#if $dismissedScenes.has(scene.id)}
	<div class="feed-card bg-base-200 flex items-center justify-between gap-3 rounded-xl px-4 py-3 text-sm">
		<span class="text-base-content/60">Hidden. You'll see less like this.</span>
		<button type="button" class="font-semibold" onclick={undoDismiss}>Undo</button>
	</div>
{:else}
<div class="feed-card group flex flex-col gap-2">
	<div
		class="relative"
		role="presentation"
		use:longPress={openActions}
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
				<img
					src={imageSrc}
					alt={scene.title}
					loading="lazy"
					use:fadeInImage
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
					src={scene.previewVideo ?? undefined}
					class="absolute inset-0 h-full w-full object-cover"
					onloadedmetadata={(event) => (clipSeconds = event.currentTarget.duration)}
					autoplay
					muted
					loop
					playsinline
				></video>
			{/if}

			{#if preferredTag}
				<span
					class="bg-primary/85 absolute top-1.5 left-1.5 rounded px-1.5 py-0.5 text-[10px] font-semibold text-white shadow"
					title="Matches your taste"
				>
					{preferredTag}
				</span>
			{/if}

			{#if durationLabel}
				<span
					class="absolute right-1.5 bottom-1.5 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
				>
					{durationLabel}
				</span>
			{/if}

			{#if opening}
				<span class="absolute inset-0 flex items-center justify-center bg-black/60" aria-hidden="true">
					<span class="loading loading-sm text-white"></span>
				</span>
			{/if}
		</button>

		{#if hovered && actionable}
			<CardHoverPlayer
				sceneId={scene.id}
				title={scene.title}
				frameCount={frames.length}
				{clipSeconds}
				onshow={(shown) => (hoverPlaying = shown)}
				onopen={onclick}
			/>
		{/if}

		<!-- Actions appear on hover or keyboard/TV focus only. -->
		<div
			class="absolute top-1.5 right-1.5 z-20 flex gap-1 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100"
			class:hidden={!actionable}
		>
			<button
				type="button"
				onclick={save}
				aria-label="Add to watchlist"
				title="Add to watchlist"
				class="flex h-8 w-8 items-center justify-center rounded-md bg-black/70 text-white backdrop-blur-sm transition-colors hover:bg-black/90"
				class:text-primary={saved}
			>
				{#if saved}
					<CheckIcon size={16} weight="bold" />
				{:else}
					<PlusIcon size={16} />
				{/if}
			</button>
			<button
				type="button"
				onclick={download}
				disabled={downloading}
				aria-label="Download"
				title="Download"
				class="flex h-8 w-8 items-center justify-center rounded-md bg-black/70 text-white backdrop-blur-sm transition-colors hover:bg-black/90 disabled:opacity-60"
			>
				{#if downloading}
					<SpinnerGapIcon size={16} class="animate-spin" />
				{:else}
					<DownloadSimpleIcon size={16} />
				{/if}
			</button>
		</div>
	</div>

	<CardCaption
		title={scene.title}
		byline={sceneByline(scene)}
		avatarUrl={scene.studioImageUrl || scene.performerImageUrl}
		{site}
		detail={ageLabel}
		{onclick}
		onmore={$isCompact && actionable ? openActions : undefined}
	/>
</div>
{/if}
