<!--
	The desktop top bar's search field. What is typed stays in the field and only
	drives suggestions (recent searches, tags, performers, studios from
	internal/suggest); the search itself runs on Enter, on choosing a suggestion
	or on the search button, so typing never fires a search.
-->
<script lang="ts">
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { searchQuery } from '$lib/stores/searchQuery';
	import {
		forgetSearch,
		openSuggestion,
		submitSearch,
		type SearchSuggestion
	} from '$lib/searchSuggestions';
	import { createSuggestionFeed } from '$lib/suggestionFeed.svelte';
	import SearchSuggestionIcon from './SearchSuggestionIcon.svelte';

	/** Element id the Ctrl+K shortcut focuses. */
	const SEARCH_INPUT_ID = 'global-search';

	let input = $state<HTMLInputElement | null>(null);
	// What is typed. It follows the shared searchQuery when a search is set from
	// elsewhere (Home restoring ?q=), but typing never writes back to it.
	let draft = $derived($searchQuery);
	let focused = $state(false);
	// Suggestion highlighted with the arrow keys; -1 is the typed text itself.
	let activeIndex = $state(-1);

	const feed = createSuggestionFeed({ query: () => draft, active: () => focused });
	let suggestions = $derived(feed.items);
	let open = $derived(focused && suggestions.length > 0);

	/** Runs the typed text as a search and closes the suggestions. */
	function submit() {
		submitSearch(draft);
		input?.blur();
	}

	/** Opens the chosen suggestion and closes the list. */
	function choose(suggestion: SearchSuggestion) {
		openSuggestion(suggestion);
		input?.blur();
	}

	/** Empties the field and the running search, and keeps the cursor in the field. */
	function clear() {
		draft = '';
		searchQuery.set('');
		input?.focus();
	}

	/** Drops a recent search from the list and the backend. */
	async function forget(suggestion: SearchSuggestion) {
		feed.remove(suggestion);
		await forgetSearch(suggestion.text).catch(() => {});
	}

	/** Moves the highlight with the arrow keys; Enter takes it, Escape closes the list. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			const step = event.key === 'ArrowDown' ? 1 : -1;
			activeIndex = Math.max(-1, Math.min(suggestions.length - 1, activeIndex + step));
			return;
		}
		if (event.key === 'Escape') {
			input?.blur();
			return;
		}
		if (event.key !== 'Enter') return;
		if (activeIndex >= 0) {
			choose(suggestions[activeIndex]);
			return;
		}
		submit();
	}
</script>

<div class="flex flex-1 items-center">
	<div
		class="border-base-300 bg-base-200 focus-within:border-base-content/30 relative flex h-10 flex-1 items-center rounded-l-full border"
	>
		<input
			id={SEARCH_INPUT_ID}
			bind:this={input}
			bind:value={draft}
			oninput={() => (activeIndex = -1)}
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			onkeydown={onKeydown}
			type="search"
			autocomplete="off"
			role="combobox"
			aria-expanded={open}
			aria-controls="search-suggestions"
			placeholder="Search"
			class="h-full w-full bg-transparent pr-10 pl-5 text-[15px] outline-none"
		/>
		{#if draft.length > 0}
			<button
				type="button"
				aria-label="Clear search"
				onclick={clear}
				class="text-base-content/50 hover:text-base-content absolute right-3 flex items-center"
			>
				<XIcon size={18} />
			</button>
		{/if}

		{#if open}
			<!-- Pressing a row must not blur the field before the click lands. -->
			<ul
				id="search-suggestions"
				role="listbox"
				onmousedown={(event) => event.preventDefault()}
				class="border-base-300 bg-base-200 z-overlay absolute top-full right-0 left-0 mt-2 flex max-h-[28rem] flex-col overflow-y-auto rounded-2xl border py-1 shadow-xl"
			>
				{#each suggestions as suggestion, index (suggestion.kind + suggestion.text)}
					<li
						role="option"
						aria-selected={index === activeIndex}
						class={['hover:bg-base-content/6 flex items-center', { 'bg-base-content/6': index === activeIndex }]}
					>
						<button
							type="button"
							onclick={() => choose(suggestion)}
							class="flex min-w-0 flex-1 items-center gap-3 py-2 pl-4 text-left"
						>
							<SearchSuggestionIcon {suggestion} />
							<span class="flex min-w-0 flex-col">
								<span class="truncate text-[15px]">{suggestion.text}</span>
								{#if suggestion.detail}
									<span class="text-base-content/50 truncate text-xs">{suggestion.detail}</span>
								{/if}
							</span>
						</button>
						{#if suggestion.kind === 'RECENT'}
							<button
								type="button"
								onclick={() => forget(suggestion)}
								aria-label="Remove {suggestion.text} from recent searches"
								class="text-base-content/40 hover:text-base-content flex h-10 w-10 shrink-0 items-center justify-center"
							>
								<XIcon size={16} />
							</button>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>
	<button
		type="button"
		onclick={submit}
		aria-label="Search"
		class="border-base-300 bg-base-300 hover:bg-base-content/15 flex h-10 w-16 items-center justify-center rounded-r-full border border-l-0 transition-colors"
	>
		<MagnifyingGlassIcon size={20} />
	</button>
</div>
