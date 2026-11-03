<!--
	A page of one media kind beyond videos (audio, stories): the library's,
	newest first, or with a search (?q=) what the sites that list the kind
	answer, as they answer. "Browse a site" stores a page of one site's
	catalog.
-->
<script lang="ts" generics="Item extends { id: string }">
	import { untrack, type Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { browse } from '$lib/library';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { searchPage, type SearchBatchRow, type SearchPage } from '$lib/search';

	const PAGE_SIZE = 40;

	interface Props {
		title: string;
		kind: 'AUDIO' | 'STORY';
		// What a search asks the sites for, by name ("stories").
		noun: string;
		// A page of the library's items.
		load: (limit: number, offset: number) => Promise<Item[]>;
		// The items of this kind in one site's search answer.
		pick: (batch: SearchBatchRow) => Item[];
		// Smallest card width in px.
		cardWidth: number;
		card: Snippet<[Item]>;
	}

	let { title, kind, noun, load, pick, cardWidth, card }: Props = $props();

	let items = $state<Item[]>([]);
	let loading = $state(false);
	let exhausted = $state(false);
	let error = $state<string | null>(null);
	let browsePlugin = $state('');
	let browsing = $state(false);
	let typed = $state('');
	let sitesAsked = $state(0);
	let sitesAnswered = $state(0);

	// The running search page, the sites that answered and those with nothing more.
	let running: SearchPage | null = null;
	let searchPageNumber = $state(0);
	let searchSites = new Set<string>();
	let drySites = new Set<string>();
	// Counts loads started, so a late answer to an earlier one is dropped.
	let generation = 0;

	let query = $derived(page.url.searchParams.get('q') ?? '');
	let kindSites = $derived(listers($sites, kind));

	$effect(() => {
		const current = query;
		untrack(() => {
			typed = current;
			void restart(current);
		});
		return () => running?.cancel();
	});

	/** Starts over: the library's first page, or the search's. */
	async function restart(searchText: string) {
		running?.cancel();
		running = null;
		generation++;
		items = [];
		exhausted = false;
		error = null;
		if (searchText === '') {
			await loadMore();
			return;
		}
		searchPageNumber = 0;
		searchSites = new Set();
		drySites = new Set();
		await runSearch(searchText, null);
	}

	/** Appends the next page of the library or of the search. */
	async function loadMore() {
		if (loading || exhausted) return;
		if (query !== '') {
			searchNextPage();
			return;
		}
		loading = true;
		const started = generation;
		try {
			const nextPage = await load(PAGE_SIZE, items.length);
			if (started !== generation) return;
			items = [...items, ...nextPage];
			exhausted = nextPage.length < PAGE_SIZE;
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Asks the sites with more results for their next page. */
	function searchNextPage() {
		const remaining = [...searchSites].filter((site) => !drySites.has(site));
		if (remaining.length === 0) {
			exhausted = true;
			return;
		}
		searchPageNumber++;
		void runSearch(query, remaining);
	}

	/** Asks the sites (only plugins, when given) for one page of the search. */
	async function runSearch(searchText: string, plugins: string[] | null) {
		loading = true;
		sitesAsked = 0;
		sitesAnswered = 0;
		const started = generation;
		const current = searchPage(kind, searchText, searchPageNumber, plugins, null, (batch) => {
			if (started === generation) receive(batch);
		});
		running = current;
		try {
			await current.done;
		} catch (failure) {
			error = String(failure);
		}
		if (running !== current) return;
		running = null;
		loading = false;
		exhausted = [...searchSites].every((site) => drySites.has(site));
	}

	/** Takes part of a site's answer, skipping what's shown already. */
	function receive(batch: SearchBatchRow) {
		sitesAsked = batch.sites;
		searchSites.add(batch.plugin);
		const shown = new Set(items.map((item) => item.id));
		const fresh = pick(batch).filter((item) => !shown.has(item.id));
		items = [...items, ...fresh];
		if (!batch.final) return;
		sitesAnswered++;
		if (batch.exhausted || batch.returned === 0) drySites.add(batch.plugin);
	}

	/** Puts the typed search in the URL, which starts it. */
	function submitSearch(event: SubmitEvent) {
		event.preventDefault();
		const url = new URL(page.url);
		const searchText = typed.trim();
		if (searchText === '') {
			url.searchParams.delete('q');
		} else {
			url.searchParams.set('q', searchText);
		}
		goto(url, { keepFocus: true });
	}

	/** Stores a page of the picked site's catalog, then shows the library again. */
	async function browseSite() {
		if (!browsePlugin) return;
		browsing = true;
		error = null;
		try {
			await browse(browsePlugin, kind);
			if (query !== '') {
				goto(page.url.pathname);
				return;
			}
			await restart('');
		} catch (failure) {
			error = String(failure);
		} finally {
			browsing = false;
		}
	}
</script>

<svelte:head><title>{title} · Veil</title></svelte:head>

<div class="flex flex-col gap-5">
	<PageHeader {title}>
		{#snippet actions()}
			{#if kindSites.length > 0}
				<select
					bind:value={browsePlugin}
					class="bg-base-200 h-9 rounded-lg px-2 text-sm"
					aria-label="Site to browse"
				>
					<option value="">Browse a site…</option>
					{#each kindSites as site (site.name)}
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

	<form class="flex max-w-xl items-center gap-2" onsubmit={submitSearch}>
		<label class="bg-base-200 flex h-10 flex-1 items-center gap-2 rounded-lg px-3">
			<MagnifyingGlassIcon size={16} class="text-base-content/50" />
			<input
				type="search"
				bind:value={typed}
				placeholder="Search the sites for {noun}"
				class="min-w-0 flex-1 bg-transparent text-sm outline-none"
			/>
		</label>
		{#if query !== ''}
			<button
				type="button"
				class="text-base-content/60 hover:text-base-content h-10 px-2 text-sm"
				onclick={() => goto(page.url.pathname)}>Clear</button
			>
		{/if}
	</form>

	{#if query !== '' && loading && searchPageNumber === 0 && sitesAsked > 0}
		<p class="text-base-content/50 -mt-2 text-sm">{sitesAnswered} of {sitesAsked} sites answered</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	<div
		class="grid gap-x-4 gap-y-6"
		style:grid-template-columns="repeat(auto-fill, minmax({cardWidth}px, 1fr))"
	>
		{#each items as item (item.id)}
			{@render card(item)}
		{/each}
	</div>

	{#if !loading && items.length === 0}
		{#if query !== ''}
			<p class="text-base-content/50 text-sm">No site found {noun} for this search.</p>
		{:else if kindSites.length === 0}
			<p class="text-base-content/50 text-sm">
				No {noun} yet, and no installed plugin lists them. Install one on the plugins page.
			</p>
		{:else}
			<p class="text-base-content/50 text-sm">No {noun} yet. Search or browse a site above.</p>
		{/if}
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && items.length > 0} done={exhausted} />
</div>
