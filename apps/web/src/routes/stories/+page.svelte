<!--
	Stories: Continue reading (straight into the reader), newest releases,
	authors and the most common tags, then everything most recently added
	first; a top bar search asks the story sites.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import StoryCard from '$lib/components/StoryCard.svelte';
	import MediaOverview from '$lib/components/MediaOverview.svelte';
	import SiteMediaList from '$lib/components/SiteMediaList.svelte';
	import { browseStories, fetchContinueReading, fetchStories } from '$lib/media';
	import { storyReaderUrl, storyUrl } from '$lib/routes';

	const CARD_WIDTH = 150;
</script>

<SiteMediaList
	title="Stories"
	kind="STORY"
	noun="stories"
	load={fetchStories}
	browsePage={browseStories}
	pick={(batch) => batch.listed.stories}
	cardWidth={CARD_WIDTH}
>
	{#snippet card(story)}
		<StoryCard {story} onclick={() => goto(storyUrl(story.id))} />
	{/snippet}
	{#snippet overview()}
		<MediaOverview
			kind="story"
			continueTitle="Continue reading"
			peopleTitle="Authors"
			fetchContinue={fetchContinueReading}
			load={fetchStories}
			itemWidth={CARD_WIDTH}
		>
			{#snippet mediaCard(story)}
				<StoryCard {story} onclick={() => goto(storyUrl(story.id))} />
			{/snippet}
			{#snippet continueCard(story)}
				<StoryCard {story} onclick={() => goto(storyReaderUrl(story.id))} />
			{/snippet}
		</MediaOverview>
	{/snippet}
</SiteMediaList>
