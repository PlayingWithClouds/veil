<!--
	The top recommended videos at the head of home and search: a heading over
	the picks as large video cards side by side, filling the width (four
	across a desktop window), each noting why it's picked. The first pick's
	cover lights the app's shared glow.
-->
<script lang="ts">
	import SceneCard from '$lib/components/SceneCard.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import type { RecommendedVideo } from '$lib/home';
	import { cacheUrl } from '$lib/img';
	import type { VideoCardRow } from '$lib/library';

	interface Props {
		// The picks, best first, each with why it's picked.
		picks: RecommendedVideo[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		// The row's title, e.g. "Top picks for you", "Best match".
		label?: string;
		onopen: (video: VideoCardRow) => void;
	}

	let { picks, surface, label = 'Top picks for you', onopen }: Props = $props();

	showAmbientGlow(() => cacheUrl(picks[0]?.video.cover?.uri));
</script>

<section class="flex flex-col gap-3">
	<h2 class="text-lg font-semibold">{label}</h2>
	<div class="grid grid-cols-4 gap-4 max-xl:grid-cols-2 max-md:grid-cols-1">
		{#each picks as pick, position (pick.video.id)}
			<SceneCard
				video={pick.video}
				note={pick.reason}
				shown={{ surface, position }}
				onclick={() => onopen(pick.video)}
			/>
		{/each}
	</div>
</section>
