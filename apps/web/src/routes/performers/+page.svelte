<script lang="ts">
	import { goto } from '$app/navigation';
	import PerformerCard from '$lib/components/PerformerCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { loadPerformers, type PerformerCard as PerformerCardData } from '$lib/performerDetail';

	// The loader brings the first page; the rest load as you scroll. The
	// backend serves at most 100 per page (media.maxPageLimit).
	const PAGE_SIZE = 100;

	let { data } = $props();
	// Follows the loader's data; loadMore appends to it.
	let performers = $derived<PerformerCardData[]>(data.performers);
	let loading = $state(false);
	let done = $derived(data.performers.length < PAGE_SIZE);

	/** Appends the next page of performers. */
	async function loadMore() {
		if (loading || done) return;
		loading = true;
		const next = await loadPerformers(PAGE_SIZE, performers.length);
		performers = [...performers, ...next];
		done = next.length < PAGE_SIZE;
		loading = false;
	}
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
		<ScrollSentinel onvisible={loadMore} {loading} {done} />
	{/if}
</div>
