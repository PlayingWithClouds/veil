<!--
	A cam recording running now: the room's live picture with a REC badge and
	how long it has been recording.
-->
<script lang="ts">
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import type { ActiveRecordingRow } from '$lib/cams';

	interface Props {
		recording: ActiveRecordingRow;
		onclick: () => void;
	}

	let { recording, onclick }: Props = $props();

	let room = $derived(recording.camModel);
	let name = $derived(room?.displayName ?? 'Recording');
</script>

<div class="group flex w-full flex-col gap-2">
	<CoverFrame
		src={cacheUrl(room?.live?.thumbnail?.uri ?? room?.thumbnail?.uri)}
		alt={name}
		aspect="16 / 9"
		placeholder={VideoCameraIcon}
		{onclick}
	>
		<CardBadge corner="top-left" live>● REC</CardBadge>
		<CardBadge corner="bottom-right">{formatClock(recording.durationSeconds) ?? '0:00'}</CardBadge>
	</CoverFrame>
	<button type="button" {onclick} class="flex min-w-0 flex-col text-left">
		<span class="truncate text-sm font-medium">{name}</span>
		{#if recording.quality}
			<span class="text-base-content/50 text-xs">{recording.quality}</span>
		{/if}
	</button>
</div>
