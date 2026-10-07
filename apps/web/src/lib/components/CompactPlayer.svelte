<script lang="ts">
	// Small player chrome around a Playback's element: skip, play/pause, mute and
	// a timeline, shown on hover. It only mounts the element; loading, stopping
	// and handing it on stay with the owner (mini player, card hover player).
	import { untrack, type Snippet } from 'svelte';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import ClockClockwiseIcon from 'phosphor-svelte/lib/ClockClockwiseIcon';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import SpeakerXIcon from 'phosphor-svelte/lib/SpeakerXIcon';
	import type { Playback } from '$lib/playback';

	export interface CompactPlayerState {
		position: number;
		duration: number;
		playing: boolean;
		muted: boolean;
	}

	interface Props {
		playback: Playback;
		// cover crops to fill (cards); contain letterboxes (mini player).
		fit?: 'contain' | 'cover';
		// False drops the hover controls, for owners that draw their own.
		chrome?: boolean;
		// Top row, e.g. title and close button.
		header?: Snippet;
		// Right end of the control row, e.g. an expand button.
		actions?: Snippet;
		onvideoclick?: () => void;
		// Called on play/pause/mute changes, seeks and time updates.
		onstatechange?: (state: CompactPlayerState) => void;
	}

	let {
		playback,
		fit = 'contain',
		chrome = true,
		header,
		actions,
		onvideoclick,
		onstatechange
	}: Props = $props();

	const SKIP_SECONDS = 10;

	let slot = $state<HTMLDivElement | undefined>();
	let playing = $state(false);
	let muted = $state(false);
	let position = $state(0);
	let duration = $state(0);
	let playedPercent = $derived(duration > 0 ? (position / duration) * 100 : 0);

	// Mount the element and mirror its state; parked on unmount so whoever takes
	// it next finds it still playing.
	$effect(() => {
		const container = slot;
		const video = playback.video;
		if (!container) return;
		video.className = 'h-full w-full';
		video.classList.toggle('object-cover', fit === 'cover');
		video.classList.toggle('object-contain', fit === 'contain');
		playback.mount(container);
		// The owner's callback may read state of its own; keep it out of this effect.
		untrack(() => readState(video));
		const events = ['play', 'pause', 'timeupdate', 'durationchange', 'volumechange', 'seeked'];
		const onEvent = () => readState(video);
		for (const name of events) video.addEventListener(name, onEvent);
		return () => {
			for (const name of events) video.removeEventListener(name, onEvent);
			if (playback.isMountedIn(container)) playback.park();
		};
	});

	/** Copies the element's state and reports it. */
	function readState(video: HTMLVideoElement) {
		playing = !video.paused && !video.ended;
		muted = video.muted;
		position = video.currentTime;
		duration = isFinite(video.duration) ? video.duration : 0;
		onstatechange?.({ position, duration, playing, muted });
	}

	function togglePlay() {
		const video = playback.video;
		if (video.paused) {
			video.play().catch(() => {});
			return;
		}
		video.pause();
	}

	/** Jumps by seconds (negative = back), clamped to the video. */
	function skip(seconds: number) {
		const video = playback.video;
		let target = Math.max(0, video.currentTime + seconds);
		if (duration > 0) {
			target = Math.min(target, duration);
		}
		video.currentTime = target;
	}

	function toggleMute() {
		playback.video.muted = !playback.video.muted;
	}

	function seek(event: Event) {
		playback.video.currentTime = Number((event.target as HTMLInputElement).value);
	}
</script>

<div class="group/player relative h-full w-full bg-black">
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div bind:this={slot} class="h-full w-full cursor-pointer" onclick={onvideoclick}></div>
	<div
		class="pointer-events-none absolute inset-0 flex flex-col justify-between bg-gradient-to-b from-black/70 via-transparent to-black/70 p-2 opacity-0 transition-opacity group-focus-within/player:opacity-100 group-hover/player:opacity-100"
		class:hidden={!chrome}
	>
		<div class="pointer-events-auto">
			{@render header?.()}
		</div>
		<div class="pointer-events-auto flex flex-col gap-1">
			<input
				type="range"
				min="0"
				max={duration || 0}
				value={position}
				step="0.5"
				class="player-range w-full"
				style="--fill:{playedPercent}%"
				oninput={seek}
				aria-label="Seek"
			/>
			<div class="flex items-center gap-1 text-white">
				<button
					type="button"
					class="btn btn-square btn-ghost btn-sm"
					aria-label="Back {SKIP_SECONDS} seconds"
					onclick={() => skip(-SKIP_SECONDS)}
				>
					<ClockCounterClockwiseIcon size={18} />
				</button>
				<button
					type="button"
					class="btn btn-square btn-ghost btn-sm"
					aria-label={playing ? 'Pause' : 'Play'}
					onclick={togglePlay}
				>
					{#if playing}
						<PauseIcon size={18} weight="fill" />
					{:else}
						<PlayIcon size={18} weight="fill" />
					{/if}
				</button>
				<button
					type="button"
					class="btn btn-square btn-ghost btn-sm"
					aria-label="Forward {SKIP_SECONDS} seconds"
					onclick={() => skip(SKIP_SECONDS)}
				>
					<ClockClockwiseIcon size={18} />
				</button>
				<button
					type="button"
					class="btn btn-square btn-ghost btn-sm"
					aria-label={muted ? 'Unmute' : 'Mute'}
					onclick={toggleMute}
				>
					{#if muted}
						<SpeakerXIcon size={18} />
					{:else}
						<SpeakerHighIcon size={18} />
					{/if}
				</button>
				<div class="ml-auto flex items-center gap-1">
					{@render actions?.()}
				</div>
			</div>
		</div>
	</div>
</div>
