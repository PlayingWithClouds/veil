<!--
	Photos full screen, over a blur of the one shown. Arrows (buttons, arrow
	keys, a swipe) step through them, a counter and a strip of thumbnails say
	where you are, the next photos load ahead. Clicking the photo switches
	between fitting the screen and full size (scrollable); the slideshow steps
	on by itself until it reaches the end. Escape closes.
-->
<script lang="ts" module>
	/** One photo of the lightbox. */
	export interface LightboxImage {
		src: string;
		// A smaller copy for the thumbnail strip; the photo itself when absent.
		thumbnail?: string;
	}
</script>

<script lang="ts">
	import { tick } from 'svelte';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';

	interface Props {
		images: LightboxImage[];
		// The photo shown.
		index: number;
		alt?: string;
		// Starts the slideshow on opening.
		playing?: boolean;
		onclose: () => void;
	}

	let {
		images,
		index = $bindable(),
		alt = '',
		playing = $bindable(false),
		onclose
	}: Props = $props();

	const SLIDE_MS = 4000;
	const PRELOAD_AHEAD = 3;
	// How far a swipe travels before it steps, in px.
	const SWIPE_DISTANCE = 50;

	let zoomed = $state(false);
	let strip = $state<HTMLElement>();
	let swipeStart: number | null = null;

	let current = $derived(images[index]);
	let atEnd = $derived(index >= images.length - 1);

	$effect(() => {
		preload(index);
		void scrollThumbnailIntoView(index);
	});

	$effect(() => {
		if (!playing) return;
		const timer = setInterval(() => {
			if (index >= images.length - 1) {
				playing = false;
				return;
			}
			show(index + 1);
		}, SLIDE_MS);
		return () => clearInterval(timer);
	});

	/** Shows the photo at next, within bounds, fitted to the screen. */
	function show(next: number) {
		index = Math.max(0, Math.min(next, images.length - 1));
		zoomed = false;
	}

	/** Loads the next few photos so stepping shows them at once. */
	function preload(from: number) {
		for (const image of images.slice(from + 1, from + 1 + PRELOAD_AHEAD)) {
			new Image().src = image.src;
		}
	}

	/** Keeps the shown photo's thumbnail in view. */
	async function scrollThumbnailIntoView(shown: number) {
		await tick();
		strip?.children[shown]?.scrollIntoView({
			block: 'nearest',
			inline: 'center',
			behavior: 'smooth'
		});
	}

	/** Steps with the arrow keys, plays or pauses with space, closes with Escape. */
	function onKeydown(event: KeyboardEvent) {
		const handlers: Record<string, () => void> = {
			Escape: onclose,
			ArrowRight: () => show(index + 1),
			ArrowLeft: () => show(index - 1),
			Home: () => show(0),
			End: () => show(images.length - 1),
			' ': () => (playing = !playing)
		};
		const handle = handlers[event.key];
		if (!handle) return;
		event.preventDefault();
		event.stopPropagation();
		handle();
	}

	/** Remembers where a touch swipe started. */
	function onPointerDown(event: PointerEvent) {
		if (event.pointerType === 'mouse' || zoomed) return;
		swipeStart = event.clientX;
	}

	/** Steps when a swipe went far enough. */
	function onPointerUp(event: PointerEvent) {
		if (swipeStart === null) return;
		const distance = event.clientX - swipeStart;
		swipeStart = null;
		if (distance <= -SWIPE_DISTANCE) show(index + 1);
		if (distance >= SWIPE_DISTANCE) show(index - 1);
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div
	class="fixed inset-0 z-[100] flex flex-col bg-black text-white select-none"
	role="dialog"
	aria-modal="true"
	aria-label={alt || 'Photos'}
>
	{#if current}
		<img
			src={current.thumbnail ?? current.src}
			alt=""
			aria-hidden="true"
			class="pointer-events-none absolute inset-0 h-full w-full scale-110 object-cover opacity-30 blur-3xl"
		/>
	{/if}

	<div class="relative flex items-center gap-2 px-4 py-3">
		<span class="text-sm text-white/80 tabular-nums">{index + 1} / {images.length}</span>
		<span class="flex-1"></span>
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10 transition-colors hover:bg-white/20"
			aria-label={playing ? 'Pause slideshow' : 'Play slideshow'}
			aria-pressed={playing}
			disabled={atEnd && !playing}
			onclick={() => (playing = !playing)}
		>
			{#if playing}<PauseIcon size={18} weight="fill" />{:else}<PlayIcon
					size={18}
					weight="fill"
				/>{/if}
		</button>
		{#if current}
			<a
				href={current.src}
				target="_blank"
				rel="noreferrer"
				class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10 transition-colors hover:bg-white/20"
				aria-label="Open the original"
			>
				<ArrowSquareOutIcon size={18} />
			</a>
		{/if}
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10 transition-colors hover:bg-white/20"
			aria-label="Close"
			onclick={onclose}
		>
			<XIcon size={20} />
		</button>
	</div>

	<div
		class="relative min-h-0 flex-1"
		class:overflow-auto={zoomed}
		role="presentation"
		onpointerdown={onPointerDown}
		onpointerup={onPointerUp}
		onpointercancel={() => (swipeStart = null)}
	>
		{#if current}
			<button
				type="button"
				class="flex min-h-full min-w-full items-center justify-center"
				class:h-full={!zoomed}
				class:w-full={!zoomed}
				class:p-4={!zoomed}
				class:cursor-zoom-in={!zoomed}
				class:cursor-zoom-out={zoomed}
				aria-label={zoomed ? 'Fit to the screen' : 'Full size'}
				onclick={() => (zoomed = !zoomed)}
			>
				<img
					src={current.src}
					{alt}
					class="rounded-lg shadow-2xl"
					class:max-h-full={!zoomed}
					class:max-w-full={!zoomed}
					class:object-contain={!zoomed}
					class:max-w-none={zoomed}
				/>
			</button>
		{/if}

		{#if !zoomed}
			<button
				type="button"
				class="absolute top-1/2 left-3 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-black/40 transition-colors hover:bg-black/60 disabled:opacity-0 max-md:hidden"
				aria-label="Previous photo"
				disabled={index === 0}
				onclick={() => show(index - 1)}
			>
				<CaretLeftIcon size={24} weight="bold" />
			</button>
			<button
				type="button"
				class="absolute top-1/2 right-3 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-black/40 transition-colors hover:bg-black/60 disabled:opacity-0 max-md:hidden"
				aria-label="Next photo"
				disabled={atEnd}
				onclick={() => show(index + 1)}
			>
				<CaretRightIcon size={24} weight="bold" />
			</button>
		{/if}
	</div>

	{#if images.length > 1}
		<div bind:this={strip} class="relative flex gap-1.5 overflow-x-auto px-4 py-3">
			{#each images as image, position (position)}
				<button
					type="button"
					class="h-16 w-12 shrink-0 overflow-hidden rounded opacity-50 transition-opacity hover:opacity-90"
					class:opacity-100={position === index}
					class:ring-2={position === index}
					class:ring-white={position === index}
					aria-label="Photo {position + 1}"
					aria-current={position === index}
					onclick={() => show(position)}
				>
					<img
						src={image.thumbnail ?? image.src}
						alt=""
						loading="lazy"
						class="h-full w-full object-cover"
					/>
				</button>
			{/each}
		</div>
	{/if}
</div>
