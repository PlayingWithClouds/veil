<!--
	A category as a wide tile: its picture (its cover, else the newest cover
	among its media) darkened, with the name and how much of one kind of
	media it holds written over it. Empty categories show faded.
-->
<script lang="ts">
	import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import { cacheUrl } from '$lib/img';
	import type { CategoryTileRow } from '$lib/categories';

	interface Props {
		tile: CategoryTileRow;
		// What's counted, by name: "videos", "galleries".
		noun: string;
		onclick: () => void;
	}

	let { tile, noun, onclick }: Props = $props();
</script>

<div class="group w-full" class:opacity-50={tile.count === 0}>
	<CoverFrame
		src={cacheUrl(tile.preview?.uri)}
		alt={tile.category.name}
		aspect="16 / 9"
		placeholder={SquaresFourIcon}
		{onclick}
	>
		<span
			class="absolute inset-0 flex flex-col justify-end bg-linear-to-t from-black/80 via-black/20 to-transparent p-3 text-left"
		>
			<span class="text-base font-semibold text-white">{tile.category.name}</span>
			<span class="text-xs text-white/70">{tile.count} {noun}</span>
		</span>
	</CoverFrame>
</div>
