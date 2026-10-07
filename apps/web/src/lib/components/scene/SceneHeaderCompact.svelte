<!--
	Phone watch page header: the title (two lines, tap to read it all) and one
	meta line with the channel (studio, else site), age, length, views and rating.
-->
<script lang="ts">
	import { cacheUrl } from '$lib/img';
	import { studioUrl } from '$lib/routes';
	import type { SceneDetail, SceneSite } from '$lib/sceneDetail';
	import { formatAbsoluteDate, formatCount, formatDuration, formatRelativeDate } from './format';
	import SiteIcon from './SiteIcon.svelte';

	interface Props {
		scene: SceneDetail;
		site: SceneSite | null;
	}

	let { scene, site }: Props = $props();

	let titleExpanded = $state(false);
	let studioAvatar = $derived(cacheUrl(scene.studio?.imagePath));
	let relativeDate = $derived(formatRelativeDate(scene.date));
	let duration = $derived(formatDuration(scene.durationSeconds));
</script>

<div class="flex flex-col gap-1.5">
	<button type="button" class="text-left" onclick={() => (titleExpanded = !titleExpanded)}>
		<h1 class="text-lg leading-snug font-bold" class:line-clamp-2={!titleExpanded}>{scene.title}</h1>
	</button>
	<p class="text-base-content/60 no-scrollbar flex min-w-0 items-center gap-1.5 overflow-x-auto text-[13px] whitespace-nowrap">
		{#if scene.studio}
			<a href={studioUrl(scene.studio.id)} class="text-base-content/85 flex min-w-0 items-center gap-1.5 font-medium">
				{#if studioAvatar}
					<img src={studioAvatar} alt="" class="h-5 w-5 shrink-0 rounded-full object-cover" />
				{:else if site}
					<SiteIcon iconUrl={site.iconUrl} class="h-4 w-4" />
				{/if}
				<span class="truncate">{scene.studio.name}</span>
			</a>
		{:else if site}
			<a
				href={scene.sourceUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="text-base-content/85 flex min-w-0 items-center gap-1.5 font-medium"
			>
				<SiteIcon iconUrl={site.iconUrl} class="h-4 w-4" />
				<span class="truncate">{site.name}</span>
			</a>
		{/if}
		{#if relativeDate}
			<span aria-hidden="true">·</span>
			<span class="shrink-0" title={formatAbsoluteDate(scene.date)}>{relativeDate}</span>
		{/if}
		{#if duration}
			<span aria-hidden="true">·</span>
			<span class="shrink-0">{duration}</span>
		{/if}
		{#if scene.viewCount > 0}
			<span aria-hidden="true">·</span>
			<span class="shrink-0">{formatCount(scene.viewCount)} views</span>
		{/if}
		{#if scene.rating && scene.rating > 0}
			<!-- Site ratings are stored on a 0–10 scale; a percentage reads unambiguously. -->
			<span aria-hidden="true">·</span>
			<span class="shrink-0">{Math.round(scene.rating * 10)}% liked</span>
		{/if}
	</p>
</div>
