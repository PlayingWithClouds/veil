<!--
	Home without a search: the top recommendation featured, then Continue
	watching, Live now, For you (three rows of six), the user's performers and
	the sites. Recommendations follow the library filter; while one is set,
	only the featured video and For you show. Videos being watched are never
	recommended here: they have their own row. Each row loads on its own and
	stays hidden while empty.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import type { FilterInput } from '@playingwithclouds/veil-sdk/generated';
	import Shelf from '$lib/components/Shelf.svelte';
	import FeaturedVideo from '$lib/components/FeaturedVideo.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import RecordingCard from '$lib/components/RecordingCard.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import SiteIcon from '$lib/components/scene/SiteIcon.svelte';
	import {
		openVideo,
		fetchPerformers,
		type VideoCardRow,
		type VideoDetail,
		type MemberRow
	} from '$lib/library';
	import {
		fetchContinueWatching,
		fetchRecommendedVideos,
		fetchFavoritePerformers,
		type RecommendedVideo
	} from '$lib/home';
	import {
		fetchLiveCams,
		fetchActiveRecordings,
		type CamRow,
		type ActiveRecordingRow
	} from '$lib/cams';
	import { listers } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { videoUrl, camUrl } from '$lib/routes';

	interface Props {
		// The library filter from the URL; recommendations follow it.
		filter: FilterInput;
		// Whether filter narrows anything: then only the recommendations show.
		filtered: boolean;
	}

	let { filter, filtered }: Props = $props();

	// Card widths per row, in px.
	const VIDEO_WIDTH = 280;
	const LIVE_WIDTH = 220;
	const PERFORMER_WIDTH = 140;
	const SITE_WIDTH = 150;
	// For you fills three rows of six under the featured video.
	const FOR_YOU_COUNT = 18;
	// Asked for beyond what's shown, to make up for videos being watched.
	const RECOMMENDATION_SLACK = 20;

	/** One card of the Live row: a recording running now, or a live room. */
	type LiveItem =
		{ recording: ActiveRecordingRow; cam?: never } | { cam: CamRow; recording?: never };

	let continueWatching = $state<VideoCardRow[]>([]);
	let recommended = $state<RecommendedVideo[]>([]);
	let featuredDetail = $state<VideoDetail | null>(null);
	let liveCams = $state<CamRow[]>([]);
	let recordings = $state<ActiveRecordingRow[]>([]);
	let performers = $state<MemberRow[]>([]);

	// Serialised so recommendations reload only when the filter itself changes.
	let filterKey = $derived(JSON.stringify(filter));
	let freshPicks = $derived.by(() => {
		const watching = new Set(continueWatching.map((video) => video.id));
		return recommended.filter((item) => !watching.has(item.video.id));
	});
	let featured = $derived(freshPicks[0] ?? null);
	let forYou = $derived(freshPicks.slice(1, FOR_YOU_COUNT + 1));
	let live = $derived<LiveItem[]>([
		...recordings.map((recording) => ({ recording })),
		...liveCams.map((cam) => ({ cam }))
	]);
	let siteTiles = $derived(listers($sites, 'VIDEO'));

	$effect(() => {
		load(fetchContinueWatching(), (found) => (continueWatching = found));
		load(fetchLiveCams(), (found) => (liveCams = found));
		load(fetchActiveRecordings(), (found) => (recordings = found));
		load(loadPerformers(), (found) => (performers = found));
	});

	$effect(() => {
		const current: FilterInput = JSON.parse(filterKey);
		load(fetchRecommendedVideos(FOR_YOU_COUNT + 1 + RECOMMENDATION_SLACK, current), (found) => {
			recommended = found;
		});
	});

	// Reads the featured video's page up front, so its credits, tags and description show.
	$effect(() => {
		const id = featured?.video.id;
		featuredDetail = null;
		if (!id) return;
		load(openVideo(id), (found) => {
			if (found?.id === id) featuredDetail = found;
		});
	});

	/** The user's favourite performers, else the library's first ones. */
	async function loadPerformers(): Promise<MemberRow[]> {
		const favorites = await fetchFavoritePerformers();
		if (favorites.length > 0) return favorites;
		return fetchPerformers('', 20);
	}

	/** Fills a row when its data arrives; a row that fails stays hidden and says why in the console. */
	function load<T>(request: Promise<T>, fill: (found: T) => void) {
		request.then(fill, (failure) => console.warn('home row failed to load', failure));
	}

	/** Opens a video's page. */
	function openPage(video: VideoCardRow) {
		goto(videoUrl(video.id));
	}

	/** Opens the room a recording is made from. */
	function openRecording(recording: ActiveRecordingRow) {
		if (!recording.camModel) return;
		goto(camUrl(recording.camModel.id));
	}

	/** The key of a Live row card. */
	function liveKey(item: LiveItem): string {
		if (item.recording)
			return `recording:${item.recording.camModel?.id ?? item.recording.startedAt}`;
		return `cam:${item.cam.id}`;
	}
