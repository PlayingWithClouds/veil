<!--
	Sticky chip row under the top bar (YouTube's filter chips): "All sites" plus
	one chip per search plugin that offers the selected media type (videos or
	images). The selection persists across pages.
-->
<script lang="ts">
	import { browser } from '$app/environment';
	import type { Snippet } from 'svelte';
	import {
		carriesContent,
		fetchSearchPlugins,
		supportsMediaFilter,
		type ContentKind,
		type SearchPlugin
	} from '$lib/search';
	import { mediaFilter } from '$lib/stores/mediaFilter';
	import { chromeHidden, isCompact } from '$lib/stores/viewport';
	import WarningCircleIcon from 'phosphor-svelte/lib/WarningCircleIcon';
	import type { SiteStatus } from '$lib/liveSearch.svelte';

	interface Props {
		// Selected plugin names (empty = all), two-way bound to the parent.
		selectedPlugins?: string[];
		// Chips placed before "All sites" (Home's media chip on phones).
		leading?: Snippet;
		// While a search runs: each searched site's state and result count.
		status?: Map<string, SiteStatus>;
		// The kind of content a search asks for: only the sites carrying it get a chip.
		content?: ContentKind | null;
	}

	let { selectedPlugins = $bindable([]), leading, status, content = null }: Props = $props();

	const SOURCES_STORAGE_KEY = 'veil:search-sources';

	if (browser && selectedPlugins.length === 0) {
		selectedPlugins = loadSelectedPlugins();
	}

	let searchPlugins = $state<SearchPlugin[]>([]);

	// Sites that don't offer the selected media type get no chip. A hidden site
	// stays in the stored selection, so it comes back with the media type.
	let visiblePlugins = $derived(
		searchPlugins.filter(
			(plugin) => supportsMediaFilter(plugin, $mediaFilter) && carriesContent(plugin, content)
		)
	);
	let hasVisibleSelection = $derived(
		visiblePlugins.some((plugin) => selectedPlugins.includes(plugin.name))
	);

	/** Reads the stored selection, ignoring anything malformed. */
	function loadSelectedPlugins(): string[] {
		try {
			const stored = localStorage.getItem(SOURCES_STORAGE_KEY);
			if (!stored) return [];
			const parsed = JSON.parse(stored);
			if (!Array.isArray(parsed)) return [];
			return parsed.filter((entry) => typeof entry === 'string');
		} catch {
			return [];
		}
	}

	/** The plugin's display label. */
	function pluginLabel(plugin: SearchPlugin): string {
		if (plugin.displayName) return plugin.displayName;
		return plugin.name;
	}

	/** Hover text for a site's search status, undefined without a search. */
	function statusTitle(siteStatus: SiteStatus | undefined): string | undefined {
		if (!siteStatus) return undefined;
		if (siteStatus.state === 'searching') return 'Searching…';
		if (siteStatus.state === 'failed') return 'This site could not be searched';
		if (siteStatus.count === 1) return '1 result';
		return `${siteStatus.count} results`;
	}

	/** Adds or removes one site from the selection. */
	function togglePlugin(name: string) {
		if (selectedPlugins.includes(name)) {
			selectedPlugins = selectedPlugins.filter((entry) => entry !== name);
			return;
		}
		selectedPlugins = [...selectedPlugins, name];
	}

	$effect(() => {
		fetchSearchPlugins().then((plugins) => {
			searchPlugins = plugins
				.filter((plugin) => plugin.enabled)
				.sort((first, second) => pluginLabel(first).localeCompare(pluginLabel(second)));
		});
	});

	$effect(() => {
		if (!browser) return;
		localStorage.setItem(SOURCES_STORAGE_KEY, JSON.stringify(selectedPlugins));
	});
</script>

<div
	class="site-chips bg-base-100 top-chrome sticky z-20 -mx-4 px-4 py-3 max-md:-mx-3 max-md:px-3"
	class:site-chips-hidden={$isCompact && $chromeHidden}
>
	<div class="no-scrollbar flex items-center gap-3 overflow-x-auto">
		{@render leading?.()}
		<button
			type="button"
			aria-pressed={!hasVisibleSelection}
			onclick={() => (selectedPlugins = [])}
			class="site-chip"
			class:site-chip-active={!hasVisibleSelection}
		>
			All sites
		</button>
		{#each visiblePlugins as plugin (plugin.name)}
			{@const siteStatus = status?.get(plugin.name)}
			<button
				type="button"
				aria-pressed={selectedPlugins.includes(plugin.name)}
				onclick={() => togglePlugin(plugin.name)}
				class="site-chip"
				class:site-chip-active={selectedPlugins.includes(plugin.name)}
				class:site-chip-empty={siteStatus &&
					siteStatus.state !== 'searching' &&
					siteStatus.count === 0}
				title={statusTitle(siteStatus)}
			>
				{#if plugin.iconUrl}
					<img src={plugin.iconUrl} alt="" class="h-4 w-4 rounded-sm" />
				{/if}
				{pluginLabel(plugin)}
				{#if siteStatus?.state === 'searching'}
					<span class="loading loading-spinner loading-xs opacity-60"></span>
				{:else if siteStatus?.state === 'failed'}
					<WarningCircleIcon size={14} class="text-warning" />
				{:else if siteStatus && siteStatus.count > 0}
					<span class="site-chip-count">{siteStatus.count}</span>
				{/if}
			</button>
		{/each}
	</div>
</div>

<style>
	.site-chips {
		transition: translate 280ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	/* Slides up with the top bar while the page scrolls down. */
	.site-chips-hidden {
		translate: 0 -100%;
	}

	@media (prefers-reduced-motion: reduce) {
		.site-chips {
			transition: none;
		}
	}

	.site-chip {
		display: flex;
		flex-shrink: 0;
		align-items: center;
		gap: 0.4rem;
		height: 2rem;
		padding-inline: 0.75rem;
		border-radius: 0.5rem;
		background-color: var(--color-base-200);
		font-size: 14px;
		font-weight: 500;
		transition: background-color 150ms;
	}

	.site-chip:hover {
		background-color: var(--color-base-300);
	}

	.site-chip-active,
	.site-chip-active:hover {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}

	/* A site whose search came back with nothing. */
	.site-chip-empty:not(.site-chip-active) {
		opacity: 0.45;
	}

	.site-chip-count {
		font-size: 12px;
		font-variant-numeric: tabular-nums;
		opacity: 0.6;
	}
</style>
