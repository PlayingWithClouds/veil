<!--
	A site group (movie, series, compilation) in a grid or row: its box-art
	poster, title and studio.
-->
<script lang="ts">
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import { cacheUrl } from '$lib/img';
	import type { GroupCardRow } from '$lib/media';

	interface Props {
		group: GroupCardRow;
		onclick: () => void;
	}

	let { group, onclick }: Props = $props();
</script>

<div class="group flex w-full flex-col gap-2">
	<CoverFrame
		src={cacheUrl(group.cover?.uri)}
		alt={group.title}
		aspect="2 / 3"
		placeholder={FilmStripIcon}
		{onclick}
	/>
	<button type="button" {onclick} class="flex min-w-0 flex-col text-left">
		<span class="line-clamp-2 text-sm font-medium">{group.title}</span>
		{#if group.studio}
			<span class="text-base-content/50 truncate text-xs">{group.studio.name}</span>
		{/if}
	</button>
</div>
