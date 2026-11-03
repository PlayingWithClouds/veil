<!--
	Who's in a video, as large cards: the credited performers (✕ takes one
	off, as "Not her": when a site profile of theirs credited them, that
	profile moves away), the suggested ones (named in the title, …) to add
	with a click, and a card to add anyone else, or someone new by name. Changes are the user's: the
	sites leave the credits as set.
-->
<script lang="ts">
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import RecordPickerDialog from '$lib/components/RecordPickerDialog.svelte';
	import {
		createPerformer,
		fetchPerformers,
		fetchVideoDetail,
		setVideoPerformers,
		type MemberRow,
		type VideoDetail
	} from '$lib/library';
	import { readPerformer, unassignPerformer } from '$lib/performers';
	import { showUndo } from '$lib/stores/undoToast';
	import { cacheUrl } from '$lib/img';
	import { performerUrl } from '$lib/routes';
	import {
		FACE_MODEL,
		faceTracking,
		fetchFaceMatches,
		learnFaces,
		type FaceTracking,
		type VideoFaceGuesses
	} from '$lib/faces';

	/** A performer as a suggestion shows them. */
	interface SuggestedPerformer {
		id: string;
		name: string;
		image: { uri: string } | null;
	}

	/** Someone probably in the video, and why. */
	interface Suggestion {
		performer: SuggestedPerformer;
		reason: string;
	}

	interface Props {
		video: VideoDetail;
		// Who the face scan recognised; null before a scan.
		faces?: VideoFaceGuesses | null;
		onchange: (video: VideoDetail) => void;
	}

	let { video, faces = null, onchange }: Props = $props();

	// What the face scan says it's doing, by phase.
	const FACE_PHASES: Record<FaceTracking['phase'], string> = {
		pictures: 'Learning the performers’ faces',
		stills: 'Looking for faces in the stills',
		playback: 'Looking for faces as the video plays',
		waiting: 'Waiting for the video to pause',
		frames: 'Looking for faces in the video',
		saving: 'Saving the faces'
	};

	let adding = $state(false);
	let busy = $state(false);
	// Suggestions the user waved away on this visit.
	let dismissed = $state<string[]>([]);

	let suggestions = $derived(
		suggestionsOf(video, faces).filter((suggestion) => !dismissed.includes(suggestion.performer.id))
	);

	/**
	 * The title's suggestions, then the faces recognised but not credited,
	 * each performer once.
	 */
	function suggestionsOf(shown: VideoDetail, found: VideoFaceGuesses | null): Suggestion[] {
		const credited = new Set(shown.performers.map((performer) => performer.id));
		const all: Suggestion[] = [...shown.suggestedPerformers];
		for (const guess of found?.guesses ?? []) {
			if (guess.credited || credited.has(guess.performer.id)) continue;
			all.push({
				performer: guess.performer,
				reason: `Face match ${Math.round(guess.similarity * 100)}%`
			});
		}
		return all.filter(
			(suggestion, index) =>
				all.findIndex((other) => other.performer.id === suggestion.performer.id) === index
		);
	}

	/** Credits a performer on the video, then learns their face from it. */
	async function add(performer: { id: string; name: string }) {
		adding = false;
		if (video.performers.some((credited) => credited.id === performer.id)) return;
		busy = true;
		try {
			const ids = [...video.performers.map((credited) => credited.id), performer.id];
			const saved = await setVideoPerformers(video.id, ids);
			if (!saved) return;
			onchange(saved);
			void suggestElsewhere(saved.id, performer);
		} finally {
			busy = false;
		}
	}

	/**
	 * Learns the video's credited faces again and, when that taught the app
	 * a face, says how many other videos probably show the performer.
	 */
	async function suggestElsewhere(videoId: string, performer: { id: string; name: string }) {
		const learned = await learnFaces(videoId, FACE_MODEL);
		if (learned === 0) return;
		const matches = await fetchFaceMatches(performer.id);
		if (matches.length === 0) return;
		let count = `${matches.length} other videos`;
		if (matches.length === 1) count = 'one other video';
		showUndo({
			message: `${performer.name} is probably also in ${count}`,
			link: { label: 'Show', href: `${performerUrl(performer.id)}#face-matches` }
		});
	}

	/**
	 * Creates a performer by name and credits them, then has the sites look
	 * for their profile in the background; the video shows what they found.
	 */
	async function create(name: string) {
		adding = false;
		const videoId = video.id;
		const performer = await createPerformer(name);
		await add(performer);
		await readPerformer(performer.id);
		// The page may show another video by now.
		if (video.id !== videoId) return;
		const saved = await fetchVideoDetail(videoId);
		if (saved) onchange(saved);
	}

	/** Takes a performer off the video; a profile that credited them moves away. */
	async function remove(performer: MemberRow) {
		busy = true;
		try {
			const other = await unassignPerformer(video.id, performer.id);
			const saved = await fetchVideoDetail(video.id);
			if (saved) onchange(saved);
			if (other) {
				showUndo({
					message: `A site's ${performer.name} was someone else: that profile and its videos moved to a new performer`
				});
			}
		} finally {
			busy = false;
		}
	}
