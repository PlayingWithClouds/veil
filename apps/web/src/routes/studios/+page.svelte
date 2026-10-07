<script lang="ts">
	import { goto } from '$app/navigation';
	import { cacheUrl } from '$lib/img';

	let { data } = $props();
	let studios = $derived(data.studios);
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
					<div class="tv-card bg-base-200 flex aspect-video w-full items-center justify-center overflow-hidden rounded-xl">
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
	{/if}
</div>
