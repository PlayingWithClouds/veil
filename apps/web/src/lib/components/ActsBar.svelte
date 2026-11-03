<!--
	What happens when in a video, under the player: a slim timeline of its
	segments (a site's or the user's markers, named by tag or title) in soft
	colours, one per act, the one playing now lit up; under it every segment
	as a chip with its colour, name and start, in time order. Clicking either
	seeks there. Markers without an end run to the next one.
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
		end: number;
		left: number;
		width: number;
		hue: number;
	}

	interface Props {
		segments: SegmentRow[];
		durationSeconds: number;
		// Where playback is now, in seconds; null before it starts.
		currentSeconds?: number | null;
		onseek: (seconds: number) => void;
	}

	let { segments, durationSeconds, currentSeconds = null, onseek }: Props = $props();

	let placed = $derived(place(segments, durationSeconds));
	let playingId = $derived(playingSegment(placed, currentSeconds));

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
				end,
				left: segment.startSeconds / duration,
				width: (end - segment.startSeconds) / duration,
				hue: hueOf(label)
			};
		});
	}

	/** The segment playing at seconds, the latest started when they overlap. */
	function playingSegment(all: Placed[], seconds: number | null): string | null {
		if (seconds === null) return null;
		let found: string | null = null;
		for (const segment of all) {
			if (segment.start <= seconds && seconds < segment.end) found = segment.id;
		}
		return found;
	}

	/** A steady colour per act name. */
	function hueOf(label: string): number {
		let hash = 0;
		for (const character of label) hash = (hash * 31 + character.charCodeAt(0)) % 360;
		return hash;
	}
</script>

{#if placed.length > 0}
	<div class="flex flex-col gap-2.5">
		<div class="relative h-2 w-full rounded-full bg-white/[0.06]" role="group" aria-label="Acts">
			{#each placed as segment (segment.id)}
				<button
					type="button"
					class="act-segment absolute inset-y-0 rounded-full"
					class:act-playing={segment.id === playingId}
					style:left="calc({segment.left * 100}% + 1px)"
					style:width="max(4px, calc({segment.width * 100}% - 2px))"
					style:--act-color="hsl({segment.hue} 70% 68%)"
					title="{segment.label} · {formatClock(segment.start)}"
					aria-label="{segment.label} at {formatClock(segment.start)}"
					onclick={() => onseek(segment.start)}
				></button>
			{/each}
			{#if currentSeconds !== null && durationSeconds > 0}
				<span
					class="pointer-events-none absolute -inset-y-1 w-0.5 rounded-full bg-white shadow-[0_0_6px_rgba(255,255,255,0.7)]"
					style:left="{Math.min(1, currentSeconds / durationSeconds) * 100}%"
				></span>
			{/if}
		</div>

		<div class="no-scrollbar fade-right-edge -mx-1 flex gap-2 overflow-x-auto px-1 py-0.5">
			{#each placed as segment (segment.id)}
				<button
					type="button"
					class="chip chip-sm"
					class:chip-on={segment.id === playingId}
					onclick={() => onseek(segment.start)}
				>
					<span
						class="h-2 w-2 shrink-0 rounded-full"
						style:background-color="hsl({segment.hue} 70% 68%)"
					></span>
					{segment.label}
					<span class="tabular-nums opacity-50">{formatClock(segment.start) ?? '0:00'}</span>
				</button>
			{/each}
		</div>
	</div>
{/if}

<style>
	.act-segment {
		background-color: color-mix(in oklab, var(--act-color) 55%, transparent);
		transition:
			background-color 150ms,
			transform 150ms,
			box-shadow 150ms;
	}
	.act-segment:hover {
		background-color: var(--act-color);
		transform: scaleY(1.6);
	}
	.act-playing {
		background-color: var(--act-color);
		box-shadow: 0 0 10px color-mix(in oklab, var(--act-color) 60%, transparent);
	}
</style>
