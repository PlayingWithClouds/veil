<script lang="ts">
	// A recording in progress on the Cams page: the room playing as a muted live
	// preview with REC, the running time and the quality over it; under it the
	// model's name, the size so far and Stop. The preview opens the room.
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import LiveRoomPreview from './LiveRoomPreview.svelte';
	import {
		formatRecordingClock,
		formatRecordingSize,
		stopCamRecording,
		type CamRecording
	} from '$lib/cams';
	import { camUrl } from '$lib/routes';
	import { refreshCamRecordings } from '$lib/stores/camRecordings';
	import { notifications } from '$lib/stores/notifications';

	interface Props {
		recording: CamRecording;
	}

	let { recording }: Props = $props();

	let now = $state(Date.now());
	let stopping = $state(false);
	let runningSeconds = $derived(Math.max(0, (now - Date.parse(recording.startedAt)) / 1000));

	$effect(() => {
		const timer = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(timer);
	});

	/** Stops the recording; it becomes a scene. */
	async function stop() {
		if (stopping) return;
		stopping = true;
		try {
			const finished = await stopCamRecording(recording.camModelId);
			if (finished && !finished.sceneId) {
				notifications.push('The recording was too short to keep', 'error');
			}
		} catch {
			notifications.push('Could not stop the recording', 'error');
		} finally {
			stopping = false;
			await refreshCamRecordings();
		}
	}
</script>

<div class="flex w-80 shrink-0 flex-col gap-2">
	<a
		href={camUrl(recording.camModelId)}
		class="relative block aspect-video overflow-hidden rounded-xl ring-1 ring-red-500/40"
		title="Open {recording.displayName}"
	>
		<LiveRoomPreview camModelId={recording.camModelId} thumbnailUrl={recording.thumbnailUrl} />
		<span
			class="absolute top-2 left-2 flex items-center gap-1.5 rounded bg-black/75 px-1.5 py-0.5 text-[11px] font-bold text-white tabular-nums"
		>
			<span class="h-2 w-2 animate-pulse rounded-full bg-red-500"></span>
			REC {formatRecordingClock(runningSeconds)}
		</span>
		{#if recording.quality}
			<span class="absolute top-2 right-2 rounded bg-black/75 px-1.5 py-0.5 text-[11px] font-semibold text-white">
				{recording.quality}
			</span>
		{/if}
	</a>
	<div class="flex items-center justify-between gap-3">
		<a href={camUrl(recording.camModelId)} class="flex min-w-0 flex-col hover:underline">
			<span class="truncate text-sm font-semibold">{recording.displayName}</span>
			<span class="text-base-content/50 text-xs tabular-nums">
				{formatRecordingSize(recording.sizeBytes)} so far
			</span>
		</a>
		<button
			type="button"
			class="flex shrink-0 items-center gap-1.5 rounded-full bg-white/10 px-3 py-1.5 text-xs font-semibold transition-colors hover:bg-white/20 disabled:opacity-50"
			disabled={stopping}
			onclick={stop}
			aria-label="Stop recording {recording.displayName}"
		>
			<StopIcon size={12} weight="fill" class="text-red-500" />
			Stop
		</button>
	</div>
</div>
