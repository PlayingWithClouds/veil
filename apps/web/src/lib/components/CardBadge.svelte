<!--
	A small label in a corner of a card's picture: a duration, a count, Live.
	"live" draws it red; the others are dark glass.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		corner: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
		live?: boolean;
		children: Snippet;
	}

	let { corner, live = false, children }: Props = $props();
</script>

<span
	class="absolute flex items-center gap-1 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
	class:top-1.5={corner === 'top-left' || corner === 'top-right'}
	class:bottom-1.5={corner === 'bottom-left' || corner === 'bottom-right'}
	class:left-1.5={corner === 'top-left' || corner === 'bottom-left'}
	class:right-1.5={corner === 'top-right' || corner === 'bottom-right'}
	class:badge-live={live}
>
	{@render children()}
</span>

<style>
	/* Scoped, so it outranks the dark default fill. */
	.badge-live {
		background-color: var(--color-error);
	}
</style>
