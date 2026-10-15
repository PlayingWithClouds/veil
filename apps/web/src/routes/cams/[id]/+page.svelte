<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import CaretUpIcon from 'phosphor-svelte/lib/CaretUpIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import Toggle from '$lib/components/Toggle.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import { isCompact } from '$lib/stores/viewport';
	import { activeCamRecordings, refreshCamRecordings } from '$lib/stores/camRecordings';
	import {
		fetchCamModel,
		fetchCamSources,
		fetchCamModelRecordings,
		fetchRecordedSceneCards,
		fetchCamStream,
		formatRecordingClock,
		formatRecordingSize,
		formatViewers,
		loadWatchQuality,
		qualityName,
		recordQualityFor,
		saveWatchQuality,
		setCamAutoRecord,
		setCamModelFollowed,
		startCamRecording,
		stopCamRecording,
		type CamModel,
		type RecordedSceneCard,
		type CamSource
	} from '$lib/cams';
	import { cacheUrl } from '$lib/img';
	import type { PlaybackLevel } from '$lib/playback';
	import { camsUrl, sceneUrl } from '$lib/routes';
	import { pluginIconUrl } from '$lib/server';
	import { notifications } from '$lib/stores/notifications';
	import { formatRelativeTime } from '$lib/time';

	let camModel = $state<CamModel | null>(null);
	let source = $state<CamSource | null>(null);
	let streamUrl = $state<string | null>(null);
	let loading = $state(true);
	let notFound = $state(false);
	let offline = $state(false);
	let followBusy = $state(false);
	// Bumped to start playback again (retry after the room went offline).
	let attempt = $state(0);
	let recordBusy = $state(false);
	// The video height watched at (0 = automatic), remembered between visits.
	let watchHeight = $state(loadWatchQuality());
	// The stream's qualities, tallest first, for the Record menu.
	let streamLevels = $state<PlaybackLevel[]>([]);
	let recordMenuOpen = $state(false);
	// How much of the stream can be rewound to, and how far behind its newest
	// media the playhead is (seconds); the Record menu offers starts in the past.
	let rewindBufferedSeconds = $state(0);
	let rewindLagSeconds = $state(0);

	// A start in the past is offered for these distances when that much is buffered.
	const RECORD_FROM_PAST_OPTIONS = [
		{ label: 'From 1 min ago', seconds: 60 },
		{ label: 'From 3 min ago', seconds: 180 }
	];
	// "Record from here" appears once the playhead is this far behind the newest media.
	const RECORD_FROM_HERE_MINIMUM_SECONDS = 10;
	// A buffer this close to the requested distance still counts as holding it.
	const RECORD_FROM_PAST_SLACK_SECONDS = 10;
	let canRecordFromPast = $derived(
		rewindLagSeconds >= RECORD_FROM_HERE_MINIMUM_SECONDS ||
			rewindBufferedSeconds + RECORD_FROM_PAST_SLACK_SECONDS >= RECORD_FROM_PAST_OPTIONS[0].seconds
	);

	// Recording qualities auto-record can be set to.
	const AUTO_RECORD_QUALITIES = ['source', '1080', '720', '480', '360', '240'];
	let autoRecordBusy = $state(false);
	// Finished recordings of this model, newest first.
	let pastRecordings = $state<RecordedSceneCard[]>([]);
	// Ticks every second while recording, for the timer.
	let now = $state(Date.now());

	// A dropped stream is retried with a freshly resolved URL (the site's stream
	// tokens expire within about a minute) this many times before the room
	// counts as offline; a stream that held for STABLE_STREAM_MS starts over.
	const MAX_RECONNECTS = 3;
	const STABLE_STREAM_MS = 60_000;
	let reconnects = 0;
	let reconnecting = false;
	let stableTimer: ReturnType<typeof setTimeout> | null = null;

	let roomId = $derived(page.params.id ?? '');
	let recording = $derived($activeCamRecordings.find((entry) => entry.camModelId === roomId));
	let recordedSeconds = $derived.by(() => {
		if (!recording) return 0;
		return Math.max(0, (now - Date.parse(recording.startedAt)) / 1000);
	});
	let thumbnail = $derived(cacheUrl(camModel?.thumbnailUrl));
	let sourceIcon = $derived.by(() => {
		if (!source?.iconUrl) return null;
		return cacheUrl(source.iconUrl);
	});

	// Load the model and, when it is live, its stream.
	$effect(() => {
		void attempt;
		const id = roomId;
		let cancelled = false;
		loading = true;
		offline = false;
		notFound = false;
		streamUrl = null;
		reconnects = 0;
		start(id, () => cancelled);
		loadPastRecordings(id);
		return () => {
			cancelled = true;
		};
	});

	/** Loads the model, then asks for its live stream. */
	async function start(id: string, isCancelled: () => boolean) {
		const loaded = await fetchCamModel(id);
		if (isCancelled()) return;
		if (!loaded) {
			notFound = true;
			loading = false;
			return;
		}
		camModel = loaded;
		const sources = await fetchCamSources();
		if (isCancelled()) return;
		source = sources.find((entry) => entry.name === loaded.pluginName) ?? null;
		if (source) source = { ...source, iconUrl: pluginIconUrl(source.iconUrl) };
		if (!loaded.online) {
			offline = true;
			loading = false;
			return;
		}
		try {
			const stream = await fetchCamStream(id);
			if (isCancelled()) return;
			streamUrl = stream.url;
			watchStability();
		} catch {
			if (isCancelled()) return;
			offline = true;
		}
		loading = false;
	}

	/** Forgets earlier reconnects once the current stream has held for a while. */
	function watchStability() {
		if (stableTimer) clearTimeout(stableTimer);
		stableTimer = setTimeout(() => (reconnects = 0), STABLE_STREAM_MS);
	}

	/**
	 * The stream failed: ask for a fresh one (the old URL's token has expired)
	 * and show the offline state when the room does not come back.
	 */
	async function onStreamError() {
		if (reconnecting) return;
		reconnecting = true;
		try {
			if (reconnects >= MAX_RECONNECTS) throw new Error('the stream keeps failing');
			reconnects += 1;
			const stream = await fetchCamStream(roomId);
			streamUrl = stream.url;
			watchStability();
		} catch {
			streamUrl = null;
			offline = true;
		} finally {
			reconnecting = false;
		}
	}

	/** Loads the finished recordings of the model. */
	async function loadPastRecordings(id: string) {
		const finished = (await fetchCamModelRecordings(id)).filter(
			(entry) => entry.status === 'finished'
		);
		const cards = await fetchRecordedSceneCards(finished);
		if (id !== roomId) return;
		pastRecordings = cards;
	}

	// Tick the timer while a recording runs; reload the list when one ends.
	$effect(() => {
		if (!recording) return;
		const timer = setInterval(() => (now = Date.now()), 1000);
		const id = roomId;
		return () => {
			clearInterval(timer);
			void loadPastRecordings(id);
		};
	});

	/** Remembers the quality picked in the player. */
	function onWatchQualityChange(height: number) {
		watchHeight = height;
		saveWatchQuality(height);
	}

	/**
	 * Stops the running recording, or starts one at quality ("source" or a
	 * height); without one it records what is being watched (the source on Auto).
	 */
	async function toggleRecording(quality?: string, fromSecondsAgo?: number) {
		if (!camModel || recordBusy) return;
		recordMenuOpen = false;
		recordBusy = true;
		try {
			if (recording) {
				const finished = await stopCamRecording(camModel.id);
				if (finished && !finished.sceneId)
					notifications.push('The recording was too short to keep', 'error');
			} else {
				await startCamRecording(
					camModel.id,
					quality ?? recordQualityFor(watchHeight),
					fromSecondsAgo
				);
				now = Date.now();
			}
		} catch {
			notifications.push('Could not update the recording', 'error');
		} finally {
			recordBusy = false;
			await refreshCamRecordings();
		}
	}

	/** Turns automatic recording on or off, and sets the quality it records at. */
	async function updateAutoRecord(enabled: boolean, quality?: string) {
		if (!camModel || autoRecordBusy) return;
		autoRecordBusy = true;
		try {
			const updated = await setCamAutoRecord(camModel.id, enabled, quality);
			camModel = {
				...camModel,
				followed: updated.followed,
				autoRecord: updated.autoRecord,
				recordQuality: updated.recordQuality
			};
		} catch {
			notifications.push('Could not update auto-record', 'error');
		} finally {
			autoRecordBusy = false;
		}
	}

	/** Follows or unfollows the model. */
	async function toggleFollow() {
		if (!camModel || followBusy) return;
		followBusy = true;
		const wanted = !camModel.followed;
		try {
			const updated = await setCamModelFollowed(camModel.id, wanted);
			camModel = { ...camModel, followed: updated.followed, autoRecord: updated.autoRecord };
		} catch {
			notifications.push('Could not update follow', 'error');
		} finally {
			followBusy = false;
		}
	}
