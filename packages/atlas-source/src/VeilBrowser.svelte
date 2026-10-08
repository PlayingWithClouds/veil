<script lang="ts">
	// Searchable thumbnail grid over veil's galleries or scenes; opens one by id.
	import { getContext, onMount } from 'svelte';
	import type { Context } from '@neoworks/extension-system';

	interface ListedItem {
		id: string;
		title: string;
		thumbnail?: string;
	}

	let {
		kind,
		label,
		open
	}: {
		kind: string;
		label: string;
		open: (params: Record<string, unknown>) => Promise<void>;
	} = $props();

	const PAGE_SIZE = 40;
	// Same key the host shell provides in @atlas/web/kernel.
	const ctx = getContext<Context>(Symbol.for('@atlas/web:kernel-context'));

	let items = $state<ListedItem[]>([]);
	let search = $state('');
	let loading = $state(false);
	let hasMore = $state(false);
	let openingId = $state<string | null>(null);
	let error = $state('');

	onMount(() => {
		void load(true);
	});

	function messageOf(failure: unknown): string {
		if (failure instanceof Error) {
			return failure.message;
		}
		return String(failure);
	}

	async function load(reset: boolean) {
		loading = true;
		error = '';
		const offset = reset ? 0 : items.length;
		try {
			const query = `search=${encodeURIComponent(search)}&limit=${PAGE_SIZE}&offset=${offset}`;
			const listing = await ctx.api.get<{ items: ListedItem[] }>(`/sources/veil/${kind}/items?${query}`);
			items = reset ? listing.items : [...items, ...listing.items];
			hasMore = listing.items.length === PAGE_SIZE;
		} catch (failure) {
			error = messageOf(failure);
		} finally {
			loading = false;
		}
	}

	async function openItem(item: ListedItem) {
		openingId = item.id;
		error = '';
		try {
			await open({ id: item.id });
		} catch (failure) {
			error = messageOf(failure);
			openingId = null;
		}
	}
</script>

{#if error}<div class="alert alert-error mb-4 text-sm">{error}</div>{/if}

<div class="mb-4 flex items-center gap-2">
	<input
		class="input input-sm max-w-xs flex-1"
		placeholder={`Search ${label}…`}
		bind:value={search}
		onkeydown={(event) => event.key === 'Enter' && load(true)}
	/>
	<button type="button" class="btn btn-sm" disabled={loading} onclick={() => load(true)}>Search</button>
</div>

{#if !items.length && !loading}
	<div class="text-dim py-16 text-center text-sm">Nothing found.</div>
{:else}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-3">
		{#each items as item (item.id)}
			<button
				type="button"
				onclick={() => openItem(item)}
				class="border-line bg-surface hover:border-line-strong group flex flex-col overflow-hidden rounded-lg border text-left transition"
			>
				<div class="bg-base-100 relative aspect-[4/3] overflow-hidden">
					{#if item.thumbnail}
						<img src={item.thumbnail} alt="" class="h-full w-full object-cover transition group-hover:opacity-90" loading="lazy" />
					{/if}
					{#if openingId === item.id}
						<div class="absolute inset-0 flex items-center justify-center bg-black/50"><span class="loading"></span></div>
					{/if}
				</div>
				<div class="p-2"><span class="text-default block truncate text-xs font-medium">{item.title}</span></div>
			</button>
		{/each}
	</div>
{/if}

{#if loading}
	<div class="text-dim flex justify-center py-8"><span class="loading"></span></div>
{:else if hasMore}
	<div class="flex justify-center py-4">
		<button type="button" class="btn btn-sm" onclick={() => load(false)}>Load more</button>
	</div>
{/if}
