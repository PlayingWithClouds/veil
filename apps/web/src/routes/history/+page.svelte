<script lang="ts">
	import NsfwGrid from '$lib/components/NsfwGrid.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import { fetchWatchedScenes, type PluginSearchResult } from '$lib/search';

	let items = $state<PluginSearchResult[]>([]);
	let loading = $state(true);

	$effect(() => {
		fetchWatchedScenes().then((found) => {
			items = found;
			loading = false;
		});
	});
</script>

<div class="flex flex-col gap-6">
	<PageHeader title="History" description="What you watched, most recent first." />

	{#if !loading && items.length === 0}
		<div class="flex flex-col items-center gap-3 py-24 text-center">
			<span class="text-6xl opacity-10">🕑</span>
			<p class="text-base-content/40">Nothing watched yet.</p>
		</div>
	{:else}
		<NsfwGrid {items} loading={loading && items.length === 0} ariaLabel="History" />
	{/if}
</div>
