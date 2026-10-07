<script lang="ts">
	// Round avatar for a subscription: the studio logo / performer photo, their
	// initial when there is none, or a tag / search icon.
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import TagSimpleIcon from 'phosphor-svelte/lib/TagSimpleIcon';
	import { cacheUrl } from '$lib/img';
	import { subscriptionName, type SearchSubscription } from '$lib/searchSubscriptions';

	interface Props {
		subscription: SearchSubscription;
		// Diameter in pixels.
		size?: number;
	}

	let { subscription, size = 24 }: Props = $props();

	let imageUrl = $derived(cacheUrl(subscription.target?.imageUrl));
	let iconSize = $derived(Math.round(size * 0.55));
</script>

<span
	class="bg-base-300 flex shrink-0 items-center justify-center overflow-hidden rounded-full font-semibold uppercase"
	style:width="{size}px"
	style:height="{size}px"
	style:font-size="{Math.round(size * 0.45)}px"
>
	{#if imageUrl}
		<img src={imageUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
	{:else if subscription.kind === 'SEARCH'}
		<MagnifyingGlassIcon size={iconSize} />
	{:else if subscription.kind === 'TAG'}
		<TagSimpleIcon size={iconSize} />
	{:else}
		{subscriptionName(subscription).charAt(0)}
	{/if}
</span>
