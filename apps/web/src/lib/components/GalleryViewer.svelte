<script lang="ts">
	// Full-screen viewer for a gallery's images: arrows step through them, a
	// click zooms in where you clicked (again to fit), the wheel zooms freely,
	// and a zoomed image pans with the pointer. Escape closes.
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import CornersOutIcon from 'phosphor-svelte/lib/CornersOutIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import UserFocusIcon from 'phosphor-svelte/lib/UserFocusIcon';
	import { cacheUrl } from '$lib/img';
	import type { Image } from '$lib/images';

	interface Props {
		images: Image[];
		index: number;
		title: string;
		onclose: () => void;
		/** Asks who is in the image on screen. */
		onidentify: (image: Image) => void;
	}

	let { images, index = $bindable(), title, onclose, onidentify }: Props = $props();

	// A click zooms to CLICK_ZOOM; the wheel moves between 1 and MAX_ZOOM.
	const CLICK_ZOOM = 2.5;
	const MAX_ZOOM = 5;
	const WHEEL_STEP = 0.0015;

	let viewer = $state<HTMLElement | null>(null);
	let zoom = $state(1);
	// The zoom's focus point in percent of the image, following the pointer.
	let originX = $state(50);
	let originY = $state(50);
	let zoomed = $derived(zoom > 1);
	let image = $derived(images[index]);

	/** Moves by delta images (wrapping) and resets the zoom. */
	function step(delta: number) {
		if (images.length === 0) return;
		index = (index + delta + images.length) % images.length;
		resetZoom();
	}

	/** Back to the fitted image. */
	function resetZoom() {
		zoom = 1;
		originX = 50;
		originY = 50;
	}

	/** Points the zoom at the pointer's spot on the image. */
	function follow(event: PointerEvent | MouseEvent | WheelEvent) {
		const bounds = (event.currentTarget as HTMLElement).getBoundingClientRect();
		originX = ((event.clientX - bounds.left) / bounds.width) * 100;
		originY = ((event.clientY - bounds.top) / bounds.height) * 100;
	}

	/** Click: zoom in at the clicked spot, or back out. */
	function toggleZoom(event: MouseEvent) {
		event.stopPropagation();
		if (zoomed) {
			resetZoom();
			return;
		}
		follow(event);
		zoom = CLICK_ZOOM;
	}

	/** Wheel: zoom in or out around the pointer. */
	function wheelZoom(event: WheelEvent) {
		event.preventDefault();
		if (!zoomed) follow(event);
		zoom = Math.min(MAX_ZOOM, Math.max(1, zoom * (1 - event.deltaY * WHEEL_STEP)));
	}

	/** Pans a zoomed image with the pointer. */
	function pan(event: PointerEvent) {
		if (zoomed) follow(event);
	}

	/** Shows the viewer on the whole screen. */
	function enterFullscreen(event: MouseEvent) {
		event.stopPropagation();
		viewer?.requestFullscreen?.().catch(() => {});
	}

	/** Escape closes (or unzooms), arrows step. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			if (zoomed) resetZoom();
			else onclose();
		}
		if (event.key === 'ArrowRight') step(1);
		if (event.key === 'ArrowLeft') step(-1);
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div
	bind:this={viewer}
	class="fixed inset-0 z-[60] flex items-center justify-center overflow-hidden bg-black"
	role="dialog"
	aria-modal="true"
	aria-label="{title} {index + 1}"
>
	<button type="button" class="absolute inset-0" aria-label="Close" onclick={onclose}></button>

	{#if image}
		<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
		<img
			src={cacheUrl(image.filePath)}
			alt="{title} {index + 1}"
			class="relative max-h-screen max-w-full object-contain transition-transform duration-150 ease-out select-none"
			class:cursor-zoom-in={!zoomed}
			class:cursor-zoom-out={zoomed}
			style:transform="scale({zoom})"
			style:transform-origin="{originX}% {originY}%"
			draggable="false"
			onclick={toggleZoom}
			onpointermove={pan}
			onwheel={wheelZoom}
		/>
	{/if}

	<div class="absolute top-4 right-4 flex gap-2">
		{#if image}
			<button
				type="button"
				class="flex h-10 items-center gap-1.5 rounded-full bg-white/10 px-4 text-sm font-semibold text-white hover:bg-white/20"
				onclick={(event) => {
					event.stopPropagation();
					onidentify(image);
				}}
			>
				<UserFocusIcon size={18} />
				Who is this?
			</button>
		{/if}
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
			onclick={enterFullscreen}
			aria-label="Full screen"
		>
			<CornersOutIcon size={20} weight="bold" />
		</button>
		<button
			type="button"
			class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
			onclick={onclose}
			aria-label="Close"
		>
			<XIcon size={20} weight="bold" />
		</button>
	</div>
	<button
		type="button"
		class="absolute left-4 flex h-12 w-12 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
		onclick={(event) => {
			event.stopPropagation();
			step(-1);
		}}
		aria-label="Previous"
	>
		<CaretLeftIcon size={24} weight="bold" />
	</button>
	<button
		type="button"
		class="absolute right-4 flex h-12 w-12 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
		onclick={(event) => {
			event.stopPropagation();
			step(1);
		}}
		aria-label="Next"
	>
		<CaretRightIcon size={24} weight="bold" />
	</button>
	<span
		class="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-black/70 px-3 py-1 text-xs text-white"
	>
		{index + 1} / {images.length}
	</span>
</div>
