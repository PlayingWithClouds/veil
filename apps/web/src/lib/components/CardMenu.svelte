<!--
	A card's ⋮ menu. On a desktop it drops down from the ⋮ button, moved to
	<body> so rows that scroll sideways and hover-scaled cards don't clip it;
	on a phone it is the action sheet. Clicking outside, Escape, scrolling or
	resizing closes it. Render it only while open.
-->
<script lang="ts">
	import { fade } from 'svelte/transition';
	import ActionSheet, { type SheetAction } from './ActionSheet.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import { portal } from '$lib/portal';
	import { pushOverlay } from '$lib/stores/overlayStack';

	interface Props {
		// The ⋮ button the menu drops down from.
		anchor: HTMLElement;
		title: string;
		subtitle?: string | null;
		imageUrl?: string | null;
		actions: SheetAction[];
		onclose: () => void;
	}

	let { anchor, title, subtitle = null, imageUrl = null, actions, onclose }: Props = $props();

	const MENU_WIDTH = 256;
	const VIEWPORT_MARGIN = 8;

	let menu = $state<HTMLElement | null>(null);
	let position = $state({ top: 0, left: 0 });

	$effect(() => {
		if ($isCompact) return;
		const box = anchor.getBoundingClientRect();
		const left = Math.min(box.right - MENU_WIDTH, window.innerWidth - MENU_WIDTH - VIEWPORT_MARGIN);
		position = { top: box.bottom + 4, left: Math.max(VIEWPORT_MARGIN, left) };
	});

	$effect(() => {
		if ($isCompact) return;
		return pushOverlay(onclose);
	});

	// Opening above the anchor when the menu would run off the bottom.
	$effect(() => {
		if (!menu) return;
		const height = menu.offsetHeight;
		if (position.top + height <= window.innerHeight - VIEWPORT_MARGIN) return;
		const above = anchor.getBoundingClientRect().top - height - 4;
		position = { ...position, top: Math.max(VIEWPORT_MARGIN, above) };
	});

	/** Closes the menu, then runs the action. */
	function choose(action: SheetAction) {
		onclose();
		action.run();
	}

	/** Closes the drop-down; the phone's sheet handles its own dismissal. */
	function onDesktopOnly() {
		if (!$isCompact) onclose();
	}

	/** Closes on a press outside the menu and its button. */
	function onWindowPointerDown(event: PointerEvent) {
		if ($isCompact) return;
		const target = event.target as Node;
		if (menu?.contains(target) || anchor.contains(target)) return;
		onclose();
	}

	/** Closes on Escape. */
	function onWindowKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onDesktopOnly();
	}
</script>

<svelte:window
	onpointerdown={onWindowPointerDown}
	onkeydown={onWindowKeydown}
	onresize={onDesktopOnly}
	onscrollcapture={onDesktopOnly}
/>

{#if $isCompact}
	<ActionSheet {title} {subtitle} {imageUrl} {actions} {onclose} />
{:else}
	<div
		use:portal
		bind:this={menu}
		class="menu-panel z-modal fixed flex flex-col gap-0.5 p-1.5"
		style:top="{position.top}px"
		style:left="{position.left}px"
		style:width="{MENU_WIDTH}px"
		role="menu"
		aria-label={title}
		out:fade={{ duration: 100 }}
	>
		{#each actions as action (action.label)}
			<button
				type="button"
				role="menuitem"
				class="menu-item shrink-0"
				class:menu-item-danger={action.destructive}
				onclick={() => choose(action)}
			>
				<action.icon size={18} class="opacity-80" />
				<span class="truncate">{action.label}</span>
			</button>
		{/each}
	</div>
{/if}
