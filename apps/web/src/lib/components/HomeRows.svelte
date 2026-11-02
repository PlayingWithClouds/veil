<!--
	Home without a search: one video in focus (the one to resume, else the top
	recommendation) with "Up next" beside it, then a row per thing to watch,
	personal first, live next, then every kind of media. Each row loads on its
	own and stays hidden while empty.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import Shelf from '$lib/components/Shelf.svelte';
	import HomeFocus, { type FocusItem } from '$lib/components/HomeFocus.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import GalleryCard from '$lib/components/GalleryCard.svelte';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import StoryCard from '$lib/components/StoryCard.svelte';
	import PostCard from '$lib/components/PostCard.svelte';
	import GroupCard from '$lib/components/GroupCard.svelte';
	import CategoryCard from '$lib/components/CategoryCard.svelte';
	import CollectionCard from '$lib/components/CollectionCard.svelte';
	import CamCard from '$lib/components/CamCard.svelte';
	import RecordingCard from '$lib/components/RecordingCard.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import { fetchGalleries, type VideoCardRow, type GalleryCardRow, type MemberRow } from '$lib/library';
	import {
		fetchContinueWatching,
		fetchWatchAgain,
		fetchRecommendedVideos,
		fetchFollowedVideos,
		fetchFavoritePerformers,
		type RecommendedVideo
	} from '$lib/home';
	import {
		fetchAudios,
		fetchStories,
		fetchPosts,
		fetchGroups,
		type AudioCardRow,
		type StoryCardRow,
		type PostCardRow,
		type GroupCardRow
	} from '$lib/media';
	import {
		fetchCategories,
		fetchCollections,
		fetchWatchlist,
		type CategoryCardRow,
		type CollectionCardRow
	} from '$lib/curation';
	import {
		fetchLiveCams,
		fetchActiveRecordings,
		fetchRecordedVideos,
		type CamRow,
		type ActiveRecordingRow
	} from '$lib/cams';
	import {
		videoUrl,
		galleryUrl,
		audioUrl,
		storyUrl,
		postUrl,
		groupUrl,
		categoryUrl,
		collectionUrl,
		camUrl
	} from '$lib/routes';

	// Card widths per row, in px.
	const VIDEO_WIDTH = 280;
	const LIVE_WIDTH = 220;
	const PERFORMER_WIDTH = 140;
	const WIDE_TILE_WIDTH = 280;
	const COLLECTION_WIDTH = 240;
	const POSTER_WIDTH = 160;
	const GALLERY_WIDTH = 180;
	const SQUARE_WIDTH = 200;
	const UP_NEXT_COUNT = 4;

	/** One card of the Live row: a recording running now, or a live room. */
	type LiveItem = { recording: ActiveRecordingRow; cam?: never } | { cam: CamRow; recording?: never };

	let continueWatching = $state<VideoCardRow[]>([]);
	let watchlist = $state<VideoCardRow[]>([]);
	let watchAgain = $state<VideoCardRow[]>([]);
	let recommended = $state<RecommendedVideo[]>([]);
	let followed = $state<VideoCardRow[]>([]);
	let liveCams = $state<CamRow[]>([]);
	let recordings = $state<ActiveRecordingRow[]>([]);
	let recordedVideos = $state<VideoCardRow[]>([]);
	let performers = $state<MemberRow[]>([]);
	let categories = $state<CategoryCardRow[]>([]);
	let collections = $state<CollectionCardRow[]>([]);
	let groups = $state<GroupCardRow[]>([]);
	let galleries = $state<GalleryCardRow[]>([]);
	let audios = $state<AudioCardRow[]>([]);
	let stories = $state<StoryCardRow[]>([]);
	let posts = $state<PostCardRow[]>([]);

	// Resume when something is unfinished, else the top recommendation.
	let focus = $derived.by<FocusItem | null>(() => {
		if (continueWatching.length > 0) return { video: continueWatching[0], label: 'Continue' };
		if (recommended.length > 0) return { video: recommended[0].video, label: 'For you' };
		return null;
	});
	let focusReason = $derived.by(() => {
		if (!focus || focus.label !== 'For you') return null;
		return recommended[0]?.reason ?? null;
	});
	let upNext = $derived.by(() => {
		const candidates: FocusItem[] = [
			...continueWatching.map((video) => ({ video, label: 'Continue' })),
			...watchlist.map((video) => ({ video, label: 'Watchlist' })),
			...recommended.map((item) => ({ video: item.video, label: 'For you' }))
		];
		const seen = new Set([focus?.video.id]);
		return candidates
			.filter((item) => {
				if (seen.has(item.video.id)) return false;
				seen.add(item.video.id);
				return true;
			})
			.slice(0, UP_NEXT_COUNT);
	});
	let live = $derived<LiveItem[]>([
		...recordings.map((recording) => ({ recording })),
		...liveCams.map((cam) => ({ cam }))
	]);

	$effect(() => {
		load(fetchContinueWatching(), (found) => (continueWatching = found));
		load(fetchWatchlist(), (found) => (watchlist = found));
		load(fetchWatchAgain(), (found) => (watchAgain = found));
		load(fetchRecommendedVideos(), (found) => (recommended = found));
		load(fetchFollowedVideos(), (found) => (followed = found));
		load(fetchLiveCams(), (found) => (liveCams = found));
		load(fetchActiveRecordings(), (found) => (recordings = found));
		load(fetchRecordedVideos(), (found) => (recordedVideos = found));
		load(fetchFavoritePerformers(), (found) => (performers = found));
		load(fetchCategories(), (found) => (categories = found));
		load(fetchCollections(), (found) => (collections = found));
		load(fetchGroups(), (found) => (groups = found));
		load(fetchGalleries({}, 20), (found) => (galleries = found));
		load(fetchAudios(), (found) => (audios = found));
		load(fetchStories(), (found) => (stories = found));
		load(fetchPosts(), (found) => (posts = found));
	});

	/** Fills a row when its data arrives; a row that fails stays hidden and says why in the console. */
	function load<T>(request: Promise<T>, fill: (found: T) => void) {
		request.then(fill, (failure) => console.warn('home row failed to load', failure));
	}

	/** Opens a video's page. */
	function openVideo(video: VideoCardRow) {
		goto(videoUrl(video.id));
	}

	/** Opens the room a recording is made from. */
	function openRecording(recording: ActiveRecordingRow) {
		if (!recording.camModel) return;
		goto(camUrl(recording.camModel.id));
	}

	/** The key of a Live row card. */
	function liveKey(item: LiveItem): string {
		if (item.recording) return `recording:${item.recording.camModel?.id ?? item.recording.startedAt}`;
		return `cam:${item.cam.id}`;
	}
