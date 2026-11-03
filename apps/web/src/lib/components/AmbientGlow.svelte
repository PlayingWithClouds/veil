<!--
	The app's shared glow: the picture the page features (ambientGlow),
	blurred into soft colour behind the top bar, the sidebar and the top of
	the page. A new picture fades in over the old one; scrolling the page
	fades the glow out.
-->
<script lang="ts">
	import { ambientGlow } from '$lib/ambientGlow.svelte';

	interface Props {
		// How far the page has scrolled, in px.
		scrollTop: number;
	}

	let { scrollTop }: Props = $props();

	// Scrolled this far, the glow is gone.
	const FADE_DISTANCE = 700;

	let strength = $derived(Math.max(0, 1 - scrollTop / FADE_DISTANCE));
</script>

<div
	class="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[85vh] overflow-hidden"
	style:opacity={strength}
	aria-hidden="true"
>
	{#key $ambientGlow}
		{#if $ambientGlow}
			<img
				src={$ambientGlow}
				alt=""
				class="glow-picture absolute -inset-x-[10%] -top-[30%] h-[130%] w-[120%] max-w-none object-cover blur-[100px] saturate-150"
			/>
		{/if}
	{/key}
</div>

<style>
	.glow-picture {
		opacity: 0.38;
		mask-image: radial-gradient(ellipse 70% 60% at 50% 30%, black 20%, transparent 75%);
		animation: glow-in 700ms ease-out both;
	}

	@keyframes glow-in {
		from {
			opacity: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.glow-picture {
			animation: none;
		}
	}
</style>