</script>

<div class="flex flex-col gap-10">
	{#if featured}
		<FeaturedVideo
			video={featured.video}
			detail={featuredDetail}
			reason={featured.reason}
			shown={{ surface: 'home', position: 0 }}
			onopen={() => openPage(featured.video)}
		/>
	{/if}

	{#if !filtered}
		<Shelf
			title="Continue watching"
			items={continueWatching}
			key={(video) => video.id}
			itemWidth={VIDEO_WIDTH}
		>
			{#snippet card(video, position)}
				<SceneCard
					{video}
					shown={{ surface: 'continue', position }}
					onclick={() => openPage(video)}
				/>
			{/snippet}
		</Shelf>

		<Shelf title="Live now" items={live} key={liveKey} itemWidth={LIVE_WIDTH}>
			{#snippet card(item)}
				{#if item.recording}
					<RecordingCard recording={item.recording} onclick={() => openRecording(item.recording)} />
				{:else}
					<CamCard cam={item.cam} onclick={() => goto(camUrl(item.cam.id))} />
				{/if}
			{/snippet}
		</Shelf>
	{/if}

	{#if forYou.length > 0}
		<section class="flex flex-col gap-4">
			<h2 class="text-lg font-semibold">For you</h2>
			<!-- The engine's own picks, so their surface is "home": the one it learns from.
			     Positions continue after the featured video's 0. -->
			<div
				class="grid grid-cols-6 gap-x-4 gap-y-8 max-xl:grid-cols-4 max-lg:grid-cols-3 max-md:grid-cols-1"
			>
				{#each forYou as item, index (item.video.id)}
					<SceneCard
						video={item.video}
						shown={{ surface: 'home', position: index + 1 }}
						onclick={() => openPage(item.video)}
					/>
				{/each}
			</div>
		</section>
	{/if}

	{#if !filtered}
		<Shelf
			title="Your performers"
			items={performers}
			key={(performer) => performer.id}
			itemWidth={PERFORMER_WIDTH}
			href="/performers"
		>
			{#snippet card(performer)}
				<MemberCard kind="performer" member={performer} />
			{/snippet}
		</Shelf>

		<Shelf title="Sites" items={siteTiles} key={(site) => site.name} itemWidth={SITE_WIDTH}>
			{#snippet card(site)}
				<a
					href="/?{new URLSearchParams({ site: site.name })}"
					class="bg-base-200 hover:bg-base-300 flex flex-col items-center gap-3 rounded-xl px-3 py-5 transition-colors"
				>
					<SiteIcon iconUrl={site.iconUrl} class="h-12 w-12 rounded-lg" />
					<span class="w-full truncate text-center text-sm font-medium">
						{site.displayName ?? site.name}
					</span>
				</a>
			{/snippet}
		</Shelf>
	{/if}
</div>
