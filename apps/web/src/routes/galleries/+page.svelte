<!--
	Galleries in the library, newest first, plus "Browse a site" to store a
	page of one site's photo sets.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { fetchGalleries, browse, type GalleryCardRow } from '$lib/library';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { galleryUrl } from '$lib/routes';

	const PAGE_SIZE = 40;

	let galleries = $state<GalleryCardRow[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let browsePlugin = $state('');
	let browsing = $state(false);

	let gallerySites = $derived(listers($sites, 'GALLERY'));

	$effect(() => {
		void loadFirst();
	});

	/** Loads the first page of the library's galleries. */
	async function loadFirst() {
		loading = true;
		error = null;
		try {
			galleries = await fetchGalleries({}, PAGE_SIZE);
			exhausted = galleries.length < PAGE_SIZE;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Appends the next page. */
	async function loadMore() {
		if (loading || exhausted) return;
		loading = true;
		try {
			const nextPage = await fetchGalleries({}, PAGE_SIZE, galleries.length);
			galleries = [...galleries, ...nextPage];
			exhausted = nextPage.length < PAGE_SIZE;
		} finally {
			loading = false;
		}
	}

	/** Stores a page of the picked site's galleries, then reloads. */
	async function browseSite() {
		if (!browsePlugin) return;
		browsing = true;
		error = null;
		try {
			await browse(browsePlugin, 'GALLERY');
			await loadFirst();
		} catch (failure) {
			error = String(failure);
		} finally {
			browsing = false;
		}
	}
</script>

<svelte:head><title>Galleries · Veil</title></svelte:head>

<div class="flex flex-col gap-5">
	<PageHeader title="Galleries">
		{#snippet actions()}
			{#if gallerySites.length > 0}
				<select
					bind:value={browsePlugin}
					class="bg-base-200 h-9 rounded-lg px-2 text-sm"
					aria-label="Site to browse"
				>
					<option value="">Browse a site…</option>
					{#each gallerySites as site (site.name)}
						<option value={site.name}>{site.displayName ?? site.name}</option>
					{/each}
				</select>
				<button
					type="button"
					class="bg-primary text-primary-content h-9 rounded-lg px-3 text-sm font-medium disabled:opacity-50"
					disabled={!browsePlugin || browsing}
					onclick={browseSite}
				>
					{browsing ? 'Loading…' : 'Browse'}
				</button>
			{/if}
		{/snippet}
	</PageHeader>

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	<div class="grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-x-4 gap-y-6">
		{#each galleries as gallery (gallery.id)}
			<GalleryCard {gallery} onclick={() => goto(galleryUrl(gallery.id))} />
		{/each}
	</div>

	{#if !loading && galleries.length === 0}
		<p class="text-base-content/50 text-sm">No galleries yet. Browse a site above.</p>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && galleries.length > 0} done={exhausted} />
</div>
