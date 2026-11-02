<!--
	What the recommendations think you like and dislike: tags, performers,
	studios and sites, from what you watched, liked and skipped.
-->
<script lang="ts">
	import PageHeader from '$lib/components/PageHeader.svelte';
	import TasteList from '$lib/components/TasteList.svelte';
	import { fetchTaste, type Taste } from '$lib/taste';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { performerUrl, studioUrl, tagFilterUrl } from '$lib/routes';

	let taste = $state<Taste | null>(null);
	let error = $state<string | null>(null);

	// Sites are named by plugin; show their display name.
	let siteEntries = $derived(
		(taste?.sites ?? []).map((entry) => ({ ...entry, name: siteDisplayName(entry.id, $sites) }))
	);

	$effect(() => {
		fetchTaste()
			.then((found) => (taste = found))
			.catch((failure) => (error = String(failure)));
	});
</script>

<svelte:head><title>Taste · Veil</title></svelte:head>

<div class="flex flex-col gap-6">
	<PageHeader
		title="Your taste"
		description="What recommendations are based on: what you watch to the end, like, skip and pass over."
	/>

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{:else if taste && taste.signalCount === 0}
		<p class="text-base-content/50 text-sm">
			Nothing to go on yet. Watch a few videos and this fills in.
		</p>
	{:else if taste}
		<p class="text-base-content/50 -mt-3 text-sm">
			Based on {taste.signalCount}
			{taste.signalCount === 1 ? 'signal' : 'signals'}.
		</p>
		<div class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6">
			<TasteList title="Tags" entries={taste.tags} href={(entry) => tagFilterUrl(entry.id)} />
			<TasteList
				title="Performers"
				entries={taste.performers}
				href={(entry) => performerUrl(entry.id)}
			/>
			<TasteList title="Studios" entries={taste.studios} href={(entry) => studioUrl(entry.id)} />
			<TasteList title="Sites" entries={siteEntries} />
		</div>
	{/if}
</div>
