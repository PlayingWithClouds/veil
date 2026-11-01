<!--
	One video. Opening it reads the pages behind its sources that haven't been
	read yet (credits, description, streams); then the first stream plays.
	The player's quality menu switches between every source's streams.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import VideoPlayer, { type PlayerSource } from '$lib/components/VideoPlayer.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import { openVideo, play, type VideoDetail, type StreamRow } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';

	/** What the player is loading. */
	interface ActiveStream {
		streamId: string;
		src: string;
		mimeType: string;
	}

	let video = $state<VideoDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let active = $state<ActiveStream | null>(null);
	let starting = $state(false);

	let id = $derived(page.params.id ?? '');
	let playerSources = $derived<PlayerSource[]>(
		(video?.streams ?? []).map((stream) => ({
			id: stream.id,
			provider: stream.source?.plugin ?? 'site',
			quality: stream.quality ?? 'auto'
		}))
	);
	let durationLabel = $derived(formatClock(video?.durationSeconds));
	let poster = $derived(cacheUrl(video?.cover?.uri));

	$effect(() => {
		void load(id);
	});

	/** Opens the video and starts its first stream. */
	async function load(videoId: string) {
		video = null;
		active = null;
		error = null;
		loading = true;
		try {
			video = await openVideo(videoId);
			if (video && video.streams.length > 0) await start(video.streams[0]);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Resolves a stream to a playable URL and hands it to the player. */
	async function start(stream: StreamRow) {
		starting = true;
		try {
			const playable = await play(stream.url);
			if (!playable) {
				error = 'This stream could not be resolved.';
				return;
			}
			active = { streamId: stream.id, src: playable.url, mimeType: playable.mimeType };
		} catch (failure) {
			error = String(failure);
		} finally {
			starting = false;
		}
	}

	/** Switches to the stream picked in the player's menu. */
	function selectSource(source: PlayerSource) {
		const stream = video?.streams.find((candidate) => candidate.id === source.id);
		if (stream) void start(stream);
	}

	/** Resolves a stream for the player's quality upgrade. */
	async function resolveSource(sourceId: string) {
		const stream = video?.streams.find((candidate) => candidate.id === sourceId);
		if (!stream) return null;
		return play(stream.url);
	}
</script>

<svelte:head><title>{video?.title ?? 'Video'} · Veil</title></svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-4 pt-2">
	{#if active && video}
		<VideoPlayer
			src={active.src}
			mimeType={active.mimeType}
			title={video.title}
			sources={playerSources}
			activeSourceId={active.streamId}
			onselectsource={selectSource}
			onresolvesource={resolveSource}
			inline
			square={$isCompact}
			{poster}
			onclose={() => (active = null)}
		/>
	{:else}
		<div
			class="bg-base-200 relative flex aspect-video w-full items-center justify-center overflow-hidden rounded-xl"
		>
			{#if poster}
				<img src={poster} alt="" class="absolute inset-0 h-full w-full object-cover opacity-40" />
			{/if}
			<span class="text-base-content/70 relative flex items-center gap-2 text-sm">
				{#if loading || starting}
					<span class="loading loading-spinner"></span>
					{loading ? 'Reading the page…' : 'Starting the stream…'}
				{:else if video && video.streams.length > 0}
					<button
						type="button"
						class="bg-primary text-primary-content flex items-center gap-2 rounded-full px-4 py-2"
						onclick={() => video && start(video.streams[0])}
					>
						<PlayIcon size={18} weight="fill" /> Play
					</button>
				{:else if video}
					No stream found on any site.
				{/if}
			</span>
		</div>
	{/if}

	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}

	{#if video}
		<div class="flex flex-col gap-3">
			<h1 class="text-xl font-semibold">{video.title}</h1>
			<div class="text-base-content/60 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
				{#if video.studio}
					<a href={studioUrl(video.studio.id)} class="text-base-content font-medium hover:underline"
						>{video.studio.name}</a
					>
				{/if}
				{#if video.date}<span>{video.date}</span>{/if}
				{#if durationLabel}<span>{durationLabel}</span>{/if}
			</div>

			{#if video.performers.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each video.performers as performer (performer.id)}
						<a
							href={performerUrl(performer.id)}
							class="bg-base-200 hover:bg-base-300 flex items-center gap-2 rounded-full py-1 pr-3 pl-1 text-sm transition-colors"
						>
							<span class="bg-base-300 h-7 w-7 overflow-hidden rounded-full">
								{#if performer.image}
									<img
										src={cacheUrl(performer.image.uri)}
										alt=""
										class="h-full w-full object-cover"
									/>
								{/if}
							</span>
							{performer.name}
						</a>
					{/each}
				</div>
			{/if}

			{#if video.details}
				<p class="text-base-content/80 text-sm whitespace-pre-line">{video.details}</p>
			{/if}

			<TagChips tags={video.tags} />

			<div class="flex flex-col gap-1">
				<h2 class="text-base-content/50 text-xs font-semibold tracking-wide uppercase">Sources</h2>
				{#each video.sources as source (source.url)}
					<a
						href={source.url}
						target="_blank"
						rel="noreferrer"
						class="text-base-content/70 hover:text-base-content flex items-center gap-2 text-sm"
					>
						<ArrowSquareOutIcon size={14} />
						{siteDisplayName(source.plugin, $sites)}
						{#if !source.fetchedAt}<span class="text-base-content/40">(page not read)</span>{/if}
					</a>
				{/each}
			</div>
		</div>
	{:else if !loading && !error}
		<p class="text-base-content/50 text-sm">This video isn't in the library.</p>
		<button type="button" class="text-primary text-sm" onclick={() => goto('/')}
			>Back to videos</button
		>
	{/if}
</div>
