<script lang="ts">
	// YouTube-style description panel under the channel row: a meta line (date,
	// runtime, views, rating, site), performers, tags, collapsible details, the
	// marker chapters and the per-site source list.
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import { cacheUrl } from '$lib/img';
	import { performerUrl } from '$lib/routes';
	import type { SceneMarker } from '$lib/markers';
	import type { TagRef } from '$lib/tags';
	import type { SceneDetail, SceneSite, SceneStream, StreamGroup } from '$lib/sceneDetail';
	import {
		formatAbsoluteDate,
		formatCount,
		formatDuration,
		formatRelativeDate
	} from './format';
	import SiteIcon from './SiteIcon.svelte';
	import SceneChapters from './SceneChapters.svelte';
	import SceneSources from './SceneSources.svelte';

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
		onplay: (stream: SceneStream) => void;
		onjump: (seconds: number) => void;
		onremovemarker: (markerId: string) => void;
		onblocktag: (tag: TagRef) => void;
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
		onplay,
		onjump,
		onremovemarker,
		onblocktag
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
		{#if site}
			<span class="opacity-30">·</span>
			<a
				href={scene.sourceUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="flex items-center gap-1.5 hover:underline"
				title="Open on {site.name}"
			>
				<SiteIcon iconUrl={site.iconUrl} />
				{site.name}
			</a>
		{/if}
	</p>

	{#if scene.performers.length > 0}
		<ul class="flex flex-wrap gap-2">
			{#each scene.performers as performer (performer.id)}
				<li>
					<a
						href={performerUrl(performer.id)}
						class="bg-base-300 flex items-center gap-2 rounded-full py-1 pr-3 pl-1 transition-colors hover:bg-white/15"
					>
						<span
							class="bg-base-100 flex h-7 w-7 items-center justify-center overflow-hidden rounded-full"
						>
							{#if cacheUrl(performer.imagePath)}
								<img
									src={cacheUrl(performer.imagePath)}
									alt=""
									loading="lazy"
									class="h-full w-full object-cover"
								/>
							{:else}
								<UserIcon size={14} class="opacity-40" />
							{/if}
						</span>
						<span class="font-medium">{performer.name}</span>
					</a>
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

	<div class="flex flex-col gap-2">
		<h2 class="font-semibold">Sources</h2>
		{#if streamGroups.length > 0}
			<SceneSources groups={streamGroups} {pluginIcons} {activeSourceId} {resolvingId} {onplay} />
		{:else if sourcesPending}
			<p class="text-base-content/40 flex items-center gap-2">
				<span class="loading loading-xs"></span>
				Fetching sources…
			</p>
		{:else}
			<p class="text-base-content/40">No sources available yet. Try “Find sources”.</p>
		{/if}
	</div>
</section>
