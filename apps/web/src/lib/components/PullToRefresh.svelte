<!--
	Pull-to-refresh for the page's scroll container (the layout's <main>): pulling
	down while scrolled to the top draws a spinner down after the finger; letting
	go past the threshold runs `onrefresh` and spins until it settles. Touch only;
	place it anywhere in the page.
-->
<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';

	interface Props {
		onrefresh: () => Promise<void>;
	}

	let { onrefresh }: Props = $props();

	// Pull distance (after resistance) that triggers a refresh on release.
	const THRESHOLD = 72;
	// The indicator never travels further than this.
	const MAX_PULL = 120;
	// Finger travel per pixel of indicator travel: the pull feels heavy.
	const RESISTANCE = 0.5;
	// Movement before a touch counts as a pull or a sideways swipe.
	const SLOP_PIXELS = 10;
	// Where the indicator rests while the refresh runs.
	const REFRESHING_PULL = 56;

	let anchor = $state<HTMLElement | null>(null);
	let pull = $state(0);
	let refreshing = $state(false);
	// A finger is moving the indicator; it follows without easing.
	let tracking = $state(false);

	let startX = 0;
	let startY = 0;
	// Null until the touch leaves the slop radius, then whether it is a pull.
	let pulling: boolean | null = null;

	let armed = $derived(pull >= THRESHOLD);
	// The arrow winds up with the pull; the spinner takes over while refreshing.
	let arrowRotation = $derived.by(() => {
		if (refreshing) return 0;
		return (pull / THRESHOLD) * 300;
	});

	/** Starts watching a touch that begins while the page is at its top. */
	function onTouchStart(container: HTMLElement, event: TouchEvent) {
		pulling = null;
		if (refreshing || event.touches.length !== 1 || container.scrollTop > 0) {
			pulling = false;
			return;
		}
		startX = event.touches[0].clientX;
		startY = event.touches[0].clientY;
	}

	/** Decides once the finger leaves the slop radius whether this is a downward pull. */
	function decide(deltaX: number, deltaY: number) {
		if (Math.abs(deltaX) < SLOP_PIXELS && Math.abs(deltaY) < SLOP_PIXELS) return;
		pulling = deltaY > 0 && deltaY > Math.abs(deltaX);
		tracking = pulling;
	}

	function onTouchMove(event: TouchEvent) {
		if (pulling === false) return;
		const deltaX = event.touches[0].clientX - startX;
		const deltaY = event.touches[0].clientY - startY;
		if (pulling === null) decide(deltaX, deltaY);
		if (!pulling) return;
		if (event.cancelable) event.preventDefault();
		const next = Math.min(MAX_PULL, Math.max(0, deltaY * RESISTANCE));
		if (next >= THRESHOLD && pull < THRESHOLD) navigator.vibrate?.(8);
		pull = next;
	}

	async function onTouchEnd() {
		if (!pulling) return;
		pulling = null;
		tracking = false;
		if (pull < THRESHOLD) {
			pull = 0;
			return;
		}
		refreshing = true;
		pull = REFRESHING_PULL;
		try {
			await onrefresh();
		} finally {
			refreshing = false;
			pull = 0;
		}
	}

	$effect(() => {
		const container = anchor?.closest('main');
		if (!container) return;
		const start = (event: TouchEvent) => onTouchStart(container, event);
		container.addEventListener('touchstart', start, { passive: true });
		// Not passive: the pull cancels the browser's own overscroll.
		container.addEventListener('touchmove', onTouchMove, { passive: false });
		container.addEventListener('touchend', onTouchEnd);
		container.addEventListener('touchcancel', onTouchEnd);
		return () => {
			container.removeEventListener('touchstart', start);
			container.removeEventListener('touchmove', onTouchMove);
			container.removeEventListener('touchend', onTouchEnd);
			container.removeEventListener('touchcancel', onTouchEnd);
		};
	});
</script>

<div bind:this={anchor} class="hidden"></div>
<div
	class="pull-indicator bg-base-200 ring-base-content/10 pointer-events-none fixed left-1/2 z-40 flex h-10 w-10 items-center justify-center rounded-full shadow-lg ring-1"
	class:pull-indicator-tracking={tracking}
	class:text-primary={armed || refreshing}
	style:translate="-50% {pull - 48}px"
	style:opacity={Math.min(1, pull / (THRESHOLD * 0.6))}
	aria-hidden={!refreshing}
	role="status"
>
	<span class:animate-spin={refreshing} class="flex">
		<span class="flex" style:rotate="{arrowRotation}deg">
			<ArrowClockwiseIcon size={20} weight="bold" />
		</span>
	</span>
</div>

<style>
	.pull-indicator {
		top: var(--chrome-offset);
		transition:
			translate 360ms cubic-bezier(0.22, 1, 0.36, 1),
			opacity 200ms ease-out,
			color 150ms;
	}

	.pull-indicator-tracking {
		transition: color 150ms;
	}
</style>
