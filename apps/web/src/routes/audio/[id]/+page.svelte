<!--
	One audio. Opening it reads the pages behind its sources that weren't read
	yet (voice, tags, streams); then its first stream loads into the player,
	resuming where listening stopped. A banner over the blurred cover shows
	the cover, title, voice, length, date and source, and the player. Below:
	the description, and the tags folded away. Beside it on wide screens
	(below on narrow ones): more by the voice and audio sharing its tags.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import AudioPlayer from '$lib/components/AudioPlayer.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import {
		fetchAudiosBy,
		fetchSimilarAudios,
		openAudio,
		type AudioCardRow,
		type AudioDetail
	} from '$lib/media';
	import { play } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { audioUrl, performerUrl, studioUrl } from '$lib/routes';
	import { formatClock } from '$lib/time';
	import { logEvent } from '$lib/events';
	import { trackWatch } from '$lib/watchLog';

	// Audio in each side list.
	const SIDE_AUDIOS = 8;
	// Longer descriptions are cut until "Show more".
	const LONG_DETAILS = 280;

	/** What the player plays. */
	interface ActiveStream {
		src: string;
		mimeType: string;
	}

	let audio = $state<AudioDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let active = $state<ActiveStream | null>(null);
	let fullDetails = $state(false);
	let byVoice = $state<AudioCardRow[]>([]);
	let similar = $state<AudioCardRow[]>([]);

	let id = $derived(page.params.id ?? '');
	let cover = $derived(cacheUrl(audio?.cover?.uri));
	let durationLabel = $derived(formatClock(audio?.durationSeconds));
	let voice = $derived(audio?.performers[0] ?? null);
	let siteName = $derived.by(() => {
		const source = audio?.sources[0];
		if (!source) return 'The site';
		return siteDisplayName(source.plugin, $sites);
	});

	$effect(() => {
		void load(id);
	});

	/** Opens the audio, resolves its first stream and loads the audio to show beside it. */
	async function load(audioId: string) {
		audio = null;
		active = null;
		byVoice = [];
		similar = [];
		error = null;
		loading = true;
		fullDetails = false;
		void fetchSimilarAudios(audioId, SIDE_AUDIOS).then((found) => {
			if (audioId === id) similar = found;
		});
		try {
			const found = await openAudio(audioId);
			if (audioId !== id) return;
			audio = found;
			if (!found) return;
			logEvent({ kind: 'view', subject: { kind: 'audio', id: found.id } });
			void loadByVoice(found);
			await start(found);
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** The voice's other audio. */
	async function loadByVoice(found: AudioDetail) {
		const voiceId = found.performers[0]?.id;
		if (!voiceId) return;
		const audios = await fetchAudiosBy(voiceId, SIDE_AUDIOS + 1);
		if (found.id !== id) return;
		byVoice = audios.filter((other) => other.id !== found.id).slice(0, SIDE_AUDIOS);
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

<div class="mx-auto grid w-full max-w-[1700px] gap-8 pt-2 xl:grid-cols-[minmax(0,1fr)_380px]">
	<div class="flex min-w-0 flex-col gap-8">
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
			<section class="bg-base-200 relative overflow-hidden rounded-3xl">
				{#if cover}
					<img
						src={cover}
						alt=""
						aria-hidden="true"
						class="absolute inset-0 h-full w-full scale-125 object-cover opacity-40 blur-3xl"
					/>
				{/if}
				<div
					class="from-base-200 via-base-200/85 to-base-200/50 absolute inset-0 bg-linear-to-r"
				></div>

				<div class="relative flex gap-8 p-8 max-md:flex-col max-md:gap-5 max-md:p-5">
					<div
						class="bg-base-300 flex aspect-square w-64 shrink-0 items-center justify-center self-start overflow-hidden rounded-xl shadow-2xl max-md:w-full"
					>
						{#if cover}
							<img src={cover} alt={audio.title} class="h-full w-full object-cover" />
						{:else}
							<HeadphonesIcon size={64} class="text-base-content/20" />
						{/if}
					</div>

					<div class="flex min-w-0 flex-1 flex-col gap-4">
						<div class="flex flex-col gap-1.5">
							<h1 class="text-3xl leading-tight font-bold tracking-tight max-md:text-2xl">
								{audio.title}
							</h1>
							{#if voice || audio.studio}
								<p class="text-base-content/70">
									{#if voice}
										by
										<a
											href={performerUrl(voice.id)}
											class="text-base-content font-medium hover:underline">{voice.name}</a
										>
									{/if}
									{#if voice && audio.studio}·{/if}
									{#if audio.studio}
										<a href={studioUrl(audio.studio.id)} class="hover:underline"
											>{audio.studio.name}</a
										>
									{/if}
								</p>
							{/if}
						</div>

						<div class="flex flex-wrap items-center gap-2 text-xs">
							{#if durationLabel}
								<span class="bg-base-content/10 rounded-full px-3 py-1">{durationLabel}</span>
							{/if}
							{#if audio.date}
								<span class="bg-base-content/10 rounded-full px-3 py-1">{audio.date}</span>
							{/if}
							{#each audio.sources as source (source.url)}
								<a
									href={source.url}
									target="_blank"
									rel="noreferrer"
									class="bg-base-content/10 hover:bg-base-content/20 flex items-center gap-1 rounded-full px-3 py-1"
								>
									{siteDisplayName(source.plugin, $sites)}
									<ArrowSquareOutIcon size={12} />
								</a>
							{/each}
						</div>

						<div class="mt-auto pt-2">
							{#if active}
								<AudioPlayer
									src={active.src}
									mimeType={active.mimeType}
									startPosition={audio.userState.progress ?? 0}
									onattach={(media) => trackWatch(media, id, 'audio')}
								/>
							{:else if loading}
								<p class="text-base-content/60 text-sm">Starting the stream…</p>
							{:else if audio.streams.length === 0}
								<div class="bg-base-content/5 flex flex-col gap-1 rounded-2xl p-4 text-sm">
									<span class="font-medium">Not playable yet</span>
									<span class="text-base-content/60">
										{siteName} offers no stream for this yet; it may still be processing the upload.
									</span>
								</div>
							{/if}
						</div>
					</div>
				</div>
			</section>

			{#if audio.details}
				<section class="flex max-w-prose flex-col items-start gap-2">
					<h2 class="text-lg font-semibold">About</h2>
					<p
						class="text-base-content/85 leading-relaxed whitespace-pre-line"
						class:line-clamp-6={!fullDetails}
					>
						{audio.details}
					</p>
					{#if audio.details.length > LONG_DETAILS}
						<button
							type="button"
							class="text-base-content/60 hover:text-base-content text-sm"
							onclick={() => (fullDetails = !fullDetails)}
						>
							{fullDetails ? 'Show less' : 'Show more'}
						</button>
					{/if}
				</section>
			{/if}

			{#if audio.tags.length > 0}
				<details class="group">
					<summary class="flex cursor-pointer list-none items-center gap-2 text-lg font-semibold">
						Tags
						<span class="text-base-content/50 text-sm font-normal">{audio.tags.length}</span>
						<CaretDownIcon size={14} class="transition-transform group-open:rotate-180" />
					</summary>
					<div class="pt-3">
						<TagChips tags={audio.tags} />
					</div>
				</details>
			{/if}
		{/if}
	</div>

	{#if byVoice.length > 0 || similar.length > 0}
		<aside class="flex flex-col gap-8">
			{#if voice && byVoice.length > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-lg font-semibold">More by {voice.name}</h2>
					<div
						class="grid grid-cols-2 gap-x-4 gap-y-6 max-xl:grid-cols-[repeat(auto-fill,minmax(160px,1fr))]"
					>
						{#each byVoice as other (other.id)}
							<AudioCard audio={other} onclick={() => goto(audioUrl(other.id))} />
						{/each}
					</div>
				</section>
			{/if}
			{#if similar.length > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-lg font-semibold">Similar audio</h2>
					<div
						class="grid grid-cols-2 gap-x-4 gap-y-6 max-xl:grid-cols-[repeat(auto-fill,minmax(160px,1fr))]"
					>
						{#each similar as other (other.id)}
							<AudioCard audio={other} onclick={() => goto(audioUrl(other.id))} />
						{/each}
					</div>
				</section>
			{/if}
		</aside>
	{/if}
</div>
