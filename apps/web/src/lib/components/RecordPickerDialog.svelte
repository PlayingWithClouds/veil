<!--
	A dialog to pick a performer or studio: a search field over a grid of
	portraits (performers) or logos (studios). Before anything is typed it
	shows the suggestions given, then the library's first matches; typing
	narrows them. Arrow keys move, Enter or a click picks. Records listed in
	taken show as already added and can't be picked again. With oncreate, a
	typed name no result has exactly ends the grid as a card that creates it.
	An optional clear action (e.g. "Remove the studio") sits beside Cancel.
-->
<script lang="ts">
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import BuildingsIcon from 'phosphor-svelte/lib/BuildingsIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import { cacheUrl } from '$lib/img';
	import type { MemberRow } from '$lib/library';

	/** A record the dialog can show: what a card needs. */
	interface Pickable {
		id: string;
		name: string;
		image: { uri: string } | null;
		videoCount?: number;
	}

	/** A suggestion shown first, with why. */
	interface Suggested {
		record: Pickable;
		reason: string;
	}

	interface Props {
		kind: 'performer' | 'studio';
		title: string;
		placeholder: string;
		search: (text: string) => Promise<MemberRow[]>;
		suggested?: Suggested[];
		// Ids already picked, shown as added.
		taken?: string[];
		clearLabel?: string;
		onclear?: () => void;
		onpick: (record: Pickable) => void;
		// Creates a record with the typed name.
		oncreate?: (name: string) => void;
		onclose: () => void;
	}

	let {
		kind,
		title,
		placeholder,
		search,
		suggested = [],
		taken = [],
		clearLabel,
		onclear,
		onpick,
		oncreate,
		onclose
	}: Props = $props();

	// How long typing rests before the search runs.
	const SEARCH_DELAY_MS = 200;
	const RESULTS = 24;

	let text = $state('');
	let results = $state<MemberRow[]>([]);
	let searching = $state(false);
	let highlighted = $state(0);
	let timer: ReturnType<typeof setTimeout> | null = null;

	let typed = $derived(text.trim());
	// Suggestions only while nothing is typed; then the search alone.
	let shownSuggestions = $derived(typed === '' ? suggested : []);
	let shownResults = $derived(
		results.filter(
			(result) => !shownSuggestions.some((suggestion) => suggestion.record.id === result.id)
		)
	);
	let pickable = $derived([
		...shownSuggestions.map((suggestion) => suggestion.record),
		...shownResults
	]);
	// Offered once the search for what's typed is back without that exact name.
	let creatable = $derived(
		oncreate !== undefined &&
			typed !== '' &&
			!searching &&
			!pickable.some((record) => sameName(record.name, typed))
	);
	// The create card follows the records: highlighted at index pickable.length.
	let choices = $derived(creatable ? pickable.length + 1 : pickable.length);

	$effect(() => {
		const query = typed;
		if (timer) clearTimeout(timer);
		searching = true;
		timer = setTimeout(() => void run(query), SEARCH_DELAY_MS);
	});

	/** Searches, keeping the answer only while it's still what's typed. */
	async function run(query: string) {
		const found = await search(query);
		if (query !== typed) return;
		results = found.slice(0, RESULTS);
		highlighted = 0;
		searching = false;
	}

	/** Opens the dialog as a modal once it is in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Picks a record unless it's added already. */
	function pick(record: Pickable) {
		if (taken.includes(record.id)) return;
		onpick(record);
	}

	/** Whether two names are the same apart from case and outer spaces. */
	function sameName(name: string, other: string) {
		return name.trim().toLowerCase() === other.trim().toLowerCase();
	}

	/** Creates a record with the typed name. */
	function create() {
		if (!creatable || oncreate === undefined) return;
		oncreate(typed);
	}

	/** Arrow keys move through the grid, Enter picks or creates. */
	function onKeydown(event: KeyboardEvent) {
		if (choices === 0) return;
		const steps: Record<string, number> = {
			ArrowRight: 1,
			ArrowDown: 1,
			ArrowLeft: -1,
			ArrowUp: -1
		};
		const step = steps[event.key];
		if (step !== undefined) {
			event.preventDefault();
			highlighted = (highlighted + step + choices) % choices;
			return;
		}
		if (event.key !== 'Enter') return;
		event.preventDefault();
		if (highlighted === pickable.length) {
			create();
			return;
		}
		pick(pickable[highlighted]);
	}

	/** Closes on a click beside the box (the dialog itself fills the screen). */
	function closeOnBackdrop(event: MouseEvent) {
		if (event.target === event.currentTarget) onclose();
	}

	/** A steady colour per name, for the initial of a record without a picture. */
	function hueOf(name: string): number {
		let hash = 0;
		for (const character of name) hash = (hash * 31 + character.charCodeAt(0)) % 360;
		return hash;
	}

	/** Focuses the search field once it's in the page. */
	function focus(input: HTMLInputElement) {
		input.focus();
	}
</script>

