<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import { miniPlayer, takeMiniPlayback } from '$lib/stores/miniPlayer';
	import { takeSceneHandoff, type PlaybackHandoff } from '$lib/playback';
	import { untrack } from 'svelte';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import AlikeSourcesDialog from '$lib/components/AlikeSourcesDialog.svelte';
	import { activateFocusZone } from '$lib/components/FocusList.svelte';
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import ChannelRow from '$lib/components/scene/ChannelRow.svelte';
	import PlayerPoster from '$lib/components/scene/PlayerPoster.svelte';
	import SceneActions from '$lib/components/scene/SceneActions.svelte';
	import SceneDescription from '$lib/components/scene/SceneDescription.svelte';
	import SceneRail from '$lib/components/scene/SceneRail.svelte';
	import SceneSkeleton from '$lib/components/scene/SceneSkeleton.svelte';
	import SceneHeaderCompact from '$lib/components/scene/SceneHeaderCompact.svelte';
	import ScenePills from '$lib/components/scene/ScenePills.svelte';
	import SceneAboutCompact from '$lib/components/scene/SceneAboutCompact.svelte';
	import SceneUpNextFeed from '$lib/components/scene/SceneUpNextFeed.svelte';
	import { RelatedFeed } from '$lib/components/scene/relatedFeed.svelte';
	import { chromeHidden, immersiveTop, isCompact } from '$lib/stores/viewport';
	import {
		fetchSceneMarkers,
		createSceneMarker,
		deleteSceneMarker,
		markerTitle,
		type SceneMarker
	} from '$lib/markers';
	import { addBlock } from '$lib/blocklist';
	import { cacheUrl } from '$lib/img';
	import { sceneUrl } from '$lib/routes';
	import type { TagRef } from '$lib/tags';
	import { notifications } from '$lib/stores/notifications';
	import { downloads } from '$lib/stores/downloads';
	import { sidebarFocus } from '$lib/stores/focus';
	import { watchProgress } from '$lib/stores/watchProgress';
	import {
		ingestPluginResult,
		fetchPluginIcons,
		fetchSearchPlugins,
		pluginForUrl,
		subscribeRelated,
		type PluginSearchResult,
		type SearchPlugin
	} from '$lib/search';
	import {
		subscribeStreams,
		ensureSceneStreams,
		loadSceneDetail,
		resolveStream,
		queueDownload,
		fetchSceneSite,
		groupStreams,
		qualityLabel,
		streamHeight,
		streamLabel,
		type SceneDetail,
		type SceneSite,
		type SceneStream
	} from '$lib/sceneDetail';
	import { findAlikeSources, type AlikeCandidate } from '$lib/alike';

	type Direction = 'left' | 'right' | 'up' | 'down';

	let { data } = $props();
	// The first visit fetches the scene's details from its site (ensureSceneStreams),
	// after SSR loaded the stub; the refetched scene replaces the stale snapshot.
	let refreshedScene = $state<typeof data.scene | null>(null);
	let scene = $derived(refreshedScene ?? data.scene);

	// Sources arrive from SSR load, then stream in live as plugins resolve them.
	let streams = $state<SceneStream[]>([]);
	let ensuring = $state(false);
	// Skeletons stand in for studio/performers/tags until the detail fetch lands.
	let detailPending = $derived(ensuring && refreshedScene === null);

	// data.streams changes only on navigation (fresh SSR load), so this doubles as
	// the "new scene" reset: clear stale inline playback so the reused component
	// doesn't keep showing the previous video and autoplay can fire again.
	$effect(() => {
		streams = data.streams;
		refreshedScene = null;
		activeStream = null;
		resolving = null;
		queueing = null;
		pendingSeek = null;
		alikeOpen = false;
		alikeCandidates = [];
		alikeFetchedId = null;
		alikeLoading = false;
		// Opened from the mini player or a card's hover player: take its running
		// video over as is (the seek only applies when it has to reload, e.g.
		// after a page reload).
		let handoff: PlaybackHandoff | null = takeMiniPlayback(data.id);
		if (!handoff) {
			handoff = takeSceneHandoff(data.id);
		}
		if (handoff) {
			pendingSeek = handoff.position;
			activeStream = {
				src: handoff.src,
				mimeType: handoff.mimeType,
				title: handoff.title,
				sourceId: handoff.sourceId
			};
			autoPlayedFor = data.id;
		}
	});

	// Phones: the player starts right under the status bar; the top bar only
	// comes back when scrolling up mid-page.
	$effect(() => {
		if (!$isCompact) return;
		immersiveTop.set(true);
		return () => {
			immersiveTop.set(false);
			chromeHidden.set(false);
		};
	});

	/** The handler on phones only; desktop keeps its own top bar and no swipes. */
	function compactOnly(handler: () => void): (() => void) | undefined {
		if (!$isCompact) return undefined;
		return handler;
	}

	/**
	 * Leaves the watch page (back, or Home without history). A playing video
	 * carries on in the mini player.
	 */
	function leavePage() {
		if (history.length > 1) {
			history.back();
			return;
		}
		goto('/');
	}

	// Leaving for a non-scene page while playing keeps the video going in the
	// mini player.
	beforeNavigate(({ to }) => {
		const routeId = to?.route.id;
		if (!routeId || routeId === '/scene/[id]' || !activeStream || !playerRef) return;
		const { position, playing, muted } = playerRef.playback();
		if (!playing) return;
		miniPlayer.set({
			sceneId: data.id,
			title: scene?.title ?? activeStream.title,
			src: activeStream.src,
			mimeType: activeStream.mimeType,
			sourceId: activeStream.sourceId,
			position,
			playing,
			muted
		});
	});
	$effect(() => {
		if (!data.id) return;
		return subscribeStreams(data.id, (live) => (streams = live));
	});

	// Stub scenes (discovered but never scraped) load with no streams. Scrape the
	// origin URL on demand so the page has something to play. Guard on the live
	// stream list + a per-id flag so this fires once per scene, not in a loop.
	let ensuredId = $state<string | null>(null);
	$effect(() => {
		if (!data.id || ensuring || ensuredId === data.id) return;
		if (streams.length > 0) {
			ensuredId = data.id;
			return;
		}
		ensuring = true;
		const id = data.id;
		ensureSceneStreams(id)
			.then((live) => {
				if (live.length > 0) streams = live;
				return reloadScene(id);
			})
			.finally(() => {
				ensuring = false;
				ensuredId = id;
			});
	});

	/** Refetches the scene after its details were fetched, unless the user navigated away. */
	async function reloadScene(id: string) {
		const detail = await loadSceneDetail(id);
		if (data.id !== id || !detail.scene) return;
		refreshedScene = detail.scene;
	}

	let best = $derived<SceneStream | null>(streams.length > 0 ? streams[0] : null);
	let streamGroups = $derived(groupStreams(streams));

	// The site the scene was scraped from, for the channel row and description.
	let site = $state<SceneSite | null>(null);
	let sourceUrl = $derived(scene?.sourceUrl);
	$effect(() => {
		const url = sourceUrl;
		site = null;
		if (!url) return;
		fetchSceneSite(url).then((found) => {
			if (sourceUrl === url) site = found;
		});
	});

	let pluginIcons = $state<Record<string, string | null>>({});
	$effect(() => {
		fetchPluginIcons().then((icons) => (pluginIcons = icons));
	});

	// Sites, to label the fullscreen up-next cards by their source URL.
	let searchPlugins = $state<SearchPlugin[]>([]);
	$effect(() => {
		fetchSearchPlugins().then((plugins) => (searchPlugins = plugins));
	});

	// --- Find-alike ----------------------------------------------------------
	// Alternate copies of this scene on other sites, matched by
	// performers/duration/poster rather than title.
	let alikeOpen = $state(false);
	let alikeCandidates = $state<AlikeCandidate[]>([]);
	let alikeLoading = $state(false);
	// Which scene the loaded candidates belong to, so the fetch runs once per scene.
	let alikeFetchedId = $state<string | null>(null);
	let alikeCount = $derived.by(() => {
		if (alikeFetchedId !== data.id) return 0;
		return alikeCandidates.length;
	});

	/** Searches alternate sources for a scene, once per scene. */
	async function loadAlike(id: string) {
		if (alikeLoading || alikeFetchedId === id) return;
		alikeLoading = true;
		try {
			const found = await findAlikeSources(id);
			if (data.id === id) {
				alikeCandidates = found;
				alikeFetchedId = id;
			}
		} finally {
			alikeLoading = false;
		}
	}

	/** Opens the find-sources dialog and starts its search. */
	function openAlike() {
		alikeOpen = true;
		if (data.id) loadAlike(data.id);
	}

	// Auto-hint: once source resolution has settled with nothing to play, search
	// for alternates in the background so the button can advertise a count.
	$effect(() => {
		if (!data.id || ensuring || ensuredId !== data.id) return;
		if (streams.length > 0 || alikeFetchedId === data.id || alikeLoading) return;
		loadAlike(data.id);
	});

	/** After attaching an alternate source, plays the new provider's best quality. */
	function onAlikeAttached(live: SceneStream[], pluginName: string) {
		streams = live;
		const fromPlugin = live
			.filter((stream) => stream.pluginName === pluginName)
			.sort((first, second) => streamHeight(second) - streamHeight(first));
		const target = fromPlugin[0] ?? live[0];
		if (target) play(target);
	}

	// --- Playback ------------------------------------------------------------
	/** The player preview frame: hqporner "_main" posters have a "_1" first-frame sibling. */
	function firstFrame(url: string | null | undefined): string | null {
		if (!url) return null;
		return url.replace(/_main\.jpg$/i, '_1.jpg');
	}
	let previewFrame = $derived(cacheUrl(firstFrame(scene?.posterPath)));

	// Inline playback: null until a source is resolved to a playable URL.
	let activeStream = $state<{
		src: string;
		mimeType: string;
		title: string;
		sourceId: string;
	} | null>(null);
	let resolving = $state<string | null>(null);
	let queueing = $state<string | null>(null);
	let activeSourceId = $derived.by(() => {
		if (!activeStream) return null;
		return activeStream.sourceId;
	});
	let playerRef = $state<
		| {
				seekTo: (seconds: number) => void;
				playback: () => { position: number; playing: boolean; muted: boolean };
		  }
		| undefined
	>();
	let pendingSeek = $state<number | null>(null);

	// A marker jump wins over the saved resume position.
	let startPosition = $derived.by(() => {
		if (pendingSeek !== null) return pendingSeek;
		if (!data.id) return 0;
		const progress = $watchProgress.get(data.id);
		if (!progress) return 0;
		return progress.progressSeconds;
	});

	// Flat source list handed to the player so quality/provider can be switched
	// from inside the player controls.
	let playerSources = $derived(
		streamGroups.flatMap((group) =>
			group.streams.map((stream) => ({
				id: stream.id,
				provider: group.provider,
				quality: qualityLabel(stream),
				verified: stream.verified,
				speedBps: stream.expectedSpeedBps
			}))
		)
	);

	/** Plays the source the user picked in the player's source menu. */
	function selectPlayerSource(sourceId: string) {
		const stream = streams.find((entry) => entry.id === sourceId);
		if (stream) play(stream);
	}

	/** Resolves a source id to a playable URL, for the player's adaptive prebuffering. */
	async function resolvePlayerSource(sourceId: string) {
		const stream = streams.find((entry) => entry.id === sourceId);
		if (!stream) return null;
		try {
			const resolved = await resolveStream(stream.url);
			if (!resolved) return null;
			return { url: resolved.url, mimeType: resolved.mimeType };
		} catch {
			return null;
		}
	}

	/** Resolves a source to a playable URL and starts it in the inline player. */
	async function play(stream: SceneStream) {
		resolving = stream.id;
		try {
			const resolved = await resolveStream(stream.url);
			if (!resolved) throw new Error('no stream');
			activeStream = {
				src: resolved.url,
				mimeType: resolved.mimeType,
				title: `${scene?.title ?? ''} – ${streamLabel(stream)}`,
				sourceId: stream.id
			};
		} catch {
			notifications.push('Could not resolve stream', 'error');
		} finally {
			resolving = null;
		}
	}

	/** Queues a source for download to the NAS and tracks the job. */
	async function download(stream: SceneStream) {
		queueing = stream.id;
		try {
			const jobId = await queueDownload(
				stream.url,
				`${scene?.title ?? ''} – ${streamLabel(stream)}`
			);
			if (data.id) downloads.start(jobId, data.id);
			notifications.push(`Download started: ${scene?.title}`, 'success');
		} catch (error) {
			notifications.push(String(error), 'error');
		} finally {
			queueing = null;
		}
	}

	// Autoplay once a playable source is known, once per scene. Closing the player
	// won't retrigger it (the per-id guard stays set).
	let autoPlayedFor = $state<string | null>(null);
	$effect(() => {
		if (!best || activeStream || autoPlayedFor === data.id) return;
		autoPlayedFor = data.id;
		play(best);
	});

	// --- Scene markers -------------------------------------------------------
	let markers = $state<SceneMarker[]>([]);
	let markersLoadedFor = $state<string | null>(null);
	$effect(() => {
		const id = data.id;
		if (!id || markersLoadedFor === id) return;
		markersLoadedFor = id;
		fetchSceneMarkers(id).then((list) => (markers = list));
	});

	let playerMarkers = $derived(
		markers.map((marker) => ({
			id: marker.id,
			label: markerTitle(marker),
			seconds: marker.seconds,
			endSeconds: marker.endSeconds,
			personal: marker.personal
		}))
	);

	/** Asks for a tag (or, without one, a note) and adds a personal marker at the given time. */
	async function addMarker(seconds: number) {
		if (!data.id) return;
		const clock = formatClock(seconds) ?? '0:00';
		const tagName = window.prompt(
			`Marker at ${clock} — tag (e.g. a position; leave empty for a note)`,
			''
		);
		if (tagName === null) return;
		let label: string | null = null;
		if (!tagName.trim()) {
			label = window.prompt(`Marker at ${clock} — note`, clock);
			if (!label) return;
		}
		const created = await createSceneMarker(data.id, seconds, {
			tagName: tagName.trim() || undefined,
			label: label ?? undefined
		});
		markers = [...markers, created].sort((first, second) => first.seconds - second.seconds);
	}

	/** Deletes a personal marker. */
	async function removeMarker(markerId: string) {
		await deleteSceneMarker(markerId);
		markers = markers.filter((marker) => marker.id !== markerId);
	}

	/** Seeks a playing stream to a marker, or starts playback there. */
	function jumpToMarker(seconds: number) {
		if (activeStream && playerRef) {
			playerRef.seekTo(seconds);
			return;
		}
		pendingSeek = seconds;
		if (best) play(best);
	}

	// --- Blocklist -----------------------------------------------------------
	/** Blocks a tag, performer or studio from recommendations. */
	async function blockEntity(kind: 'tag' | 'performer' | 'studio', id: string, label: string) {
		await addBlock(kind, id, label);
		notifications.push(`Blocked ${label}`, 'success');
	}

	/** Blocks a tag picked from the description's tag chips. */
	function blockTag(tag: TagRef) {
		blockEntity('tag', tag.id, tag.name);
	}

	// --- Up-next rail --------------------------------------------------------
	/** Whether a feed item is the scene currently on screen (never recommended). */
	function isCurrentScene(item: PluginSearchResult): boolean {
		if (item.sceneId && item.sceneId === data.id) return true;
		const seedTitle = (scene?.title ?? '').trim().toLowerCase();
		return item.title.trim().toLowerCase() === seedTitle;
	}

	const feed = new RelatedFeed(isCurrentScene);
	let openingUrl = $state<string | null>(null);

	// Reset and kick off the feed whenever the viewed scene changes.
	$effect(() => {
		if (!data.id) return;
		untrack(() => {
			feed.reset();
			feed.loadMore();
		});
	});

	// Related scenes arrive over a subscription: the stored list right away, then
	// again as the visit links the site's list and fallback-search hits.
	$effect(() => {
		const id = data.id;
		if (!id) return;
		return subscribeRelated(id, (items) => feed.setRelated(items));
	});

	// Adapts the rail feed to the player's fullscreen up-next row.
	let playerRecs = $derived(feed.items.map(toPlayerRec));

	/** One up-next card for the fullscreen player. */
	function toPlayerRec(item: PluginSearchResult) {
		const itemSite = pluginForUrl(item.sourceUrl, searchPlugins);
		let siteName: string | null = null;
		if (itemSite) siteName = itemSite.displayName || itemSite.name;
		return {
			id: item.externalId,
			title: item.title,
			image: cacheUrl(item.posterUrl),
			durationSeconds: item.durationSeconds,
			siteName,
			siteIconUrl: itemSite?.iconUrl
		};
	}

	/** Opens the rail item picked from the fullscreen player's grid. */
	function onSelectPlayerRec(rec: { id: string }) {
		const original = feed.items.find((item) => item.externalId === rec.id);
		if (original) openRecommendation(original);
	}

	/** Navigates to a feed item, ingesting fresh plugin results first. */
	async function openRecommendation(item: PluginSearchResult) {
		if (item.sceneId) {
			goto(sceneUrl(item.sceneId));
			return;
		}
		if (openingUrl !== null) return;
		openingUrl = item.sourceUrl;
		try {
			const mediaId = await ingestPluginResult(item.plugin, item.sourceUrl, item.posterUrl);
			goto(sceneUrl(mediaId));
		} catch {
			openingUrl = null;
			notifications.push('Could not open this title', 'error');
		}
	}

	// --- TV / D-pad ----------------------------------------------------------
	let actionsRef = $state<{ focusFirst: () => void } | undefined>();
	let railRef = $state<{ focusFirst: () => void } | undefined>();

	/** Right off the action pills enters the rail; left opens the sidebar. */
	function onActionsEdge(direction: 'left' | 'right') {
		if (direction === 'right') {
			railRef?.focusFirst();
			return;
		}
		sidebarFocus.openSidebar();
	}

	/** Left off the rail returns to the action pills, releasing the rail's arrow keys. */
	function onRailEdge(direction: Direction) {
		if (direction !== 'left') return;
		activateFocusZone(Symbol('scene-actions'));
		actionsRef?.focusFirst();
	}
