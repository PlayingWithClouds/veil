<script lang="ts">
	// The o-counter pill. Hold it to charge: it fills with liquid from the
	// bottom, trembles and buzzes harder as the charge builds; letting go (or a
	// plain tap) bursts it with a spring and a spray of drops, rolls the count
	// up and plays the full-screen splash, bigger the longer it was held.
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import DropIcon from 'phosphor-svelte/lib/DropIcon';
	import DripOverlay from './DripOverlay.svelte';
	import { portal } from '$lib/portal';
	import { notifications } from '$lib/stores/notifications';
	import { fetchOCount, incrementOCount, decrementOCount } from '$lib/oCount';

	interface Props {
		mediaId: string;
		// The tab bar's glass instead of the solid pill (phone watch page).
		glass?: boolean;
		class?: string;
	}

	let { mediaId, glass = false, class: className = '' }: Props = $props();

	// Holding this long charges fully.
	const FULL_CHARGE_MS = 1600;
	// A tap still gives a decent splash.
	const TAP_INTENSITY = 0.55;
	const BURST_DROPS = 10;

	let count = $state(0);
	let busy = $state(false);
	// 0–1 while held; drives the fill, the tremble and the buzz.
	let charge = $state(0);
	let charging = $state(false);
	// Restarts the splash, the burst and the pop on every increment; 0 = idle.
	let splashKey = $state(0);
	let splashIntensity = $state(1);
	let chargeStartedAt = 0;
	let chargeFrame = 0;
	let buzzTimer: ReturnType<typeof setTimeout> | undefined;
	// The click that follows a charged release must not count again.
	let releasedByPointer = false;
	let buttonElement = $state<HTMLButtonElement | null>(null);
	// Screen point the burst flies from; fixed so a scrolling row can't clip it.
	let burstOriginX = $state(0);
	let burstOriginY = $state(0);

	// Load the current count; re-checks when the media changes.
	$effect(() => {
		if (!mediaId) {
			count = 0;
			return;
		}
		let cancelled = false;
		fetchOCount(mediaId).then((value) => {
			if (!cancelled) count = value;
		});
		return () => {
			cancelled = true;
		};
	});

	/** Starts charging while the button is held. */
	function startCharge(event: PointerEvent) {
		if (busy || event.button > 0) return;
		charging = true;
		charge = 0;
		chargeStartedAt = performance.now();
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
		chargeFrame = requestAnimationFrame(growCharge);
		buzz();
	}

	/** Raises the charge each frame until it's full. */
	function growCharge(now: number) {
		if (!charging) return;
		charge = Math.min(1, (now - chargeStartedAt) / FULL_CHARGE_MS);
		chargeFrame = requestAnimationFrame(growCharge);
	}

	/** Short buzzes that come faster and stronger as the charge builds. */
	function buzz() {
		if (!charging) return;
		navigator.vibrate?.(Math.round(8 + charge * 22));
		buzzTimer = setTimeout(buzz, 260 - charge * 190);
	}

	/** Stops charging without counting (finger slid off or was cancelled). */
	function cancelCharge() {
		charging = false;
		charge = 0;
		cancelAnimationFrame(chargeFrame);
		clearTimeout(buzzTimer);
	}

	/** Letting go counts, with a splash as big as the charge. */
	function releaseCharge() {
		if (!charging) return;
		const intensity = Math.max(TAP_INTENSITY, charge);
		cancelCharge();
		releasedByPointer = true;
		increment(intensity);
	}

	/** Keyboard/TV presses count at tap strength; pointer releases already did. */
	function onClick() {
		if (releasedByPointer) {
			releasedByPointer = false;
			return;
		}
		increment(TAP_INTENSITY);
	}

	async function increment(intensity: number) {
		if (!mediaId || busy) return;
		busy = true;
		const previous = count;
		count = previous + 1;
		splashIntensity = intensity;
		if (buttonElement) {
			const rect = buttonElement.getBoundingClientRect();
			burstOriginX = rect.left + 24;
			burstOriginY = rect.top + rect.height / 2;
		}
		splashKey = performance.now();
		try {
			count = await incrementOCount(mediaId);
		} catch {
			count = previous;
			notifications.push('Could not update the o-counter', 'error');
		} finally {
			busy = false;
		}
	}

	/** Undoes the most recent event, for owners that offer undo elsewhere (glass pill). */
	export function undo() {
		decrement();
	}

	/** The current count, for owners deciding whether to offer undo. */
	export function currentCount(): number {
		return count;
	}

	// Undo the most recent event.
	async function decrement(event?: MouseEvent) {
		event?.stopPropagation();
		if (!mediaId || busy || count === 0) return;
		busy = true;
		const previous = count;
		count = previous - 1;
		try {
			count = await decrementOCount(mediaId);
		} catch {
			count = previous;
			notifications.push('Could not update the o-counter', 'error');
		} finally {
			busy = false;
		}
	}

	/** Where each drop of the release burst flies, spread in a fan upwards. */
	function burstDrops(): { x: number; y: number; size: number; delay: number }[] {
		const drops: { x: number; y: number; size: number; delay: number }[] = [];
		for (let index = 0; index < BURST_DROPS; index++) {
			const angle = -Math.PI / 2 + (index / (BURST_DROPS - 1) - 0.5) * 2.4;
			const distance = 28 + Math.random() * 26;
			drops.push({
				x: Math.cos(angle) * distance,
				y: Math.sin(angle) * distance,
				size: 4 + Math.random() * 5,
				delay: Math.random() * 60
			});
		}
		return drops;
	}
