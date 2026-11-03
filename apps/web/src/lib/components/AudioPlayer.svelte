<!--
	Plays one audio stream on its page: play/pause, a seek bar, skipping back
	15 and ahead 30 seconds, speed and volume. Plays through its own Playback
	(HLS or a plain file), which stops when the page is left. Space plays and
	pauses, the arrow keys skip.
-->
<script lang="ts">
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import ArrowCounterClockwiseIcon from 'phosphor-svelte/lib/ArrowCounterClockwiseIcon';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import SpeakerXIcon from 'phosphor-svelte/lib/SpeakerXIcon';
	import { Playback } from '$lib/playback';
	import { formatClock } from '$lib/time';

	const BACK_SECONDS = 15;
	const AHEAD_SECONDS = 30;
	const SPEEDS = [0.75, 1, 1.25, 1.5, 2];
	// A resume point this close to the end starts over instead.
	const RESUME_END_MARGIN = 10;

	interface Props {
		src: string;
		mimeType: string;
		// Seconds to resume from; 0 starts at the beginning.
		startPosition?: number;
		// Called with the element that plays, e.g. to log the listen; returns
		// what to run when the player lets go of it.
		onattach?: (media: HTMLMediaElement) => () => void;
	}

	let { src, mimeType, startPosition = 0, onattach }: Props = $props();

	const playback = new Playback();
	const media = playback.video;

	let paused = $state(true);
	let currentTime = $state(0);
	let duration = $state(0);
	let volume = $state(1);
	let muted = $state(false);
	let speed = $state(1);
	let error = $state<string | null>(null);

	let progress = $derived(duration > 0 ? currentTime / duration : 0);

	$effect(() => {
		playback.park();
		const handlers: Record<string, () => void> = {
			play: () => (paused = false),
			pause: () => (paused = true),
			timeupdate: () => (currentTime = media.currentTime),
			durationchange: () => (duration = Number.isFinite(media.duration) ? media.duration : 0),
			volumechange: () => {
				volume = media.volume;
				muted = media.muted;
			},
			ratechange: () => (speed = media.playbackRate)
		};
		for (const [name, handler] of Object.entries(handlers)) media.addEventListener(name, handler);
		const detach = onattach?.(media);
		return () => {
			detach?.();
			for (const [name, handler] of Object.entries(handlers))
				media.removeEventListener(name, handler);
			playback.stop();
			media.remove();
		};
	});

	$effect(() => {
		error = null;
		media.addEventListener('loadedmetadata', resume, { once: true });
		playback.load(src, mimeType, {}, (message) => (error = message));
	});

	/** Jumps to the resume point once the length is known. */
	function resume() {
		if (startPosition <= 0) return;
		if (Number.isFinite(media.duration) && startPosition > media.duration - RESUME_END_MARGIN)
			return;
		media.currentTime = startPosition;
	}

	/** Plays or pauses. */
	function toggle() {
		if (media.paused) {
			void media.play().catch((failure) => (error = String(failure)));
			return;
		}
		media.pause();
	}

	/** Moves the playhead by seconds, within the audio. */
	function skip(seconds: number) {
		media.currentTime = Math.min(Math.max(media.currentTime + seconds, 0), duration || Infinity);
	}

	/** Seeks to where the seek bar was set. */
	function seek(event: Event) {
		const fraction = Number((event.currentTarget as HTMLInputElement).value);
		media.currentTime = fraction * duration;
	}

	/** Steps to the next speed, back to the slowest after the fastest. */
	function nextSpeed() {
		const index = SPEEDS.indexOf(speed);
		media.playbackRate = SPEEDS[(index + 1) % SPEEDS.length];
	}

	/** Sets the volume from the slider; moving it unmutes. */
	function setVolume(event: Event) {
		media.volume = Number((event.currentTarget as HTMLInputElement).value);
		media.muted = false;
	}

	/** Space plays and pauses, arrows skip; not while typing. */
	function onKeydown(event: KeyboardEvent) {
		const target = event.target as HTMLElement | null;
		if (target?.closest('input, textarea, select, [contenteditable]')) return;
		if (event.key === ' ') {
			event.preventDefault();
			toggle();
		}
		if (event.key === 'ArrowLeft') skip(-BACK_SECONDS);
		if (event.key === 'ArrowRight') skip(AHEAD_SECONDS);
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="bg-base-content/5 flex flex-col gap-3 rounded-2xl p-4 backdrop-blur-md">
	<div class="flex items-center gap-3">
		<span class="text-base-content/60 w-14 text-right font-mono text-xs tabular-nums">
			{formatClock(currentTime) ?? '0:00'}
		</span>
		<input
			type="range"
			min="0"
			max="1"
			step="0.001"
			value={progress}
			oninput={seek}
			disabled={duration === 0}
			aria-label="Position"
			class="accent-base-content h-1 flex-1"
		/>
		<span class="text-base-content/60 w-14 font-mono text-xs tabular-nums">
			{formatClock(duration) ?? '–:––'}
		</span>
	</div>

	<div class="flex items-center justify-center gap-4">
		<button
			type="button"
			class="text-base-content/70 hover:text-base-content flex h-10 w-12 items-center justify-center rounded-full text-sm font-medium tabular-nums"
			onclick={nextSpeed}
			aria-label="Speed">{speed}×</button
		>
		<button
			type="button"
			class="text-base-content/70 hover:text-base-content relative flex h-11 w-11 items-center justify-center rounded-full"
			onclick={() => skip(-BACK_SECONDS)}
			aria-label="Back {BACK_SECONDS} seconds"
		>
			<ArrowCounterClockwiseIcon size={26} />
			<span class="absolute text-[9px] font-semibold">{BACK_SECONDS}</span>
		</button>
		<button
			type="button"
			class="bg-primary text-primary-content flex h-14 w-14 items-center justify-center rounded-full"
			onclick={toggle}
			aria-label={paused ? 'Play' : 'Pause'}
		>
			{#if paused}
				<PlayIcon size={26} weight="fill" />
			{:else}
				<PauseIcon size={26} weight="fill" />
			{/if}
		</button>
		<button
			type="button"
			class="text-base-content/70 hover:text-base-content relative flex h-11 w-11 items-center justify-center rounded-full"
			onclick={() => skip(AHEAD_SECONDS)}
			aria-label="Ahead {AHEAD_SECONDS} seconds"
		>
			<ArrowClockwiseIcon size={26} />
			<span class="absolute text-[9px] font-semibold">{AHEAD_SECONDS}</span>
		</button>
		<div class="flex w-12 items-center">
			<button
				type="button"
				class="text-base-content/70 hover:text-base-content flex h-10 w-10 items-center justify-center rounded-full"
				onclick={() => (media.muted = !media.muted)}
				aria-label={muted ? 'Unmute' : 'Mute'}
			>
				{#if muted || volume === 0}
					<SpeakerXIcon size={20} />
				{:else}
					<SpeakerHighIcon size={20} />
				{/if}
			</button>
		</div>
	</div>
	<input
		type="range"
		min="0"
		max="1"
		step="0.01"
		value={muted ? 0 : volume}
		oninput={setVolume}
		aria-label="Volume"
		class="accent-base-content h-1 w-32 self-end max-md:hidden"
	/>

	{#if error}
		<p class="text-error text-center text-xs">{error}</p>
	{/if}
</div>
