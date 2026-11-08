<!--
	What the user can do with a video, as a row of glass buttons under its
	title: like or dislike (which hides it from recommendations, with Undo),
	favourite, the O counter (an O marks the playing moment), a star rating, the watchlist, download for keeps, and a ⋮ menu
	(not interested, blocks, open on the site).
-->
<script lang="ts">
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import DropIcon from 'phosphor-svelte/lib/DropIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import ThumbsDownIcon from 'phosphor-svelte/lib/ThumbsDownIcon';
	import ThumbsUpIcon from 'phosphor-svelte/lib/ThumbsUpIcon';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import RatingStars from '$lib/components/RatingStars.svelte';
	import { addToWatchlist } from '$lib/curation';
	import { flushEvents, logEvent } from '$lib/events';
	import { cacheUrl } from '$lib/img';
	import { downloadVideo, type VideoDetail } from '$lib/library';
	import { recordActions } from '$lib/recordActions';
	import { showUndo } from '$lib/stores/undoToast';
	import { logNotInterested, setFavorite, setHidden, setRating } from '$lib/userState';

	interface Props {
		video: VideoDetail;
		// Where playback is now; an O logged then marks the video's best moment.
		currentSeconds?: number | null;
	}

	let { video, currentSeconds = null }: Props = $props();

	// The user's marks as changed here; the video row is the page's.
	let favorite = $state(false);
	let rating = $state<number | null>(null);
	let verdict = $state<'liked' | 'disliked' | null>(null);
	let onWatchlist = $state(false);
	let downloading = $state(false);
	// Os logged here on top of the stored count.
	let addedOs = $state(0);
	let menuAnchor = $state<HTMLElement | null>(null);

	let downloaded = $derived(video.files.some((file) => !file.missing));
	let menuActions = $derived(
		recordActions(
			{
				kind: 'video',
				id: video.id,
				title: video.title,
				favorite,
				performers: video.performers,
				studio: video.studio,
				sourceUrl: video.sources[0]?.url ?? null
			},
			(value) => (favorite = value)
		)
	);

	// A new video starts from its own stored marks.
	$effect(() => {
		favorite = video.userState.favorite;
		rating = video.userState.rating;
		verdict = null;
		onWatchlist = false;
		downloading = false;
		addedOs = 0;
	});

	/** Counts an O, at the playing moment when something plays. */
	function logO() {
		addedOs += 1;
		const data: Record<string, unknown> = {};
		if (currentSeconds !== null) data.position = currentSeconds;
		logEvent({ kind: 'o', subject: { kind: 'video', id: video.id }, data });
		void flushEvents();
	}

	/** Thumbs up: the engine recommends more like it. */
	function like() {
		if (verdict === 'liked') return;
		verdict = 'liked';
		logEvent({ kind: 'like', subject: { kind: 'video', id: video.id }, data: { value: 1 } });
	}

	/** Thumbs down: hidden from recommendations for good once Undo has passed. */
	function dislike() {
		if (verdict === 'disliked') return;
		const videoId = video.id;
		verdict = 'disliked';
		setHidden('video', videoId, true);
		showUndo({
			message: 'You won’t be recommended this video again.',
			undo: () => {
				setHidden('video', videoId, false);
				verdict = null;
			},
			settle: () => logNotInterested('video', videoId)
		});
	}

	/** Flips the favourite mark. */
	async function toggleFavorite() {
		const next = !favorite;
		favorite = next;
		await setFavorite('video', video.id, next);
	}

	/** Rates the video; null clears the rating. */
	async function rate(value: number | null) {
		rating = value;
		await setRating('video', video.id, value);
	}

	/** Puts the video on the watchlist. */
	async function saveForLater() {
		onWatchlist = true;
		await addToWatchlist(video.id);
	}

	/** Downloads the video's best stream in the background. */
	async function download() {
		downloading = true;
		try {
			await downloadVideo(video.id);
			showUndo({ message: 'Downloading. It plays from your library once it’s done.' });
		} catch (failure) {
			downloading = false;
			showUndo({ message: `Couldn't download: ${failure}` });
		}
	}
</script>

<div class="flex flex-wrap items-center gap-2">
	<div class="glass-pill flex h-10 items-center rounded-full">
		<button
			type="button"
			class="flex h-full items-center gap-2 rounded-l-full pr-3 pl-4 text-sm font-medium transition-colors hover:bg-white/8"
			aria-pressed={verdict === 'liked'}
			title="More like this"
			onclick={like}
		>
			<ThumbsUpIcon size={18} weight={verdict === 'liked' ? 'fill' : 'regular'} />
			Like
		</button>
		<span class="h-5 w-px bg-white/15"></span>
		<button
			type="button"
			class="flex h-full items-center rounded-r-full pr-4 pl-3 transition-colors hover:bg-white/8"
			aria-pressed={verdict === 'disliked'}
			aria-label="Not for me"
			title="Not for me"
			onclick={dislike}
		>
			<ThumbsDownIcon size={18} weight={verdict === 'disliked' ? 'fill' : 'regular'} />
		</button>
	</div>

	<button
		type="button"
		class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
		aria-pressed={favorite}
		onclick={toggleFavorite}
	>
		<span class="flex" class:text-error={favorite}>
			<HeartIcon size={18} weight={favorite ? 'fill' : 'regular'} />
		</span>
		Favourite
	</button>

	<button
		type="button"
		class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-90"
		title="Count an O at this moment"
		onclick={logO}
	>
		<DropIcon size={18} weight={video.oCount + addedOs > 0 ? 'fill' : 'regular'} />
		O
		{#if video.oCount + addedOs > 0}
			<span class="tabular-nums opacity-60">{video.oCount + addedOs}</span>
		{/if}
	</button>

	<div class="glass-pill flex h-10 items-center rounded-full px-3">
		<RatingStars {rating} onrate={(value) => void rate(value)} />
	</div>

	<button
		type="button"
		class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-60"
		disabled={onWatchlist}
		onclick={saveForLater}
	>
		{#if onWatchlist}
			<CheckIcon size={18} /> On watchlist
		{:else}
			<PlusIcon size={18} /> Watchlist
		{/if}
	</button>

	<button
		type="button"
		class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95 disabled:opacity-60"
		disabled={downloaded || downloading}
		onclick={download}
	>
		{#if downloaded}
			<CheckIcon size={18} /> Downloaded
		{:else if downloading}
			<span class="loading loading-spinner loading-xs"></span> Downloading
		{:else}
			<DownloadSimpleIcon size={18} /> Download
		{/if}
	</button>

	<button
		type="button"
		class="glass-round flex h-10 w-10 items-center justify-center rounded-full"
		aria-label="More"
		onclick={(event) => (menuAnchor = event.currentTarget)}
	>
		<DotsThreeVerticalIcon size={20} weight="bold" />
	</button>
</div>

{#if menuAnchor}
	<CardMenu
		anchor={menuAnchor}
		title={video.title}
		subtitle={video.studio?.name}
		imageUrl={cacheUrl(video.cover?.uri)}
		actions={menuActions}
		onclose={() => (menuAnchor = null)}
	/>
{/if}
