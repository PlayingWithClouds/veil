<!--
	Three-way toggle in the top bar: show videos, images, or both on Home.
-->
<script lang="ts">
	import SquaresFourIcon from 'phosphor-svelte/lib/SquaresFourIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import type { Component } from 'svelte';
	import { mediaFilter, type MediaFilter } from '$lib/stores/mediaFilter';

	type Option = { value: MediaFilter; label: string; icon: Component };

	const options: Option[] = [
		{ value: 'all', label: 'Videos and images', icon: SquaresFourIcon },
		{ value: 'videos', label: 'Videos only', icon: VideoCameraIcon },
		{ value: 'images', label: 'Images only', icon: ImagesIcon }
	];
</script>

<div class="bg-base-200 border-base-300 flex h-10 shrink-0 items-center rounded-full border p-1" role="radiogroup" aria-label="Media type">
	{#each options as option (option.value)}
		<button
			type="button"
			role="radio"
			aria-checked={$mediaFilter === option.value}
			aria-label={option.label}
			title={option.label}
			onclick={() => mediaFilter.set(option.value)}
			class="text-base-content/60 hover:text-base-content flex h-8 w-9 items-center justify-center rounded-full transition-colors"
			class:media-filter-active={$mediaFilter === option.value}
		>
			<option.icon size={18} weight={$mediaFilter === option.value ? 'fill' : 'regular'} />
		</button>
	{/each}
</div>

<style>
	.media-filter-active {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}

	.media-filter-active:hover {
		color: var(--color-base-100);
	}
</style>
