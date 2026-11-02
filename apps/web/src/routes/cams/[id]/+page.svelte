<!-- One cam room: its live stream, played as soon as the page opens. -->
<script lang="ts">
	import { page } from '$app/state';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import { fetchCam, watchCam, type CamRow } from '$lib/cams';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { isCompact } from '$lib/stores/viewport';

	let cam = $state<CamRow | null>(null);
	let stream = $state<{ url: string; mimeType: string } | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let id = $derived(page.params.id ?? '');

	$effect(() => {
		void load(id);
	});

	/** Loads the room and, while it's live, its stream. */
	async function load(camId: string) {
		cam = null;
		stream = null;
		error = null;
		loading = true;
		try {
			cam = await fetchCam(camId);
			if (cam?.live) stream = await watchCam(camId);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>{cam?.displayName ?? 'Cam'} · Veil</title></svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-4 pt-2">
	{#if stream && cam}
		<VideoPlayer
			src={stream.url}
			mimeType={stream.mimeType}
			title={cam.displayName}
			poster={cacheUrl(cam.live?.thumbnail?.uri ?? cam.thumbnail?.uri)}
			live
			inline
			square={$isCompact}
			onclose={() => (stream = null)}
		/>
	{:else}
		<div
			class="bg-base-200 text-base-content/70 flex aspect-video w-full items-center justify-center rounded-xl text-sm"
		>
			{#if loading}
				<span class="loading loading-spinner"></span>
			{:else if cam && !cam.live}
				{cam.displayName} is offline.
			{:else if !cam}
				This room isn't known.
			{/if}
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if cam}
		<div class="flex flex-col gap-1">
			<h1 class="text-xl font-semibold">{cam.displayName}</h1>
			<p class="text-base-content/60 text-sm">
				{siteDisplayName(cam.plugin, $sites)}
				{#if cam.live}· {cam.live.viewers} watching{/if}
			</p>
			{#if cam.live?.subject}
				<p class="text-base-content/80 text-sm">{cam.live.subject}</p>
			{/if}
		</div>
	{/if}
</div>
