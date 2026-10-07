<script lang="ts">
	import { browser } from '$app/environment';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { fetchSearchPlugins, type SearchPlugin } from '$lib/search';

	interface Props {
		// The query text, two-way bound to the parent.
		value?: string;
		// Selected source plugin names (empty = all), two-way bound to the parent.
		selectedPlugins?: string[];
		placeholder?: string;
		// Fired on Enter — used by pages that navigate to Explore rather than
		// searching in place.
		onsubmit?: () => void;
	}

	let {
		value = $bindable(''),
		selectedPlugins = $bindable([]),
		placeholder = 'Search performers, studios, scenes, galleries…',
		onsubmit
	}: Props = $props();

	const SOURCES_STORAGE_KEY = 'veil:search-sources';

	// Seed the source filter from storage so it stays consistent across pages.
	if (browser && selectedPlugins.length === 0) {
		selectedPlugins = loadSelectedPlugins();
	}

	let searchPlugins = $state<SearchPlugin[]>([]);

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

	function pluginLabel(plugin: SearchPlugin): string {
		if (plugin.displayName) return plugin.displayName;
		return plugin.name;
	}

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
				.sort((a, b) => pluginLabel(a).localeCompare(pluginLabel(b)));
		});
	});

	$effect(() => {
		if (!browser) return;
		localStorage.setItem(SOURCES_STORAGE_KEY, JSON.stringify(selectedPlugins));
	});

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') onsubmit?.();
	}
</script>

<style>
	.source-chip-active {
		border-color: color-mix(in oklab, var(--color-primary) 60%, transparent);
		background-color: color-mix(in oklab, var(--color-primary) 18%, transparent);
		color: var(--color-base-content);
	}
</style>

<div class="sticky top-0 z-20 -mx-2 -mt-2 px-2 pt-0 pb-1">
	<div
		class="border-base-content/10 bg-base-200/90 flex flex-col gap-3 rounded-xl border p-3 shadow-lg shadow-black/40 backdrop-blur-md"
	>
		<div class="relative">
			<span class="text-base-content/40 pointer-events-none absolute inset-y-0 left-3 flex items-center">
				<MagnifyingGlassIcon size={18} />
			</span>
			<input
				type="search"
				bind:value
				onkeydown={onKeydown}
				{placeholder}
				class="h-11 w-full bg-transparent pr-10 pl-10 text-sm outline-none"
			/>
			{#if value.trim().length > 0}
				<button
					type="button"
					aria-label="Clear search"
					onclick={() => (value = '')}
					class="text-base-content/40 hover:text-base-content absolute inset-y-0 right-3 flex items-center transition-colors"
				>
					<XIcon size={18} />
				</button>
			{/if}
		</div>

		{#if searchPlugins.length > 0}
			<div class="flex flex-wrap items-center gap-2">
				<!-- No selection means every site; this chip makes that visible. -->
				<button
					type="button"
					aria-pressed={selectedPlugins.length === 0}
					onclick={() => (selectedPlugins = [])}
					class="border-base-300 text-base-content/50 hover:text-base-content flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors"
					class:source-chip-active={selectedPlugins.length === 0}
				>
					All sites
				</button>
				{#each searchPlugins as plugin (plugin.name)}
					<button
						type="button"
						aria-pressed={selectedPlugins.includes(plugin.name)}
						onclick={() => togglePlugin(plugin.name)}
						class="border-base-300 text-base-content/50 hover:text-base-content flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors"
						class:source-chip-active={selectedPlugins.includes(plugin.name)}
					>
						{#if plugin.iconUrl}
							<img src={plugin.iconUrl} alt="" class="h-3.5 w-3.5 rounded-sm" />
						{/if}
						{pluginLabel(plugin)}
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>
