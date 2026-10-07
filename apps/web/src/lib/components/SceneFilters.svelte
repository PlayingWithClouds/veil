<script lang="ts">
	import FunnelIcon from 'phosphor-svelte/lib/FunnelIcon';
	import FloppyDiskIcon from 'phosphor-svelte/lib/FloppyDiskIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import { notifications } from '$lib/stores/notifications';
	import {
		fetchFilterOptions,
		fetchSavedFilters,
		createSavedFilter,
		deleteSavedFilter,
		isFilterActive,
		type SceneFilter,
		type FilterOption,
		type SavedFilter
	} from '$lib/filters';

	interface Props {
		filter: SceneFilter;
		oncommit?: () => void;
	}

	let { filter = $bindable(), oncommit }: Props = $props();

	let tags = $state<FilterOption[]>([]);
	let studios = $state<FilterOption[]>([]);
	let performers = $state<FilterOption[]>([]);
	let savedFilters = $state<SavedFilter[]>([]);
	let open = $state(false);

	$effect(() => {
		fetchFilterOptions().then((options) => {
			tags = options.tags;
			studios = options.studios;
			performers = options.performers;
		});
	});

	$effect(() => {
		fetchSavedFilters().then((list) => (savedFilters = list));
	});

	let active = $derived(isFilterActive(filter));

	function commit() {
		oncommit?.();
	}

	// Duration presets map to a min/max seconds pair.
	const durationPresets = [
		{ label: 'Any length', min: null, max: null },
		{ label: 'Under 10 min', min: null, max: 600 },
		{ label: '10–30 min', min: 600, max: 1800 },
		{ label: '30–60 min', min: 1800, max: 3600 },
		{ label: 'Over 60 min', min: 3600, max: null }
	];
	let durationIndex = $derived(
		durationPresets.findIndex(
			(preset) =>
				(preset.min ?? null) === (filter.minDuration ?? null) &&
				(preset.max ?? null) === (filter.maxDuration ?? null)
		)
	);
	function setDuration(event: Event) {
		const preset = durationPresets[Number((event.target as HTMLSelectElement).value)];
		filter.minDuration = preset.min;
		filter.maxDuration = preset.max;
		commit();
	}

	function clearAll() {
		filter.studioId = null;
		filter.performerId = null;
		filter.tagId = null;
		filter.minRating = null;
		filter.minDuration = null;
		filter.maxDuration = null;
		filter.dateFrom = null;
		filter.dateTo = null;
		filter.sort = null;
		commit();
	}

	function applySaved(saved: SavedFilter) {
		filter.search = saved.filter.search ?? filter.search;
		filter.studioId = saved.filter.studioId ?? null;
		filter.performerId = saved.filter.performerId ?? null;
		filter.tagId = saved.filter.tagId ?? null;
		filter.minRating = saved.filter.minRating ?? null;
		filter.minDuration = saved.filter.minDuration ?? null;
		filter.maxDuration = saved.filter.maxDuration ?? null;
		filter.dateFrom = saved.filter.dateFrom ?? null;
		filter.dateTo = saved.filter.dateTo ?? null;
		filter.sort = saved.filter.sort ?? null;
		commit();
	}

	async function saveCurrent() {
		if (!active) {
			notifications.push('Set a filter first', 'error');
			return;
		}
		const name = window.prompt('Name this filter');
		if (!name) return;
		const created = await createSavedFilter(name, filter);
		savedFilters = [created, ...savedFilters];
		notifications.push(`Saved "${name}"`, 'success');
	}

	async function removeSaved(saved: SavedFilter) {
		await deleteSavedFilter(saved.id);
		savedFilters = savedFilters.filter((entry) => entry.id !== saved.id);
	}

	const selectClass =
		'bg-base-200 border-base-300 rounded-lg border px-3 py-1.5 text-sm focus:border-primary focus:outline-none';
</script>

