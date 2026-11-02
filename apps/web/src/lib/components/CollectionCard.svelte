<!--
	A collection as a stack: the covers of its first items fanned behind each
	other (its own cover first, when set), its name and how much it holds.
-->
<script lang="ts">
	import StackIcon from 'phosphor-svelte/lib/StackIcon';
	import { cacheUrl } from '$lib/img';
	import type { CollectionCardRow } from '$lib/curation';

	interface Props {
		collection: CollectionCardRow;
		onclick: () => void;
	}

	let { collection, onclick }: Props = $props();

	// Front cover first; the ones behind peek out above it.
	let stack = $derived.by(() => {
		const uris = [...collection.covers];
		if (collection.cover) uris.unshift(collection.cover.uri);
		return uris.slice(0, 3).map((uri) => cacheUrl(uri));
	});
</script>

<button type="button" {onclick} class="group flex w-full flex-col gap-2 text-left">
	<span class="relative block aspect-video w-full pt-3">
		{#each stack.toReversed() as src, depth (depth)}
			{@const behind = stack.length - 1 - depth}
			<span
				class="tv-card bg-base-300 absolute inset-x-0 bottom-0 block overflow-hidden ring-1 ring-white/5 transition-transform duration-200 group-hover:scale-[1.02]"
				style:top="{12 - behind * 6}px"
				style:margin-inline="{behind * 10}px"
				style:opacity={1 - behind * 0.25}
			>
				{#if src}
					<img {src} alt="" loading="lazy" class="h-full w-full object-cover" />
				{/if}
			</span>
		{/each}
		{#if stack.length === 0}
			<span class="tv-card bg-base-300 absolute inset-0 flex items-center justify-center">
				<StackIcon size={40} class="opacity-20" />
			</span>
		{/if}
	</span>
	<span class="flex min-w-0 flex-col">
		<span class="line-clamp-1 text-sm font-medium">{collection.name}</span>
		<span class="text-base-content/50 text-xs">{collection.itemCount} items</span>
	</span>
</button>
