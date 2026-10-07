<script lang="ts">
	// Scene markers as a chapter list: timestamp + title, click to jump there.
	// Personal markers are highlighted and can be deleted.
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import { markerTitle, type SceneMarker } from '$lib/markers';

	interface Props {
		markers: SceneMarker[];
		onjump: (seconds: number) => void;
		onremove: (markerId: string) => void;
	}

	let { markers, onjump, onremove }: Props = $props();

	/** "1:05" or "1:05–2:30" for a marker's point or span. */
	function markerTime(marker: SceneMarker): string {
		const start = formatClock(marker.seconds) ?? '0:00';
		if (!marker.endSeconds) return start;
		return `${start}–${formatClock(marker.endSeconds)}`;
	}
</script>

<ul class="grid grid-cols-1 gap-1 sm:grid-cols-2">
	{#each markers as marker (marker.id)}
		<li class="group flex items-center gap-1 rounded-lg transition-colors hover:bg-white/5">
			<button
				type="button"
				class="flex min-w-0 flex-1 items-center gap-3 px-2 py-1.5 text-left text-sm"
				title="Jump to {markerTitle(marker)}"
				onclick={() => onjump(marker.seconds)}
			>
				<span
					class="shrink-0 font-medium tabular-nums"
					class:text-info={!marker.personal}
					class:text-warning={marker.personal}
				>
					{markerTime(marker)}
				</span>
				<span class="truncate">{markerTitle(marker)}</span>
			</button>
			{#if marker.personal}
				<button
					type="button"
					class="text-base-content/30 hover:text-error flex h-7 w-7 shrink-0 items-center justify-center opacity-0 transition-opacity group-hover:opacity-100 focus:opacity-100"
					onclick={() => onremove(marker.id)}
					aria-label="Delete marker"
				>
					<TrashIcon size={14} />
				</button>
			{/if}
		</li>
	{/each}
</ul>
