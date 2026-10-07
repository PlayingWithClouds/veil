<script lang="ts">
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import ImageIcon from 'phosphor-svelte/lib/ImageIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import type { Component } from 'svelte';
	import { cacheUrl } from '$lib/img';
	import type { CollectionMember } from '$lib/collections';

	let { member, onclick }: { member: CollectionMember; onclick?: () => void } = $props();

	const typeIcons: Record<string, Component> = {
		scene: FilmStripIcon,
		gallery: ImagesIcon,
		image: ImageIcon,
		performer: UserIcon,
		studio: FilmSlateIcon
	};

	let TypeIcon = $derived(typeIcons[member.mediaType] ?? FilmStripIcon);
	let poster = $derived(cacheUrl(member.posterPath));
	// Performers read best as portraits; everything else is landscape.
	let portrait = $derived(member.mediaType === 'performer');
</script>

<button type="button" class="flex w-full flex-col gap-2 text-left" {onclick}>
	<div
		class="tv-card bg-base-200 relative flex w-full items-center justify-center overflow-hidden rounded-xl {portrait
			? 'aspect-[2/3]'
			: 'aspect-video'}"
	>
		{#if poster}
			<img src={poster} alt={member.title} loading="lazy" class="h-full w-full object-cover" />
		{:else}
			<TypeIcon size={32} class="opacity-15" />
		{/if}
		<span
			class="absolute left-1.5 top-1.5 flex items-center gap-1 rounded-md bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white/80"
		>
			<TypeIcon size={11} />
			{member.mediaType}
		</span>
	</div>
	<span class="line-clamp-1 text-sm font-medium">{member.title}</span>
</button>
