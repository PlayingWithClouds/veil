<!--
	The live cam recordings in progress, one row each: the model's avatar and
	name, a pulsing red dot with the running time, and a Stop button that ends
	that recording in place. A row opens the room. Nothing renders while no
	recording runs.
-->
<script lang="ts">
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import { formatRecordingClock, stopCamRecording } from '$lib/cams';
	import { camUrl } from '$lib/routes';
	import { activeCamRecordings, refreshCamRecordings } from '$lib/stores/camRecordings';
	import { notifications } from '$lib/stores/notifications';
	import CamAvatar from './CamAvatar.svelte';

	// Ticks every second while something records, for the running times.
	let now = $state(Date.now());
	// Models whose Stop is in flight.
	let stopping = $state<string[]>([]);

	$effect(() => {
		if ($activeCamRecordings.length === 0) return;
		const timer = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(timer);
	});

	/** Seconds a recording has run. */
	function runningSeconds(startedAt: string): number {
		return Math.max(0, (now - Date.parse(startedAt)) / 1000);
	}

	/** Stops the model's recording and files it as a scene. */
	async function stop(camModelId: string) {
		if (stopping.includes(camModelId)) return;
		stopping = [...stopping, camModelId];
		try {
			await stopCamRecording(camModelId);
		} catch {
			notifications.push('Could not stop the recording', 'error');
		} finally {
			stopping = stopping.filter((id) => id !== camModelId);
			await refreshCamRecordings();
		}
	}
</script>

{#if $activeCamRecordings.length > 0}
	<section class="flex flex-col gap-0.5" aria-label="Recording now">
		<p class="text-base-content/40 px-3 pb-1 text-xs font-semibold tracking-wide uppercase">
			Recording
		</p>
		{#each $activeCamRecordings as recording (recording.id)}
			<div class="hover:bg-base-200 flex items-center gap-1 rounded-lg pr-1 transition-colors">
				<a
					href={camUrl(recording.camModelId)}
					class="flex min-w-0 flex-1 items-center gap-3 px-3 py-1.5"
					title="Open {recording.displayName}"
				>
					<CamAvatar
						camModel={{
							displayName: recording.displayName,
							thumbnailUrl: recording.thumbnailUrl,
							online: false
						}}
					/>
					<span class="flex min-w-0 flex-1 flex-col leading-tight">
						<span class="truncate text-sm">{recording.displayName}</span>
						<span class="flex items-center gap-1.5 text-xs text-red-400 tabular-nums">
							<span class="relative flex h-2 w-2">
								<span
									class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500 opacity-75"
								></span>
								<span class="relative inline-flex h-2 w-2 rounded-full bg-red-500"></span>
							</span>
							{formatRecordingClock(runningSeconds(recording.startedAt))}
							{#if recording.quality}
								<span class="text-base-content/50">{recording.quality}</span>
							{/if}
						</span>
					</span>
				</a>
				<button
					type="button"
					class="text-base-content/70 hover:text-base-content hover:bg-base-300 flex h-8 w-8 shrink-0 items-center justify-center rounded-full disabled:opacity-50"
					disabled={stopping.includes(recording.camModelId)}
					onclick={() => stop(recording.camModelId)}
					aria-label="Stop recording {recording.displayName}"
					title="Stop recording"
				>
					<StopIcon size={16} weight="fill" />
				</button>
			</div>
		{/each}
	</section>
	<hr class="border-base-300 my-3" />
{/if}