</script>

<div class="flex flex-col gap-4">
	<a href={camsUrl()} class="text-base-content/50 hover:text-base-content w-fit text-sm">
		&larr; All cams
	</a>

	{#if notFound}
		<p class="text-base-content/40 py-24 text-center">This room is not known.</p>
	{:else}
		<!-- Phones: edge to edge with square corners, like the scene player. -->
		<div
			class="bg-base-200 relative aspect-video w-full overflow-hidden"
			class:rounded-xl={!$isCompact}
		>
			{#if streamUrl}
				{#key streamUrl}
					<VideoPlayer
						src={streamUrl}
						mimeType="application/x-mpegURL"
						title={camModel?.displayName ?? ''}
						poster={thumbnail}
						inline
						live
						square={$isCompact}
						levelHeight={watchHeight}
						onlevels={(levels) => (streamLevels = levels)}
						onlevelchange={onWatchQualityChange}
						controls={recordControl}
						onstreamerror={onStreamError}
						onrewindwindow={(bufferedSeconds, lagSeconds) => {
							rewindBufferedSeconds = bufferedSeconds;
							rewindLagSeconds = lagSeconds;
						}}
						onclose={() => {}}
					/>
				{/key}
			{:else if offline}
				{#if thumbnail}
					<img
						src={thumbnail}
						alt=""
						class="absolute inset-0 h-full w-full object-cover opacity-20"
					/>
				{/if}
				<div class="absolute inset-0 flex flex-col items-center justify-center gap-3 text-center">
					<p class="text-lg font-semibold">{camModel?.displayName} is offline</p>
					{#if camModel?.lastSeenOnlineAt}
						<p class="text-base-content/50 text-sm">
							Last seen {formatRelativeTime(camModel.lastSeenOnlineAt)}
						</p>
					{/if}
					<button type="button" class="btn btn-sm" onclick={() => (attempt += 1)}>
						<ArrowClockwiseIcon size={16} />
						Check again
					</button>
				</div>
			{:else if loading}
				<div class="absolute inset-0 flex items-center justify-center">
					<span class="loading loading-spinner text-base-content/40"></span>
				</div>
			{/if}
		</div>

		{#if camModel}
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 flex-1 basis-80 flex-col gap-2">
					<div class="flex items-center gap-2">
						{#if sourceIcon}
							<img src={sourceIcon} alt={source?.displayName ?? ''} class="h-5 w-5 rounded-sm" />
						{/if}
						<h1 class="truncate text-2xl font-bold">{camModel.displayName}</h1>
						{#if streamUrl}
							<span
								class="flex items-center gap-1 rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold tracking-wide text-white uppercase"
							>
								<span class="h-1.5 w-1.5 rounded-full bg-white"></span>
								Live
							</span>
						{/if}
					</div>
					{#if camModel.subject}
						<p class="text-base-content/70 text-sm">{camModel.subject}</p>
					{/if}
					<div class="text-base-content/50 flex items-center gap-3 text-sm">
						{#if streamUrl}
							<span class="flex items-center gap-1">
								<EyeIcon size={14} weight="fill" />
								{formatViewers(camModel.viewers)} watching
							</span>
						{/if}
						<a
							href={camModel.sourceUrl}
							target="_blank"
							rel="noreferrer noopener"
							class="hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={14} />
							Open on {source?.displayName ?? 'site'}
						</a>
					</div>
					{#if camModel.tags.length > 0}
						<div class="flex flex-wrap gap-1.5">
							{#each camModel.tags as tag (tag)}
								<a
									href="{camsUrl()}?tag={encodeURIComponent(tag)}"
									class="border-base-300 bg-base-200 text-base-content/70 hover:text-base-content rounded-full border px-2.5 py-1 text-xs transition-colors"
								>
									{tag}
								</a>
							{/each}
						</div>
					{/if}
				</div>
				<div class="flex flex-col items-end gap-2">
					<button
						type="button"
						class="btn"
						class:btn-primary={!camModel.followed}
						disabled={followBusy}
						onclick={toggleFollow}
					>
						<HeartIcon size={18} weight={camModel.followed ? 'fill' : 'regular'} />
						{camModel.followed ? 'Following' : 'Follow'}
					</button>
					<label class="flex items-center gap-2 text-sm">
						<RecordIcon size={16} weight="fill" class="text-red-500" />
						<span class="text-base-content/70">Auto-record when live</span>
						<Toggle
							checked={camModel.autoRecord}
							disabled={autoRecordBusy}
							ariaLabel="Auto-record when live"
							onchange={(enabled) => updateAutoRecord(enabled)}
						/>
					</label>
					<label class="flex items-center gap-2 text-sm">
						<span class="text-base-content/70">Auto-record quality</span>
						<select
							class="bg-base-200 border-base-300 rounded-md border px-2 py-1 text-sm"
							value={camModel.recordQuality}
							disabled={autoRecordBusy}
							onchange={(event) =>
								updateAutoRecord(camModel!.autoRecord, event.currentTarget.value)}
						>
							{#each AUTO_RECORD_QUALITIES as quality (quality)}
								<option value={quality}>{qualityName(quality)}</option>
							{/each}
						</select>
					</label>
				</div>
			</div>

			{#if recording || pastRecordings.length > 0}
				<section class="flex flex-col gap-2">
					<h2 class="text-lg font-semibold">Recordings</h2>
					{#if recording}
						<p class="text-sm text-red-400">
							Recording now{recording.quality ? ` at ${recording.quality}` : ''}: {formatRecordingClock(
								recordedSeconds
							)}, {formatRecordingSize(recording.sizeBytes)}
						</p>
					{/if}
					{#if pastRecordings.length > 0}
						<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
							{#each pastRecordings as past (past.recording.id)}
								<SceneCard
									item={past.card}
									instantHover
									onclick={() => past.card.id && goto(sceneUrl(past.card.id))}
								/>
							{/each}
						</div>
					{/if}
				</section>
			{/if}
		{/if}
	{/if}
</div>

{#snippet recordControl()}
	<div class="relative flex flex-none items-center">
		<button
			type="button"
			class="flex h-9 flex-none items-center gap-1.5 rounded-full px-3 text-xs font-semibold text-white hover:bg-white/10 disabled:opacity-50"
			disabled={recordBusy}
			onclick={() => toggleRecording()}
			aria-label={recording ? 'Stop recording' : 'Record'}
			title={recording ? 'Stop recording' : 'Record this room'}
		>
			{#if recording}
				<StopIcon size={16} weight="fill" class="text-red-500" />
				{#if recording.quality}<span class="text-white/70">{recording.quality}</span>{/if}
				<span class="tabular-nums">{formatRecordingClock(recordedSeconds)}</span>
			{:else}
				<RecordIcon size={16} weight="fill" class="text-red-500" />
				Record
			{/if}
		</button>
		{#if !recording}
			<button
				type="button"
				class="flex h-9 w-6 flex-none items-center justify-center text-white/70 hover:text-white"
				class:rotate-180={recordMenuOpen}
				onclick={() => (recordMenuOpen = !recordMenuOpen)}
				aria-label="Choose the recording quality"
				aria-expanded={recordMenuOpen}
			>
				<CaretUpIcon size={14} weight="bold" />
			</button>
		{/if}
		{#if recordMenuOpen && !recording}
			<div
				class="player-record-menu absolute right-0 bottom-full mb-2 w-52 rounded-2xl p-1.5"
				role="menu"
			>
				{#if canRecordFromPast}
					<p
						class="px-3 pt-1.5 pb-1 text-[11px] font-semibold tracking-wider text-white/50 uppercase"
					>
						Start in the past
					</p>
				{/if}
				{#if rewindLagSeconds >= RECORD_FROM_HERE_MINIMUM_SECONDS}
					<button
						type="button"
						role="menuitem"
						class="flex min-h-10 w-full items-center justify-between rounded-xl px-3 text-left text-sm text-white hover:bg-white/10"
						onclick={() => toggleRecording(undefined, Math.round(rewindLagSeconds))}
					>
						From here
						<span class="text-white/50 tabular-nums">-{formatRecordingClock(rewindLagSeconds)}</span
						>
					</button>
				{/if}
				{#each RECORD_FROM_PAST_OPTIONS as option (option.seconds)}
					{#if rewindBufferedSeconds + RECORD_FROM_PAST_SLACK_SECONDS >= option.seconds}
						<button
							type="button"
							role="menuitem"
							class="flex min-h-10 w-full items-center justify-between rounded-xl px-3 text-left text-sm text-white hover:bg-white/10"
							onclick={() => toggleRecording(undefined, option.seconds)}
						>
							{option.label}
						</button>
					{/if}
				{/each}
				<p
					class="px-3 pt-1.5 pb-1 text-[11px] font-semibold tracking-wider text-white/50 uppercase"
				>
					Record at
				</p>
				<button
					type="button"
					role="menuitem"
					class="flex min-h-10 w-full items-center justify-between rounded-xl px-3 text-left text-sm text-white hover:bg-white/10"
					onclick={() => toggleRecording('source')}
				>
					Source (best)
					{#if recordQualityFor(watchHeight) === 'source'}<CheckIcon size={14} weight="bold" />{/if}
				</button>
				{#each streamLevels as level (level.index)}
					{#if level.height > 0}
						<button
							type="button"
							role="menuitem"
							class="flex min-h-10 w-full items-center justify-between rounded-xl px-3 text-left text-sm text-white hover:bg-white/10"
							onclick={() => toggleRecording(String(level.height))}
						>
							{level.label}
							{#if recordQualityFor(watchHeight) === String(level.height)}
								<CheckIcon size={14} weight="bold" />
							{/if}
						</button>
					{/if}
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

<style>
	.player-record-menu {
		background-color: color-mix(in oklab, var(--color-base-200) 70%, transparent);
		backdrop-filter: blur(24px) saturate(1.6);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
	}
</style>
