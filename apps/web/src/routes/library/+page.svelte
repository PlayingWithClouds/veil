<!--
	What the user keeps: the watchlist, the play queue, their collections and
	their downloads, one tab each (?tab=, the watchlist by default). A tab
	loads the first time it's opened and stays, so switching back is instant.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import CollectionGrid from '$lib/components/CollectionGrid.svelte';
	import CollectionList from '$lib/components/CollectionList.svelte';
	import DownloadList from '$lib/components/DownloadList.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import Segmented from '$lib/components/Segmented.svelte';
	import { fetchBuiltInCollection } from '$lib/curation';
	import { libraryUrl, type LibraryTab } from '$lib/routes';

	const TABS: { value: LibraryTab; label: string }[] = [
		{ value: 'watchlist', label: 'Watchlist' },
		{ value: 'queue', label: 'Queue' },
		{ value: 'collections', label: 'Collections' },
		{ value: 'downloads', label: 'Downloads' }
	];

	// What the watchlist and the queue say while empty.
	const EMPTY = {
		watchlist: 'Nothing on your watchlist. Add videos from their page or card menu.',
		queue: 'Nothing queued up next.'
	};

	let tab = $derived(tabOf(page.url.searchParams.get('tab')));
	// The tabs opened so far; each stays in the page, hidden while another shows.
	let opened = $state<Set<LibraryTab>>(new Set());

	// Only the tab reruns this, not the set it adds to.
	$effect(() => {
		const current = tab;
		untrack(() => {
			if (!opened.has(current)) opened = new Set([...opened, current]);
		});
	});

	/** The tab a ?tab= value names, the watchlist when it names none. */
	function tabOf(value: string | null): LibraryTab {
		const found = TABS.find((choice) => choice.value === value);
		if (!found) return 'watchlist';
		return found.value;
	}
</script>

<svelte:head><title>Library · Veil</title></svelte:head>

{#snippet builtIn(kind: 'watchlist' | 'queue')}
	{#await fetchBuiltInCollection(kind)}
		<div class="flex justify-center py-10">
			<span class="loading loading-spinner"></span>
		</div>
	{:then collection}
		<CollectionGrid {collection} surface={kind} empty={EMPTY[kind]} />
	{:catch failure}
		<p class="text-error text-sm">{String(failure)}</p>
	{/await}
{/snippet}

<div class="flex flex-col gap-6">
	<PageHeader
		title="Library"
		description="What you keep: to watch later, up next, collected and downloaded."
	/>

	<div class="self-start">
		<Segmented
			label="Library"
			choices={TABS}
			current={tab}
			onchoose={(choice) => goto(libraryUrl(choice), { replaceState: true, noScroll: true })}
		/>
	</div>

	{#each TABS as choice (choice.value)}
		{#if opened.has(choice.value)}
			<div class="flex flex-col gap-6" class:hidden={tab !== choice.value}>
				{#if choice.value === 'watchlist' || choice.value === 'queue'}
					{@render builtIn(choice.value)}
				{:else if choice.value === 'collections'}
					<CollectionList />
				{:else}
					<DownloadList />
				{/if}
			</div>
		{/if}
	{/each}
</div>
