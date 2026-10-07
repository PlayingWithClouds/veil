<!--
	Phone drawer: a floating glass panel holding the phone menu. Opens from the
	top bar's menu button or a swipe (swipeGestures.ts drives drawerDrag while
	the finger moves). The layout reads the same progress to push the page back
	behind it.
-->
<script lang="ts">
	import DrawerMenu from './DrawerMenu.svelte';
	import { drawerDrag, drawerOpen, drawerProgress } from '$lib/stores/viewport';

	let dragging = $derived($drawerDrag !== null);
</script>

<button
	type="button"
	class="drawer-backdrop z-modal fixed inset-0 bg-black/40"
	class:dragging
	class:pointer-events-none={$drawerProgress === 0}
	style:opacity={$drawerProgress}
	tabindex="-1"
	aria-label="Close menu"
	onclick={() => drawerOpen.set(false)}
></button>
<div
	class="drawer-panel z-modal fixed left-2 flex w-72 max-w-[calc(100vw-4rem)] flex-col overflow-hidden rounded-[1.75rem]"
	class:dragging
	class:drawer-panel-shown={$drawerProgress > 0}
	style:transform="translateX(calc({$drawerProgress - 1} * (100% + 1rem)))"
	data-drawer-panel
	inert={!$drawerOpen && !dragging}
>
	<DrawerMenu open={$drawerOpen && !dragging} {dragging} />
</div>

<style>
	.drawer-backdrop {
		backdrop-filter: blur(3px);
		transition: opacity 300ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	/* Same frosted glass as the tab bar, a touch more opaque for the longer text. */
	.drawer-panel {
		top: calc(0.5rem + var(--safe-area-inset-top, env(safe-area-inset-top, 0px)));
		bottom: calc(0.5rem + var(--safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
		background-color: color-mix(in oklab, var(--color-base-200) 80%, transparent);
		backdrop-filter: blur(24px) saturate(1.8);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		transition:
			transform 340ms cubic-bezier(0.22, 1, 0.36, 1),
			box-shadow 340ms ease-out;
	}

	.drawer-panel-shown {
		box-shadow:
			0 20px 50px rgba(0, 0, 0, 0.55),
			inset 0 1px 0 color-mix(in oklab, var(--color-base-content) 8%, transparent);
	}

	/* The finger drives the drawer directly; easing would make it lag behind. */
	.dragging {
		transition: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.drawer-panel,
		.drawer-backdrop {
			transition-duration: 1ms;
		}
	}
</style>
