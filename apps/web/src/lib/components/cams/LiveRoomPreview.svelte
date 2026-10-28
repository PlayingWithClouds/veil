<script lang="ts">
	// A room's live stream as a small muted preview at its lowest quality, over
	// the room picture until it plays. The stream URL's token expires within
	// about a minute, so a failed stream asks for a fresh one a few times before
	// staying on the picture.
	import type Hls from 'hls.js';
	import { fetchCamStream } from '$lib/cams';
	import { cacheUrl } from '$lib/img';

	interface Props {
		camModelId: string;
		thumbnailUrl: string | null;
	}

	let { camModelId, thumbnailUrl }: Props = $props();

	const MAX_RETRIES = 3;

	let video = $state<HTMLVideoElement | null>(null);
	let playing = $state(false);

	$effect(() => {
		const element = video;
		const id = camModelId;
		if (!element) return;
		let instance: Hls | null = null;
		let retries = 0;
		let stopped = false;

		/** Resolves a fresh stream URL and plays it at the lowest quality. */
		async function attach() {
			let url: string;
			try {
				url = (await fetchCamStream(id)).url;
			} catch {
				return;
			}
			if (stopped || !element) return;
			const { default: HlsClass } = await import('hls.js');
			if (stopped) return;
			if (!HlsClass.isSupported()) {
				element.src = url;
				return;
			}
			instance?.destroy();
			instance = new HlsClass({ startLevel: 0, capLevelToPlayerSize: true, liveSyncDurationCount: 2 });
			instance.on(HlsClass.Events.MANIFEST_PARSED, () => {
				if (instance) instance.currentLevel = 0;
			});
			instance.on(HlsClass.Events.ERROR, (_event, data) => {
				if (!data.fatal || retries >= MAX_RETRIES) return;
				retries++;
				attach();
			});
			instance.loadSource(url);
			instance.attachMedia(element);
		}

		attach();
		return () => {
			stopped = true;
			instance?.destroy();
		};
	});
</script>

<span class="relative block h-full w-full bg-black">
	{#if cacheUrl(thumbnailUrl)}
		<img
			src={cacheUrl(thumbnailUrl)}
			alt=""
			class="absolute inset-0 h-full w-full object-cover transition-opacity"
			class:opacity-0={playing}
		/>
	{/if}
	<!-- svelte-ignore a11y_media_has_caption -->
	<video
		bind:this={video}
		class="absolute inset-0 h-full w-full object-cover"
		muted
		autoplay
		playsinline
		onplaying={() => (playing = true)}
	></video>
</span>
