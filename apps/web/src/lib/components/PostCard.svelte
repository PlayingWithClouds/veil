<!--
	A creator post in a grid or row: square cover, a lock while it's
	paywalled, and the caption with its creator.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import NoteIcon from 'phosphor-svelte/lib/NoteIcon';
	import LockSimpleIcon from 'phosphor-svelte/lib/LockSimpleIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import { cacheUrl } from '$lib/img';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import type { PostCardRow } from '$lib/media';

	interface Props {
		post: PostCardRow;
		onclick: () => void;
	}

	let { post, onclick }: Props = $props();

	let site = $derived.by(() => {
		const sourceUrl = post.sources[0]?.url;
		if (!sourceUrl) return null;
		return pluginForUrl(sourceUrl, $sites);
	});
</script>

<div class="group flex w-full flex-col gap-3">
	<CoverFrame src={cacheUrl(post.cover?.uri)} alt={post.title} aspect="1 / 1" placeholder={NoteIcon} {onclick}>
		{#if post.locked}
			<CardBadge corner="top-right"><LockSimpleIcon size={12} weight="fill" /> Locked</CardBadge>
		{/if}
	</CoverFrame>
	<CardCaption
		title={post.title}
		byline={post.creator?.name}
		avatarUrl={cacheUrl(post.creator?.image?.uri)}
		{site}
		detail={post.date ? dayjs(post.date).fromNow() : null}
		{onclick}
	/>
</div>
