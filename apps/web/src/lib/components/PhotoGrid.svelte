<!--
	A gallery's photos in justified rows: in reading order (so the lightbox
	steps through them as they appear), each keeping its proportions. A photo
	whose size the site didn't give counts as portrait until it loads.
-->
<script lang="ts" module>
	/** A photo to lay out. */
	export interface GridPhoto {
		id: string;
		src: string;
		width: number | null;
		height: number | null;
	}
</script>

<script lang="ts">
	import { justifiedRows } from '$lib/justify';

	interface Props {
		photos: GridPhoto[];
		onopen: (index: number) => void;
	}

	let { photos, onopen }: Props = $props();

	const GAP = 6;
	const TARGET_HEIGHT = 280;
	const COMPACT_TARGET_HEIGHT = 170;
	const COMPACT_WIDTH = 640;
	// Assumed until a photo without a known size loads.
	const DEFAULT_RATIO = 2 / 3;

	let width = $state(0);
	// Ratios learned from loaded photos the site gave no size for, by id.
	let loadedRatios = $state<Record<string, number>>({});

	let ratios = $derived(photos.map(ratioOf));
	let targetHeight = $derived(width < COMPACT_WIDTH ? COMPACT_TARGET_HEIGHT : TARGET_HEIGHT);
	let rows = $derived(justifiedRows(ratios, width, targetHeight, GAP));

	/** A photo's width over height, as far as it's known. */
	function ratioOf(photo: GridPhoto): number {
		if (photo.width && photo.height) return photo.width / photo.height;
		return loadedRatios[photo.id] ?? DEFAULT_RATIO;
	}

	/** Learns the size of a loaded photo the site gave none for. */
	function onLoad(photo: GridPhoto, event: Event) {
		if (photo.width && photo.height) return;
		const image = event.currentTarget as HTMLImageElement;
		if (!image.naturalWidth || !image.naturalHeight) return;
		loadedRatios[photo.id] = image.naturalWidth / image.naturalHeight;
	}
</script>

<div bind:clientWidth={width} class="flex flex-col" style:gap="{GAP}px">
	{#each rows as row (row.start)}
		<div class="flex" style:gap="{GAP}px" style:height="{row.height}px">
			{#each photos.slice(row.start, row.end) as photo, offset (photo.id)}
				{@const index = row.start + offset}
				<button
					type="button"
					class="bg-base-200 relative h-full min-w-0 overflow-hidden rounded-md focus-visible:ring-2"
					style:flex={row.full ? `${ratios[index]} ${ratios[index]} 0` : 'none'}
					style:width={row.full ? null : `${ratios[index] * row.height}px`}
					aria-label="Photo {index + 1}"
					onclick={() => onopen(index)}
				>
					<img
						src={photo.src}
						alt=""
						loading="lazy"
						decoding="async"
						onload={(event) => onLoad(photo, event)}
						class="h-full w-full object-cover transition-transform duration-300 hover:scale-[1.03]"
					/>
				</button>
			{/each}
		</div>
	{/each}
</div>
