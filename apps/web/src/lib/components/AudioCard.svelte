<!--
	An audio in a grid or row: square cover with its duration, and the caption
	with its studio or voices and site.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import type { AudioCardRow } from '$lib/media';

	interface Props {
		audio: AudioCardRow;
		onclick: () => void;
	}

	let { audio, onclick }: Props = $props();

	let durationLabel = $derived(formatClock(audio.durationSeconds));
	// The studio, else the first voices.
	let byline = $derived.by(() => {
		if (audio.studio) return audio.studio.name;
		return audio.performers
			.slice(0, 2)
			.map((performer) => performer.name)
			.join(', ');
	});
	let site = $derived.by(() => {
		const sourceUrl = audio.sources[0]?.url;
		if (!sourceUrl) return null;
		return pluginForUrl(sourceUrl, $sites);
	});
</script>

<div class="group flex w-full flex-col gap-3">
	<CoverFrame
		src={cacheUrl(audio.cover?.uri)}
		alt={audio.title}
		aspect="1 / 1"
		placeholder={HeadphonesIcon}
		{onclick}
	>
		{#if durationLabel}
			<CardBadge corner="bottom-right">{durationLabel}</CardBadge>
		{/if}
	</CoverFrame>
	<CardCaption
		title={audio.title}
		{byline}
		avatarUrl={cacheUrl(audio.studio?.image?.uri ?? audio.performers[0]?.image?.uri)}
		{site}
		detail={audio.date ? dayjs(audio.date).fromNow() : null}
		{onclick}
	/>
</div>
