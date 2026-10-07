<!--
	Phone full-screen search, opened from the top bar's search pill. Before
	typing it shows recent searches and taste picks; while typing, suggestions
	from the backend (internal/suggest). Choosing a query runs it on Home;
	choosing a performer, studio or tag opens its page.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import ArrowUpLeftIcon from 'phosphor-svelte/lib/ArrowUpLeftIcon';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import TagSimpleIcon from 'phosphor-svelte/lib/TagSimpleIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { searchQuery } from '$lib/stores/searchQuery';
	import { searchOverlayOpen } from '$lib/stores/searchOverlay';
	import { pushOverlay } from '$lib/stores/overlayStack';
	import {
		fetchSearchSuggestions,
		forgetSearch,
		recordSearch,
		suggestionUrl,
		type SearchSuggestion
	} from '$lib/searchSuggestions';

	// Wait this long after the last keystroke before asking the backend.
	const SUGGEST_DEBOUNCE_MS = 120;

	let input = $state<HTMLInputElement | null>(null);
	let text = $state($searchQuery);
	let suggestions = $state<SearchSuggestion[]>([]);
	// Answers can arrive out of order; only the newest request may land.
	let requestToken = 0;

	let recent = $derived(suggestions.filter((suggestion) => suggestion.kind === 'RECENT'));
	let picks = $derived(suggestions.filter((suggestion) => suggestion.kind !== 'RECENT'));
	let typing = $derived(text.trim().length > 0);

	$effect(() => {
		input?.focus();
	});

	$effect(() => pushOverlay(close));

	$effect(() => {
		const query = text.trim();
		const token = ++requestToken;
		const timer = setTimeout(async () => {
			try {
				const next = await fetchSearchSuggestions(query);
				if (token === requestToken) suggestions = next;
			} catch {
				if (token === requestToken) suggestions = [];
			}
		}, SUGGEST_DEBOUNCE_MS);
		return () => clearTimeout(timer);
	});

	/** Closes the overlay without searching. */
	function close() {
		searchOverlayOpen.set(false);
	}

	/** Runs a search on Home and remembers it. */
	function runSearch(query: string) {
		const trimmed = query.trim();
		close();
		searchQuery.set(trimmed);
		if (!trimmed) {
			goto('/');
			return;
		}
		recordSearch(trimmed);
		goto(`/?q=${encodeURIComponent(trimmed)}`);
	}

	/** Opens an entity's page, or searches for the suggestion's text. */
	function choose(suggestion: SearchSuggestion) {
		const url = suggestionUrl(suggestion);
		if (!url) {
			runSearch(suggestion.text);
			return;
		}
		recordSearch(suggestion.text);
		close();
		goto(url);
	}

	/** Puts the suggestion into the field to keep typing from it. */
	function fillIn(suggestion: SearchSuggestion) {
		text = `${suggestion.text} `;
		input?.focus();
	}

	/** Empties the field and keeps the keyboard up. */
	function clearText() {
		text = '';
		input?.focus();
	}

	/** Drops a recent search from the list and the backend. */
	async function forget(suggestion: SearchSuggestion) {
		suggestions = suggestions.filter((entry) => entry !== suggestion);
		await forgetSearch(suggestion.text).catch(() => {});
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') runSearch(text);
		if (event.key === 'Escape') close();
	}

	/**
	 * Splits the suggestion text around the typed prefix so the part the user
	 * hasn't typed yet can be emphasised, like most search boxes do.
	 */
	function splitOnTyped(value: string): { typed: string; rest: string } {
		const query = text.trim().toLowerCase();
		if (query && value.toLowerCase().startsWith(query)) {
			return { typed: value.slice(0, query.length), rest: value.slice(query.length) };
		}
		return { typed: '', rest: value };
	}
</script>

<button
	type="button"
	class="z-modal fixed inset-0 bg-black/40"
	aria-label="Close search"
	tabindex="-1"
	onclick={close}
	transition:fade={{ duration: 200 }}
></button>
<div
	class="search-overlay z-modal fixed inset-x-2 flex flex-col overflow-hidden rounded-[1.75rem]"
	role="dialog"
	aria-modal="true"
	aria-label="Search"
	data-swipe-ignore
	transition:scale={{ start: 0.96, duration: 240, easing: cubicOut }}