</script>

<div
	class="relative flex items-center {className}"
	class:glass-pill={glass}
	class:rounded-full={glass}
	class:overflow-hidden={glass}
>
	<button
		bind:this={buttonElement}
		type="button"
		aria-label="Increment o-counter (hold to charge)"
		title="O-counter — hold to charge"
		disabled={busy}
		onpointerdown={startCharge}
		onpointerup={releaseCharge}
		onpointercancel={cancelCharge}
		onlostpointercapture={cancelCharge}
		oncontextmenu={(event) => event.preventDefault()}
		onclick={onClick}
		class="o-button relative flex h-10 items-center gap-1.5 overflow-hidden py-2.5 pl-4 text-sm font-semibold transition-colors select-none disabled:opacity-70"
		class:text-white={!glass}
		class:bg-primary={count > 0 && !glass}
		class:bg-white-10={count === 0 && !glass}
		class:glass-pill-on={count > 0 && glass}
		class:rounded-full={count === 0 || glass}
		class:rounded-l-full={count > 0 && !glass}
		class:pr-2={count > 0 && !glass}
		class:pr-4={count === 0 || glass}
		class:o-charging={charging}
		style:--charge={charge}
	>
		<!-- The liquid rising inside while held. -->
		<span class="o-fill pointer-events-none absolute inset-x-0 bottom-0" aria-hidden="true"></span>
		{#key splashKey}
			<span class="relative flex items-center gap-1.5" class:o-pop={splashKey > 0}>
				<DropIcon size={18} weight={count > 0 || charging ? 'fill' : 'regular'} />
				<span class="relative inline-grid overflow-hidden tabular-nums">
					{#key count}
						<span
							class="col-start-1 row-start-1"
							in:fly={{ y: 14, duration: 320, easing: cubicOut }}
							out:fly={{ y: -14, duration: 220 }}
						>
							{count}
						</span>
					{/key}
				</span>
			</span>
		{/key}
	</button>
	<!-- The glass pill offers undo from its owner's menu instead. -->
	{#if count > 0 && !glass}
		<button
			type="button"
			aria-label="Undo o-counter"
			disabled={busy}
			onclick={decrement}
			class="bg-primary flex h-10 items-center rounded-r-full py-2.5 pr-3 pl-2 text-sm font-semibold text-white transition-colors disabled:opacity-50"
		>
			−
		</button>
	{/if}

	<!-- Drops bursting out of the button on release. -->
	<!-- Moved to <body>: the glass pill's backdrop blur would otherwise turn
	     these fixed layers into button-sized ones clipped by the pill. -->
	{#if splashKey}
		{#key splashKey}
			<div use:portal>
				<span
					class="z-modal pointer-events-none fixed"
					style:left="{burstOriginX}px"
					style:top="{burstOriginY}px"
					aria-hidden="true"
				>
					{#each burstDrops() as drop, index (index)}
						<span
							class="o-burst-drop"
							style:--burst-x="{drop.x}px"
							style:--burst-y="{drop.y}px"
							style:width="{drop.size}px"
							style:height="{drop.size}px"
							style:animation-delay="{drop.delay}ms"
						></span>
					{/each}
				</span>
				<DripOverlay intensity={splashIntensity} onend={() => (splashKey = 0)} />
			</div>
		{/key}
	{/if}
</div>

<style>
	.bg-white-10 {
		background-color: rgb(255 255 255 / 0.1);
	}

	.o-fill {
		height: calc(var(--charge) * 100%);
		background: linear-gradient(to top, rgb(244 240 228 / 0.85), rgb(244 240 228 / 0.55));
		border-radius: 40% 40% 0 0 / 12% 12% 0 0;
	}

	/* Trembles harder as the charge builds. */
	.o-charging {
		animation: o-tremble 90ms linear infinite;
		scale: calc(0.96 + var(--charge) * 0.06);
	}

	@keyframes o-tremble {
		0%,
		100% {
			translate: 0 0;
		}
		25% {
			translate: calc(var(--charge) * 1.5px) calc(var(--charge) * -1px);
		}
		75% {
			translate: calc(var(--charge) * -1.5px) calc(var(--charge) * 1px);
		}
	}

	/* The release: squash, then spring past full size and settle. */
	.o-pop {
		animation: o-pop 520ms cubic-bezier(0.22, 1.6, 0.36, 1);
	}

	@keyframes o-pop {
		0% {
			scale: 0.7;
		}
		100% {
			scale: 1;
		}
	}

	.o-burst-drop {
		position: absolute;
		left: 0;
		top: 0;
		border-radius: 50% 50% 50% 50% / 60% 60% 40% 40%;
		background: rgb(244 240 228);
		box-shadow: inset -1px -1px 2px rgb(0 0 0 / 0.25);
		opacity: 0;
		animation: o-burst 560ms cubic-bezier(0.22, 1, 0.36, 1) forwards;
	}

	@keyframes o-burst {
		0% {
			opacity: 1;
			translate: 0 0;
			scale: 0.4;
		}
		70% {
			opacity: 1;
			translate: var(--burst-x) var(--burst-y);
			scale: 1;
		}
		100% {
			opacity: 0;
			translate: var(--burst-x) calc(var(--burst-y) + 14px);
			scale: 0.8;
		}
	}
</style>
