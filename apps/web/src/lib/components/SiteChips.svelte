<!--
	Sticky chip row under the top bar (YouTube's filter chips): "All sites" plus
	one chip per search plugin. The selection persists across pages.
-->
<script lang="ts">
	import { browser } from '$app/environment';
	import type { Snippet } from 'svelte';
	import { fetchSearchPlugins, type SearchPlugin } from '$lib/search';
	import { chromeHidden, isCompact } from '$lib/stores/viewport';

	interface Props {
		// Selected plugin names (empty = all), two-way bound to the parent.
		selectedPlugins?: string[];
		// Chips placed before "All sites" (Home's media chip on phones).
		leading?: Snippet;
	}

	let { selectedPlugins = $bindable([]), leading }: Props = $props();

	const SOURCES_STORAGE_KEY = 'veil:search-sources';

	if (browser && selectedPlugins.length === 0) {
		selectedPlugins = loadSelectedPlugins();
	}

	let searchPlugins = $state<SearchPlugin[]>([]);

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
			aria-pressed={selectedPlugins.length === 0}
			onclick={() => (selectedPlugins = [])}
			class="site-chip"
			class:site-chip-active={selectedPlugins.length === 0}
		>
			All sites
		</button>
		{#each searchPlugins as plugin (plugin.name)}
			<button
				type="button"
				aria-pressed={selectedPlugins.includes(plugin.name)}
				onclick={() => togglePlugin(plugin.name)}
				class="site-chip"
				class:site-chip-active={selectedPlugins.includes(plugin.name)}
			>
				{#if plugin.iconUrl}
					<img src={plugin.iconUrl} alt="" class="h-4 w-4 rounded-sm" />
				{/if}
				{pluginLabel(plugin)}
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
</style>
