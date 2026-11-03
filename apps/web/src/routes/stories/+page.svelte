<!-- Stories in the library, newest first; a search asks the story sites. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import StoryCard from '$lib/components/StoryCard.svelte';
	import SiteMediaList from '$lib/components/SiteMediaList.svelte';
	import { fetchStories } from '$lib/media';
	import { storyUrl } from '$lib/routes';
</script>

<SiteMediaList
	title="Stories"
	kind="STORY"
	noun="stories"
	load={fetchStories}
	pick={(batch) => batch.listed.stories}
	cardWidth={150}
>
	{#snippet card(story)}
		<StoryCard {story} onclick={() => goto(storyUrl(story.id))} />
	{/snippet}
</SiteMediaList>