</script>

{#if !scene}
	<div class="flex h-64 items-center justify-center">
		<span class="text-base-content/40">Title not found</span>
	</div>
{:else if $isCompact}
	<div class="flex flex-col gap-4">
		<!-- The player stays in view while scrolling, just under the top bar
		     (or the status bar once the bar hides). -->
		<div class="sticky-player sticky z-20 -mx-3" data-swipe-ignore>
			{@render player(scene)}
		</div>

		<SceneHeaderCompact {scene} {site} />

		<ScenePills
			{scene}
			posterUrl={previewFrame}
			{streamGroups}
			{pluginIcons}
			queueingId={queueing}
			{activeSourceId}
			resolvingId={resolving}
			{alikeCount}
			ondownload={download}
			onplay={play}
			onfindsources={openAlike}
			onblock={blockEntity}
		/>

		{#if detailPending}
			<SceneSkeleton variant="description" />
		{:else}
			<SceneAboutCompact
				{scene}
				{markers}
				onjump={jumpToMarker}
				onremovemarker={removeMarker}
				onblocktag={blockTag}
			/>
		{/if}

		<SceneUpNextFeed
			{feed}
			sceneId={scene.id}
			performers={scene.performers}
			{openingUrl}
			onopen={openRecommendation}
		/>
	</div>
{:else}
	<div class="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(300px,400px)]">
		<div class="flex min-w-0 flex-col gap-3">
			{@render player(scene)}

			<h1 class="text-xl leading-snug font-bold sm:text-2xl">{scene.title}</h1>

			<div class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3">
				{#if detailPending}
					<SceneSkeleton variant="channel" />
				{:else}
					<ChannelRow studio={scene.studio} {site} />
				{/if}
				<SceneActions
					bind:this={actionsRef}
					sceneId={scene.id}
					sceneTitle={scene.title}
					{streamGroups}
					{pluginIcons}
					queueingId={queueing}
					{alikeCount}
					studio={scene.studio}
					performers={scene.performers}
					ondownload={download}
					onfindsources={openAlike}
					onblock={blockEntity}
					onedge={onActionsEdge}
				/>
			</div>

			{#if detailPending}
				<SceneSkeleton variant="description" />
			{:else}
				<SceneDescription
					{scene}
					{site}
					{streamGroups}
					{pluginIcons}
					sourcesPending={ensuring}
					{activeSourceId}
					resolvingId={resolving}
					{markers}
					onplay={play}
					onjump={jumpToMarker}
					onremovemarker={removeMarker}
					onblocktag={blockTag}
				/>
			{/if}
		</div>

		<SceneRail
			bind:this={railRef}
			{feed}
			{openingUrl}
			onopen={openRecommendation}
			onedge={onRailEdge}
		/>
	</div>
{/if}

{#if scene}
	<AlikeSourcesDialog
		open={alikeOpen}
		sceneId={scene.id}
		sceneTitle={scene.title}
		candidates={alikeCandidates}
		loading={alikeLoading}
		onClose={() => (alikeOpen = false)}
		onAttached={onAlikeAttached}
	/>
{/if}

{#snippet player(current: SceneDetail)}
	<!-- Phones: edge to edge with square corners. -->
	<div class="overflow-hidden bg-black" class:rounded-xl={!$isCompact}>
		{#if activeStream}
			<VideoPlayer
				bind:this={playerRef}
				src={activeStream.src}
				mimeType={activeStream.mimeType}
				title={activeStream.title}
				sources={playerSources}
				activeSourceId={activeStream.sourceId}
				onselectsource={(source) => selectPlayerSource(source.id)}
				onresolvesource={resolvePlayerSource}
				mediaId={data.id}
				{startPosition}
				recommendations={playerRecs}
				onselectrec={onSelectPlayerRec}
				onloadmore={() => feed.loadMore()}
				markers={playerMarkers}
				onaddmarker={addMarker}
				inline
				square={$isCompact}
				poster={previewFrame}
				onswipedown={compactOnly(leavePage)}
				onback={compactOnly(leavePage)}
				shared
				onclose={() => (activeStream = null)}
			/>
		{:else}
			<PlayerPoster
				imageUrl={previewFrame}
				title={current.title}
				durationLabel={formatClock(current.durationSeconds)}
				onback={compactOnly(leavePage)}
				canPlay={best !== null}
				loading={best !== null && resolving === best.id}
				{ensuring}
				onplay={() => {
					if (best) play(best);
				}}
			/>
		{/if}
	</div>
{/snippet}

<style>
	/* Under the top bar while it shows, under the status bar once it hides. A
	   shadow in the page colour covers the strip above it, so cards scrolling
	   past don't show through there. */
	.sticky-player {
		top: max(var(--chrome-offset), var(--safe-area-inset-top, env(safe-area-inset-top, 0px)));
		background-color: var(--color-base-100);
		box-shadow: 0 -4rem 0 var(--color-base-100);
	}
</style>
