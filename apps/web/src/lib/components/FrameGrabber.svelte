<!--
	Plays one video so a frame can be picked: seek to the moment, then "Use
	this frame" hands it over as a JPEG file. Streams come through the
	backend's proxy, which allows reading their pixels (crossorigin).
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import type Hls from 'hls.js';
	import { openVideo, play } from '$lib/library';
	import { isHlsSource } from '$lib/playback';

	interface Props {
		videoId: string;
		onpick: (frame: File) => void;
	}

	let { videoId, onpick }: Props = $props();

	let videoElement = $state<HTMLVideoElement>();
	let error = $state<string | null>(null);
	let loading = $state(true);
	let hls: Hls | null = null;

	onMount(() => {
		void load();
		return () => hls?.destroy();
	});

	/** Resolves the video's first stream and plays it in the element. */
	async function load() {
		const video = await openVideo(videoId);
		const stream = video?.streams[0];
		if (!stream) {
			error = 'This video has no stream to pick a frame from.';
			loading = false;
			return;
		}
		const playable = await play(stream.url);
		if (!playable || !videoElement) {
			error = 'The stream could not be opened.';
			loading = false;
			return;
		}
		await attach(videoElement, playable.url, playable.mimeType);
		loading = false;
	}

	/** Plays src in the element, through hls.js where the browser can't play HLS itself. */
	async function attach(element: HTMLVideoElement, src: string, mimeType: string) {
		const nativeHls = element.canPlayType('application/vnd.apple.mpegurl') !== '';
		if (!isHlsSource(src, mimeType) || nativeHls) {
			element.src = src;
			return;
		}
		const { default: HlsClass } = await import('hls.js');
		hls = new HlsClass();
		hls.loadSource(src);
		hls.attachMedia(element);
	}

	/** Draws the frame on screen and hands it over as a JPEG. */
	function useFrame() {
		const element = videoElement;
		if (!element || element.videoWidth === 0) return;
		element.pause();
		const canvas = document.createElement('canvas');
		canvas.width = element.videoWidth;
		canvas.height = element.videoHeight;
		canvas.getContext('2d')?.drawImage(element, 0, 0);
		try {
			canvas.toBlob(
				(blob) => {
					if (blob) onpick(new File([blob], 'frame.jpg', { type: 'image/jpeg' }));
				},
				'image/jpeg',
				0.92
			);
		} catch {
			error = 'This stream doesn’t allow taking frames from it.';
		}
	}
</script>

<div class="flex flex-col gap-3">
	<!-- svelte-ignore a11y_media_has_caption -->
	<video
		bind:this={videoElement}
		class="aspect-video w-full rounded-xl bg-black"
		crossorigin="anonymous"
		controls
		muted
		playsinline
	></video>
	{#if error}
		<p class="text-error text-sm">{error}</p>
	{/if}
	<div class="flex items-center justify-between gap-3">
		<span class="text-base-content/50 text-xs">
			{loading ? 'Opening the video…' : 'Seek to the moment, then take the frame.'}
		</span>
		<button type="button" class="btn btn-primary btn-sm" disabled={loading} onclick={useFrame}>
			Use this frame
		</button>
	</div>
</div>
