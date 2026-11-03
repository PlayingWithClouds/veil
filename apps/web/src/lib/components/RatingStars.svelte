<!--
	Five stars to rate a record. Ratings are 0 to 10, so a star is worth two;
	hovering previews, clicking the star of the current rating clears it.
-->
<script lang="ts">
	import StarIcon from 'phosphor-svelte/lib/StarIcon';

	interface Props {
		// 0 to 10, null when unrated.
		rating: number | null;
		onrate: (rating: number | null) => void;
	}

	let { rating, onrate }: Props = $props();

	const STARS = [1, 2, 3, 4, 5];

	let hovered = $state<number | null>(null);

	let shownStars = $derived.by(() => {
		if (hovered !== null) return hovered;
		if (rating === null) return 0;
		return Math.round(rating / 2);
	});

	/** Rates with stars, or clears the rating when it already is that. */
	function rate(stars: number) {
		if (rating !== null && Math.round(rating / 2) === stars) {
			onrate(null);
			return;
		}
		onrate(stars * 2);
	}
</script>

<div class="flex items-center" role="group" aria-label="Rating" onpointerleave={() => (hovered = null)}>
	{#each STARS as stars (stars)}
		<button
			type="button"
			class="text-warning p-0.5 transition-transform hover:scale-110"
			aria-label="{stars} of 5"
			aria-pressed={rating !== null && Math.round(rating / 2) === stars}
			onpointerenter={() => (hovered = stars)}
			onclick={() => rate(stars)}
		>
			<StarIcon size={20} weight={stars <= shownStars ? 'fill' : 'regular'} />
		</button>
	{/each}
</div>
