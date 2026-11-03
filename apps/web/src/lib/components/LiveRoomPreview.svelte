<!--
	A room's live stream as a small muted preview at its lowest quality,
	covering a card's picture once it plays. The stream URL's token expires
	within about a minute, so a failed stream asks for a fresh one a few times
	before giving up and leaving the picture.
-->
<script lang="ts">
	import type Hls from 'hls.js';
	import { watchCam } from '$lib/cams';

	interface Props {
		camId: string;
	}

	let { camId }: Props = $props();

	const MAX_RETRIES = 3;

	let video = $state<HTMLVideoElement | null>(null);
	let playing = $state(false);

	$effect(() => {
		const element = video;
		const id = camId;
		if (!element) return;
		let instance: Hls | null = null;
		let retries = 0;
		let stopped = false;

		/** Resolves a fresh stream URL and plays it at the lowest quality. */
		async function attach() {
			let url: string;
			try {
				url = (await watchCam(id)).url;
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
			instance = new HlsClass({
				startLevel: 0,
				capLevelToPlayerSize: true,
				liveSyncDurationCount: 2
			});
			instance.on(HlsClass.Events.MANIFEST_PARSED, () => {
				if (instance) instance.currentLevel = 0;
			});
			instance.on(HlsClass.Events.ERROR, (_event, data) => {
				if (!data.fatal || retries >= MAX_RETRIES) return;
				retries++;
				void attach();
			});
			instance.loadSource(url);
			instance.attachMedia(element);
		}

		void attach();
		return () => {
			stopped = true;
			instance?.destroy();
		};
	});
</script>

<video
	bind:this={video}
	class="pointer-events-none absolute inset-0 h-full w-full object-cover transition-opacity"
	class:opacity-0={!playing}
	muted
	autoplay
	playsinline
	onplaying={() => (playing = true)}
></video>
