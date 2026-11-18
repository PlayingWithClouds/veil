<!--
	One uploader of a tube site (?plugin=, ?url=, ?name=): the videos on their
	channel page, read from the site a page at a time and stored as they come.
	Uploaders aren't studios and have no record, so the page is addressed by
	the site and their page.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import MediaFeed from '$lib/components/MediaFeed.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { videoUrl } from '$lib/routes';
	import { uploaderFeed } from '$lib/uploader';

	let plugin = $derived(page.url.searchParams.get('plugin') ?? '');
	let url = $derived(page.url.searchParams.get('url') ?? '');
	let name = $derived(page.url.searchParams.get('name') ?? '');
	let site = $derived(siteDisplayName(plugin, $sites));
</script>

<svelte:head><title>{name || 'Uploader'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6">
	<PageHeader title={name || 'Uploader'} description={`Uploader on ${site}`}>
		{#snippet actions()}
			{#if url}
				<a href={url} target="_blank" rel="noreferrer" class="chip chip-sm">
					<ArrowSquareOutIcon size={13} /> Open on {site}
				</a>
			{/if}
		{/snippet}
	</PageHeader>

	{#if plugin && url}
		{#key `${plugin}|${url}`}
			<MediaFeed
				feed={() => uploaderFeed(plugin, url, name)}
				cardWidth={260}
				surface="uploader"
				empty="No videos found on this uploader's page."
			>
				{#snippet card(video, shown)}
					<SceneCard {video} {shown} onclick={() => goto(videoUrl(video.id))} />
				{/snippet}
			</MediaFeed>
		{/key}
	{:else}
		<p class="text-base-content/50 text-sm">This uploader has no page to read videos from.</p>
	{/if}
</div>
