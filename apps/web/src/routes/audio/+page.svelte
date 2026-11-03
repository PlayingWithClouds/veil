<!--
	Audio: Continue listening, newest releases, voices and the most common
	tags, then everything most recently added first; a top bar search asks
	the audio sites.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import MediaOverview from '$lib/components/MediaOverview.svelte';
	import SiteMediaList from '$lib/components/SiteMediaList.svelte';
	import { fetchAudios, fetchContinueListening } from '$lib/media';
	import { audioUrl } from '$lib/routes';

	const CARD_WIDTH = 180;
</script>

<SiteMediaList
	title="Audio"
	kind="AUDIO"
	noun="audio"
	load={fetchAudios}
	pick={(batch) => batch.listed.audios}
	cardWidth={CARD_WIDTH}
>
	{#snippet card(audio)}
		<AudioCard {audio} onclick={() => goto(audioUrl(audio.id))} />
	{/snippet}
	{#snippet overview()}
		<MediaOverview
			kind="audio"
			continueTitle="Continue listening"
			peopleTitle="Voices"
			fetchContinue={fetchContinueListening}
			load={fetchAudios}
			itemWidth={CARD_WIDTH}
		>
			{#snippet mediaCard(audio)}
				<AudioCard {audio} onclick={() => goto(audioUrl(audio.id))} />
			{/snippet}
		</MediaOverview>
	{/snippet}
</SiteMediaList>
