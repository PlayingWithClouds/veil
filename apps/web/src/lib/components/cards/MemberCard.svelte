<!--
	A performer or studio in a grid: its picture (portrait for performers,
	landscape for studios), name and what the library holds of it.
-->
<script lang="ts">
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl } from '$lib/routes';
	import type { MemberRow } from '$lib/library';

	interface Props {
		kind: 'performer' | 'studio';
		member: MemberRow;
	}

	let { kind, member }: Props = $props();

	let picture = $derived(cacheUrl(member.image?.uri));
	let portrait = $derived(kind === 'performer');
	let href = $derived.by(() => {
		if (kind === 'performer') return performerUrl(member.id);
		return studioUrl(member.id);
	});
	// "12 videos · 3 galleries", leaving out what's zero.
	let counts = $derived.by(() => {
		const parts: string[] = [];
		if (member.videoCount > 0) parts.push(`${member.videoCount} videos`);
		if (member.galleryCount > 0) parts.push(`${member.galleryCount} galleries`);
		return parts.join(' · ');
	});
</script>

<a {href} class="flex w-full flex-col gap-2 text-left">
	<div
		class="tv-card bg-base-200 relative flex w-full items-center justify-center overflow-hidden rounded-xl"
		style:aspect-ratio={portrait ? '2 / 3' : '16 / 9'}
	>
		{#if picture}
			<img src={picture} alt={member.name} loading="lazy" class="h-full w-full object-cover" />
		{:else if portrait}
			<UserIcon size={32} class="opacity-15" />
		{:else}
			<FilmSlateIcon size={32} class="opacity-15" />
		{/if}
	</div>
	<span class="flex min-w-0 flex-col">
		<span class="line-clamp-1 text-sm font-medium">{member.name}</span>
		{#if counts}
			<span class="text-base-content/50 text-xs">{counts}</span>
		{/if}
	</span>
</a>
