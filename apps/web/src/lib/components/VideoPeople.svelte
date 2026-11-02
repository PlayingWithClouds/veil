<!--
	Who's in a video, as large cards: the credited performers (✕ takes one
	off, as "Not her": when a site profile of theirs credited them, that
	profile moves away), the suggested ones (named in the title, …) to add
	with a click, and a card to add anyone else. Changes are the user's: the
	sites leave the credits as set.
-->
<script lang="ts">
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import RecordPicker from '$lib/components/RecordPicker.svelte';
	import {
		fetchPerformers,
		fetchVideoDetail,
		setVideoPerformers,
		type MemberRow,
		type VideoDetail
	} from '$lib/library';
	import { unassignPerformer } from '$lib/performers';
	import { showUndo } from '$lib/stores/undoToast';
	import { cacheUrl } from '$lib/img';
	import { performerUrl } from '$lib/routes';
	import { faceTracking, type VideoFaceGuesses } from '$lib/faces';

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
			all.push({ performer: guess.performer, reason: `Face match ${Math.round(guess.similarity * 100)}%` });
		}
		const seen = new Set<string>();
		return all.filter((suggestion) => {
			if (seen.has(suggestion.performer.id)) return false;
			seen.add(suggestion.performer.id);
			return true;
		});
	}

	/** Credits a performer on the video. */
	async function add(performer: { id: string }) {
		adding = false;
		if (video.performers.some((credited) => credited.id === performer.id)) return;
		busy = true;
		try {
			const ids = [...video.performers.map((credited) => credited.id), performer.id];
			const saved = await setVideoPerformers(video.id, ids);
			if (saved) onchange(saved);
		} finally {
			busy = false;
		}
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
	<span class="bg-base-200 flex aspect-[3/4] w-full items-center justify-center overflow-hidden rounded-xl">
		{#if member.image}
			<img src={cacheUrl(member.image.uri)} alt="" loading="lazy" class="h-full w-full object-cover" />
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
					<button type="button" class="btn btn-xs flex-1" disabled={busy} onclick={() => add(suggestion.performer)}>
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

		{#if adding}
			<RecordPicker
				placeholder="Find a performer…"
				search={(text) => fetchPerformers(text, 8)}
				onpick={add}
				onclose={() => (adding = false)}
			/>
		{:else}
			<button
				type="button"
				class="bg-base-200 hover:bg-base-300 text-base-content/60 flex aspect-[3/4] w-28 flex-col items-center justify-center gap-1 rounded-xl text-xs"
				onclick={() => (adding = true)}
			>
				<PlusIcon size={20} />
				Add performer
			</button>
		{/if}
	</div>
	{#if $faceTracking && $faceTracking.subject.kind === 'video' && $faceTracking.subject.id === video.id}
		<span class="text-base-content/50 flex items-center gap-2 text-xs">
			<span class="loading loading-spinner loading-xs"></span>
			Looking for faces… {Math.round($faceTracking.progress * 100)}%
		</span>
	{/if}
</section>
