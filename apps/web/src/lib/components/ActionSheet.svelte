<!--
	Phone action sheet: floats up from the bottom as glass like the tab bar, with
	a header (image, title, subtitle) and a list of actions. Tap outside, drag it
	down or press back to dismiss. Render it only while open.
-->
<script module lang="ts">
	import type { Component, Snippet } from 'svelte';

	export type SheetAction = {
		label: string;
		icon: Component;
		run: () => void;
		// Drawn in the error colour.
		destructive?: boolean;
	};
</script>

<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { pushOverlay } from '$lib/stores/overlayStack';

	interface Props {
		title: string;
		subtitle?: string | null;
		imageUrl?: string | null;
		// How the header image is cropped: a 16:9 thumbnail or a round avatar.
		imageShape?: 'thumbnail' | 'avatar';
		actions?: SheetAction[];
		// Custom content below the actions, e.g. a list of choices.
		children?: Snippet;
		onclose: () => void;
	}

	let {
		title,
		subtitle = null,
		imageUrl = null,
		imageShape = 'thumbnail',
		actions = [],
		children,
		onclose
	}: Props = $props();

	// Downward drag (px) or speed (px/ms) that dismisses the sheet on release.
	const DISMISS_DISTANCE = 90;
	const DISMISS_VELOCITY = 0.5;

	let dragOffset = $state(0);
	let dragging = $state(false);
	let dragStartY = 0;
	let lastY = 0;
	let lastTime = 0;
	let velocity = 0;

	$effect(() => pushOverlay(onclose));

	/** Closes the sheet, then runs the action. */
	function choose(action: SheetAction) {
		onclose();
		action.run();
	}

	function onTouchStart(event: TouchEvent) {
		dragging = true;
		dragStartY = event.touches[0].clientY;
		lastY = dragStartY;
		lastTime = event.timeStamp;
		velocity = 0;
	}

	function onTouchMove(event: TouchEvent) {
		if (!dragging) return;
		const y = event.touches[0].clientY;
		const elapsed = event.timeStamp - lastTime;
		if (elapsed > 0) velocity = (y - lastY) / elapsed;
		lastY = y;
		lastTime = event.timeStamp;
		// Upward drags stretch only a little.
		const delta = y - dragStartY;
		if (delta < 0) {
			dragOffset = delta * 0.15;
		} else {
			dragOffset = delta;
		}
	}

	function onTouchEnd() {
		if (!dragging) return;
		dragging = false;
		if (dragOffset > DISMISS_DISTANCE || velocity > DISMISS_VELOCITY) onclose();
		dragOffset = 0;
	}
</script>

<button
	type="button"
	class="z-modal fixed inset-0 bg-black/30"
	aria-label="Close"
	tabindex="-1"
	onclick={onclose}
	transition:fade={{ duration: 200 }}
></button>
<div
	class="action-sheet z-modal fixed inset-x-2 mx-auto flex max-w-lg flex-col overflow-hidden rounded-[1.75rem] pb-2"
	class:action-sheet-dragging={dragging}
	style:translate="0 {dragOffset}px"
	role="dialog"
	aria-modal="true"
	aria-label={title}
	data-swipe-ignore
	transition:fly={{ y: 320, duration: 320, easing: cubicOut, opacity: 1 }}
>
	<div
		class="flex flex-col"
		role="presentation"
		ontouchstart={onTouchStart}
		ontouchmove={onTouchMove}
		ontouchend={onTouchEnd}
		ontouchcancel={onTouchEnd}
	>
		<span class="bg-base-content/25 mx-auto mt-2.5 h-1 w-9 rounded-full"></span>
		<div class="flex items-center gap-3 px-4 pt-3 pb-3">
			{#if imageUrl}
				<img
					src={imageUrl}
					alt=""
					class="shrink-0 object-cover"
					class:sheet-thumbnail={imageShape === 'thumbnail'}
					class:sheet-avatar={imageShape === 'avatar'}
				/>
			{/if}
			<div class="flex min-w-0 flex-col gap-0.5">
				<span class="line-clamp-2 text-[15px] leading-5 font-medium">{title}</span>
				{#if subtitle}
					<span class="text-base-content/60 truncate text-[13px]">{subtitle}</span>
				{/if}
			</div>
		</div>
	</div>
	<hr class="border-base-content/10 mx-4" />
	<div class="flex max-h-[70vh] flex-col overflow-y-auto overscroll-contain px-2 pt-2">
		{#each actions as action (action.label)}
			<button
				type="button"
				class="flex h-12 shrink-0 items-center gap-4 rounded-2xl px-3 text-left text-[15px]"
				class:text-error={action.destructive}
				onclick={() => choose(action)}
			>
				<action.icon size={22} />
				{action.label}
			</button>
		{/each}
		{@render children?.()}
	</div>
</div>

<style>
	.action-sheet {
		bottom: calc(0.5rem + var(--safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
		background-color: color-mix(in oklab, var(--color-base-200) 55%, transparent);
		backdrop-filter: blur(28px) saturate(1.8);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow:
			0 -10px 40px rgba(0, 0, 0, 0.5),
			inset 0 1px 0 color-mix(in oklab, var(--color-base-content) 8%, transparent);
		transition: translate 300ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	/* The finger drives the sheet directly. */
	.action-sheet-dragging {
		transition: none;
	}

	.sheet-thumbnail {
		width: 6rem;
		aspect-ratio: 16 / 9;
		border-radius: 0.5rem;
	}

	.sheet-avatar {
		width: 3rem;
		height: 3rem;
		border-radius: 9999px;
	}
</style>
