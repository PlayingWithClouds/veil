<!--
	The chip row over the library grid, one line that scrolls sideways: the
	Filters button, then what narrows the grid now (each with ×), the pinned
	saved filters, recent searches, and the tags most common in the grid. A tag
	chip adds the tag to the filter.
-->
<script lang="ts">
	import FunnelSimpleIcon from 'phosphor-svelte/lib/FunnelSimpleIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import type { SavedFilterChip } from '$lib/home';
	import { savedFilterUrl } from '$lib/routes';
	import {
		toggledListParam,
		clearedFilterUrl,
		isFiltered,
		type TagCountRow
	} from '$lib/libraryFilter';

	interface Props {
		url: URL;
		// Names of the tags, performers and studios the URL filters on.
		names: Map<string, string>;
		tagCounts: TagCountRow[];
		pinnedFilters: SavedFilterChip[];
		recentSearches: string[];
		// How many panel settings are on, shown on the Filters button.
		panelCount: number;
		panelOpen: boolean;
		ontogglepanel: () => void;
	}

	let {
		url,
		names,
		tagCounts,
		pinnedFilters,
		recentSearches,
		panelCount,
		panelOpen,
		ontogglepanel
	}: Props = $props();

	let savedId = $derived(url.searchParams.get('saved'));
	let activeChips = $derived.by(() => {
		const chips: { param: string; id: string; label: string; excluded: boolean }[] = [];
		for (const param of ['performer', 'studio', 'tag', 'not']) {
			for (const id of url.searchParams.getAll(param)) {
				chips.push({ param, id, label: names.get(id) ?? '…', excluded: param === 'not' });
			}
		}
		return chips;
	});
</script>

{#snippet chipLink(href: string, label: string, active: boolean)}
	<a
		{href}
		data-sveltekit-noscroll
		class="shrink-0 rounded-full px-3 py-1 text-sm whitespace-nowrap hover:brightness-125"
		class:bg-base-200={!active}
		class:bg-base-content={active}
		class:text-base-100={active}
	>
		{label}
	</a>
{/snippet}

<div class="no-scrollbar fade-right-edge -mx-1 flex items-center gap-2 overflow-x-auto px-1 py-0.5">
	<button
		type="button"
		onclick={ontogglepanel}
		aria-expanded={panelOpen}
		class="flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-sm hover:brightness-125"
		class:bg-base-200={!panelOpen}
		class:bg-base-content={panelOpen}
		class:text-base-100={panelOpen}
	>
		<FunnelSimpleIcon size={16} />
		Filters
		{#if panelCount > 0}
			<span class="bg-base-content/15 rounded-full px-1.5 text-xs tabular-nums">{panelCount}</span>
		{/if}
	</button>

	{#each activeChips as chip (chip.param + chip.id)}
		<a
			href={toggledListParam(url, chip.param, chip.id)}
			data-sveltekit-noscroll
			class="flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-sm whitespace-nowrap"
			class:bg-base-content={!chip.excluded}
			class:text-base-100={!chip.excluded}
			class:bg-error={chip.excluded}
			class:text-error-content={chip.excluded}
			aria-label="Remove {chip.label}"
		>
			{#if chip.excluded}
				<span class="line-through">{chip.label}</span>
			{:else}
				{chip.label}
			{/if}
			<XIcon size={14} />
		</a>
	{/each}
	{#if isFiltered(url.searchParams)}
		<a
			href={clearedFilterUrl(url)}
			data-sveltekit-noscroll
			class="text-base-content/60 hover:text-base-content shrink-0 px-1 text-sm whitespace-nowrap"
		>
			Clear all
		</a>
	{/if}

	{#each pinnedFilters as saved (saved.id)}
		{@render chipLink(
			savedId === saved.id ? '/' : savedFilterUrl(saved.id),
			`★ ${saved.name}`,
			savedId === saved.id
		)}
	{/each}

	{#each recentSearches as text (text)}
		<a
			href="/?q={encodeURIComponent(text)}"
			class="bg-base-200 hover:bg-base-300 flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-sm whitespace-nowrap"
		>
			<MagnifyingGlassIcon size={14} class="opacity-60" />
			{text}
		</a>
	{/each}

	{#each tagCounts as entry (entry.tag.id)}
		{@render chipLink(toggledListParam(url, 'tag', entry.tag.id), entry.tag.name, false)}
	{/each}
</div>
