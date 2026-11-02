<!--
	What happens when in a video, as a bar under the player: each segment
	(a site's or the user's marker, named by its tag or title) placed along
	the timeline, one colour per act. Clicking one seeks there. Markers
	without an end run to the next one.
-->
<script lang="ts">
	import type { VideoDetail } from '$lib/library';
	import { formatClock } from '$lib/time';
	import { tagLabel } from '$lib/components/TagChips.svelte';

	type SegmentRow = VideoDetail['segments'][number];

	/** A segment placed on the bar, as shares of the duration. */
	interface Placed {
		id: string;
		label: string;
		start: number;
		left: number;
		width: number;
		hue: number;
	}

	interface Props {
		segments: SegmentRow[];
		durationSeconds: number;
		onseek: (seconds: number) => void;
	}

	let { segments, durationSeconds, onseek }: Props = $props();

	let placed = $derived(place(segments, durationSeconds));
	let acts = $derived([...new Map(placed.map((segment) => [segment.label, segment.hue])).entries()]);

	/** Lays the named segments out along the duration, in order. */
	function place(all: SegmentRow[], duration: number): Placed[] {
		const named = all
			.filter((segment) => (segment.tag?.name ?? segment.title) && segment.startSeconds < duration)
			.sort((first, second) => first.startSeconds - second.startSeconds);
		return named.map((segment, index) => {
			const label = tagLabel(segment.tag?.name ?? segment.title ?? '');
			let end = segment.endSeconds ?? named[index + 1]?.startSeconds ?? duration;
			end = Math.min(Math.max(end, segment.startSeconds + 1), duration);
			return {
				id: segment.id,
				label,
				start: segment.startSeconds,
				left: segment.startSeconds / duration,
				width: (end - segment.startSeconds) / duration,
				hue: hueOf(label)
			};
		});
	}

	/** A steady colour per act name. */
	function hueOf(label: string): number {
		let hash = 0;
		for (const character of label) hash = (hash * 31 + character.charCodeAt(0)) % 360;
		return hash;
	}
</script>

{#if placed.length > 0}
	<div class="flex flex-col gap-1.5">
		<div class="bg-base-200 relative h-7 w-full overflow-hidden rounded-lg" role="group" aria-label="Acts">
			{#each placed as segment (segment.id)}
				<button
					type="button"
					class="absolute inset-y-0 flex items-center overflow-hidden border-r border-black/40 px-1.5 text-[11px] font-medium text-black/80 hover:brightness-110"
					style:left="{segment.left * 100}%"
					style:width="{segment.width * 100}%"
					style:background-color="hsl({segment.hue} 60% 65%)"
					title="{segment.label} at {formatClock(segment.start)}"
					onclick={() => onseek(segment.start)}
				>
					<span class="truncate">{segment.label}</span>
				</button>
			{/each}
		</div>
		<div class="text-base-content/60 flex flex-wrap gap-x-3 gap-y-1 text-xs">
			{#each acts as [label, hue] (label)}
				<span class="flex items-center gap-1.5">
					<span class="h-2 w-2 rounded-full" style:background-color="hsl({hue} 60% 65%)"></span>
					{label}
				</span>
			{/each}
		</div>
	</div>
{/if}