<div class="border-base-300 bg-base-100 flex flex-col gap-3 rounded-xl border p-3">
	<div class="flex flex-wrap items-center gap-2">
		<button
			type="button"
			class="text-base-content/70 hover:text-base-content flex items-center gap-2 text-sm font-medium"
			onclick={() => (open = !open)}
		>
			<FunnelIcon size={16} weight={active ? 'fill' : 'regular'} />
			Filters
			{#if active}<span class="bg-primary/20 text-primary rounded-full px-2 py-0.5 text-xs">on</span>{/if}
		</button>

		<div class="flex-1"></div>

		{#if savedFilters.length > 0}
			<select
				class={selectClass}
				onchange={(event) => {
					const saved = savedFilters.find((entry) => entry.id === (event.target as HTMLSelectElement).value);
					if (saved) applySaved(saved);
					(event.target as HTMLSelectElement).value = '';
				}}
			>
				<option value="">Saved filters…</option>
				{#each savedFilters as saved (saved.id)}
					<option value={saved.id}>{saved.name}</option>
				{/each}
			</select>
		{/if}
		<button
			type="button"
			class="text-base-content/60 hover:text-base-content flex items-center gap-1.5 text-sm"
			onclick={saveCurrent}
			title="Save current filter"
		>
			<FloppyDiskIcon size={15} /> Save
		</button>
		{#if active}
			<button
				type="button"
				class="text-base-content/60 hover:text-error flex items-center gap-1.5 text-sm"
				onclick={clearAll}
			>
				<XIcon size={15} /> Clear
			</button>
		{/if}
	</div>

	{#if open}
		<div class="flex flex-wrap gap-2 pt-1">
			<select class={selectClass} bind:value={filter.sort} onchange={commit}>
				<option value={null}>Newest</option>
				<option value="date">Release date</option>
				<option value="rating">Top rated</option>
				<option value="views">Most viewed</option>
				<option value="title">Title A–Z</option>
			</select>

			<select class={selectClass} bind:value={filter.tagId} onchange={commit}>
				<option value={null}>Any tag</option>
				{#each tags as tag (tag.id)}
					<option value={tag.id}>{tag.name}</option>
				{/each}
			</select>

			<select class={selectClass} bind:value={filter.studioId} onchange={commit}>
				<option value={null}>Any studio</option>
				{#each studios as studio (studio.id)}
					<option value={studio.id}>{studio.name}</option>
				{/each}
			</select>

			<select class={selectClass} bind:value={filter.performerId} onchange={commit}>
				<option value={null}>Any performer</option>
				{#each performers as performer (performer.id)}
					<option value={performer.id}>{performer.name}</option>
				{/each}
			</select>

			<select class={selectClass} bind:value={filter.minRating} onchange={commit}>
				<option value={null}>Any rating</option>
				<option value={6}>6+ ★</option>
				<option value={7}>7+ ★</option>
				<option value={8}>8+ ★</option>
				<option value={9}>9+ ★</option>
			</select>

			<select class={selectClass} value={durationIndex < 0 ? 0 : durationIndex} onchange={setDuration}>
				{#each durationPresets as preset, index (preset.label)}
					<option value={index}>{preset.label}</option>
				{/each}
			</select>

			<label class="text-base-content/50 flex items-center gap-1.5 text-sm">
				From
				<input type="date" class={selectClass} bind:value={filter.dateFrom} onchange={commit} />
			</label>
			<label class="text-base-content/50 flex items-center gap-1.5 text-sm">
				To
				<input type="date" class={selectClass} bind:value={filter.dateTo} onchange={commit} />
			</label>
		</div>

		{#if savedFilters.length > 0}
			<div class="border-base-300 flex flex-wrap gap-2 border-t pt-2">
				{#each savedFilters as saved (saved.id)}
					<span class="bg-base-200 group flex items-center gap-1.5 rounded-full py-1 pl-3 pr-1.5 text-xs">
						<button type="button" class="hover:text-primary" onclick={() => applySaved(saved)}>{saved.name}</button>
						<button
							type="button"
							class="text-base-content/30 hover:text-error"
							onclick={() => removeSaved(saved)}
							aria-label="Delete saved filter {saved.name}"
						>
							<TrashIcon size={12} />
						</button>
					</span>
				{/each}
			</div>
		{/if}
	{/if}
</div>
