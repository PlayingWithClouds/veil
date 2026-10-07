<!--
	Phone replacement for the top bar's MediaFilterToggle: one chip at the start
	of Home's chip row that steps through all → videos → images on each tap.
-->
<script lang="ts">
	import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import CaretUpDownIcon from 'phosphor-svelte/lib/CaretUpDownIcon';
	import type { Component } from 'svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { mediaFilter, type MediaFilter } from '$lib/stores/mediaFilter';

	type Option = { value: MediaFilter; label: string; icon: Component };

	const options: Option[] = [
		{ value: 'all', label: 'All', icon: SquaresFourIcon },
		{ value: 'videos', label: 'Videos', icon: VideoCameraIcon },
		{ value: 'images', label: 'Images', icon: ImagesIcon }
	];

	let currentIndex = $derived(
		Math.max(
			0,
			options.findIndex((option) => option.value === $mediaFilter)
		)
	);
	let current = $derived(options[currentIndex]);

	/** Moves on to the next media type, wrapping around. */
	function cycle() {
		mediaFilter.set(options[(currentIndex + 1) % options.length].value);
	}
</script>

<button
	type="button"
	onclick={cycle}
	aria-label="Media type: {current.label}. Tap to change."
	class="bg-base-200 ring-base-content/15 flex h-8 shrink-0 items-center gap-1.5 overflow-hidden rounded-lg pr-2.5 pl-3 text-sm font-semibold ring-1 ring-inset"
>
	{#key current.value}
		<span class="flex items-center gap-1.5" in:fly={{ y: 10, duration: 220, easing: cubicOut }}>
			<current.icon size={16} weight="fill" />
			{current.label}
		</span>
	{/key}
	<CaretUpDownIcon size={12} weight="bold" class="opacity-60" />
</button>
