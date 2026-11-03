<!--
	A room of the user's as a round picture with its name under it: a red
	ring and the running time while it's being recorded, a breathing pink
	ring while it's live, greyed with when it was last seen while it's not.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import type { CamRow } from '$lib/cams';
	import { fadeInImage } from '$lib/fadeInImage';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';

	interface Props {
		cam: CamRow;
		// How long the room's running recording has been going; null when none runs.
		recordingSeconds?: number | null;
		onclick: () => void;
	}

	let { cam, recordingSeconds = null, onclick }: Props = $props();

	let picture = $derived(cacheUrl(cam.live?.thumbnail?.uri ?? cam.thumbnail?.uri));
	let recording = $derived(recordingSeconds !== null);
	let detail = $derived(offlineDetail(cam));

	/** When an offline room was last seen; null while it's live. */
	function offlineDetail(room: CamRow): string | null {
		if (room.live) return null;
		if (room.lastSeenOnlineAt) return `seen ${dayjs(room.lastSeenOnlineAt).fromNow()}`;
		return 'offline';
	}
</script>

<button
	type="button"
	class="group flex w-full flex-col items-center gap-2 text-center"
	{onclick}
	title={cam.displayName}
>
	<span
		class="bg-base-300 relative block aspect-square w-full overflow-hidden rounded-full transition-transform duration-200 group-hover:scale-105"
		class:rec-ring={recording}
		class:live-ring={!recording && cam.live}
		class:grayscale={!recording && !cam.live}
		class:opacity-60={!recording && !cam.live}
	>
		{#if picture}
			<img src={picture} alt="" loading="lazy" use:fadeInImage class="h-full w-full object-cover" />
		{:else}
			<span class="flex h-full w-full items-center justify-center">
				<VideoCameraIcon size={28} class="opacity-30" />
			</span>
		{/if}
	</span>
	<span class="flex w-full min-w-0 flex-col">
		<span class="truncate text-sm font-medium">{cam.displayName}</span>
		{#if recording}
			<span class="text-error text-xs font-semibold tabular-nums">
				REC {formatClock(recordingSeconds ?? 0) ?? '0:00'}
			</span>
		{:else if detail}
			<span class="text-base-content/45 truncate text-xs">{detail}</span>
		{:else}
			<span
				class="text-base-content/60 flex items-center justify-center gap-1.5 text-xs font-medium"
			>
				<span class="live-dot" style:width="6px" style:height="6px"></span> Live
			</span>
		{/if}
	</span>
</button>

<style>
	.rec-ring {
		box-shadow:
			0 0 0 2px var(--color-base-100),
			0 0 0 4px var(--color-error);
	}
</style>
