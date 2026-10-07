<!--
	A small live line of the download speed over the last minute, filled
	below, scaled to the fastest second in view. Newest sample on the right.
-->
<script lang="ts">
	interface Props {
		// Bytes per second, oldest first.
		samples: number[];
		// How many samples the width holds.
		capacity: number;
	}

	let { samples, capacity }: Props = $props();

	const WIDTH = 100;
	const HEIGHT = 32;

	let peak = $derived(Math.max(1, ...samples));
	let linePath = $derived(buildLine(samples));

	/** The x of sample index, so the newest sits at the right edge. */
	function xAt(index: number): number {
		const offset = capacity - samples.length;
		return ((offset + index) / Math.max(1, capacity - 1)) * WIDTH;
	}

	/** The y of a rate, leaving a hairline at the bottom for zero. */
	function yAt(rate: number): number {
		return HEIGHT - 1 - (rate / peak) * (HEIGHT - 2);
	}

	/** An SVG path through every sample; empty with fewer than two. */
	function buildLine(list: number[]): string {
		if (list.length < 2) return '';
		const points = list.map((rate, index) => `${xAt(index).toFixed(2)},${yAt(rate).toFixed(2)}`);
		return `M${points.join(' L')}`;
	}
</script>

<svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" class="h-full w-full" aria-hidden="true">
	<defs>
		<linearGradient id="speed-fill" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0%" stop-color="var(--color-success)" stop-opacity="0.35" />
			<stop offset="100%" stop-color="var(--color-success)" stop-opacity="0" />
		</linearGradient>
	</defs>
	{#if linePath}
		<path d="{linePath} L{WIDTH},{HEIGHT} L{xAt(0).toFixed(2)},{HEIGHT} Z" fill="url(#speed-fill)" />
		<path d={linePath} fill="none" stroke="var(--color-success)" stroke-width="1.5" vector-effect="non-scaling-stroke" />
	{/if}
</svg>
