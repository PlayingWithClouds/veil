<!--
	An audio in a grid or row: square cover (a generated one when the site has
	none) with its audience tag ("F4M") and duration, and the caption with the
	title (its bracketed tags left out), the voice and the site.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import GeneratedCover from '$lib/components/GeneratedCover.svelte';
	import { cacheUrl } from '$lib/img';
	import { formatClock } from '$lib/time';
	import { pluginForUrl } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { splitTitle } from '$lib/titleTags';
	import type { AudioCardRow } from '$lib/media';

	interface Props {
		audio: AudioCardRow;
		onclick: () => void;
	}

	let { audio, onclick }: Props = $props();

	let durationLabel = $derived(formatClock(audio.durationSeconds));
	let parts = $derived(splitTitle(audio.title));
	let voice = $derived(audio.performers[0] ?? null);
	// The voices, else the studio.
	let byline = $derived.by(() => {
		if (audio.performers.length > 0) {
			return audio.performers
				.slice(0, 2)
				.map((performer) => performer.name)
				.join(', ');
		}
		return audio.studio?.name ?? null;
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
		{#snippet fallback()}
			<GeneratedCover title={parts.title} {byline} kind="audio" />
		{/snippet}
		{#if parts.audiences.length > 0}
			<CardBadge corner="top-left">{parts.audiences[0]}</CardBadge>
		{/if}
		{#if durationLabel}
			<CardBadge corner="bottom-right">{durationLabel}</CardBadge>
		{/if}
	</CoverFrame>
	<CardCaption
		title={parts.title}
		{byline}
		avatarUrl={cacheUrl(voice?.image?.uri ?? audio.studio?.image?.uri)}
		siteAvatar={false}
		{site}
		detail={audio.date ? dayjs(audio.date).fromNow() : null}
		{onclick}
	/>
</div>
