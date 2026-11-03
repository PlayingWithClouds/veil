<!--
	One audio. Opening it reads the pages behind its sources that weren't read
	yet (voices, tags, streams); then its first stream loads into the player,
	resuming where listening stopped. Beside the cover: the title, studio,
	voices, tags, description and sources.
-->
<script lang="ts">
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
	import AudioPlayer from '$lib/components/AudioPlayer.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import { openAudio, type AudioDetail } from '$lib/media';
	import { play } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl } from '$lib/routes';
	import { formatClock } from '$lib/time';
	import { logEvent } from '$lib/events';
	import { trackWatch } from '$lib/watchLog';

	/** What the player plays. */
	interface ActiveStream {
		src: string;
		mimeType: string;
	}

	let audio = $state<AudioDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let active = $state<ActiveStream | null>(null);

	let id = $derived(page.params.id ?? '');
	let cover = $derived(cacheUrl(audio?.cover?.uri));
	let durationLabel = $derived(formatClock(audio?.durationSeconds));

	$effect(() => {
		void load(id);
	});

	/** Opens the audio and resolves its first stream. */
	async function load(audioId: string) {
		audio = null;
		active = null;
		error = null;
		loading = true;
		try {
			const found = await openAudio(audioId);
			if (audioId !== id) return;
			audio = found;
			if (!found) return;
			logEvent({ kind: 'view', subject: { kind: 'audio', id: found.id } });
			await start(found);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Resolves the audio's first stream to a playable URL. */
	async function start(found: AudioDetail) {
		const stream = found.streams[0];
		if (!stream) return;
		const playable = await play(stream.url);
		if (!playable) {
			error = 'This stream could not be resolved.';
			return;
		}
		if (found.id === id) active = { src: playable.url, mimeType: playable.mimeType };
	}
</script>

<svelte:head><title>{audio?.title ?? 'Audio'} · Veil</title></svelte:head>

<div class="mx-auto flex max-w-4xl flex-col gap-6 pt-2">
	{#if loading && !audio}
		<p class="text-base-content/50 flex items-center gap-2 text-sm">
			<span class="loading loading-spinner"></span> Reading the page…
		</p>
	{/if}
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	{#if !loading && !audio && !error}
		<p class="text-base-content/50 text-sm">This audio isn't in the library.</p>
	{/if}

	{#if audio}
		<div class="flex gap-6 max-md:flex-col">
			<div
				class="bg-base-200 flex aspect-square w-64 shrink-0 items-center justify-center self-start overflow-hidden rounded-2xl max-md:w-full"
			>
				{#if cover}
					<img src={cover} alt={audio.title} class="h-full w-full object-cover" />
				{:else}
					<HeadphonesIcon size={64} class="text-base-content/20" />
				{/if}
			</div>

			<div class="flex min-w-0 flex-1 flex-col gap-3">
				<h1 class="text-2xl font-semibold">{audio.title}</h1>
				<div class="text-base-content/60 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
					{#if audio.studio}
						<a
							href={studioUrl(audio.studio.id)}
							class="text-base-content font-medium hover:underline">{audio.studio.name}</a
						>
					{/if}
					{#if audio.date}<span>{audio.date}</span>{/if}
					{#if durationLabel}<span>{durationLabel}</span>{/if}
					{#each audio.sources as source (source.url)}
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
				{#if audio.performers.length > 0}
					<div class="flex flex-wrap gap-2 text-sm">
						{#each audio.performers as performer (performer.id)}
							<a
								href={performerUrl(performer.id)}
								class="bg-base-200 hover:bg-base-300 rounded-full px-3 py-1">{performer.name}</a
							>
						{/each}
					</div>
				{/if}
				<TagChips tags={audio.tags} />

				<div class="mt-auto">
					{#if active}
						<AudioPlayer
							src={active.src}
							mimeType={active.mimeType}
							startPosition={audio.userState.progress ?? 0}
							onattach={(media) => trackWatch(media, id, 'audio')}
						/>
					{:else if loading}
						<p class="text-base-content/50 text-sm">Starting the stream…</p>
					{:else if audio.streams.length === 0}
						<p class="text-base-content/50 text-sm">No stream found on any site.</p>
					{/if}
				</div>
			</div>
		</div>

		{#if audio.details}
			<p class="text-base-content/80 max-w-prose text-sm leading-relaxed whitespace-pre-line">
				{audio.details}
			</p>
		{/if}
	{/if}
</div>
