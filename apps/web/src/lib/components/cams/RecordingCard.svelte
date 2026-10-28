<script lang="ts">
	// One finished recording on its room's page: the frame it opened with, its
	// length and quality on the thumbnail, then when it was made, its size and
	// why it ended. The model's name is left out, the page already says it.
	import { formatClock } from '$lib/components/SceneCard.svelte';
	import { cacheUrl } from '$lib/img';
	import { formatRecordingSize, type RecordedSceneCard } from '$lib/cams';

	interface Props {
		recorded: RecordedSceneCard;
		onclick: () => void;
	}

	let { recorded, onclick }: Props = $props();

	const STOP_REASONS: Record<string, string> = {
		stopped: 'Stopped by you',
		offline: 'Room went offline',
		max_duration: 'Reached the length limit',
		restart: 'Cut by a restart',
		error: 'Ended by an error'
	};

	let recording = $derived(recorded.recording);
	let posterUrl = $derived.by(() => {
		if (recorded.card.bannerUrl) return recorded.card.bannerUrl;
		return cacheUrl(recording.thumbnailUrl);
	});
	let durationLabel = $derived(formatClock(recording.durationSeconds));
	let startedLabel = $derived(
		new Date(recording.startedAt).toLocaleString(undefined, {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		})
	);
	let details = $derived(recordingDetails());

	/** The line under the date: size, then why the recording ended. */
	function recordingDetails(): string {
		const parts = [formatRecordingSize(recording.sizeBytes)];
		if (recording.stopReason && STOP_REASONS[recording.stopReason]) {
			parts.push(STOP_REASONS[recording.stopReason]);
		}
		return parts.join(' · ');
	}
</script>

<button type="button" class="group flex flex-col gap-2 text-left" {onclick}>
	<span class="bg-base-300 relative block aspect-video w-full overflow-hidden rounded-lg">
		{#if posterUrl}
			<img
				src={posterUrl}
				alt=""
				loading="lazy"
				class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
			/>
		{/if}
		{#if recording.quality}
			<span class="absolute top-1.5 left-1.5 rounded bg-black/75 px-1.5 py-0.5 text-[11px] font-semibold text-white">
				{recording.quality}
			</span>
		{/if}
		{#if durationLabel}
			<span
				class="absolute right-1.5 bottom-1.5 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
			>
				{durationLabel}
			</span>
		{/if}
	</span>
	<span class="flex flex-col">
		<span class="text-sm font-medium group-hover:underline">{startedLabel}</span>
		<span class="text-base-content/50 text-xs">{details}</span>
	</span>
</button>
