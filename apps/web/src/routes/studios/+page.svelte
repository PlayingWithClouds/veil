<script lang="ts">
	import { goto } from '$app/navigation';
	import { cacheUrl } from '$lib/img';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { loadStudios, type StudioCard } from '$lib/studioDetail';

	// The loader brings the first page; the rest load as you scroll. The
	// backend serves at most 100 per page (media.maxPageLimit).
	const PAGE_SIZE = 100;

	let { data } = $props();
	// Follows the loader's data; loadMore appends to it.
	let studios = $derived<StudioCard[]>(data.studios);
	let loading = $state(false);
	let done = $derived(data.studios.length < PAGE_SIZE);

	/** Appends the next page of studios. */
	async function loadMore() {
		if (loading || done) return;
		loading = true;
		const next = await loadStudios(PAGE_SIZE, studios.length);
		studios = [...studios, ...next];
		done = next.length < PAGE_SIZE;
		loading = false;
	}
</script>

<div class="flex flex-col gap-6">
	<h1 class="text-3xl font-bold">Studios</h1>

	{#if studios.length === 0}
		<div class="flex flex-col items-center gap-4 py-24 text-center">
			<span class="text-6xl opacity-10">🎬</span>
			<p class="text-base-content/40">No studios yet</p>
		</div>
	{:else}
		<div class="grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
			{#each studios as studio (studio.id)}
				<button
					type="button"
					class="flex flex-col gap-2 text-left"
					onclick={() => goto(`/studios/${studio.id}`)}
				>
					<div
						class="tv-card bg-base-200 flex aspect-video w-full items-center justify-center overflow-hidden rounded-xl"
					>
						{#if cacheUrl(studio.imagePath)}
							<img
								src={cacheUrl(studio.imagePath)}
								alt={studio.name}
								loading="lazy"
								class="h-full w-full object-cover"
							/>
						{:else}
							<span class="text-4xl opacity-10">🎬</span>
						{/if}
					</div>
					<span class="line-clamp-1 text-sm font-medium">{studio.name}</span>
					<span class="text-base-content/40 text-xs">{studio.sceneCount} scenes</span>
				</button>
			{/each}
		</div>
		<ScrollSentinel onvisible={loadMore} {loading} {done} />
	{/if}
</div>
