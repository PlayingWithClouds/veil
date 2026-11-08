<!--
	One video. Opening it reads the pages behind its sources that haven't been
	read yet (credits, description, streams, comments); then the first stream
	plays. The player's quality menu switches between every source's streams.
	Its cover lights the app's shared glow. Under the player: the acts bar
	(segments: what happens when, the playing one lit), the title, studio
	(changeable), what the user can do with it (like, favourite, rate,
	watchlist, download, ⋮), the performers as large cards with suggestions and
	adding (VideoPeople; a background face scan adds face matches), their cam
	rooms and whether they're live now, the description, the main tags,
	sources and the
	sites' comments. Videos like it (relatedFeed, endless) run alongside on
	wide screens, below on narrow ones.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import ImagesIcon from 'phosphor-svelte/lib/ImagesIcon';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import VideoPlayer, { type PlayerSource } from '$lib/components/VideoPlayer.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import VideoActions from '$lib/components/VideoActions.svelte';
	import VideoListRow from '$lib/components/VideoListRow.svelte';
	import VideoPeople from '$lib/components/VideoPeople.svelte';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';
	import VideoComments from '$lib/components/VideoComments.svelte';
	import ActsBar from '$lib/components/ActsBar.svelte';
	import RecordPickerDialog from '$lib/components/RecordPickerDialog.svelte';
	import VideoCamRooms from '$lib/components/VideoCamRooms.svelte';
	import { fetchVideoRooms, type CamRow } from '$lib/cams';
	import { formatClock } from '$lib/time';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import type { Feed } from '$lib/feed';
	import { relatedFeed } from '$lib/relatedFeed';
	import {
		createStudio,
		fetchStudios,
		openVideo,
		play,
		playableFile,
		setVideoStudio,
		type LocalFileRow,
		type StreamRow,
		type VideoCardRow,
		type VideoDetail
	} from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';
	import { galleryUrl, studioUrl, videoUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';
	import { logEvent } from '$lib/events';
	import { trackWatch } from '$lib/watchLog';
	import { qualityLabel } from '$lib/videoCard';
	import { fetchVideoFaces, trackVideo, type VideoFaceGuesses } from '$lib/faces';

	/** What the player is loading. */
	interface ActiveStream {
		streamId: string;
		src: string;
		mimeType: string;
	}

	// Tags shown before "+N more": the rest is mostly the sites' keyword noise.
	const MAIN_TAGS = 12;
	// Related videos added each time the list's end comes near.
	const RELATED_PAGE = 20;

	let video = $state<VideoDetail | null>(null);
	let related = $state<VideoCardRow[]>([]);
	// The rest of the related list after what's shown, and whether it's loading or through.
	let relatedNext: Feed<VideoCardRow> = async () => [];
	let relatedLoading = $state(false);
	let relatedDone = $state(false);
	let camRooms = $state<CamRow[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let active = $state<ActiveStream | null>(null);
	let starting = $state(false);
	let player = $state<VideoPlayer>();
	let allTags = $state(false);
	let fullDetails = $state(false);
	let changingStudio = $state(false);
	let faces = $state<VideoFaceGuesses | null>(null);
	// The element the player plays in, for the face scan; not reactive, only read.
	let playerElement: HTMLVideoElement | null = null;
	// Where playback is now, for the acts bar; null while nothing plays.
	let playbackSeconds = $state<number | null>(null);

	let id = $derived(page.params.id ?? '');
	let localFiles = $derived((video?.files ?? []).filter((file) => !file.missing));
	let playerSources = $derived<PlayerSource[]>([
		...localFiles.map((file) => ({
			id: file.id,
			provider: 'local',
			quality: file.quality ?? 'source'
		})),
		...(video?.streams ?? []).map((stream) => ({
			id: stream.id,
			provider: stream.source?.plugin ?? 'site',
			quality: stream.quality ?? 'auto'
		}))
	]);
	let canPlay = $derived(localFiles.length > 0 || (video?.streams.length ?? 0) > 0);
	let durationLabel = $derived(formatClock(video?.durationSeconds));
	let poster = $derived(cacheUrl(video?.cover?.uri));
	let shownTags = $derived(allTags ? (video?.tags ?? []) : (video?.tags ?? []).slice(0, MAIN_TAGS));
	let quality = $derived(video ? qualityLabel(video.width, video.height) : null);

	showAmbientGlow(() => poster);

	$effect(() => {
		void load(id);
	});

	/** Opens the video, starts its first stream and loads videos like it. */
	async function load(videoId: string) {
		video = null;
		related = [];
		camRooms = [];
		active = null;
		error = null;
		loading = true;
		allTags = false;
		fullDetails = false;
		faces = null;
		const opening = openVideo(videoId);
		relatedNext = relatedFeed(videoId, opening);
		relatedLoading = false;
		relatedDone = false;
		void loadMoreRelated();
		void watchFaces(videoId);
		try {
			video = await opening;
			// After opening: reading the source pages credits the cam site
			// (the studio) that names the room. Not awaited, as it asks the
			// cam sites whether the rooms are live.
			if (video) void loadCamRooms(videoId);
			if (video) logEvent({ kind: 'view', subject: { kind: 'video', id: video.id } });
			await startFirst();
		} catch (failure) {
			error = String(failure);
		} finally {
			loading = false;
		}
	}

	/** Appends the next related videos, unless the page moved on to another video. */
	async function loadMoreRelated() {
		if (relatedLoading || relatedDone) return;
		const feed = relatedNext;
		relatedLoading = true;
		try {
			const found = await feed(RELATED_PAGE);
			if (feed !== relatedNext) return;
			related = [...related, ...found];
			if (found.length === 0) relatedDone = true;
		} catch (failure) {
			console.warn('related videos', failure);
			if (feed === relatedNext) relatedDone = true;
		} finally {
			if (feed === relatedNext) relatedLoading = false;
		}
	}

	/** Loads the cam rooms of the person in the video, unless the page moved on. */
	async function loadCamRooms(videoId: string) {
		try {
			const found = await fetchVideoRooms(videoId);
			if (videoId === id) camRooms = found;
		} catch (failure) {
			console.warn('cam rooms failed to load', failure);
		}
	}

	/**
	 * Shows who earlier scans recognised, scans the video for faces in the
	 * background (also from the frames the player shows), and shows the
	 * guesses again each time the scan grows and once it's done.
	 */
	async function watchFaces(videoId: string) {
		void loadFaces(videoId);
		const outcome = await trackVideo(videoId, {
			player: () => playerElement,
			onsaved: () => void loadFaces(videoId)
		});
		if (outcome === 'scanned' || outcome === 'skipped') void loadFaces(videoId);
	}

	/** Keeps the playing element for the face scan, follows its time for the acts bar, and logs the watch. */
	function attachPlayer(element: HTMLVideoElement): () => void {
		playerElement = element;
		const stopTracking = trackWatch(element, id);
		const followTime = () => (playbackSeconds = element.currentTime);
		element.addEventListener('timeupdate', followTime);
		return () => {
			stopTracking();
			element.removeEventListener('timeupdate', followTime);
			if (playerElement === element) {
				playerElement = null;
				playbackSeconds = null;
			}
		};
	}

	/** Loads the face guesses for the video, unless the page moved on. */
	async function loadFaces(videoId: string) {
		const found = await fetchVideoFaces(videoId);
		if (videoId === id) faces = found;
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

	/** Plays a local copy when there is one, else the first site stream. */
	async function startFirst() {
		if (localFiles.length > 0) {
			startFile(localFiles[0]);
			return;
		}
		if (video && video.streams.length > 0) await start(video.streams[0]);
	}

	/** Hands a local file to the player; it needs no resolving. */
	function startFile(file: LocalFileRow) {
		const playable = playableFile(file);
		active = { streamId: file.id, src: playable.url, mimeType: playable.mimeType };
	}

	/** Switches to the file or stream picked in the player's menu. */
	function selectSource(source: PlayerSource) {
		const file = localFiles.find((candidate) => candidate.id === source.id);
		if (file) {
			startFile(file);
			return;
		}
		const stream = video?.streams.find((candidate) => candidate.id === source.id);
		if (stream) void start(stream);
	}

	/** Resolves a file or stream for the player's quality upgrade. */
	async function resolveSource(sourceId: string) {
		const file = localFiles.find((candidate) => candidate.id === sourceId);
		if (file) return playableFile(file);
		const stream = video?.streams.find((candidate) => candidate.id === sourceId);
		if (!stream) return null;
		return play(stream.url);
	}

	/** Jumps to a moment, starting playback first when nothing plays. */
	async function seek(seconds: number) {
		if (!active) await startFirst();
		player?.seekTo(seconds);
	}

	/** Sets the studio as the user's; null removes it. */
	async function pickStudio(studioId: string | null) {
		changingStudio = false;
		if (!video) return;
		const saved = await setVideoStudio(video.id, studioId);
		if (saved) video = saved;
	}

	/** Creates a studio by name and sets it as the video's. */
	async function createAndPickStudio(name: string) {
		changingStudio = false;
		const studio = await createStudio(name);
		await pickStudio(studio.id);
	}
</script>

<svelte:head><title>{video?.title ?? 'Video'} · Veil</title></svelte:head>

<div class="mx-auto grid w-full max-w-[1700px] gap-6 pt-2 xl:grid-cols-[minmax(0,1fr)_380px]">
	<div class="flex min-w-0 flex-col gap-4">
		{#if active && video}
			<div class="overflow-hidden rounded-2xl shadow-2xl ring-1 shadow-black/50 ring-white/10">
				<VideoPlayer
					bind:this={player}
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
					startPosition={video.userState.progress ?? 0}
					onattach={attachPlayer}
					onclose={() => (active = null)}
				/>
			</div>
		{:else}
			<div
				class="bg-base-300 relative flex aspect-video w-full items-center justify-center overflow-hidden rounded-2xl ring-1 ring-white/10"
			>
				{#if poster}
					<img src={poster} alt="" class="absolute inset-0 h-full w-full object-cover opacity-40" />
				{/if}
				<span class="text-base-content/70 relative flex items-center gap-2 text-sm">
					{#if loading || starting}
						<span class="loading loading-spinner"></span>
						{loading ? 'Reading the page…' : 'Starting the stream…'}
					{:else if canPlay}
						<button
							type="button"
							class="bg-primary text-primary-content flex h-12 items-center gap-2 rounded-full px-6 font-semibold shadow-xl transition-transform hover:scale-105 active:scale-95"
							onclick={startFirst}
						>
							<PlayIcon size={18} weight="fill" /> Play
						</button>
					{:else if video}
						No stream found on any site.
					{/if}
				</span>
			</div>
		{/if}

		{#if video?.durationSeconds}
			<ActsBar
				segments={video.segments}
				durationSeconds={video.durationSeconds}
				currentSeconds={playbackSeconds}
				onseek={seek}
			/>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		{#if video}
			<div class="flex flex-col gap-5">
				<div class="flex flex-col gap-3">
					<h1 class="text-2xl leading-tight font-bold tracking-tight">{video.title}</h1>
					<div class="text-base-content/60 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
						{#if video.studio}
							<a
								href={studioUrl(video.studio.id)}
								class="text-base-content flex items-center gap-2 font-semibold hover:underline"
							>
								{#if video.studio.image}
									<img
										src={cacheUrl(video.studio.image.uri)}
										alt=""
										class="h-6 w-6 rounded-full object-cover"
									/>
								{/if}
								{video.studio.name}
							</a>
						{/if}
						<button type="button" class="chip chip-sm" onclick={() => (changingStudio = true)}>
							<PencilSimpleIcon size={12} />
							{video.studio ? 'Change studio' : 'Set studio'}
						</button>
						{#if changingStudio}
							<RecordPickerDialog
								kind="studio"
								title={video.studio ? 'Change the studio' : 'Set the studio'}
								placeholder="Find a studio…"
								search={(text) => fetchStudios(text, 24)}
								taken={video.studio ? [video.studio.id] : []}
								clearLabel={video.studio ? 'Remove the studio' : undefined}
								onclear={() => void pickStudio(null)}
								onpick={(studio) => void pickStudio(studio.id)}
								oncreate={(name) => void createAndPickStudio(name)}
								onclose={() => (changingStudio = false)}
							/>
						{/if}
						{#if video.date}<span>{video.date}</span>{/if}
						{#if durationLabel}<span>{durationLabel}</span>{/if}
						{#if quality}<span>{quality}</span>{/if}
						{#if video.userState.viewCount}<span>watched {video.userState.viewCount}×</span>{/if}
					</div>
					<VideoActions {video} currentSeconds={playbackSeconds} />
				</div>

				{#if camRooms.length > 0}
					<VideoCamRooms rooms={camRooms} />
				{/if}

				<VideoPeople {video} {faces} onchange={(saved) => (video = saved)} />

				{#if video.details}
					<div class="flex flex-col items-start gap-1">
						<p
							class="text-base-content/85 max-w-4xl text-[15px] leading-relaxed whitespace-pre-line"
							class:line-clamp-4={!fullDetails}
						>
							{video.details}
						</p>
						{#if video.details.length > 300}
							<button
								type="button"
								class="text-primary text-sm"
								onclick={() => (fullDetails = !fullDetails)}
							>
								{fullDetails ? 'Show less' : 'Show more'}
							</button>
						{/if}
					</div>
				{/if}

				<div class="flex flex-wrap items-center gap-2">
					<TagChips tags={shownTags} />
					{#if video.tags.length > MAIN_TAGS}
						<button
							type="button"
							class="text-base-content/60 hover:text-base-content text-xs"
							onclick={() => (allTags = !allTags)}
						>
							{allTags ? 'Fewer tags' : `+${video.tags.length - MAIN_TAGS} more`}
						</button>
					{/if}
				</div>

				{#if video.gallery}
					<a
						href={galleryUrl(video.gallery.id)}
						class="flex items-center gap-3 self-start rounded-2xl bg-white/[0.04] p-2 pr-5 text-sm ring-1 ring-white/5 transition-all hover:-translate-y-0.5 hover:bg-white/[0.08] hover:ring-white/15"
					>
						<span
							class="bg-base-300 flex h-14 w-11 shrink-0 items-center justify-center overflow-hidden rounded-md"
						>
							{#if video.gallery.cover}
								<img
									src={cacheUrl(video.gallery.cover.uri)}
									alt=""
									class="h-full w-full object-cover"
								/>
							{:else}
								<ImagesIcon size={20} class="text-base-content/40" />
							{/if}
						</span>
						<span class="flex flex-col">
							<span class="font-medium">Photos from this scene</span>
							<span class="text-base-content/60 text-xs">{video.gallery.imageCount} photos</span>
						</span>
					</a>
				{/if}

				<div class="flex flex-wrap items-center gap-2">
					<span class="text-base-content/45 mr-1 text-xs font-semibold tracking-wide uppercase">
						Found on
					</span>
					{#each video.sources as source (source.url)}
						<a href={source.url} target="_blank" rel="noreferrer" class="chip chip-sm">
							{siteDisplayName(source.plugin, $sites)}
							{#if !source.fetchedAt}<span class="opacity-50">· page not read</span>{/if}
							<ArrowSquareOutIcon size={12} class="opacity-60" />
						</a>
					{/each}
				</div>

				<VideoComments comments={video.comments} />
			</div>
		{:else if !loading && !error}
			<p class="text-base-content/50 text-sm">This video isn't in the library.</p>
			<button type="button" class="text-primary text-sm" onclick={() => goto('/')}
				>Back to videos</button
			>
		{/if}
	</div>

	{#if related.length > 0 || relatedLoading}
		<aside class="flex flex-col gap-3">
			<h2 class="flex items-baseline gap-2 text-lg font-semibold">
				Up next
				<span class="text-base-content/45 text-sm font-normal">like this one</span>
			</h2>
			<div
				class="flex flex-col gap-1 max-xl:grid max-xl:grid-cols-[repeat(auto-fill,minmax(320px,1fr))]"
			>
				{#each related as relatedVideo, position (relatedVideo.id)}
					<div
						class="rise-in"
						style:--rise-delay="{Math.min((position % RELATED_PAGE) * 30, 300)}ms"
					>
						<VideoListRow
							video={relatedVideo}
							shown={{ surface: 'related', position }}
							onclick={() => goto(videoUrl(relatedVideo.id))}
						/>
					</div>
				{/each}
			</div>
			<ScrollSentinel onvisible={loadMoreRelated} loading={relatedLoading} done={relatedDone} />
		</aside>
	{/if}
</div>
