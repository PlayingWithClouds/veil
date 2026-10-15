<!--
	Plays a live HLS stream with hls.js: muted autoplay, sitting near the live
	edge. A fatal playback error (the broadcast ended, the token expired) is
	reported through onended so the page can show its offline state.
-->
<script lang="ts">
	interface Props {
		streamUrl: string;
		// The stream stopped for good (broadcast over or the connection was lost).
		onended: () => void;
	}

	let { streamUrl, onended }: Props = $props();

	let videoElement = $state<HTMLVideoElement | null>(null);

	// Network errors on a live edge are retried a few times before giving up.
	const MAX_NETWORK_RETRIES = 3;

	$effect(() => {
		const element = videoElement;
		if (!element) return;
		let instance: { destroy(): void } | null = null;
		let cancelled = false;

		import('hls.js').then(({ default: Hls }) => {
			if (cancelled) return;
			if (!Hls.isSupported()) {
				element.src = streamUrl;
				void element.play().catch(() => {});
				return;
			}
			let networkRetries = 0;
			const hls = new Hls({ lowLatencyMode: false, liveSyncDurationCount: 3 });
			hls.loadSource(streamUrl);
			hls.attachMedia(element);
			hls.on(Hls.Events.MANIFEST_PARSED, () => {
				void element.play().catch(() => {});
			});
			hls.on(Hls.Events.ERROR, (_event, data) => {
				if (!data.fatal) return;
				if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
					hls.recoverMediaError();
					return;
				}
				if (data.type === Hls.ErrorTypes.NETWORK_ERROR && networkRetries < MAX_NETWORK_RETRIES) {
					networkRetries += 1;
					hls.startLoad();
					return;
				}
				onended();
			});
			instance = hls;
		});

		return () => {
			cancelled = true;
			instance?.destroy();
		};
	});
</script>

<!-- svelte-ignore a11y_media_has_caption -->
<video
	bind:this={videoElement}
	class="h-full w-full bg-black"
	controls
	autoplay
	muted
	playsinline
></video>
