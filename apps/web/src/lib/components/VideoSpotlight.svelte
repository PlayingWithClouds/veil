<!--
	The top recommended videos at the head of home and search: a row of large
	video cards, each noting why it's picked. Cards stay near the size most
	covers are scraped at, so low-resolution covers aren't blown up. The first
	pick's cover lights the app's shared glow.
-->
<script lang="ts">
	import SceneCard from '$lib/components/SceneCard.svelte';
	import Shelf from '$lib/components/Shelf.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import type { RecommendedVideo } from '$lib/home';
	import { cacheUrl } from '$lib/img';
	import type { VideoCardRow } from '$lib/library';

	interface Props {
		// The picks, best first, each with why it's picked.
		picks: RecommendedVideo[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		// The row's title, e.g. "Top pick for you", "Best match".
		label?: string;
		onopen: (video: VideoCardRow) => void;
	}

	let { picks, surface, label = 'Top picks for you', onopen }: Props = $props();

	// About three cards across a desktop window; most scraped covers are 400–640px wide.
	const CARD_WIDTH = 440;

	showAmbientGlow(() => cacheUrl(picks[0]?.video.cover?.uri));
</script>

<Shelf title={label} items={picks} key={(pick) => pick.video.id} itemWidth={CARD_WIDTH}>
	{#snippet card(pick, position)}
		<SceneCard
			video={pick.video}
			note={pick.reason}
			shown={{ surface, position }}
			onclick={() => onopen(pick.video)}
		/>
	{/snippet}
</Shelf>
