<!--
	A category as a wide tile: its cover darkened, with the name and what the
	library holds of it written over it.
-->
<script lang="ts">
	import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import { cacheUrl } from '$lib/img';
	import type { CategoryCardRow } from '$lib/curation';

	interface Props {
		category: CategoryCardRow;
		onclick: () => void;
	}

	let { category, onclick }: Props = $props();

	let count = $derived(category.videoCount + category.galleryCount);
</script>

<div class="group w-full">
	<CoverFrame
		src={cacheUrl(category.cover?.uri)}
		alt={category.name}
		aspect="16 / 9"
		placeholder={SquaresFourIcon}
		{onclick}
	>
		<span
			class="absolute inset-0 flex flex-col justify-end bg-gradient-to-t from-black/80 via-black/20 to-transparent p-3 text-left"
		>
			<span class="text-base font-semibold text-white">{category.name}</span>
			{#if count > 0}
				<span class="text-xs text-white/70">{count} items</span>
			{/if}
		</span>
	</CoverFrame>
</div>
