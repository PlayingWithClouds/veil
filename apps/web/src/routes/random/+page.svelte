<script lang="ts">
	import { goto } from '$app/navigation';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { sceneUrl } from '$lib/routes';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { fetchRandomSceneHits, type SceneHit } from '$lib/explore';

	// Random mode: five picks from everything stored, reshuffled on demand.
	const COUNT = 5;

	let scenes = $state<SceneHit[]>([]);
	let loading = $state(false);
	let sites = $state<SearchPlugin[]>([]);

	/** Replaces the picks with a fresh random set. */
	async function shuffle() {
		loading = true;
		try {
			scenes = await fetchRandomSceneHits(COUNT);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		shuffle();
		fetchSearchPlugins().then((plugins) => (sites = plugins));
	});

	/** Reshuffles on "r" when no input has focus. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key !== 'r' || event.ctrlKey || event.metaKey) return;
		if (event.target instanceof HTMLInputElement) return;
		shuffle();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="flex w-full flex-col gap-6">
	<PageHeader title="Random" description="Five random picks from everything you've found. Press R to reshuffle.">
		{#snippet actions()}
		<button
			type="button"
			class="bg-base-content text-base-100 flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold transition-transform hover:scale-105 disabled:opacity-50"
			onclick={shuffle}
			disabled={loading}
		>
			{#if loading}
				<span class="loading loading-spinner loading-xs"></span>
			{:else}
				<ShuffleIcon size={16} weight="bold" />
			{/if}
			Shuffle
		</button>
		{/snippet}
	</PageHeader>

	{#if scenes.length === 0 && !loading}
		<div class="border-base-300 flex flex-col items-center gap-3 rounded-xl border border-dashed py-24 text-center">
			<span class="text-base-content/20"><ShuffleIcon size={48} /></span>
			<p class="text-base-content/50 text-sm">Nothing stored yet. Search for something first.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
			{#each scenes as scene (scene.id)}
				<ExploreSceneCard
					{scene}
					site={pluginForUrl(scene.sourceUrl, sites)}
					onclick={() => goto(sceneUrl(scene.id))}
				/>
			{/each}
		</div>
	{/if}
</div>
