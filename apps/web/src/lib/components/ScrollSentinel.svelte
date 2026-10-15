<script lang="ts">
	// Invisible marker after a list: calls onvisible when it comes within
	// LOOKAHEAD of the viewport, so the next page loads before the end shows.
	// Shows a spinner while loading; renders nothing once done.
	interface Props {
		onvisible: () => void;
		loading?: boolean;
		done?: boolean;
	}

	let { onvisible, loading = false, done = false }: Props = $props();

	const LOOKAHEAD = '800px';

	let marker = $state<HTMLElement | null>(null);

	$effect(() => {
		if (!marker || done) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) onvisible();
			},
			{ rootMargin: LOOKAHEAD }
		);
		observer.observe(marker);
		return () => observer.disconnect();
	});
</script>

{#if !done}
	<div bind:this={marker} class="flex h-12 items-center justify-center">
		{#if loading}
			<span class="loading loading-spinner text-base-content/40"></span>
		{/if}
	</div>
{/if}
