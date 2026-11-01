<!--
	A performer or studio in a grid: its picture (portrait for performers,
	landscape for studios), name and what the library holds of it.
-->
<script lang="ts">
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import { cacheUrl } from '$lib/img';

	interface Props {
		kind: 'performer' | 'studio';
		name: string;
		imageUrl: string | null;
		// A short line under the name, like "12 videos".
		detail?: string | null;
		href: string;
	}

	let { kind, name, imageUrl, detail = null, href }: Props = $props();

	let picture = $derived(cacheUrl(imageUrl));
	let portrait = $derived(kind === 'performer');
</script>

<a {href} class="flex w-full flex-col gap-2 text-left">
	<div
		class="tv-card bg-base-200 relative flex w-full items-center justify-center overflow-hidden rounded-xl"
		style:aspect-ratio={portrait ? '2 / 3' : '16 / 9'}
	>
		{#if picture}
			<img src={picture} alt={name} loading="lazy" class="h-full w-full object-cover" />
		{:else if portrait}
			<UserIcon size={32} class="opacity-15" />
		{:else}
			<FilmSlateIcon size={32} class="opacity-15" />
		{/if}
	</div>
	<span class="flex min-w-0 flex-col">
		<span class="line-clamp-1 text-sm font-medium">{name}</span>
		{#if detail}
			<span class="text-base-content/50 text-xs">{detail}</span>
		{/if}
	</span>
</a>
