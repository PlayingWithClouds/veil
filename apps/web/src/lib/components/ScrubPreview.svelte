<!--
	Hover scrubbing over a card's picture: the pointer's place across the card
	picks a frame, so any part of the video can be looked at. Segments along the
	bottom mark where the frame sits, with its time beside them. The last frame
	stays a moment after the pointer leaves, then fades. Frames come from the
	site's preview stills, else the video's trickplay sprite sheets; nothing
	loads before the first hover.
-->
<script lang="ts">
	import { formatClock } from '$lib/time';
	import { cacheUrl } from '$lib/img';
	import type { Trickplay } from '$lib/library';

	interface Props {
		frames: { uri: string }[];
		trickplay: Trickplay | null;
		durationSeconds: number | null;
	}

	let { frames, trickplay, durationSeconds }: Props = $props();

	// How long the last frame stays after the pointer leaves.
	const LINGER_MS = 900;
	// More frames than this draw one continuous bar instead of segments.
	const MAX_SEGMENTS = 24;

	let hovered = $state(false);
	let lingering = $state(false);
	// Set on the first hover so frames load only for cards someone points at.
	let armed = $state(false);
	let fraction = $state(0);
	let lingerTimer: ReturnType<typeof setTimeout> | null = null;

	let frameUrls = $derived(frames.map((frame) => cacheUrl(frame.uri) ?? frame.uri));
	let frameCount = $derived.by(() => {
		if (frameUrls.length > 0) return frameUrls.length;
		return trickplayFrameCount(trickplay, durationSeconds);
	});
	let index = $derived(Math.min(frameCount - 1, Math.floor(fraction * frameCount)));
	// Frames whose image has loaded; until the pointed-at one has, the last one shown stays.
	let loadedFrames = $state(new Set<number>());
	let shownFrame = $state<number | null>(null);

	$effect(() => {
		if (loadedFrames.has(index)) shownFrame = index;
	});

	/** Notes a loaded frame image. */
	function onFrameLoad(frame: number) {
		loadedFrames = new Set(loadedFrames).add(frame);
	}
	let shown = $derived(hovered || lingering);
	let timeLabel = $derived.by(() => {
		if (!durationSeconds) return null;
		return formatClock(((index + 0.5) / frameCount) * durationSeconds);
	});

	/** Frames the sprite sheets hold for the video's length. */
	function trickplayFrameCount(sheets: Trickplay | null, seconds: number | null): number {
		if (!sheets || sheets.intervalSeconds <= 0) return 0;
		const capacity = sheets.sheets.length * sheets.columns * sheets.rows;
		if (!seconds) return capacity;
		return Math.max(1, Math.min(capacity, Math.ceil(seconds / sheets.intervalSeconds)));
	}

	/** The CSS that shows one trickplay tile: its sheet, scaled so the tile fills the card. */
	function tileStyle(sheets: Trickplay, frame: number): string {
		const perSheet = sheets.columns * sheets.rows;
		const sheet = sheets.sheets[Math.floor(frame / perSheet)];
		const tile = frame % perSheet;
		const column = tile % sheets.columns;
		const row = Math.floor(tile / sheets.columns);
		const x = sheets.columns > 1 ? (column / (sheets.columns - 1)) * 100 : 0;
		const y = sheets.rows > 1 ? (row / (sheets.rows - 1)) * 100 : 0;
		const url = cacheUrl(sheet.uri) ?? sheet.uri;
		return `background-image: url("${url}"); background-size: ${sheets.columns * 100}% ${sheets.rows * 100}%; background-position: ${x}% ${y}%;`;
	}

	/** Starts scrubbing for a mouse or pen; a finger scrolling past doesn't count. */
	function onPointerEnter(event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		stopLinger();
		armed = true;
		hovered = true;
		follow(event);
	}

	/** Picks the frame under the pointer. */
	function follow(event: PointerEvent) {
		if (!hovered) return;
		const box = (event.currentTarget as HTMLElement).getBoundingClientRect();
		fraction = Math.min(0.999, Math.max(0, (event.clientX - box.left) / box.width));
	}

	/** Keeps the last frame up a moment before it fades. */
	function onPointerLeave() {
		if (!hovered) return;
		hovered = false;
		lingering = true;
		lingerTimer = setTimeout(() => (lingering = false), LINGER_MS);
	}

	/** Cancels a pending fade. */
	function stopLinger() {
		if (lingerTimer !== null) clearTimeout(lingerTimer);
		lingerTimer = null;
		lingering = false;
	}
</script>

{#if frameCount > 0}
	<span
		class="absolute inset-0 transition-opacity duration-200"
		class:opacity-0={!shown}
		onpointerenter={onPointerEnter}
		onpointermove={follow}
		onpointerleave={onPointerLeave}
		role="presentation"
	>
		{#if armed}
			{#if frameUrls.length > 0}
				{#each frameUrls as url, frame (frame)}
					<img
						src={url}
						alt=""
						onload={() => onFrameLoad(frame)}
						class="absolute inset-0 h-full w-full object-cover"
						class:invisible={frame !== shownFrame}
					/>
				{/each}
			{:else if trickplay}
				<span class="absolute inset-0 bg-no-repeat" style={tileStyle(trickplay, index)}></span>
			{/if}

			<span class="absolute inset-x-1.5 bottom-1.5 flex items-center gap-1.5">
				{#if timeLabel}
					<span
						class="rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
					>
						{timeLabel}
					</span>
				{/if}
				{#if frameCount <= MAX_SEGMENTS}
					<span class="flex h-1 flex-1 gap-0.5">
						{#each { length: frameCount }, segment (segment)}
							<span
								class="h-full flex-1 rounded-full bg-white shadow-sm"
								class:opacity-40={segment !== index}
							></span>
						{/each}
					</span>
				{:else}
					<span class="h-1 flex-1 rounded-full bg-white/35">
						<span
							class="block h-full rounded-full bg-white"
							style:width="{((index + 1) / frameCount) * 100}%"
						></span>
					</span>
				{/if}
			</span>
		{/if}
	</span>
{/if}
