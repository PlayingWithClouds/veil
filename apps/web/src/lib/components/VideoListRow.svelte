<!--
	A video as one row of a side list: its picture with the length on it, the
	title over two lines, and studio · age. Logs being shown and picked.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import { logChosen, type Shown } from '$lib/events';
	import { fadeInImage } from '$lib/fadeInImage';
	import { cacheUrl } from '$lib/img';
	import { impression } from '$lib/impression';
	import type { VideoCardRow } from '$lib/library';
	import { formatClock } from '$lib/time';

	interface Props {
		video: VideoCardRow;
		shown?: Shown | null;
		onclick: () => void;
	}

	let { video, shown = null, onclick }: Props = $props();

	let subject = $derived({ kind: 'video', id: video.id } as const);
	let cover = $derived(cacheUrl(video.cover?.uri));
	let length = $derived(formatClock(video.durationSeconds));
	let meta = $derived(metaLine(video));

	/** "Studio · 3 days ago", leaving out what's unknown. */
	function metaLine(row: VideoCardRow): string {
		const parts: string[] = [];
		if (row.studio) parts.push(row.studio.name);
		if (row.date) parts.push(dayjs(row.date).fromNow());
		return parts.join(' · ');
	}

	/** Logs the pick where the row is shown, then opens the video. */
	function open() {
		if (shown) logChosen(subject, shown);
		onclick();
	}
</script>

<button
	type="button"
	class="group flex w-full items-start gap-3 rounded-xl p-1.5 text-left transition-colors hover:bg-white/5"
	use:impression={{ subject, shown }}
	onclick={open}
>
	<span
		class="bg-base-300 relative aspect-video w-40 shrink-0 overflow-hidden rounded-lg ring-1 ring-white/5"
	>
		{#if cover}
			<img
				src={cover}
				alt=""
				loading="lazy"
				use:fadeInImage
				class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
			/>
		{:else}
			<span class="flex h-full w-full items-center justify-center">
				<VideoCameraIcon size={24} class="opacity-25" />
			</span>
		{/if}
		{#if length}
			<span
				class="absolute right-1 bottom-1 rounded bg-black/75 px-1 py-px text-[11px] font-medium text-white tabular-nums"
			>
				{length}
			</span>
		{/if}
	</span>
	<span class="flex min-w-0 flex-col gap-0.5 pt-0.5">
		<span class="line-clamp-2 text-sm leading-snug font-medium">{video.title}</span>
		{#if meta}
			<span class="text-base-content/50 truncate text-xs">{meta}</span>
		{/if}
	</span>
</button>
