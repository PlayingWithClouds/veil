<!--
	Leading icon of a search suggestion: the entity's picture when it has one,
	else an icon for its kind (tag, recent search, performer/studio initial, or
	a magnifier for plain query completions).
-->
<script lang="ts">
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import TagSimpleIcon from 'phosphor-svelte/lib/TagSimpleIcon';
	import type { SearchSuggestion } from '$lib/searchSuggestions';

	let { suggestion }: { suggestion: SearchSuggestion } = $props();
</script>

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
