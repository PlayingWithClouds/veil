<!-- One studio: logo, description, network, and its videos and galleries. -->
<script lang="ts">
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import RecordMedia from '$lib/components/RecordMedia.svelte';
	import { fetchStudio, type StudioDetail } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { studioUrl } from '$lib/routes';

	let studio = $state<StudioDetail | null>(null);
	let loading = $state(true);

	let id = $derived(page.params.id ?? '');

	$effect(() => {
		loading = true;
		fetchStudio(id)
			.then((found) => (studio = found))
			.finally(() => (loading = false));
	});
</script>

<svelte:head><title>{studio?.name ?? 'Studio'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6 pt-2">
	{#if studio}
		<div class="flex items-center gap-4">
			<div
				class="bg-base-200 flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-full"
			>
				{#if studio.image}
					<img
						src={cacheUrl(studio.image.uri)}
						alt={studio.name}
						class="h-full w-full object-cover"
					/>
				{:else}
					<FilmSlateIcon size={32} class="opacity-20" />
				{/if}
			</div>
			<div class="flex min-w-0 flex-col gap-1">
				<h1 class="text-2xl font-bold tracking-tight">{studio.name}</h1>
				<div class="text-base-content/60 flex flex-wrap gap-x-4 gap-y-1 text-sm">
					<span>{studio.videoCount} videos</span>
					{#if studio.galleryCount > 0}<span>{studio.galleryCount} galleries</span>{/if}
					{#if studio.parent}
						<a href={studioUrl(studio.parent.id)} class="hover:text-base-content"
							>Part of {studio.parent.name}</a
						>
					{/if}
					{#if studio.url}
						<a
							href={studio.url}
							target="_blank"
							rel="noreferrer"
							class="hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={13} /> Website
						</a>
					{/if}
					{#each studio.sources as source (source.url)}
						<a
							href={source.url}
							target="_blank"
							rel="noreferrer"
							class="hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={13} />
							{siteDisplayName(source.plugin, $sites)}
						</a>
					{/each}
				</div>
			</div>
		</div>
		{#if studio.details}
			<p class="text-base-content/80 max-w-3xl text-sm whitespace-pre-line">{studio.details}</p>
		{/if}

		<RecordMedia
			filter={{ studioId: studio.id }}
			allVideosHref={`/?${new URLSearchParams({ studio: studio.id })}`}
		/>
	{:else if !loading}
		<p class="text-base-content/50 text-sm">This studio isn't in the library.</p>
	{/if}
</div>
