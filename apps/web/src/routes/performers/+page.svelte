<script lang="ts">
	import { goto } from '$app/navigation';
	import PerformerCard from '$lib/components/PerformerCard.svelte';

	let { data } = $props();
	let performers = $derived(data.performers);
</script>

<div class="flex flex-col gap-6">
	<h1 class="text-3xl font-bold">Performers</h1>

	{#if performers.length === 0}
		<div class="flex flex-col items-center gap-4 py-24 text-center">
			<span class="text-6xl opacity-10">👤</span>
			<p class="text-base-content/40">No performers yet</p>
		</div>
	{:else}
		<div class="grid grid-cols-3 gap-6 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8">
			{#each performers as performer (performer.id)}
				<PerformerCard {performer} onclick={() => goto(`/performers/${performer.id}`)} />
			{/each}
		</div>
	{/if}
</div>
