<script lang="ts">
	import { goto } from '$app/navigation';
	import TagSimpleIcon from 'phosphor-svelte/lib/TagSimpleIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import { tagUrl } from '$lib/routes';
	import { fetchTagIndex, type TagSummary } from '$lib/tags';

	let tags = $state<TagSummary[]>([]);
	let loading = $state(true);
	let search = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		load('');
	});

	async function load(query: string) {
		loading = true;
		tags = await fetchTagIndex(query);
		loading = false;
	}

	function onSearchInput() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => load(search.trim()), 250);
	}
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-center gap-3">
		<TagSimpleIcon size={26} weight="fill" class="text-base-content/70" />
		<h1 class="text-3xl font-bold">Tags</h1>
	</div>

	<label
		class="bg-base-200 border-base-300 focus-within:border-base-content/30 flex h-11 max-w-md items-center gap-2 rounded-lg border px-3"
	>
		<MagnifyingGlassIcon size={16} class="text-base-content/40" />
		<input
			type="text"
			bind:value={search}
			oninput={onSearchInput}
			placeholder="Search tags"
			class="flex-1 bg-transparent text-sm outline-none"
		/>
	</label>

	{#if loading}
		<div class="flex justify-center py-16">
			<span class="loading loading-spinner text-base-content/40"></span>
		</div>
	{:else if tags.length === 0}
		<div class="flex flex-col items-center gap-3 py-20 text-center">
			<span class="text-6xl opacity-10">🏷️</span>
			<p class="text-base-content/40">No tags yet. They appear as content is scraped.</p>
		</div>
	{:else}
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
			{#each tags as tag (tag.id)}
				<button
					type="button"
					class="tv-card border-base-300 bg-base-200 hover:border-base-content/20 flex flex-col gap-1 rounded-xl border p-4 text-left transition-colors"
					onclick={() => goto(tagUrl(tag.id))}
				>
					<span class="line-clamp-1 text-sm font-semibold">{tag.name}</span>
					<span class="text-base-content/40 text-xs">
						{tag.sceneCount}
						{tag.sceneCount === 1 ? 'scene' : 'scenes'}
					</span>
					{#if tag.category}
						<span class="badge badge-sm mt-1">{tag.category}</span>
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>