{#snippet card(record: Pickable, index: number, reason: string | null)}
	{@const added = taken.includes(record.id)}
	<button
		type="button"
		class="picker-card flex flex-col items-center gap-2 rounded-2xl px-2 pt-3 pb-2.5 text-center disabled:opacity-45"
		class:picker-card-on={index === highlighted}
		disabled={added}
		onmouseenter={() => (highlighted = index)}
		onclick={() => pick(record)}
	>
		<span
			class="relative flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden text-xl font-semibold text-black/75"
			class:rounded-full={kind === 'performer'}
			class:rounded-xl={kind === 'studio'}
			class:bg-white={kind === 'studio' && record.image !== null}
			style:background-color={record.image ? null : `hsl(${hueOf(record.name)} 50% 68%)`}
		>
			{#if record.image}
				<img
					src={cacheUrl(record.image.uri)}
					alt=""
					loading="lazy"
					class="h-full w-full"
					class:object-cover={kind === 'performer'}
					class:object-contain={kind === 'studio'}
					class:p-2={kind === 'studio'}
				/>
			{:else}
				{record.name.trim().charAt(0).toUpperCase()}
			{/if}
		</span>
		{#if added}
			<span
				class="bg-success absolute top-2 right-2 flex h-5 w-5 items-center justify-center rounded-full text-black"
				title="Added already"
			>
				<CheckIcon size={12} weight="bold" />
			</span>
		{/if}
		<span class="flex w-full min-w-0 flex-col">
			<span class="truncate text-sm font-medium">{record.name}</span>
			{#if reason}
				<span class="text-info truncate text-xs">{reason}</span>
			{:else if record.videoCount !== undefined}
				<span class="text-base-content/45 text-xs tabular-nums">
					{record.videoCount}
					{record.videoCount === 1 ? 'video' : 'videos'}
				</span>
			{/if}
		</span>
	</button>
{/snippet}

<dialog
	use:showModal
	class="modal picker-backdrop"
	{onclose}
	onclick={closeOnBackdrop}
	aria-label={title}
>
	<div class="menu-panel flex max-h-[80vh] w-[min(42rem,calc(100vw-2rem))] flex-col gap-4 p-5">
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-lg font-semibold">{title}</h2>
			{#if kind === 'performer'}
				<UserIcon size={18} class="opacity-40" />
			{:else}
				<BuildingsIcon size={18} class="opacity-40" />
			{/if}
		</div>
		<label class="picker-search flex h-12 w-full items-center gap-3 rounded-2xl px-4">
			<MagnifyingGlassIcon size={18} class="shrink-0 opacity-50" />
			<input
				use:focus
				class="h-full min-w-0 grow bg-transparent text-[15px] outline-none"
				{placeholder}
				bind:value={text}
				onkeydown={onKeydown}
			/>
			{#if searching}<span class="loading loading-spinner loading-xs"></span>{/if}
		</label>

		<div class="-mx-1 min-h-48 overflow-y-auto px-1">
			{#if shownSuggestions.length > 0}
				<h3 class="text-base-content/45 mb-2 px-1 text-xs font-semibold tracking-wider uppercase">
					Suggested
				</h3>
				<div class="mb-4 grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-1.5">
					{#each shownSuggestions as suggestion, index (suggestion.record.id)}
						{@render card(suggestion.record, index, suggestion.reason)}
					{/each}
				</div>
				{#if shownResults.length > 0}
					<h3 class="text-base-content/45 mb-2 px-1 text-xs font-semibold tracking-wider uppercase">
						In the library
					</h3>
				{/if}
			{/if}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-1.5">
				{#each shownResults as result, index (result.id)}
					{@render card(result, shownSuggestions.length + index, null)}
				{/each}
				{#if creatable}
					<button
						type="button"
						class="picker-card flex flex-col items-center gap-2 rounded-2xl px-2 pt-3 pb-2.5 text-center"
						class:picker-card-on={highlighted === pickable.length}
						onmouseenter={() => (highlighted = pickable.length)}
						onclick={create}
					>
						<span
							class="text-base-content/60 flex h-16 w-16 items-center justify-center border-2 border-dashed border-white/25"
							class:rounded-full={kind === 'performer'}
							class:rounded-xl={kind === 'studio'}
						>
							<PlusIcon size={24} />
						</span>
						<span class="flex w-full min-w-0 flex-col">
							<span class="truncate text-sm font-medium">{typed}</span>
							<span class="text-info text-xs">New {kind}</span>
						</span>
					</button>
				{/if}
			</div>
			{#if !searching && !creatable && pickable.length === 0}
				<p class="text-base-content/50 py-10 text-center text-sm">
					{#if typed === ''}Nothing in the library yet.{:else}No {kind} called “{typed}”.{/if}
				</p>
			{/if}
		</div>

		<div class="flex items-center justify-end gap-2 border-t border-white/5 pt-4">
			{#if clearLabel && onclear}
				<button type="button" class="chip chip-excluded mr-auto" onclick={onclear}>
					{clearLabel}
				</button>
			{/if}
			<span class="text-base-content/35 mr-auto text-xs max-md:hidden">
				↑↓ to move · Enter to pick
			</span>
			<button type="button" class="chip" onclick={onclose}>Cancel</button>
		</div>
	</div>
</dialog>

<style>
	.picker-backdrop::backdrop {
		background-color: rgba(0, 0, 0, 0.55);
		backdrop-filter: blur(6px);
	}

	.picker-search {
		border: 1px solid color-mix(in oklab, var(--color-base-content) 12%, transparent);
		background-color: color-mix(in oklab, var(--color-base-content) 6%, transparent);
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.picker-search:focus-within {
		border-color: color-mix(in oklab, var(--color-base-content) 28%, transparent);
		box-shadow: 0 0 0 4px color-mix(in oklab, var(--color-base-content) 6%, transparent);
	}

	.picker-card {
		position: relative;
		transition:
			background-color 120ms,
			box-shadow 120ms;
	}
	.picker-card-on {
		background-color: color-mix(in oklab, var(--color-base-content) 8%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-base-content) 14%, transparent);
	}
</style>
