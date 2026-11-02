<!--
	One dimension of the user's taste (tags, performers, studios or sites): the
	liked entries, then the disliked ones, each with a bar for how strongly.
-->
<script lang="ts">
	import { cacheUrl } from '$lib/img';
	import type { TasteEntry } from '$lib/taste';

	interface Props {
		title: string;
		entries: TasteEntry[];
		// The entry's page; entries without one aren't links.
		href?: (entry: TasteEntry) => string | null;
	}

	let { title, entries, href = () => null }: Props = $props();

	let liked = $derived(entries.filter((entry) => entry.weight > 0));
	let disliked = $derived(entries.filter((entry) => entry.weight < 0));
</script>

{#snippet row(entry: TasteEntry)}
	{@const link = href(entry)}
	{@const picture = cacheUrl(entry.image?.uri)}
	<svelte:element
		this={link ? 'a' : 'div'}
		href={link}
		class="hover:bg-base-200 flex items-center gap-3 rounded-lg px-2 py-1.5 transition-colors"
	>
		<span
			class="bg-base-300 flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full"
		>
			{#if picture}
				<img src={picture} alt="" loading="lazy" class="h-full w-full object-cover" />
			{:else}
				<span class="text-base-content/60 text-xs font-semibold"
					>{entry.name.charAt(0).toUpperCase()}</span
				>
			{/if}
		</span>
		<span class="flex min-w-0 flex-1 flex-col gap-1">
			<span class="truncate text-sm">{entry.name}</span>
			<span class="bg-base-300 h-1 w-full overflow-hidden rounded-full">
				<span
					class="block h-full rounded-full"
					class:bg-success={entry.weight > 0}
					class:bg-error={entry.weight < 0}
					style:width="{Math.abs(entry.weight) * 100}%"
				></span>
			</span>
		</span>
	</svelte:element>
{/snippet}

<section class="bg-base-100 flex flex-col gap-3 rounded-xl">
	<h2 class="text-lg font-semibold">{title}</h2>
	{#if entries.length === 0}
		<p class="text-base-content/50 text-sm">Nothing yet.</p>
	{/if}
	{#if liked.length > 0}
		<div class="flex flex-col">
			{#each liked as entry (entry.id)}
				{@render row(entry)}
			{/each}
		</div>
	{/if}
	{#if disliked.length > 0}
		<h3 class="text-base-content/50 text-xs font-semibold tracking-wide uppercase">Not for you</h3>
		<div class="flex flex-col">
			{#each disliked as entry (entry.id)}
				{@render row(entry)}
			{/each}
		</div>
	{/if}
</section>