</script>

<div class="flex flex-col gap-10">
	{#if focus}
		<HomeFocus {focus} reason={focusReason} {upNext} onopen={openVideo} />
	{/if}

	<Shelf
		title="Continue watching"
		items={continueWatching}
		key={(video) => video.id}
		itemWidth={VIDEO_WIDTH}
	>
		{#snippet card(video, position)}
			<SceneCard {video} shown={{ surface: 'continue', position }} onclick={() => openVideo(video)} />
		{/snippet}
	</Shelf>

	<Shelf title="Watchlist" items={watchlist} key={(video) => video.id} itemWidth={VIDEO_WIDTH}>
		{#snippet card(video, position)}
			<SceneCard {video} shown={{ surface: 'watchlist', position }} onclick={() => openVideo(video)} />
		{/snippet}
	</Shelf>

	<!-- The engine's own pick, so its surface is "home": the one it learns from. -->
	<Shelf title="For you" items={recommended} key={(item) => item.video.id} itemWidth={VIDEO_WIDTH}>
		{#snippet card(item, position)}
			<SceneCard
				video={item.video}
				shown={{ surface: 'home', position }}
				onclick={() => openVideo(item.video)}
			/>
		{/snippet}
	</Shelf>

	<Shelf title="Watch again" items={watchAgain} key={(video) => video.id} itemWidth={VIDEO_WIDTH}>
		{#snippet card(video, position)}
			<SceneCard {video} shown={{ surface: 'again', position }} onclick={() => openVideo(video)} />
		{/snippet}
	</Shelf>

	<Shelf title="New from follows" items={followed} key={(video) => video.id} itemWidth={VIDEO_WIDTH}>
		{#snippet card(video, position)}
			<SceneCard {video} shown={{ surface: 'follows', position }} onclick={() => openVideo(video)} />
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

	<Shelf title="Recordings" items={recordedVideos} key={(video) => video.id} itemWidth={VIDEO_WIDTH}>
		{#snippet card(video)}
			<SceneCard {video} onclick={() => openVideo(video)} />
		{/snippet}
	</Shelf>

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

	<Shelf
		title="Categories"
		items={categories}
		key={(category) => category.id}
		itemWidth={WIDE_TILE_WIDTH}
	>
		{#snippet card(category)}
			<CategoryCard {category} onclick={() => goto(categoryUrl(category.id))} />
		{/snippet}
	</Shelf>

	<Shelf
		title="Collections"
		items={collections}
		key={(collection) => collection.id}
		itemWidth={COLLECTION_WIDTH}
	>
		{#snippet card(collection)}
			<CollectionCard {collection} onclick={() => goto(collectionUrl(collection.id))} />
		{/snippet}
	</Shelf>

	<Shelf title="Movies & series" items={groups} key={(group) => group.id} itemWidth={POSTER_WIDTH}>
		{#snippet card(group)}
			<GroupCard {group} onclick={() => goto(groupUrl(group.id))} />
		{/snippet}
	</Shelf>

	<Shelf
		title="Galleries"
		items={galleries}
		key={(gallery) => gallery.id}
		itemWidth={GALLERY_WIDTH}
		href="/galleries"
	>
		{#snippet card(gallery, position)}
			<GalleryCard
				{gallery}
				shown={{ surface: 'galleries', position }}
				onclick={() => goto(galleryUrl(gallery.id))}
			/>
		{/snippet}
	</Shelf>

	<Shelf title="Audio" items={audios} key={(audio) => audio.id} itemWidth={SQUARE_WIDTH}>
		{#snippet card(audio)}
			<AudioCard {audio} onclick={() => goto(audioUrl(audio.id))} />
		{/snippet}
	</Shelf>

	<Shelf title="Stories" items={stories} key={(story) => story.id} itemWidth={POSTER_WIDTH}>
		{#snippet card(story)}
			<StoryCard {story} onclick={() => goto(storyUrl(story.id))} />
		{/snippet}
	</Shelf>

	<Shelf title="Creator posts" items={posts} key={(post) => post.id} itemWidth={SQUARE_WIDTH}>
		{#snippet card(post)}
			<PostCard {post} onclick={() => goto(postUrl(post.id))} />
		{/snippet}
	</Shelf>
</div>
