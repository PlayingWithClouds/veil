<!-- Audio in the library, newest first; a search asks the audio sites. -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import SiteMediaList from '$lib/components/SiteMediaList.svelte';
	import { fetchAudios } from '$lib/media';
	import { audioUrl } from '$lib/routes';
</script>

<SiteMediaList
	title="Audio"
	kind="AUDIO"
	noun="audio"
	load={fetchAudios}
	pick={(batch) => batch.listed.audios}
	cardWidth={180}
>
	{#snippet card(audio)}
		<AudioCard {audio} onclick={() => goto(audioUrl(audio.id))} />
	{/snippet}
</SiteMediaList>
