<!--
	A story in a grid or row: a book-like portrait cover (a generated one when
	the site has none) with its chapter count ("Ongoing" while the author
	still writes), and the caption with the title (its bracketed tags left
	out), the author and the site.
-->
<script lang="ts">
	import BookOpenTextIcon from 'phosphor-svelte/lib/BookOpenTextIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import GeneratedCover from '$lib/components/GeneratedCover.svelte';
	import { cacheUrl } from '$lib/img';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { splitTitle } from '$lib/titleTags';
	import type { StoryCardRow } from '$lib/media';

	interface Props {
		story: StoryCardRow;
		onclick: () => void;
	}

	let { story, onclick }: Props = $props();

	let parts = $derived(splitTitle(story.title));
	let author = $derived(story.performers[0]?.name ?? null);
	let chapterLabel = $derived.by(() => {
		if (story.chapterCount === 1) return '1 chapter';
		return `${story.chapterCount} chapters`;
	});
	let site = $derived.by(() => {
		const sourceUrl = story.sources[0]?.url;
		if (!sourceUrl) return null;
		return pluginForUrl(sourceUrl, $sites);
	});
</script>

<div class="group flex w-full flex-col gap-3">
	<CoverFrame
		src={cacheUrl(story.cover?.uri)}
		alt={story.title}
		aspect="2 / 3"
		placeholder={BookOpenTextIcon}
		{onclick}
	>
		{#snippet fallback()}
			<GeneratedCover title={parts.title} byline={author} kind="story" />
		{/snippet}
		{#if !story.complete}
			<CardBadge corner="top-left">Ongoing</CardBadge>
		{/if}
		{#if story.chapterCount > 1}
			<CardBadge corner="bottom-right">{chapterLabel}</CardBadge>
		{/if}
	</CoverFrame>
	<CardCaption
		title={parts.title}
		byline={author}
		avatarUrl={cacheUrl(story.performers[0]?.image?.uri)}
		siteAvatar={false}
		{site}
		{onclick}
	/>
</div>