</script>

{#snippet portrait(member: SuggestedPerformer)}
	<span
		class="bg-base-200 flex aspect-[3/4] w-full items-center justify-center overflow-hidden rounded-xl"
	>
		{#if member.image}
			<img
				src={cacheUrl(member.image.uri)}
				alt=""
				loading="lazy"
				class="h-full w-full object-cover"
			/>
		{:else}
			<UserIcon size={28} class="opacity-20" />
		{/if}
	</span>
{/snippet}

<section class="flex flex-col gap-2" aria-busy={busy}>
	<div class="flex flex-wrap items-start gap-3">
		{#each video.performers as performer (performer.id)}
			<div class="group relative flex w-28 flex-col gap-1.5">
				<a href={performerUrl(performer.id)} class="flex flex-col gap-1.5">
					{@render portrait(performer)}
					<span class="line-clamp-1 text-sm font-medium">{performer.name}</span>
					<span class="text-base-content/50 -mt-1 text-xs">{performer.videoCount} videos</span>
				</a>
				<button
					type="button"
					class="bg-base-100/80 hover:bg-error hover:text-error-content absolute top-1.5 right-1.5 flex h-7 w-7 items-center justify-center rounded-full opacity-0 transition-opacity group-hover:opacity-100 focus:opacity-100"
					aria-label="Not {performer.name}"
					title="Not {performer.name}"
					disabled={busy}
					onclick={() => remove(performer)}
				>
					<XIcon size={14} />
				</button>
			</div>
		{/each}

		{#each suggestions as suggestion (suggestion.performer.id)}
			<div class="flex w-28 flex-col gap-1.5 opacity-80">
				<a href={performerUrl(suggestion.performer.id)} class="relative flex flex-col gap-1.5">
					<span class="border-base-content/30 rounded-xl border-2 border-dashed">
						{@render portrait(suggestion.performer)}
					</span>
					<span class="line-clamp-1 text-sm font-medium">{suggestion.performer.name}</span>
				</a>
				<span class="text-base-content/50 -mt-1 text-xs">Suggested · {suggestion.reason}</span>
				<span class="flex gap-1">
					<button
						type="button"
						class="btn btn-xs flex-1"
						disabled={busy}
						onclick={() => add(suggestion.performer)}
					>
						Add
					</button>
					<button
						type="button"
						class="btn btn-ghost btn-xs"
						aria-label="Not {suggestion.performer.name}"
						onclick={() => (dismissed = [...dismissed, suggestion.performer.id])}
					>
						<XIcon size={12} />
					</button>
				</span>
			</div>
		{/each}

		<button
			type="button"
			class="text-base-content/60 hover:text-base-content flex h-9 items-center gap-1.5 self-center rounded-full border border-dashed border-white/15 px-4 text-sm transition-colors hover:border-white/35 hover:bg-white/5"
			onclick={() => (adding = true)}
		>
			<PlusIcon size={14} />
			Add performer
		</button>
	</div>
	{#if adding}
		<RecordPickerDialog
			kind="performer"
			title="Add a performer"
			placeholder="Find a performer by name…"
			search={(text) => fetchPerformers(text, 24)}
			suggested={suggestions.map((suggestion) => ({
				record: suggestion.performer,
				reason: suggestion.reason
			}))}
			taken={video.performers.map((performer) => performer.id)}
			onpick={add}
			oncreate={(name) => void create(name)}
			onclose={() => (adding = false)}
		/>
	{/if}
	{#if $faceTracking && $faceTracking.subject.kind === 'video' && $faceTracking.subject.id === video.id}
		<span class="text-base-content/50 flex items-center gap-2 text-xs">
			{#if $faceTracking.phase === 'waiting'}
				Looking through the video for faces once it's paused
			{:else}
				<span class="loading loading-spinner loading-xs"></span>
				{FACE_PHASES[$faceTracking.phase]}… {Math.round($faceTracking.progress * 100)}%
			{/if}
		</span>
	{/if}
</section>
