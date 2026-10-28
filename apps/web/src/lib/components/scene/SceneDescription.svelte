<script lang="ts">
	// YouTube-style description panel under the channel row: a meta line (date,
	// runtime, views, rating), cast cards, tags, collapsible details, the marker
	// chapters and a one-line source summary (the only place naming the site).
	import TagChips from '$lib/components/TagChips.svelte';
	import type { SceneMarker } from '$lib/markers';
	import type { TagRef } from '$lib/tags';
	import type { SceneDetail, SceneSite, SceneStream, StreamGroup } from '$lib/sceneDetail';
	import {
		formatAbsoluteDate,
		formatCount,
		formatDuration,
		formatRelativeDate
	} from './format';
	import CastCard from './CastCard.svelte';
	import SceneChapters from './SceneChapters.svelte';
	import SourceSummary from './SourceSummary.svelte';

	interface Props {
		scene: SceneDetail;
		site: SceneSite | null;
		streamGroups: StreamGroup[];
		pluginIcons: Record<string, string | null>;
		// True while the first-visit fetch is still looking for sources.
		sourcesPending: boolean;
		activeSourceId: string | null;
		resolvingId: string | null;
		markers: SceneMarker[];
		// Alternate copies found so far, for the "Find on other sites" badge.
		alikeCount: number;
		onplay: (stream: SceneStream) => void;
		onjump: (seconds: number) => void;
		onremovemarker: (markerId: string) => void;
		onblocktag: (tag: TagRef) => void;
		onfindsources: () => void;
	}

	let {
		scene,
		site,
		streamGroups,
		pluginIcons,
		sourcesPending,
		activeSourceId,
		resolvingId,
		markers,
		alikeCount,
		onplay,
		onjump,
		onremovemarker,
		onblocktag,
		onfindsources
	}: Props = $props();

	let detailsExpanded = $state(false);
	let detailsElement = $state<HTMLElement | null>(null);
	let detailsOverflow = $state(false);

	let relativeDate = $derived(formatRelativeDate(scene.date));
	let duration = $derived(formatDuration(scene.durationSeconds));

	// Only offer "more" when the clamped details actually cut text off.
	$effect(() => {
		if (!detailsElement || detailsExpanded) return;
		detailsOverflow = detailsElement.scrollHeight > detailsElement.clientHeight + 1;
	});

	/** Toggles the details between clamped and full. */
	function toggleDetails() {
		detailsExpanded = !detailsExpanded;
	}
</script>

<section class="bg-base-200 flex flex-col gap-4 rounded-xl p-4 text-sm">
	<p class="text-base-content flex flex-wrap items-center gap-x-2 gap-y-1 font-semibold">
		{#if relativeDate}
			<span title={formatAbsoluteDate(scene.date)}>{relativeDate}</span>
		{/if}
		{#if duration}
			<span class="opacity-30">·</span><span>{duration}</span>
		{/if}
		{#if scene.viewCount > 0}
			<span class="opacity-30">·</span><span>{formatCount(scene.viewCount)} views</span>
		{/if}
		{#if scene.rating && scene.rating > 0}
			<!-- Site ratings are stored on a 0–10 scale; a percentage reads unambiguously. -->
			<span class="opacity-30">·</span><span title="Rating on the site">{Math.round(scene.rating * 10)}% liked</span>
		{/if}
	</p>

	{#if scene.performers.length > 0}
		<ul class="grid grid-cols-[repeat(auto-fill,minmax(13rem,1fr))] gap-2">
			{#each scene.performers as performer (performer.id)}
				<li class="min-w-0">
					<CastCard {performer} sceneDate={scene.date} />
				</li>
			{/each}
		</ul>
	{/if}

	<TagChips tags={scene.tags} onblock={onblocktag} />

	{#if scene.details}
		<div class="flex flex-col items-start gap-1">
			<p
				bind:this={detailsElement}
				class="text-base-content/80 leading-relaxed whitespace-pre-line"
				class:line-clamp-3={!detailsExpanded}
			>
				{scene.details}
			</p>
			{#if detailsOverflow}
				<button type="button" class="font-semibold hover:underline" onclick={toggleDetails}>
					{#if detailsExpanded}Show less{:else}...more{/if}
				</button>
			{/if}
		</div>
	{/if}

	{#if markers.length > 0}
		<div class="flex flex-col gap-2">
			<h2 class="font-semibold">Chapters</h2>
			<SceneChapters {markers} {onjump} onremove={onremovemarker} />
		</div>
	{/if}

	<div class="border-t border-white/5 pt-3">
		<SourceSummary
			groups={streamGroups}
			{pluginIcons}
			{site}
			sourceUrl={scene.sourceUrl}
			{sourcesPending}
			{activeSourceId}
			{resolvingId}
			{alikeCount}
			{onplay}
			{onfindsources}
		/>
	</div>
</section>
