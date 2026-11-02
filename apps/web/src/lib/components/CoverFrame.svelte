<!--
	The clickable picture of a card: its cover at a fixed aspect ratio, a
	placeholder icon when there is none, and badges drawn over it (children).
	Lifts when its card (a `group`) is hovered.
-->
<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { fadeInImage } from '$lib/fadeInImage';

	interface Props {
		src: string | null;
		alt: string;
		// CSS aspect ratio, e.g. "16 / 9", "1 / 1", "2 / 3".
		aspect: string;
		placeholder: Component;
		onclick: () => void;
		children?: Snippet;
	}

	let { src, alt, aspect, placeholder: Placeholder, onclick, children }: Props = $props();
</script>

<button
	type="button"
	{onclick}
	class="tv-card bg-base-300 relative block w-full overflow-hidden ring-1 ring-white/5 transition-transform duration-200 group-hover:scale-[1.02] group-hover:shadow-xl focus:outline-none"
	style:aspect-ratio={aspect}
>
	{#if src}
		<img {src} {alt} loading="lazy" use:fadeInImage class="h-full w-full object-cover" />
	{:else}
		<span class="flex h-full w-full items-center justify-center">
			<Placeholder size={40} class="opacity-20" />
		</span>
	{/if}
	{@render children?.()}
</button>
