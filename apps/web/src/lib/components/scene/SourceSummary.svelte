<script lang="ts">
	// One quiet line about where the video comes from: the site (linking to the
	// scene's page there), the quality playing, how many other sites carry it,
	// and a toggle for the full per-site source list. "Find on other sites"
	// starts the alternate-source search.
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import { qualityLabel, type SceneSite, type SceneStream, type StreamGroup } from '$lib/sceneDetail';
	import SceneSources from './SceneSources.svelte';
	import SiteIcon from './SiteIcon.svelte';

	interface Props {
		groups: StreamGroup[];
		pluginIcons: Record<string, string | null>;
		site: SceneSite | null;
		sourceUrl: string;
		// True while the first-visit fetch is still looking for sources.
		sourcesPending: boolean;
		activeSourceId: string | null;
		resolvingId: string | null;
		// Alternate copies found so far (0 before a search).
		alikeCount: number;
		onplay: (stream: SceneStream) => void;
		onfindsources: () => void;
	}

	let {
		groups,
		pluginIcons,
		site,
		sourceUrl,
		sourcesPending,
		activeSourceId,
		resolvingId,
		alikeCount,
		onplay,
		onfindsources
	}: Props = $props();

	let expanded = $state(false);
	let streamCount = $derived(groups.reduce((total, group) => total + group.streams.length, 0));
	let playingLabel = $derived(activeQuality(groups, activeSourceId));
	let siteName = $derived.by(() => {
		if (site) return site.name;
		return hostOf(sourceUrl);
	});

	/** Quality label of the playing source, null when nothing plays. */
	function activeQuality(streamGroups: StreamGroup[], sourceId: string | null): string | null {
		if (!sourceId) return null;
		for (const group of streamGroups) {
			const stream = group.streams.find((entry) => entry.id === sourceId);
			if (stream) return qualityLabel(stream);
		}
		return null;
	}

	/** Host of a URL without "www.", for sites no plugin names. */
	function hostOf(url: string): string {
		try {
			return new URL(url).hostname.replace(/^www\./, '');
		} catch {
			return url;
		}
	}

	/** "2 more sites" when other sites than the first carry the video. */
	function otherSitesLabel(count: number): string {
		if (count === 1) return '1 more site';
		return `${count} more sites`;
	}
</script>

<div class="flex flex-col gap-3">
	<div class="text-base-content/70 flex flex-wrap items-center gap-x-3 gap-y-2">
		<a
			href={sourceUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="text-base-content flex items-center gap-1.5 font-medium hover:underline"
			title="Open on {siteName}"
		>
			{#if site}
				<SiteIcon iconUrl={site.iconUrl} />
			{/if}
			{siteName}
			<ArrowSquareOutIcon size={12} class="opacity-50" />
		</a>
		{#if playingLabel}
			<span>Playing {playingLabel}</span>
		{:else if sourcesPending}
			<span class="flex items-center gap-2">
				<span class="loading loading-xs"></span>
				Fetching sources…
			</span>
		{:else if streamCount === 0}
			<span class="text-base-content/40">Nothing playable here yet</span>
		{/if}
		{#if groups.length > 1}
			<span class="opacity-30">·</span>
			<span>{otherSitesLabel(groups.length - 1)}</span>
		{/if}
		{#if streamCount > 1}
			<button
				type="button"
				class="flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold hover:bg-white/10"
				aria-expanded={expanded}
				onclick={() => (expanded = !expanded)}
			>
				All sources
				<span class="flex transition-transform" class:rotate-180={expanded}>
					<CaretDownIcon size={10} weight="bold" />
				</span>
			</button>
		{/if}
		<button
			type="button"
			class="ml-auto flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-semibold hover:bg-white/10"
			class:text-primary={streamCount === 0 && alikeCount > 0}
			onclick={onfindsources}
		>
			<MagnifyingGlassIcon size={12} weight="bold" />
			Find on other sites
			{#if alikeCount > 0}
				<span class="bg-primary text-primary-content rounded-full px-1.5 text-[11px] font-bold">
					{alikeCount}
				</span>
			{/if}
		</button>
	</div>
	{#if expanded && streamCount > 0}
		<SceneSources {groups} {pluginIcons} {activeSourceId} {resolvingId} {onplay} />
	{/if}
</div>
