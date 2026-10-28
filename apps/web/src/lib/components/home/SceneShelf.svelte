<script lang="ts">
	// A titled, horizontally scrolling row of scene cards on Home ("Continue
	// watching", "More from <performer>", …). The title links to what the row
	// is about when there is a page for it.
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import { impression, type ShownEvent } from '$lib/recommendations';
	import { pluginForUrl, type SearchPlugin } from '$lib/search';
	import type { SceneHit } from '$lib/explore';

	interface Props {
		title: string;
		// Page the title links to; null leaves it plain.
		href?: string | null;
		scenes: SceneHit[];
		sites: SearchPlugin[];
		// The impression of the card at a position; an empty mediaId logs nothing.
		impressionAt?: (scene: SceneHit, position: number) => ShownEvent;
		onopen: (scene: SceneHit, position: number) => void;
	}

	let { title, href = null, scenes, sites, impressionAt, onopen }: Props = $props();

	/** The impression of a card, or an empty one (not logged) without impressionAt. */
	function eventFor(scene: SceneHit, position: number): ShownEvent {
		if (impressionAt) return impressionAt(scene, position);
		return { mediaId: '', surface: 'row', position };
	}
</script>

<section class="flex flex-col gap-3">
	<h2 class="text-lg font-semibold">
		{#if href}
			<a {href} class="group inline-flex items-center gap-1 hover:underline">
				{title}
				<CaretRightIcon size={16} weight="bold" class="opacity-50 group-hover:opacity-100" />
			</a>
		{:else}
			{title}
		{/if}
	</h2>
	<div class="no-scrollbar -mx-1 flex gap-5 overflow-x-auto px-1 pb-1">
		{#each scenes as scene, position (scene.id)}
			<div class="w-72 shrink-0" use:impression={eventFor(scene, position)}>
				<ExploreSceneCard
					{scene}
					site={pluginForUrl(scene.sourceUrl, sites)}
					onclick={() => onopen(scene, position)}
				/>
			</div>
		{/each}
	</div>
</section>
