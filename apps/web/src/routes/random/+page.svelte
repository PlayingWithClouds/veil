<script lang="ts">
	import { goto } from '$app/navigation';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import ExploreSceneCard from '$lib/components/ExploreSceneCard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { sceneUrl } from '$lib/routes';
	import { fetchSearchPlugins, pluginForUrl, type SearchPlugin } from '$lib/search';
	import { fetchRandomSceneHits, type SceneHit } from '$lib/explore';

	// Random mode: five picks from everything stored (one featured, four
	// beside it), reshuffled on demand. Earlier picks stay below, so a
	// reshuffle never loses one the user meant to open.
	const COUNT = 5;
	const EARLIER_LIMIT = 40;

	let scenes = $state<SceneHit[]>([]);
	let earlier = $state<SceneHit[]>([]);
	let loading = $state(false);
	let surprising = $state(false);
	let sites = $state<SearchPlugin[]>([]);

	let featured = $derived(scenes[0]);
	let others = $derived(scenes.slice(1));

	/** Replaces the picks with a fresh random set, moving the old ones to Earlier picks. */
	async function shuffle() {
		loading = true;
		try {
			const next = await fetchRandomSceneHits(COUNT);
			keepAsEarlier(scenes, next);
			scenes = next;
		} finally {
			loading = false;
		}
	}

	/** Puts the replaced picks first in Earlier picks, without repeats or the new picks. */
	function keepAsEarlier(replaced: SceneHit[], next: SceneHit[]) {
		const shown = new Set(next.map((scene) => scene.id));
		const kept: SceneHit[] = [];
		for (const scene of [...replaced, ...earlier]) {
			if (shown.has(scene.id)) continue;
			shown.add(scene.id);
			kept.push(scene);
		}
		earlier = kept.slice(0, EARLIER_LIMIT);
	}

	/** Opens one random scene straight away. */
	async function surpriseMe() {
		surprising = true;
		try {
			const [pick] = await fetchRandomSceneHits(1);
			if (pick) goto(sceneUrl(pick.id));
		} finally {
			surprising = false;
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
			<div class="flex items-center gap-2">
				<button
					type="button"
					class="bg-base-200 hover:bg-base-300 flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold transition-colors disabled:opacity-50"
					onclick={surpriseMe}
					disabled={surprising}
				>
					{#if surprising}
						<span class="loading loading-spinner loading-xs"></span>
					{:else}
						<PlayIcon size={16} weight="fill" />
					{/if}
					Surprise me
				</button>
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
			</div>
		{/snippet}
	</PageHeader>

	{#if scenes.length === 0 && !loading}
		<div class="border-base-300 flex flex-col items-center gap-3 rounded-xl border border-dashed py-24 text-center">
			<span class="text-base-content/20"><ShuffleIcon size={48} /></span>
			<p class="text-base-content/50 text-sm">Nothing stored yet. Search for something first.</p>
		</div>
	{:else if featured}
		<div class="grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-2 xl:grid-cols-4">
			<div class="sm:col-span-2 xl:row-span-2">
				<ExploreSceneCard
					scene={featured}
					site={pluginForUrl(featured.sourceUrl, sites)}
					onclick={() => goto(sceneUrl(featured.id))}
				/>
			</div>
			{#each others as scene (scene.id)}
				<ExploreSceneCard
					{scene}
					site={pluginForUrl(scene.sourceUrl, sites)}
					onclick={() => goto(sceneUrl(scene.id))}
				/>
			{/each}
		</div>
	{/if}

	{#if earlier.length > 0}
		<section class="flex flex-col gap-3">
			<h2 class="text-lg font-semibold">Earlier picks</h2>
			<div class="earlier-grid grid gap-x-5 gap-y-6">
				{#each earlier as scene (scene.id)}
					<ExploreSceneCard
						{scene}
						site={pluginForUrl(scene.sourceUrl, sites)}
						onclick={() => goto(sceneUrl(scene.id))}
					/>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style>
	.earlier-grid {
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
	}
</style>
