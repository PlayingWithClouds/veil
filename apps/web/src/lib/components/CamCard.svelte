<!--
	A cam room in a row: a compact live picture with the Live badge and the
	viewer count ("58K"), and the room's name and site under it.
-->
<script lang="ts">
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import { cacheUrl } from '$lib/img';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import type { CamRow } from '$lib/cams';

	interface Props {
		cam: CamRow;
		onclick: () => void;
	}

	let { cam, onclick }: Props = $props();

	const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact' });
</script>

<div class="group flex w-full flex-col gap-2">
	<!-- The live frame is fresher than the room's stored picture. -->
	<CoverFrame
		src={cacheUrl(cam.live?.thumbnail?.uri ?? cam.thumbnail?.uri)}
		alt={cam.displayName}
		aspect="16 / 9"
		placeholder={VideoCameraIcon}
		{onclick}
	>
		{#if cam.live}
			<CardBadge corner="top-left" live>● Live</CardBadge>
			<CardBadge corner="bottom-right">
				<EyeIcon size={12} />
				{compactNumber.format(cam.live.viewers)}
			</CardBadge>
		{/if}
	</CoverFrame>
	<button type="button" {onclick} class="flex min-w-0 flex-col text-left">
		<span class="truncate text-sm font-medium">{cam.displayName}</span>
		<span class="text-base-content/50 truncate text-xs">{siteDisplayName(cam.plugin, $sites)}</span>
	</button>
</div>
