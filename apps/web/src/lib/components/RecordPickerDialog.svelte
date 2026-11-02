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

	/** Focuses the search field once it's in the page. */
	function focus(input: HTMLInputElement) {
		input.focus();
	}
</script>

{#snippet card(record: Pickable, index: number, reason: string | null)}
	{@const added = taken.includes(record.id)}
	<button
		type="button"
		class="hover:bg-base-200 flex flex-col gap-1.5 rounded-xl p-1.5 text-left disabled:opacity-50"
		class:bg-base-200={index === highlighted}
		disabled={added}
		onmouseenter={() => (highlighted = index)}
		onclick={() => pick(record)}
	>
		<span
			class={[
				'bg-base-300 relative flex w-full items-center justify-center overflow-hidden rounded-lg',
				{ 'aspect-[3/4]': kind === 'performer', 'aspect-square': kind === 'studio' }
			]}
		>
			{#if record.image}
				<img
					src={cacheUrl(record.image.uri)}
					alt=""
					loading="lazy"
					class="h-full w-full"
					class:object-cover={kind === 'performer'}
					class:object-contain={kind === 'studio'}
					class:p-3={kind === 'studio'}
				/>
			{:else if kind === 'performer'}
				<UserIcon size={28} class="opacity-20" />
			{:else}
				<BuildingsIcon size={28} class="opacity-20" />
			{/if}
			{#if added}
				<span
					class="bg-base-100/80 absolute inset-x-0 bottom-0 flex items-center justify-center gap-1 py-1 text-xs"
				>
					<CheckIcon size={12} /> Added
				</span>
			{/if}
		</span>
		<span class="line-clamp-1 text-sm font-medium">{record.name}</span>
		{#if reason}
			<span class="text-primary -mt-1 line-clamp-1 text-xs">{reason}</span>
		{:else if record.videoCount !== undefined}
			<span class="text-base-content/50 -mt-1 text-xs">{record.videoCount} videos</span>
		{/if}
	</button>
{/snippet}

<dialog use:showModal class="modal" {onclose} onclick={closeOnBackdrop} aria-label={title}>
	<div class="modal-box flex max-h-[85vh] !max-w-3xl flex-col gap-4">
		<h2 class="text-lg font-semibold">{title}</h2>
		<label class="input flex w-full items-center gap-2">
			<MagnifyingGlassIcon size={16} class="opacity-50" />
			<input use:focus class="grow" {placeholder} bind:value={text} onkeydown={onKeydown} />
			{#if searching}<span class="loading loading-spinner loading-xs"></span>{/if}
		</label>

		<div class="-mx-1 min-h-48 overflow-y-auto px-1">
			{#if shownSuggestions.length > 0}
				<h3 class="text-base-content/60 mb-2 text-xs font-semibold tracking-wide uppercase">
					Suggested
				</h3>
				<div class="mb-4 grid grid-cols-[repeat(auto-fill,minmax(112px,1fr))] gap-2">
					{#each shownSuggestions as suggestion, index (suggestion.record.id)}
						{@render card(suggestion.record, index, suggestion.reason)}
					{/each}
				</div>
				{#if shownResults.length > 0}
					<h3 class="text-base-content/60 mb-2 text-xs font-semibold tracking-wide uppercase">
						In the library
					</h3>
				{/if}
			{/if}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(112px,1fr))] gap-2">
				{#each shownResults as result, index (result.id)}
					{@render card(result, shownSuggestions.length + index, null)}
				{/each}
				{#if creatable}
					<button
						type="button"
						class="hover:bg-base-200 flex flex-col gap-1.5 rounded-xl p-1.5 text-left"
						class:bg-base-200={highlighted === pickable.length}
						onmouseenter={() => (highlighted = pickable.length)}
						onclick={create}
					>
						<span
							class={[
								'border-base-content/30 text-base-content/60 flex w-full items-center justify-center rounded-lg border-2 border-dashed',
								{ 'aspect-[3/4]': kind === 'performer', 'aspect-square': kind === 'studio' }
							]}
						>
							<PlusIcon size={28} />
						</span>
						<span class="line-clamp-1 text-sm font-medium">{typed}</span>
						<span class="text-primary -mt-1 text-xs">New {kind}</span>
					</button>
				{/if}
			</div>
			{#if !searching && !creatable && pickable.length === 0}
				<p class="text-base-content/50 py-8 text-center text-sm">
					{#if typed === ''}Nothing in the library yet.{:else}No {kind} called “{typed}”.{/if}
				</p>
			{/if}
		</div>

		<div class="modal-action mt-0">
			{#if clearLabel && onclear}
				<button type="button" class="btn btn-ghost text-error mr-auto" onclick={onclear}
					>{clearLabel}</button
				>
			{/if}
			<button type="button" class="btn btn-ghost" onclick={onclose}>Cancel</button>
		</div>
	</div>
</dialog>