>
	<div class="flex h-16 shrink-0 items-center gap-2 px-2.5" in:fly={{ y: -12, duration: 260, easing: cubicOut }}>
		<button
			type="button"
			onclick={close}
			aria-label="Close search"
			class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
		>
			<ArrowLeftIcon size={22} />
		</button>
		<div class="bg-base-content/8 relative flex h-11 flex-1 items-center rounded-full">
			<input
				bind:this={input}
				bind:value={text}
				onkeydown={onKeydown}
				type="search"
				enterkeyhint="search"
				autocomplete="off"
				placeholder="Search videos, performers, tags"
				class="h-full w-full rounded-full bg-transparent pr-10 pl-4 text-base outline-none focus-visible:shadow-none"
			/>
			{#if typing}
				<button
					type="button"
					aria-label="Clear"
					onclick={clearText}
					class="text-base-content/50 absolute right-3 flex items-center"
				>
					<XIcon size={18} />
				</button>
			{/if}
		</div>
	</div>

	<div class="no-scrollbar flex-1 overflow-y-auto pb-2">
		{#if typing}
			<ul class="flex flex-col py-1">
				{#each suggestions as suggestion (suggestion.kind + suggestion.text)}
					{@const parts = splitOnTyped(suggestion.text)}
					<li class="suggestion-row flex items-center">
						<button
							type="button"
							onclick={() => choose(suggestion)}
							class="flex min-w-0 flex-1 items-center gap-4 py-2.5 pl-4 text-left"
						>
							{@render suggestionIcon(suggestion)}
							<span class="flex min-w-0 flex-col">
								<span class="truncate text-[15px]">
									<span class="text-base-content/60">{parts.typed}</span><span class="font-semibold">{parts.rest}</span>
								</span>
								{#if suggestion.detail}
									<span class="text-base-content/50 truncate text-xs">{suggestion.detail}</span>
								{/if}
							</span>
						</button>
						<button
							type="button"
							onclick={() => fillIn(suggestion)}
							aria-label="Search for {suggestion.text}"
							class="text-base-content/50 flex h-12 w-12 shrink-0 items-center justify-center"
						>
							<ArrowUpLeftIcon size={18} />
						</button>
					</li>
				{/each}
			</ul>
		{:else}
			{#if recent.length > 0}
				<h2 class="text-base-content/50 px-4 pt-3 pb-1 text-xs font-semibold tracking-wider uppercase">
					Recent
				</h2>
				<ul class="flex flex-col">
					{#each recent as suggestion (suggestion.text)}
						<li class="suggestion-row flex items-center" out:fly={{ x: -40, duration: 200 }}>
							<button
								type="button"
								onclick={() => choose(suggestion)}
								class="flex min-w-0 flex-1 items-center gap-4 py-2.5 pl-4 text-left text-[15px]"
							>
								<ClockCounterClockwiseIcon size={20} class="text-base-content/50 shrink-0" />
								<span class="truncate">{suggestion.text}</span>
							</button>
							<button
								type="button"
								onclick={() => forget(suggestion)}
								aria-label="Remove {suggestion.text} from recent searches"
								class="text-base-content/40 flex h-12 w-12 shrink-0 items-center justify-center"
							>
								<XIcon size={16} />
							</button>
						</li>
					{/each}
				</ul>
			{/if}
			{#if picks.length > 0}
				<h2 class="text-base-content/50 px-4 pt-5 pb-2 text-xs font-semibold tracking-wider uppercase">
					For you
				</h2>
				<div class="flex flex-wrap gap-2 px-4">
					{#each picks as suggestion, index (suggestion.kind + suggestion.text)}
						<button
							type="button"
							onclick={() => choose(suggestion)}
							class="pick-chip bg-base-200 ring-base-content/10 flex h-9 items-center gap-2 rounded-full pr-3.5 pl-1 text-sm ring-1 ring-inset"
							style:--stagger={index}
						>
							{@render suggestionIcon(suggestion)}
							{suggestion.text}
						</button>
					{/each}
				</div>
			{/if}
		{/if}
	</div>
</div>

{#snippet suggestionIcon(suggestion: SearchSuggestion)}
	{#if suggestion.imageUrl}
		<img src={suggestion.imageUrl} alt="" class="h-7 w-7 shrink-0 rounded-full object-cover" />
	{:else if suggestion.kind === 'TAG'}
		<span class="flex h-7 w-7 shrink-0 items-center justify-center"><TagSimpleIcon size={18} class="text-base-content/50" /></span>
	{:else if suggestion.kind === 'RECENT'}
		<span class="flex h-7 w-7 shrink-0 items-center justify-center"><ClockCounterClockwiseIcon size={18} class="text-base-content/50" /></span>
	{:else if suggestion.kind === 'PERFORMER' || suggestion.kind === 'STUDIO'}
		<span class="bg-base-300 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold uppercase">
			{suggestion.text.charAt(0)}
		</span>
	{:else}
		<span class="flex h-7 w-7 shrink-0 items-center justify-center"><MagnifyingGlassIcon size={18} class="text-base-content/50" /></span>
	{/if}
{/snippet}

<style>
	/* Floating glass like the tab bar and action sheet, clear of the screen edges. */
	.search-overlay {
		top: calc(0.5rem + var(--safe-area-inset-top, env(safe-area-inset-top, 0px)));
		bottom: calc(0.5rem + var(--safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
		background-color: color-mix(in oklab, var(--color-base-200) 72%, transparent);
		backdrop-filter: blur(28px) saturate(1.8);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow:
			0 20px 50px rgba(0, 0, 0, 0.55),
			inset 0 1px 0 color-mix(in oklab, var(--color-base-content) 8%, transparent);
	}

	.suggestion-row:active {
		background-color: color-mix(in oklab, var(--color-base-content) 6%, transparent);
	}

	/* Taste picks pop in one after another. */
	.pick-chip {
		animation: pick-in 320ms cubic-bezier(0.34, 1.4, 0.64, 1) both;
		animation-delay: calc(var(--stagger) * 30ms);
	}

	@keyframes pick-in {
		from {
			opacity: 0;
			scale: 0.85;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.pick-chip {
			animation: none;
		}
	}
</style>
